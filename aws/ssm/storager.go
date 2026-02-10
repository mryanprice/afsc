package ssm

import (
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/viant/afs/storage"
)

type Storager struct {
	region string
	client *ssm.Client
	mux    sync.Mutex
}

// Exists returns true if location exists
func (s *Storager) Exists(ctx context.Context, resourceID string, options ...storage.Option) (bool, error) {
	resource, err := newResource(resourceID)
	if err != nil {
		return false, err
	}
	client, err := s.systemManager(ctx, resource.Region)
	if err != nil {
		return false, err
	}
	param, _ := s.getParameter(ctx, client, resource)
	return param != nil, nil
}

// Get returns a file info for supplied location
func (s *Storager) Get(ctx context.Context, location string, options ...storage.Option) (os.FileInfo, error) {
	list, err := s.List(ctx, location, options...)
	if err != nil {
		return nil, err
	}
	if len(list) == 0 {
		return nil, fmt.Errorf("no found: %v", location)
	}
	return list[0], nil
}

// Delete deletes locations
func (s *Storager) Delete(ctx context.Context, location string, options ...storage.Option) error {
	return fmt.Errorf("unsupported operation")
}

// Close closes storage
func (s *Storager) Close() error {
	return nil
}

func (s *Storager) systemManager(ctx context.Context, region string) (*ssm.Client, error) {
	s.mux.Lock()
	defer s.mux.Unlock()
	if s.region == "" || s.region != region {
		s.region = region
		cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
		if err != nil {
			return nil, fmt.Errorf("failed to load AWS config: %w", err)
		}
		s.client = ssm.NewFromConfig(cfg)
		return s.client, nil
	}
	return s.client, nil
}

// NewStorager create a new secret manager storager
func NewStorager(ctx context.Context, baseURL string, options ...storage.Option) (*Storager, error) {
	result := &Storager{}
	return result, nil
}
