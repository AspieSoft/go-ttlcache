package ttlcache

import (
	"time"

	"github.com/AspieSoft/go-syncterval"
	"github.com/alphadose/haxmap"
)

type hashable interface {
	int | int8 | int16 | int32 | int64 | uint | uint8 | uint16 | uint32 | uint64 | uintptr | float32 | float64 | string | complex64 | complex128
}

type cacheItem[T any] struct {
	value T
	last int64
}

type Cache[T hashable, J any] struct {
	// Touch resets the cache expire time for an item to keep this value longer
	Touch func(key T)

	// Get returns the value of an item
	Get func(key T) (J, bool)

	// Set will either create a new value, or overwrite an existing one
	Set func(key T, value J)

	// Del will remove an item from the cache
	Del func(key T)

	// TTL will change the time to live for cache items before they expire
	// you can optionally change the delInterval for how often the cache will check to auto delete expired items
	TTL func(ttl time.Duration, delInterval ...time.Duration)

	// ForEach runs a function for each cache item in the list
	ForEach func(lambda func(key T, value J))

	// ForEachBreak is like ForEach, but will allow you to break the loop early
	// return `true` to continue the loop
	// return `false` to break the loop
	ForEachBreak func(lambda func(key T, value J) bool)

	// Len returns the number of cache items that have not expired
	Len func() uintptr

	// LenMap returns the number of cache items including those that have expired
	MapLen func() uintptr

	// Fillrate returns the fill rate of the map as a percentage integer
	// this method runs the direct function form haxmap https://github.com/alphadose/haxmap
	Fillrate func() uintptr

	// Grow resizes the hashmap to a new size, gets rounded up to next power of 2 To double the size of the hashmap use newSize 0 This function returns immediately, the resize operation is done in a goroutine No resizing is done in case of another resize operation already being in progress Growth and map bucket policy is inspired from https://github.com/cornelk/hashmap
	// this method runs the direct function form haxmap https://github.com/alphadose/haxmap
	Grow func(newSize uintptr)

	// ClearExpired clears the expired items from the cache
	// this function will automatically run on an interval unless you disable it
	ClearExpired func()

	// ClearEarly allows you to clear cache items before they expire
	// optionally set the `ttl` param to a time, and only items older than that time will be deleted
	ClearEarly func(ttl ...time.Duration)

	list *haxmap.Map[T, cacheItem[J]]
	life int64
	delInt int64
}

func New[T hashable, J any](ttl time.Duration, delInterval ...time.Duration) *Cache[T, J] {
	cache := Cache[T, J] {
		list: haxmap.New[T, cacheItem[J]](),
		life: ttl.Nanoseconds(),
	}

	if len(delInterval) != 0 {
		cache.delInt = delInterval[0].Nanoseconds()
	}else{
		cache.delInt = (1 * time.Hour).Nanoseconds()
	}

	cache.Touch = func(key T) {
		touch(&cache, key)
	}
	cache.Get = func(key T) (J, bool) {
		return get(&cache, key)
	}
	cache.Set = func(key T, value J) {
		set(&cache, key, value)
	}
	cache.Del = func(key T) {
		del(&cache, key)
	}
	cache.TTL = func(ttl time.Duration, delInterval ...time.Duration) {
		setTTL(&cache, ttl, delInterval...)
	}

	cache.ForEach = func(lambda func(key T, value J)) {
		forEach(&cache, lambda)
	}
	cache.ForEachBreak = func(lambda func(key T, value J) bool) {
		forEachBreak(&cache, lambda)
	}

	cache.Len = func() uintptr {
		return getLen(&cache)
	}
	cache.MapLen = func() uintptr {
		return getMapLen(&cache)
	}
	cache.Fillrate = func() uintptr {
		return fillrate(&cache)
	}
	cache.Grow = func(newSize uintptr) {
		grow(&cache, newSize)
	}

	cache.ClearExpired = func(){
		clearExpired(&cache)
	}

	cache.ClearEarly = func(ttl ...time.Duration){
		clearEarly(&cache, ttl...)
	}

	go func(){
		lastRun := time.Now().UnixNano()
		syncterval.New(1 * time.Second, func() {
			now := time.Now().UnixNano()
			if cache.delInt != 0 && now - cache.delInt > lastRun {
				cache.ClearExpired()
			}
		})
	}()

	return &cache
}

func touch[T hashable, J any](cache *Cache[T, J], key T){
	if item, ok := cache.list.Get(key); ok {
		if time.Now().UnixNano() - item.last > cache.life {
			return
		}
		item.last = time.Now().UnixNano()
	}
}

func get[T hashable, J any](cache *Cache[T, J], key T) (J, bool) {
	item, ok := cache.list.Get(key)
	if ok {
		if time.Now().UnixNano() - item.last > cache.life {
			return item.value, false
		}
		item.last = time.Now().UnixNano()
	}
	return item.value, ok
}

func set[T hashable, J any](cache *Cache[T, J], key T, value J){
	cache.list.Set(key, cacheItem[J]{value, time.Now().UnixNano()})
}

func del[T hashable, J any](cache *Cache[T, J], key T){
	cache.list.Del(key)
}

func setTTL[T hashable, J any](cache *Cache[T, J], ttl time.Duration, delInterval ...time.Duration){
	cache.life = ttl.Nanoseconds()

	if len(delInterval) != 0 {
		cache.delInt = delInterval[0].Nanoseconds()
	}
}

func forEach[T hashable, J any](cache *Cache[T, J], lambda func(key T, value J)){
	cache.list.ForEach(func(t T, ci cacheItem[J]) bool {
		lambda(t, ci.value)
		return true
	})
}

func forEachBreak[T hashable, J any](cache *Cache[T, J], lambda func(key T, value J) bool ){
	cache.list.ForEach(func(t T, ci cacheItem[J]) bool {
		return lambda(t, ci.value)
	})
}

func getLen[T hashable, J any](cache *Cache[T, J]) uintptr {
	l := uintptr(0)
	now := time.Now().UnixNano()
	cache.list.ForEach(func(t T, ci cacheItem[J]) bool {
		if now - ci.last <= cache.life {
			l++
		}
		return true
	})
	return l
}

func getMapLen[T hashable, J any](cache *Cache[T, J]) uintptr {
	return cache.list.Len()
}

func fillrate[T hashable, J any](cache *Cache[T, J]) uintptr {
	return cache.list.Fillrate()
}

func grow[T hashable, J any](cache *Cache[T, J], newSize uintptr){
	cache.list.Grow(newSize)
}

func clearExpired[T hashable, J any](cache *Cache[T, J]){
	now := time.Now().UnixNano()
	cache.list.ForEach(func(t T, ci cacheItem[J]) bool {
		if now - ci.last > cache.life {
			cache.list.Del(t)
		}
		return true
	})
}

func clearEarly[T hashable, J any](cache *Cache[T, J], ttl ...time.Duration){
	now := time.Now().UnixNano()

	var life int64
	if len(ttl) != 0 {
		life = ttl[0].Nanoseconds()
	}

	cache.list.ForEach(func(t T, ci cacheItem[J]) bool {
		if now - ci.last > life {
			cache.list.Del(t)
		}
		return true
	})
}
