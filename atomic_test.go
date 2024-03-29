package main

import (
	"math/rand"
	"sync"
	"sync/atomic"
	"testing"
)

func BenchmarkAtomic(b *testing.B) {
	type A struct {
		intValue int
	}

	type B struct {
		a *atomic.Pointer[A]
	}

	testStruct := &B{a: &atomic.Pointer[A]{}}

	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			newA := &A{rand.Intn(100)}
			testStruct.a.Store(newA)
		}
	})

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			testStruct.a.Load()
		}
	})
}

func BenchmarkChannel(b *testing.B) {
	type A struct {
		intValue int
	}

	type B struct {
		a *A
	}

	testStruct := &B{a: &A{}}

	done := make(chan struct{})
	getReq := make(chan struct{})
	getChan := make(chan *A)
	setChan := make(chan *A)

	go func() {
		for {
			select {
			case newA := <-setChan:
				testStruct.a = newA
			case <-getReq:
				getChan <- testStruct.a
			case <-done:
				return
			}
		}
	}()

	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			newA := &A{rand.Intn(100)}
			setChan <- newA
		}
	})

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			getReq <- struct{}{}
			<-getChan
		}
	})

	close(done)
}

func BenchmarkMutex(b *testing.B) {
	type A struct {
		intValue int
	}

	type B struct {
		a *A
	}

	mutex := sync.RWMutex{}

	testStruct := &B{a: &A{}}

	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			newA := &A{rand.Intn(100)}
			mutex.Lock()
			testStruct.a = newA
			mutex.Unlock()
		}
	})

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			mutex.RLock()
			_ = testStruct.a
			mutex.RUnlock()
		}
	})
}

func BenchmarkMapSync(b *testing.B) {
	type A struct {
		intValue int
	}

	m := sync.Map{}

	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			//Bad performance as we update the value in the map everytime
			newA := &A{rand.Intn(100)}
			m.Store("a", newA)
		}
	})

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			m.Load("a")
		}
	})
}

func BenchmarkMapSync2(b *testing.B) {
	type A struct {
		intValue atomic.Int32
	}

	m := sync.Map{}

	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			//Best performance as we're updating only the internal value of the struct pointed by the map
			if v, found := m.Load("a"); found {
				v.(*A).intValue.Store(int32(rand.Intn(100)))
			} else {
				newA := &A{}
				newA.intValue.Store(int32(rand.Intn(100)))
				m.Store("a", newA)
			}
		}
	})

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			if v, found := m.Load("a"); found {
				v.(*A).intValue.Load()
			}
		}
	})
}
