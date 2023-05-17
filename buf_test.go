package ring

import (
	"fmt"
	"os"
	"testing"
	"time"
	"unsafe"
)

var printproof = os.Getenv("PRINTPROOF") != ""

func (e entry) String() string {
	k, v := e.key, e.value
	if k == "" {
		k = "?"
	}
	if v == "" {
		v = "?"
	}
	return fmt.Sprintf("%s=%s", k, v)
}

func TestBufSize(t *testing.T) {
	i := (unsafe.Sizeof(uint64(0)))
	d := (unsafe.Sizeof(time.Duration(0)))
	e := (unsafe.Sizeof(entry{}))
	println("est: ", i+4+(e*nent)+d)
	println("est: ", est)
	println("act: ", unsafe.Sizeof(Buf{}))
	println(cacheLinePad)
	println(est + cacheLinePad)
}

func TestBuf(t *testing.T) {
	c := Buf{TTL: time.Second / 2}
	lo := []string{"a", "b", "c", "d", "e", "f", "g", "h", "i"}
	hi := []string{"A", "B", "C", "D", "E", "F", "G", "H", "I"}
	i := 0
	for i = range lo {
		c.Put(fmt.Sprint(lo[i]), hi[i])
		ch, ok := c.Get(lo[i])
		if !ok || ch != hi[i] {
			t.Fatalf("have %q, want %q", ch, hi[i])
		}
	}
	if _, ok := c.Get("A"); ok {
		t.Fatal("space: stale entry in cache")
	}
	c.Put("?", "!")
	if ch, ok := c.Get("?"); !ok || ch != "!" {
		t.Fatalf("have %q, want %q", ch, "!")
	}
	time.Sleep(time.Second)
	if _, ok := c.Get("?"); ok {
		t.Fatal("time: stale entry in cache")
	}
}

func TestBufAlign(t *testing.T) {
	n := unsafe.Sizeof(Buf{})
	if n%cacheLine != 0 {
		t.Fatalf("not aligned by %d bytes", n%cacheLine)
	}
}

func BenchmarkBufGetRecent(b *testing.B) {
	c := Buf{TTL: time.Hour}
	c.Put("x", "y")
	v, _ := c.Get("x")
	for n := 0; n < b.N; n++ {
		v, _ = c.Get("x")
	}
	_ = v

}
func BenchmarkBufGetOld(b *testing.B) {
	c := Buf{TTL: time.Hour}
	c.Put("x", "y")
	for i := 0; i < Size-4; i++ {
		c.Put("a", "b")
	}
	v, _ := c.Get("x")
	for n := 0; n < b.N; n++ {
		v, _ = c.Get("x")
	}
	_ = v

}
func BenchmarkBuf(b *testing.B) {
	c := Buf{TTL: time.Second / 2}
	lo := []string{"a", "b", "c", "d", "e", "f", "g", "h", "i"}
	hi := []string{"A", "B", "C", "D", "E", "F", "G", "H", "I"}
	i := 0

	b.Run("Put", func(b *testing.B) {
		for n := 0; n < b.N; n++ {
			if i > len(lo) {
				i = 0
			}
			c.Put(lo[i], hi[i])
		}
	})
	i = 0

	b.Run("Get", func(b *testing.B) {
		for n := 0; n < b.N; n++ {
			if i > len(lo) {
				i = 0
			}
			c.Get(lo[i])
		}
	})

	b.Run("Parallel", func(b *testing.B) {
		for _, cpu := range []int{1, 2, 4} {
			b.Run(fmt.Sprintf("%dWriters", cpu), func(b *testing.B) {
				done := make(chan bool)
				defer close(done)
				for x := 0; x < cpu; x++ {
					go func() {
						for {
							select {
							case <-done:
								return
							default:
								c.Put("x", "y")
								c.Put("y", "x")
							}
						}
					}()
				}
				b.RunParallel(func(pb *testing.PB) {
					for pb.Next() {
						c.Get("x")
					}
				})
			})
		}
	})

}

func TestParallel(t *testing.T) {
	c := &Buf{}
	go func() {
		for i := 0; i < 100; i++ {
			go func() {
				for j := 0; j < 100000; j++ {
					c.Put("x", "y")
					c.Put("y", "x")
				}
			}()
		}
	}()
	go func() {
		for i := 0; i < 100; i++ {
			go func() {
				for j := 0; j < 100000; j++ {
					c.Get("x")
					c.Get("y")
				}
			}()
		}
	}()
	for i := 0; i < 10000; i++ {
		c.Get("x")
		c.Put("x", "a")
	}

}

