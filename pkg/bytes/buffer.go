package bytes

import "sync"

// byte池

type Buffer struct {
	buf  []byte
	next *Buffer
}

func (b *Buffer) Bytes() []byte {
	return b.buf
}

type Pool struct {
	lock sync.Mutex
	free *Buffer
	max  int
	num  int
	size int
}

func NewPool(num, size int) (p *Pool) {
	p = new(Pool)
	p.init(num, size)
	return p
}

func (p *Pool) init(num, size int) {
	p.num = num
	p.size = size
	p.max = num * size
	p.grow()
}

func (p *Pool) grow() {
	buf := make([]byte, p.max)
	bs := make([]Buffer, p.num)
	p.free = &bs[0]
	b := p.free
	i := 1
	for i = 1; i < p.num; i++ {
		b.buf = buf[(i-1)*p.size : i*p.size]
		b.next = &bs[i]
		b = b.next
	}
	b.buf = buf[(i-1)*p.size : i*p.size]
	b.next = nil
}

func (p *Pool) Get() (b *Buffer) {
	p.lock.Lock()
	defer p.lock.Unlock()
	if b = p.free; b == nil {
		p.grow()
		b = p.free
	}
	p.free = b.next
	return
}

func (p *Pool) Put(b *Buffer) {
	p.lock.Lock()
	defer p.lock.Unlock()
	b.next = p.free
	p.free = b
}
