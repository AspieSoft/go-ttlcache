# Go TTL Cache

[![donation link](https://img.shields.io/badge/buy%20me%20a%20coffee-paypal-blue)](https://paypal.me/shaynejrtaylor?country.x=US&locale.x=en_US)

A Simple Cache That Expires Items.

When getting an item from the cache, the ttl is updated and reset to the current time. This helps prevent frequently accessed items from expiring.

This package uses the [haxmap](https://github.com/alphadose/haxmap) package for better performance.

## Installation

```shell script

  go get github.com/AspieSoft/go-ttlcache

```

## Usage

```go

import (
  "time"

  "github.com/AspieSoft/go-ttlcache"
)

var cache *ttlcache.Cache[string, interface{}]

func main(){
  // create a new cache with a time to live
  cache = ttlcache.New[string, interface{}](2 * time.Hour)

  // optional auto deletion interval (default: 1 hour)
  cache = ttlcache.New[string, interface{}](2 * time.Hour, 4 * time.Hour)

  cache.Set("Item1", 10)
  cache.Set("Item2", 20)

  if value, ok := cache.Get("Item1"); ok {
    // value = 10
  }

  cache.Len() // returns the number of items that have not expired

  cache.MapLen() // returns the total number of items actually stored in the cache (expired items may still be stored, but the ok value will return false if expired)

  if value, ok := cache.Get("Item2"); !ok {
    // Item2 has expired, but the cache may still hold this value and return it
  }

  // reset the cache expire time for an item to keep this value longer
  cache.Touch("Item2")

  // change the time to live
  cache.TTL(1 * time.Hour)

  // optional also change auto deletion interval
  cache.TTL(12 * time.Hour, 24 * time.Hour)

  // clear expired items from cache
  // this is done automatically on an interval, but you can run it manually if you want
  cache.ClearExpired()

  // clear items early if they were last accessed longer than a given time
  // this can be useful if you detect heavy memory usage, and need to shrink the cache sooner then usual
  cache.ClearEarly(2 * time.Hour)

  // remove the optional parameter to clear the entire cache
  cache.ClearEarly()

}

```
