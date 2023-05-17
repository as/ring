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