func BenchmarkBufHuge(b *testing.B) {
	c := Buf{TTL: time.Second / 2}
	lo := []string{"a", "b", "c", "d", "e", "f", "g", "h", "i"}
	hi := []string{"A", "B", "C", "D", "E", "F", "G", "H", "I"}
	for i := range lo {
		lo[i] = "https://example.com/user/job/" + lo[i]
	}
	for i := range hi {
		hi[i] = html + hi[i]
	}
	i := 0

	b.Run("Put", func(b *testing.B) {
		for n := 0; n < b.N; n++ {
			if i > len(lo) {
				i = 0
			}
			c.Put(lo[i], hi[i])
		}
	})
	i = 0

	b.Run("Get", func(b *testing.B) {
		for n := 0; n < b.N; n++ {
			if i > len(lo) {
				i = 0
			}
			c.Get(lo[i])
		}
	})

	b.Run("Parallel", func(b *testing.B) {
		for _, cpu := range []int{1, 2, 4, 8} {
			b.Run(fmt.Sprintf("%dWriters", cpu), func(b *testing.B) {
				done := make(chan bool)
				defer close(done)
				for x := 0; x < cpu; x++ {
					go func() {
						for {
							select {
							case <-done:
								return
							default:
								c.Put("x", "y")
								c.Put("y", "x")
							}
						}
					}()
				}
				b.RunParallel(func(pb *testing.PB) {
					for pb.Next() {
						c.Get("x")
					}
				})
			})
		}
	})

}

