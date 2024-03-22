package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/CloudNativeGame/aigc-gateway/pkg/options"
	"k8s.io/klog/v2"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/CloudNativeGame/aigc-gateway/pkg/resources"
)

const (
	KeyPrefixRes = "res"
)

var globalStorage *Storage

type Storage struct {
	redis *redis.Client
}

type ResourceStatus struct {
	Owner     string    `json:"owner"`
	Timestamp time.Time `json:"timestamp"`
}

func NewStorage(serverOpts *options.ServerOption) *Storage {
	opts := &redis.Options{
		Addr: serverOpts.RedisAddress,
	}
	client := redis.NewClient(opts)
	s := &Storage{
		redis: client,
	}
	return s
}

func (s *Storage) UpdateStatus(ctx context.Context, user string, meta *resources.ResourceMeta) error {
	hashKey := fmt.Sprintf("%s#%s#%s", KeyPrefixRes, meta.Namespace, meta.Name)

	status := &ResourceStatus{
		Owner:     user,
		Timestamp: time.Now(),
	}
	val, err := json.Marshal(status)
	if err != nil {
		return err
	}
	if err := s.redis.HSet(ctx, hashKey, meta.ID, val).Err(); err != nil {
		klog.Errorf("UpdateStatus %s:%s failed: %v", hashKey, meta.ID, err)
		return err
	}
	return nil
}

func (s *Storage) DeleteRecord(ctx context.Context, meta *resources.ResourceMeta) error {
	hashKey := fmt.Sprintf("%s#%s#%s", KeyPrefixRes, meta.Namespace, meta.Name)
	if err := s.redis.HDel(ctx, hashKey, meta.ID).Err(); err != nil {
		klog.Errorf("DeleteRecord %s:%s failed: %v", hashKey, meta.ID, err)
		return err
	}
	return nil
}

func (s *Storage) GetAllStatus(ctx context.Context, namespace, name string) (map[string]*ResourceStatus, error) {
	hashKey := fmt.Sprintf("%s#%s#%s", KeyPrefixRes, namespace, name)
	cmd := s.redis.HGetAll(ctx, hashKey)
	if err := cmd.Err(); err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, err
	}
	hashValue := cmd.Val()
	result := make(map[string]*ResourceStatus, len(hashValue))
	for id, value := range hashValue {
		status := &ResourceStatus{}
		if err := json.Unmarshal([]byte(value), status); err != nil {
			klog.Errorf("GetAllStatus Unmarshal %s:%s failed: %v", hashKey, id, err)
			continue
		}
		result[id] = status
	}
	return result, nil
}

func Initialize(serverOpts *options.ServerOption) error {
	globalStorage = NewStorage(serverOpts)
	return nil
}

func Get() *Storage {
	return globalStorage
}
