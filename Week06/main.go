package main

import (
	"fmt"
	"sync"
	"time"
)

type metrics struct {
	success int32
	fail    int32
}

type bucket struct {
	data  metrics
	start int64
}

type Rolling struct {
	mu sync.RWMutex

	size    int64
	width   int64
	tail    int64
	buckets []*bucket
}

func NewRolling(size, width int64) *Rolling {
	return &Rolling{
		size:    size,
		width:   width,
		tail:    0,
		buckets: make([]*bucket, size),
	}
}

func (r *Rolling) GetCurrent() *bucket {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now().Unix()
	last := r.buckets[r.tail]
	if r.tail == 0 && last == nil {
		r.buckets[r.tail] = &bucket{data: metrics{}, start: now}
		return r.buckets[r.tail]
	}

	if now >= last.start+r.width {
		r.tail++
		bk := &bucket{data: metrics{}, start: last.start + r.width}
		if r.tail >= r.size {
			copy(r.buckets[:], r.buckets[1:])
			r.tail--
		}
		r.buckets[r.tail] = bk
	}
	return r.buckets[r.tail]
}

func (r *Rolling) IncrSuccess() {
	bk := r.GetCurrent()

	r.mu.Lock()
	defer r.mu.Unlock()

	bk.data.success++
}

func (r *Rolling) IncrFail() {
	bk := r.GetCurrent()

	r.mu.Lock()
	defer r.mu.Unlock()

	bk.data.fail++
}

func (r *Rolling) GetSum() metrics {
	r.mu.RLock()
	defer r.mu.RUnlock()

	m := metrics{}
	for _, v := range r.buckets {
		if v == nil {
			continue
		}
		m.success += v.data.success
		m.fail += v.data.fail
	}
	return m
}

func main() {
	r := NewRolling(5, 1)
	fmt.Println(time.Now().Unix())

	r.IncrSuccess()
	time.Sleep(time.Second * 1)
	r.IncrSuccess()
	r.IncrFail()
	time.Sleep(time.Second * 1)
	r.IncrSuccess()
	time.Sleep(time.Second * 1)
	fmt.Printf("%+v\n", r.GetSum())
}
