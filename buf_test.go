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
	println(unsafe.Sizeof(Buf{}))
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
func TestBufProps(t *testing.T) {
	if nent > 256 {
		t.Skip("skipping this test, use a smaller Buf size (nent > 256)")
	}
	wrap := nent + 5
	for s := 0; s < wrap; s++ {
		for d := 0; d < wrap; d++ {
			c := Buf{TTL: 1}
			for i := 0; i < s; i++ {
				c.Put("^", "^")
			}
			c.Put("t", "f")
			for i := 0; i < d; i++ {
				c.Put("$", "$")
			}
			c.Put("t", "p")
			for i := 0; i < wrap; i++ {
				v := fmt.Sprintf("%c", 0x41+byte(i))
				have, ok := c.Get("t")
				if have == "" {
					have = "?"
				}
				k := "y"
				if !ok {
					k = "n"
				}
				if printproof {
					t.Logf("s=%02d d=%02d i=%02d x=%02d t=%s ok=%s %+v", s, d, i, c.x, have, k, c.c)
				}
				if have == "f" {
					t.Fatal("invariant violated, got f")
				}
				c.Put(v, v)
			}
		}
	}
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
		for _, cpu := range []int{0, 1, 2, 4} {
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
		for _, cpu := range []int{0, 1, 2, 4} {
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
