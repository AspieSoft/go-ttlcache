package ttlcache

import (
	"errors"
	"testing"
	"time"
)

func Test(t *testing.T) {
	cache := New[string, int](1 * time.Second)

	cache.Set("TestA", 1)

	if _, ok := cache.Get("TestA"); !ok {
		t.Error(errors.New("key TestA does not exist, but it should"))
	}

	time.Sleep(2 * time.Second)

	if _, ok := cache.Get("TestA"); ok {
		t.Error(errors.New("key TestA still exists, but it should not"))
	}

	if l := cache.Len(); l != 0 {
		t.Error(errors.New("empty cache reported length"), l)
	}

	if l := cache.MapLen(); l == 0 {
		t.Error(errors.New("cache map should not be empty yet. reported length"), l)
	}
}