const html = `
	asdflasdjflfsjgdfjgoiprgioqejgipqoerjgiopeqgierjgopqeijrgqioerjgqioerpgjq
	asdflasdjflfsjgdfjgoiprgioqejgipqoerjgiopeqgierjgopqeijrgqioerjgqioerpgjq
	asdflasdjflfsjgdfjgoiprgioqejgipqoerjgiopeqgierjgopqeijrgqioerjgqioerpgjq
	asdflasdjflfsjgdfjgoiprgioqejgipqoerjgiopeqgierjgopqeijrgqioerjgqioerpgjq
	asdflasdjflfsjgdfjgoiprgioqejgipqoerjgiopeqgierjgopqeijrgqioerjgqioerpgjq
	asdflasdjflfsjgdfjgoiprgioqejgipqoerjgiopeqgierjgopqeijrgqioerjgqioerpgjq
	asdflasdjflfsjgdfjgoiprgioqejgipqoerjgiopeqgierjgopqeijrgqioerjgqioerpgjq
	asdflasdjflfsjgdfjgoiprgioqejgipqoerjgiopeqgierjgopqeijrgqioerjgqioerpgjq
	asdflasdjflfsjgdfjgoiprgioqejgipqoerjgiopeqgierjgopqeijrgqioerjgqioerpgjq
	asdflasdjflfsjgdfjgoiprgioqejgipqoerjgiopeqgierjgopqeijrgqioerjgqioerpgjq
	asdflasdjflfsjgdfjgoiprgioqejgipqoerjgiopeqgierjgopqeijrgqioerjgqioerpgjq
	asdflasdjflfsjgdfjgoiprgioqejgipqoerjgiopeqgierjgopqeijrgqioerjgqioerpgjq
	asdflasdjflfsjgdfjgoiprgioqejgipqoerjgiopeqgierjgopqeijrgqioerjgqioerpgjq
	asdflasdjflfsjgdfjgoiprgioqejgipqoerjgiopeqgierjgopqeijrgqioerjgqioerpgjq
	asdflasdjflfsjgdfjgoiprgioqejgipqoerjgiopeqgierjgopqeijrgqioerjgqioerpgjq
	asdflasdjflfsjgdfjgoiprgioqejgipqoerjgiopeqgierjgopqeijrgqioerjgqioerpgjq
	asdflasdjflfsjgdfjgoiprgioqejgipqoerjgiopeqgierjgopqeijrgqioerjgqioerpgjq
	asdflasdjflfsjgdfjgoiprgioqejgipqoerjgiopeqgierjgopqeijrgqioerjgqioerpgjq
	asdflasdjflfsjgdfjgoiprgioqejgipqoerjgiopeqgierjgopqeijrgqioerjgqioerpgjq
	asdflasdjflfsjgdfjgoiprgioqejgipqoerjgiopeqgierjgopqeijrgqioerjgqioerpgjq
	asdflasdjflfsjgdfjgoiprgioqejgipqoerjgiopeqgierjgopqeijrgqioerjgqioerpgjq
	asdflasdjflfsjgdfjgoiprgioqejgipqoerjgiopeqgierjgopqeijrgqioerjgqioerpgjq
	asdflasdjflfsjgdfjgoiprgioqejgipqoerjgiopeqgierjgopqeijrgqioerjgqioerpgjq
	asdflasdjflfsjgdfjgoiprgioqejgipqoerjgiopeqgierjgopqeijrgqioerjgqioerpgjq
	asdflasdjflfsjgdfjgoiprgioqejgipqoerjgiopeqgierjgopqeijrgqioerjgqioerpgjq
	asdflasdjflfsjgdfjgoiprgioqejgipqoerjgiopeqgierjgopqeijrgqioerjgqioerpgjq
	asdflasdjflfsjgdfjgoiprgioqejgipqoerjgiopeqgierjgopqeijrgqioerjgqioerpgjq
	asdflasdjflfsjgdfjgoiprgioqejgipqoerjgiopeqgierjgopqeijrgqioerjgqioerpgjq
	asdflasdjflfsjgdfjgoiprgioqejgipqoerjgiopeqgierjgopqeijrgqioerjgqioerpgjq
	asdflasdjflfsjgdfjgoiprgioqejgipqoerjgiopeqgierjgopqeijrgqioerjgqioerpgjq
	asdflasdjflfsjgdfjgoiprgioqejgipqoerjgiopeqgierjgopqeijrgqioerjgqioerpgjq
	asdflasdjflfsjgdfjgoiprgioqejgipqoerjgiopeqgierjgopqeijrgqioerjgqioerpgjq
	asdflasdjflfsjgdfjgoiprgioqejgipqoerjgiopeqgierjgopqeijrgqioerjgqioerpgjq
	asdflasdjflfsjgdfjgoiprgioqejgipqoerjgiopeqgierjgopqeijrgqioerjgqioerpgjq
	asdflasdjflfsjgdfjgoiprgioqejgipqoerjgiopeqgierjgopqeijrgqioerjgqioerpgjq
	asdflasdjflfsjgdfjgoiprgioqejgipqoerjgiopeqgierjgopqeijrgqioerjgqioerpgjq
	asdflasdjflfsjgdfjgoiprgioqejgipqoerjgiopeqgierjgopqeijrgqioerjgqioerpgjq
	asdflasdjflfsjgdfjgoiprgioqejgipqoerjgiopeqgierjgopqeijrgqioerjgqioerpgjq
	asdflasdjflfsjgdfjgoiprgioqejgipqoerjgiopeqgierjgopqeijrgqioerjgqioerpgjq
	asdflasdjflfsjgdfjgoiprgioqejgipqoerjgiopeqgierjgopqeijrgqioerjgqioerpgjq
	asdflasdjflfsjgdfjgoiprgioqejgipqoerjgiopeqgierjgopqeijrgqioerjgqioerpgjq
	asdflasdjflfsjgdfjgoiprgioqejgipqoerjgiopeqgierjgopqeijrgqioerjgqioerpgjq
	asdflasdjflfsjgdfjgoiprgioqejgipqoerjgiopeqgierjgopqeijrgqioerjgqioerpgjq
	asdflasdjflfsjgdfjgoiprgioqejgipqoerjgiopeqgierjgopqeijrgqioerjgqioerpgjq
	asdflasdjflfsjgdfjgoiprgioqejgipqoerjgiopeqgierjgopqeijrgqioerjgqioerpgjq
	asdflasdjflfsjgdfjgoiprgioqejgipqoerjgiopeqgierjgopqeijrgqioerjgqioerpgjq
	asdflasdjflfsjgdfjgoiprgioqejgipqoerjgiopeqgierjgopqeijrgqioerjgqioerpgjq
	asdflasdjflfsjgdfjgoiprgioqejgipqoerjgiopeqgierjgopqeijrgqioerjgqioerpgjq
	asdflasdjflfsjgdfjgoiprgioqejgipqoerjgiopeqgierjgopqeijrgqioerjgqioerpgjq
	asdflasdjflfsjgdfjgoiprgioqejgipqoerjgiopeqgierjgopqeijrgqioerjgqioerpgjq
	asdflasdjflfsjgdfjgoiprgioqejgipqoerjgiopeqgierjgopqeijrgqioerjgqioerpgjq
	asdflasdjflfsjgdfjgoiprgioqejgipqoerjgiopeqgierjgopqeijrgqioerjgqioerpgjq
	asdflasdjflfsjgdfjgoiprgioqejgipqoerjgiopeqgierjgopqeijrgqioerjgqioerpgjq
	asdflasdjflfsjgdfjgoiprgioqejgipqoerjgiopeqgierjgopqeijrgqioerjgqioerpgjq
	asdflasdjflfsjgdfjgoiprgioqejgipqoerjgiopeqgierjgopqeijrgqioerjgqioerpgjq
	asdflasdjflfsjgdfjgoiprgioqejgipqoerjgiopeqgierjgopqeijrgqioerjgqioerpgjq
	asdflasdjflfsjgdfjgoiprgioqejgipqoerjgiopeqgierjgopqeijrgqioerjgqioerpgjq
	asdflasdjflfsjgdfjgoiprgioqejgipqoerjgiopeqgierjgopqeijrgqioerjgqioerpgjq
	asdflasdjflfsjgdfjgoiprgioqejgipqoerjgiopeqgierjgopqeijrgqioerjgqioerpgjq
	asdflasdjflfsjgdfjgoiprgioqejgipqoerjgiopeqgierjgopqeijrgqioerjgqioerpgjq
	`
