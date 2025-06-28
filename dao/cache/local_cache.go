package cache

import (
	"sync"
	"time"
)

// LocalCache 本地内存缓存，作为Redis的备用方案
type LocalCache struct {
	data map[string]interface{}
	mu   sync.RWMutex
}

var (
	localCache = &LocalCache{
		data: make(map[string]interface{}),
	}
)

// Set 设置缓存
func (lc *LocalCache) Set(key string, value interface{}, ttl time.Duration) {
	lc.mu.Lock()
	defer lc.mu.Unlock()
	lc.data[key] = value
}

// Get 获取缓存
func (lc *LocalCache) Get(key string) (interface{}, bool) {
	lc.mu.RLock()
	defer lc.mu.RUnlock()
	value, exists := lc.data[key]
	return value, exists
}

// Delete 删除缓存
func (lc *LocalCache) Delete(key string) {
	lc.mu.Lock()
	defer lc.mu.Unlock()
	delete(lc.data, key)
}
