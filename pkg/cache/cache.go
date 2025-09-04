package cache

import (
	"context"
	"sync"
	"time"
)

// CacheItem : 캐시 항목을 나타내는 구조체
type CacheItem struct {
	Value     interface{}
	ExpiresAt time.Time
}

// Cache : 메모리 기반 캐시 구조체
type Cache struct {
	mu    sync.RWMutex
	items map[string]*CacheItem
	ttl   time.Duration
}

// NewCache : 새로운 캐시 인스턴스 생성
func NewCache(ttl time.Duration) *Cache {
	cache := &Cache{
		items: make(map[string]*CacheItem),
		ttl:   ttl,
	}

	// 주기적으로 만료된 항목 정리
	go cache.cleanup()

	return cache
}

// Set : 캐시에 값을 설정합니다
func (c *Cache) Set(key string, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items[key] = &CacheItem{
		Value:     value,
		ExpiresAt: time.Now().Add(c.ttl),
	}
}

// Get : 캐시에서 값을 가져오기
func (c *Cache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, exists := c.items[key]
	if !exists {
		return nil, false
	}

	if time.Now().After(item.ExpiresAt) {
		// 만료된 항목은 삭제하고 nil 반환
		delete(c.items, key)
		return nil, false
	}

	return item.Value, true
}

// Delete : 캐시에서 특정 키를 삭제하기
func (c *Cache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.items, key)
}

// Clear : 모든 캐시를 삭제하기
func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items = make(map[string]*CacheItem)
}

// cleanup : 만료된 항목들을 주기적으로 정리하기
func (c *Cache) cleanup() {
	ticker := time.NewTicker(c.ttl / 2) // TTL의 절반 주기로 정리
	defer ticker.Stop()

	for range ticker.C {
		c.mu.Lock()
		now := time.Now()
		for key, item := range c.items {
			if now.After(item.ExpiresAt) {
				delete(c.items, key)
			}
		}
		c.mu.Unlock()
	}
}

// GetOrSet : 캐시에서 값을 가져오거나, 없으면 함수를 실행하여 설정하기
func (c *Cache) GetOrSet(key string, fn func() (interface{}, error)) (interface{}, error) {
	// 먼저 캐시에서 확인
	if value, exists := c.Get(key); exists {
		return value, nil
	}

	// 캐시에 없으면 함수 실행
	value, err := fn()
	if err != nil {
		return nil, err
	}

	// 성공하면 캐시에 저장
	c.Set(key, value)
	return value, nil
}

// WithContext : 컨텍스트와 함께 캐시 작업을 수행하기
func (c *Cache) WithContext(ctx context.Context, key string, fn func() (interface{}, error)) (interface{}, error) {
	// 컨텍스트가 이미 취소되었는지 확인
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	return c.GetOrSet(key, func() (interface{}, error) {
		// 함수 실행 중에도 컨텍스트 확인
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
			return fn()
		}
	})
}
