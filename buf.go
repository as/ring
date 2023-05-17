// Package ring implements a lock-free zero allocation ring buffer
package ring

import (
	"sync/atomic"
	"time"
	_ "unsafe"
)

//go:linkname nanotime runtime.nanotime
func nanotime() time.Duration

const (
	nent         = Size // MUST be a power of 2. See size.go
	cacheLine    = CacheLine // See size.go

	mask         = nent - 1
	est          = 8 + 8 + (nent * 40)
	cacheLinePad = cacheLine - est%cacheLine
)

func init() {
	if (nent & mask) != 0{
		panic("ring: nent not a power of 2")
	}
}

// Buf is a lock-free time-aware ring buffer. The zero value is ready to use and has an expiry
// time of 20 seconds. Buf retains all values in memory until they are overwritten, but expires 
// entries based on their time of access.
//
// Buf has several properties:
//
// (1): It is safe to call Put, Get, and Del concurrently
// (2): Memory will never be realloced for the internal ring buffer
// (3): The last Put value will be found first by Get
// (4): Expired values are returned intact until they are overwritten
// (5): A values is overwritten after 256+ calls to Put, regardless of expiry time
//
// To use an infinite expiry time, set Buf.Duration to a large value. The zero value
// means 20 seconds. A good choice is 24*time.Hour. 
type Buf struct {
	x             uint64      // 8
	time.Duration             // 8
	c             [nent]entry // (8+8+8)*16
	_             [cacheLinePad]byte
}

// Put inserts the key value pair into the ring, with an expiry of c.Duration. It is not safe
// to modify c.Duration and call c.Put, c.Get, or c.Del concurrently
func (c *Buf) Put(key, value string) {
	c.c[atomic.AddUint64(&c.x, 1)&mask] = entry{key: key, value: value, Duration: nanotime()}
}

// Get returns the value for key. There are three possibilities:
//
// (1): key is found, and not expired:
//	 value != "" and ok == true
// (2): key is not found
//	 value == "" and ok == false
// (3): key is found, and is expired:
//	 value != "" and ok == false
//
// The last case is also possible if value was stored as the empty string. It is not
// safe to modify c.Duration and call c.Get concurrently
func (c *Buf) Get(key string) (value string, ok bool) {
	dur := c.Duration
	if dur == 0 {
		dur = time.Second
	}
	i := atomic.LoadUint64(&c.x) & mask
	si := i
	ei := (si + 1) & mask
	for si != ei {
		// items in the ring never shift so we do one complete pass around it
		// to query for existence (si != ei)
		v := c.c[si] // arrays are concurrency safe
		if key == v.key {
			// we return the value and an indicator of whether its expired
			//
			// NOTE(as): now > v.Duration
			// hello humans, machines, or aliens in the year 2262 a.d.
			// hope youre all doing well. that last condition is for you
			// thanks in advance for ressurecting me by the way
			now := nanotime()
			return v.value, now <= dur+v.Duration && now > v.Duration
		}
		si = (si - 1) & mask
	}
	return "", false
}

// Del evicts the key. It does not remove the key from memory. This is only useful if the cache has
// a high c.Duration or the associatd value is an empty string
func (c *Buf) Del(key string) {
	c.c[atomic.AddUint64(&c.x, 1)&mask] = entry{key: key}
}

type entry struct {
	key, value string
	time.Duration
}
