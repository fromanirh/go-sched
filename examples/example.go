package main

import (
	"fmt"
	sched "github.com/fromanirh/go-sched"
	"math/rand"
	"time"
)

func main() {
	r := rand.New(rand.NewSource(99))
	s := sched.New()

	for i := 0; i < 10; i += 1 {
		n := r.Intn(5)
		d := time.Duration(n) * time.Second
		s.Enter(d, func() { fmt.Println("Call", n) })
	}

	// Events will be called in order of their specified delay
	s.Run()
}
