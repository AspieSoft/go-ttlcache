package ttlcache

import (
	"time"

	"github.com/AspieSoft/go-syncterval"
	"github.com/alphadose/haxmap"
)

type cacheItem[T any] struct {
	value T
	last int64
}

type hashable interface {
	int | int8 | int16 | int32 | int64 | uint | uint8 | uint16 | uint32 | uint64 | uintptr | float32 | float64 | string | complex64 | complex128
}

type Cache[T hashable, J any] struct {
	Touch func(key T)
	Get func(key T) (J, bool)
	Set func(key T, value J)
	Del func(key T)

	TTL func(ttl time.Duration, delInterval ...time.Duration)

	ForEach func(lambda func(key T, value J))

	Len func() uintptr
	MapLen func() uintptr
	Fillrate func() uintptr

	Grow func(newSize uintptr)
}

func New[T hashable, J any](ttl time.Duration, delInterval ...time.Duration) *Cache[T, J] {
	cacheList := haxmap.New[T, cacheItem[J]]()

	life := ttl.Nanoseconds()

	var delInt time.Duration
	if len(delInterval) != 0 {
		delInt = delInterval[0]
	}else{
		delInt = 1 * time.Hour
	}

	syncterval.New(delInt, func() {
		now := time.Now().UnixNano()
		cacheList.ForEach(func(t T, ci cacheItem[J]) {
			if now - ci.last > life {
				cacheList.Del(t)
			}
		})
	})

	cache := Cache[T, J] {
		Touch: func(key T) {
			if item, ok := cacheList.Get(key); ok {
				if time.Now().UnixNano() - item.last > life {
					return
				}
				item.last = time.Now().UnixNano()
			}
		},

		Get: func(key T) (J, bool) {
			item, ok := cacheList.Get(key)
			if ok {
				if time.Now().UnixNano() - item.last > life {
					return item.value, false
				}
				item.last = time.Now().UnixNano()
			}
			return item.value, ok
		},

		Set: func(key T, value J) {
			cacheList.Set(key, cacheItem[J]{value, time.Now().UnixNano()})
		},

		Del: func(key T) {
			cacheList.Del(key)
		},

		TTL: func(ttl time.Duration, delInterval ...time.Duration) {
			life = ttl.Nanoseconds()

			if len(delInterval) != 0 {
				delInt = delInterval[0]
			}
		},

		ForEach: func(lambda func(key T, value J)) {
			cacheList.ForEach(func(t T, ci cacheItem[J]) {
				lambda(t, ci.value)
			})
		},

		Len: func() uintptr {
			l := uintptr(0)
			now := time.Now().UnixNano()
			cacheList.ForEach(func(t T, ci cacheItem[J]) {
				if now - ci.last <= life {
					l++
				}
			})
			return l
		},

		MapLen: func() uintptr {
			return cacheList.Len()
		},

		Fillrate: func() uintptr {
			return cacheList.Fillrate()
		},

		Grow: func(newSize uintptr) {
			cacheList.Grow(newSize)
		},
	}

	return &cache
}
