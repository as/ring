# ring
Ring is a lock-free ring buffer without memory allocations

# usage
```
package main

import (
        "fmt"
        "time"

        "github.com/as/ring"
)

func main() {
        r := ring.Buf{TTL: time.Second}
        r.Put("a", "0")
        fmt.Println(r.Get("a")) // 0 true
        fmt.Println(r.Get("b")) //  false
        time.Sleep(time.Second)
        fmt.Println(r.Get("a")) // 0 false
        fmt.Println(r.Get("b")) //  false
        r.Del("a")
        fmt.Println(r.Get("a")) // false
}
```

# godoc

```
type Buf struct {
        TTL time.Duration
}
    Buf is a lock-free time-aware ring buffer. The zero value is ready to use
    and has an expiry time of 20 seconds. Buf retains all values in memory until
    they are overwritten, but expires entries based on their time of access.

    Buf has several properties:

    (1): It is safe to call Put, Get, and Del concurrently
    (2): Memory will never be realloced for the internal ring buffer
    (3): The last Put value will be found first by Get 
    (4): Expired values are returned intact until they are overwritten
    (5): A values is overwritten after 256+ calls to Put, regardless of expiry time

    To use an infinite expiry time, set TTL to a large value. The zero
    value means 20 seconds. A good choice is 24*time.Hour.

func (c *Buf) Del(key string)
    Del evicts the key. It does not remove the key from memory. This is only
    useful if the cache has a high TTL or the associatd value is an empty
    string

func (c *Buf) Get(key string) (value string, ok bool)
    Get returns the value for key. There are three possibilities:

    (1): key is found, and not expired:
        value != "" and ok == true

    (2): key is not found
        value == "" and ok == false

    (3): key is found, and is expired:
        value != "" and ok == false

    The last case is also possible if value was stored as the empty string. It
    is not safe to modify TTL and call Get concurrently

func (c *Buf) Put(key, value string)
    Put inserts the key value pair into the ring, with an expiry of TTL
    It is not safe to modify TTL and call Put concurrently
```

# benchmarks

## Size=256 (default)

```
goos: linux
goarch: amd64
pkg: github.com/as/ring
cpu: Intel(R) Core(TM) i7-9800X CPU @ 3.80GHz
BenchmarkBuf
BenchmarkBuf/Put
BenchmarkBuf/Put-16     41597066                28.37 ns/op            0 B/op          0 allocs/op
BenchmarkBuf/Get
BenchmarkBuf/Get-16     56317138                21.51 ns/op            0 B/op          0 allocs/op
BenchmarkBuf/Parallel
BenchmarkBuf/Parallel/0Writers
BenchmarkBuf/Parallel/0Writers-16                9960812               120.3 ns/op             0 B/op          0 allocs/op
BenchmarkBuf/Parallel/1Writers
BenchmarkBuf/Parallel/1Writers-16               131639803                9.373 ns/op           0 B/op          0 allocs/op
BenchmarkBuf/Parallel/2Writers
BenchmarkBuf/Parallel/2Writers-16               70686408                16.49 ns/op            0 B/op          0 allocs/op
BenchmarkBuf/Parallel/4Writers
BenchmarkBuf/Parallel/4Writers-16               47318340                25.19 ns/op            0 B/op          0 allocs/op
BenchmarkBufHuge
BenchmarkBufHuge/Put
BenchmarkBufHuge/Put-16                         41794219                28.07 ns/op            0 B/op          0 allocs/op
BenchmarkBufHuge/Get
BenchmarkBufHuge/Get-16                         56292502                21.32 ns/op            0 B/op          0 allocs/op
BenchmarkBufHuge/Parallel
BenchmarkBufHuge/Parallel/0Writers
BenchmarkBufHuge/Parallel/0Writers-16           27003477                38.72 ns/op            0 B/op          0 allocs/op
BenchmarkBufHuge/Parallel/1Writers
BenchmarkBufHuge/Parallel/1Writers-16           135755607                9.483 ns/op           0 B/op          0 allocs/op
BenchmarkBufHuge/Parallel/2Writers
BenchmarkBufHuge/Parallel/2Writers-16           71082435                17.62 ns/op            0 B/op          0 allocs/op
BenchmarkBufHuge/Parallel/4Writers
BenchmarkBufHuge/Parallel/4Writers-16           49010296                24.83 ns/op            0 B/op          0 allocs/op
PASS
ok      github.com/as/ring      30.182s
```

## Size=8

```
goos: linux
goarch: amd64
pkg: github.com/as/ring
cpu: Intel(R) Core(TM) i7-9800X CPU @ 3.80GHz
BenchmarkBuf
BenchmarkBuf/Put
BenchmarkBuf/Put-16     41675277                28.42 ns/op            0 B/op          0 allocs/op
BenchmarkBuf/Get
BenchmarkBuf/Get-16     55340379                21.60 ns/op            0 B/op          0 allocs/op
BenchmarkBuf/Parallel
BenchmarkBuf/Parallel/0Writers
BenchmarkBuf/Parallel/0Writers-16               312522607                3.821 ns/op           0 B/op          0 allocs/op
BenchmarkBuf/Parallel/1Writers
BenchmarkBuf/Parallel/1Writers-16               111573422               10.04 ns/op            0 B/op          0 allocs/op
BenchmarkBuf/Parallel/2Writers
BenchmarkBuf/Parallel/2Writers-16               81307870                18.64 ns/op            0 B/op          0 allocs/op
BenchmarkBuf/Parallel/4Writers
BenchmarkBuf/Parallel/4Writers-16               45487818                26.15 ns/op            0 B/op          0 allocs/op
BenchmarkBufHuge
BenchmarkBufHuge/Put
BenchmarkBufHuge/Put-16                         41687305                28.22 ns/op            0 B/op          0 allocs/op
BenchmarkBufHuge/Get
BenchmarkBufHuge/Get-16                         55196476                21.36 ns/op            0 B/op          0 allocs/op
BenchmarkBufHuge/Parallel
BenchmarkBufHuge/Parallel/0Writers
BenchmarkBufHuge/Parallel/0Writers-16           808220512                1.446 ns/op           0 B/op          0 allocs/op
BenchmarkBufHuge/Parallel/1Writers
BenchmarkBufHuge/Parallel/1Writers-16           121505556               10.88 ns/op            0 B/op          0 allocs/op
BenchmarkBufHuge/Parallel/2Writers
BenchmarkBufHuge/Parallel/2Writers-16           60795496                17.12 ns/op            0 B/op          0 allocs/op
BenchmarkBufHuge/Parallel/4Writers
BenchmarkBufHuge/Parallel/4Writers-16           43421089                26.63 ns/op            0 B/op          0 allocs/op
PASS
ok      github.com/as/ring      22.260s
```
