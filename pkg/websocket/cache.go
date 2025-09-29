package websocket

import (
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

// CacheEntry represents a cached room membership entry
type CacheEntry struct {
	IsMember  bool
	ExpiresAt time.Time
}

// RoomMembershipCache provides caching for room membership checks
type RoomMembershipCache struct {
	cache map[string]CacheEntry
	mu    sync.RWMutex
	ttl   time.Duration
}

// NewRoomMembershipCache creates a new room membership cache
func NewRoomMembershipCache(ttl time.Duration) *RoomMembershipCache {
	cache := &RoomMembershipCache{
		cache: make(map[string]CacheEntry),
		ttl:   ttl,
	}

	// Start cleanup goroutine
	go cache.cleanup()

	return cache
}

// Get retrieves cached membership status
func (c *RoomMembershipCache) Get(userUUID uuid.UUID, roomID int64) (bool, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	key := c.key(userUUID, roomID)
	entry, exists := c.cache[key]

	if !exists || time.Now().After(entry.ExpiresAt) {
		return false, false // Cache miss or expired
	}

	return entry.IsMember, true // Cache hit
}

// Set stores membership status in cache
func (c *RoomMembershipCache) Set(userUUID uuid.UUID, roomID int64, isMember bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	key := c.key(userUUID, roomID)
	c.cache[key] = CacheEntry{
		IsMember:  isMember,
		ExpiresAt: time.Now().Add(c.ttl),
	}
}

// Invalidate removes cached entry
func (c *RoomMembershipCache) Invalidate(userUUID uuid.UUID, roomID int64) {
	c.mu.Lock()
	defer c.mu.Unlock()

	key := c.key(userUUID, roomID)
	delete(c.cache, key)
}

// InvalidateUser removes all cached entries for a user
func (c *RoomMembershipCache) InvalidateUser(userUUID uuid.UUID) {
	c.mu.Lock()
	defer c.mu.Unlock()

	prefix := userUUID.String() + ":"
	for key := range c.cache {
		if len(key) > len(prefix) && key[:len(prefix)] == prefix {
			delete(c.cache, key)
		}
	}
}

// InvalidateRoom removes all cached entries for a room
func (c *RoomMembershipCache) InvalidateRoom(roomID int64) {
	c.mu.Lock()
	defer c.mu.Unlock()

	suffix := fmt.Sprintf(":%d", roomID)
	for key := range c.cache {
		if len(key) > len(suffix) && key[len(key)-len(suffix):] == suffix {
			delete(c.cache, key)
		}
	}
}

// key generates cache key for user-room pair
func (c *RoomMembershipCache) key(userUUID uuid.UUID, roomID int64) string {
	return fmt.Sprintf("%s:%d", userUUID.String(), roomID)
}

// cleanup removes expired entries periodically
func (c *RoomMembershipCache) cleanup() {
	ticker := time.NewTicker(5 * time.Minute) // Cleanup every 5 minutes
	defer ticker.Stop()

	for range ticker.C {
		c.mu.Lock()
		now := time.Now()
		for key, entry := range c.cache {
			if now.After(entry.ExpiresAt) {
				delete(c.cache, key)
			}
		}
		c.mu.Unlock()
	}
}

// CachedRoomMembershipCheckFunc wraps a room membership check function with caching
func CachedRoomMembershipCheckFunc(
	originalFunc RoomMembershipCheckFunc,
	cache *RoomMembershipCache,
) RoomMembershipCheckFunc {
	return func(userUUID uuid.UUID, roomID int64) (bool, error) {
		// Try cache first
		if isMember, found := cache.Get(userUUID, roomID); found {
			return isMember, nil
		}

		// Cache miss, call original function
		isMember, err := originalFunc(userUUID, roomID)
		if err != nil {
			return false, err
		}

		// Cache the result
		cache.Set(userUUID, roomID, isMember)

		return isMember, nil
	}
}
