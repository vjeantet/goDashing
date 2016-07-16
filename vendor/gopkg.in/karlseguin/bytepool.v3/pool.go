// Package bytepool provides a pool of []byte
package bytepool

import (
	"encoding/binary"
	"sync/atomic"
)

type Pool struct {
	depleted int64
	expanded int64
	size     int
	list     chan *Bytes
	enc      binary.ByteOrder
	stats    map[string]int64
}

// Create a new pool. The pool contains count items. Each item allocates
// an array of size bytes (but can dynamically grow)
func New(size, count int) *Pool {
	return NewEndian(size, count, binary.BigEndian)
}

func NewEndian(size, count int, enc binary.ByteOrder) *Pool {
	pool := &Pool{
		enc:   enc,
		size:  size,
		list:  make(chan *Bytes, count),
		stats: map[string]int64{"depleted": 0, "expanded": 0},
	}
	for i := 0; i < count; i++ {
		pool.list <- newPooled(pool, size, enc)
	}
	return pool
}

// Get an item from the pool
func (p *Pool) Checkout() *Bytes {
	select {
	case bytes := <-p.list:
		return bytes
	default:
		atomic.AddInt64(&p.depleted, 1)
		return NewEndianBytes(p.size, p.enc)
	}
}

// Exposes every item currently in the pool
// If an item is checkout, each won't see it.
// As such, though thread-safe, you probably only
// want to call this method on init/startup.
func (p *Pool) Each(f func(*Bytes)) {
	l := len(p.list)
	t := make([]*Bytes, l)
	defer func() {
		for i := 0; i < len(t); i++ {
			t[i].Release()
		}
	}()
	for i := 0; i < l; i++ {
		b := <-p.list
		t[i] = b
		f(b)
	}
}

// Get a count of how often Checkout() was called
// but no item was available (thus causing an item to be
// created on the fly)
// Calling this resets the counter
func (p *Pool) Depleted() int64 {
	return atomic.SwapInt64(&p.depleted, 0)
}

// Get a count of how often we had to expand an item
// beyond the initially specified size
// Calling this resets the counter
func (p *Pool) Expanded() int64 {
	return atomic.SwapInt64(&p.expanded, 0)
}

// A map containing the "expanded" and "depleted" count
// Call this resets both counters
func (p *Pool) Stats() map[string]int64 {
	p.stats["depleted"] = p.Depleted()
	p.stats["expanded"] = p.Expanded()
	return p.stats
}

func (p *Pool) onExpand() {
	atomic.AddInt64(&p.expanded, 1)
}
