package s3

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/feature/s3/transfermanager"
	"github.com/pkg/errors"
	"github.com/viant/afs/base"
	"github.com/viant/afs/option"

	"github.com/viant/afs/storage"
)

// Open return content reader and hash values if md5 or crc option is supplied or error
func (s *Storager) Open(ctx context.Context, location string, options ...storage.Option) (io.ReadCloser, error) {

	// In SDK v2, multiple functions do not handle leading slashes, so remove them here
	parsedLocation := location
	if len(parsedLocation) > 0 && parsedLocation[0] == '/' {
		parsedLocation = parsedLocation[1:]
	}

	started := time.Now()
	defer func() {
		s.logF("s3:Open %v %s\n", location, time.Since(started))
	}()

	var err error
	stream := &option.Stream{}
	key := &option.AES256Key{}
	option.Assign(options, &key, &stream)
	input := &transfermanager.GetObjectInput{
		Bucket: &s.bucket,
		Key:    &parsedLocation,
	}

	if len(key.Key) > 0 {
		stringKey := string(key.Key)
		algorithm := customEncryptionAlgorithm
		input.SSECustomerAlgorithm = &algorithm
		input.SSECustomerKey = &stringKey
		input.SSECustomerKeyMD5 = &key.Base64KeyMd5Hash
	}

	tmClient := transfermanager.New(s.Client)
	if stream.PartSize > 0 {
		objects, err := s.List(ctx, location, key)
		if err != nil {
			return nil, err
		}
		if len(objects) == 0 {
			return nil, fmt.Errorf("s3://%v/%v no found", s.bucket, location)
		}
		stream.Size = int(objects[0].Size())
		readSeeker := NewReadSeeker(ctx, input, tmClient, stream.PartSize, stream.Size)
		reader := base.NewStreamReader(stream, readSeeker)
		return reader, nil
	}

	location = strings.Trim(location, "/")
	output, err := tmClient.GetObject(ctx, input)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to download: s3://%v/%v", s.bucket, location)
	}
	data, err := io.ReadAll(output.Body)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read: s3://%v/%v", s.bucket, location)
	}
	return io.NopCloser(bytes.NewReader(data)), nil
}
