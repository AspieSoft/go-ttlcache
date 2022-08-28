# Go TTL Cache

[![donation link](https://img.shields.io/badge/buy%20me%20a%20coffee-square-blue)](https://buymeacoffee.aspiesoft.com)

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

  cache.Touch("Item2") // reset the cache expire time to keep this value longer

}

```
