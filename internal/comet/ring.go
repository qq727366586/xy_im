package comet

import (
	"xy_im/api/protocol"
	"xy_im/internal/comet/errors"
)

type Ring struct {
	rp   uint64
	wp   uint64
	num  uint64
	mask uint64
	data []protocol.Proto
}

func NewRing(num int) *Ring {
	r := new(Ring)
	r.init(uint64(num))
	return r
}

func (r *Ring) Init(num int) {
	r.init(uint64(num))
}

func (r *Ring) init(num uint64) {
	// 2^N
	if num&(num-1) != 0 {
		for num&(num-1) != 0 {
			num &= num - 1
		}
		num <<= 1
	}
	r.data = make([]protocol.Proto, num)
	r.num = num
	r.mask = r.num - 1
}

func (r *Ring) Get() (proto *protocol.Proto, err error) {
	if r.rp == r.wp {
		return nil, errors.ErrTimerEmpty
	}
	proto = &r.data[r.rp&r.mask]
	return
}

func (r *Ring) GetAdv() {
	r.rp++
}

func (r *Ring) Set() (proto *protocol.Proto, err error) {
	if r.wp-r.rp >= r.num {
		return nil, errors.ErrRingFull
	}
	proto = &r.data[r.wp&r.mask]
	return
}

func (r *Ring) SetAdv() {
	r.wp++
}

func (r *Ring) Reset() {
	r.rp, r.wp = 0, 0
}
