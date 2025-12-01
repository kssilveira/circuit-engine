// Package alu defines alithmetic and logic units.
package alu

import (
	"slices"

	"github.com/kssilveira/circuit-engine/group"
	"github.com/kssilveira/circuit-engine/lib/bus"
	"github.com/kssilveira/circuit-engine/lib/reg"
	"github.com/kssilveira/circuit-engine/lib/sum"
	"github.com/kssilveira/circuit-engine/sfmt"
	"github.com/kssilveira/circuit-engine/wire"
)

// Alu adds an artithmetic and logic unit.
func Alu(parent *group.Group, a, ai, ao, b, bi, bo, ri, ro, cin *wire.Wire) []*wire.Wire {
	group := parent.Group("ALU")
	ra := reg.Register(group, a, ai, ao)
	rb := reg.Register(group, b, bi, bo)
	rs := sum.Sum(group, ra[0], rb[0], cin)
	rr := reg.Register(group, rs[0], ri, ro)
	return slices.Concat(ra, rb, rr, []*wire.Wire{rs[1]})
}

// Alu2 adds a 2-bit arithmetic and logic unit.
func Alu2(parent *group.Group, a1, a2, ai, ao, b1, b2, bi, bo, ri, ro, cin *wire.Wire) []*wire.Wire {
	group := parent.Group("ALU2")
	r1 := Alu(group, a1, ai, ao, b1, bi, bo, ri, ro, cin)
	last := len(r1) - 1
	r2 := Alu(group, a2, ai, ao, b2, bi, bo, ri, ro, r1[last])
	return append(r1[:last], r2...)
}

// Alu4 adds a 4-bit arithmetic and logic unit.
func Alu4(parent *group.Group, a [4]*wire.Wire, ai, ao *wire.Wire, b [4]*wire.Wire, bi, bo, ri, ro, cin *wire.Wire) []*wire.Wire {
	group := parent.Group("ALU4")
	r1 := Alu2(group, a[0], a[1], ai, ao, b[0], b[1], bi, bo, ri, ro, cin)
	last := len(r1) - 1
	r2 := Alu2(group, a[2], a[3], ai, ao, b[2], b[3], bi, bo, ri, ro, r1[last])
	return append(r1[:last], r2...)
}

// Alu8 adds an 8-bit arithmetic and logic unit.
func Alu8(parent *group.Group, a [8]*wire.Wire, ai, ao *wire.Wire, b [8]*wire.Wire, bi, bo, ri, ro, cin *wire.Wire) []*wire.Wire {
	group := parent.Group("ALU8")
	r1 := Alu4(group, wire.W4(a[:4]), ai, ao, wire.W4(b[:4]), bi, bo, ri, ro, cin)
	last := len(r1) - 1
	r2 := Alu4(group, wire.W4(a[4:]), ai, ao, wire.W4(b[4:]), bi, bo, ri, ro, r1[last])
	return append(r1[:last], r2...)
}

// WithBus adds an arithmetic logic unit with a communication bus.
func WithBus(parent *group.Group, busa, ai, ao, bi, bo, ri, ro, cin *wire.Wire) []*wire.Wire {
	group := parent.Group("ALU-BUS")
	a := &wire.Wire{Name: sfmt.Sprintf("ALU-%s-a", busa.Name)}
	ra := reg.Register(group, a, ai, ao)
	b := &wire.Wire{Name: sfmt.Sprintf("ALU-%s-b", busa.Name)}
	rb := reg.Register(group, b, bi, bo)
	rs := sum.Sum(group, ra[0], rb[0], cin)
	rr := reg.Register(group, rs[0], ri, ro)
	rbus := bus.Bus(group, busa, ra[1], rb[1], rr[1], a, b)
	return slices.Concat(rbus, ra, rb, rr, []*wire.Wire{rs[1]})
}

// WithBusInputValidation validates inputs with bus.
func WithBusInputValidation(ai, ao, bi, bo, ri, ro *wire.Wire) func() bool {
	return func() bool {
		return !(ri.Bit.Get(nil) && ro.Bit.Get(nil) && (ai.Bit.Get(nil) || bi.Bit.Get(nil))) &&
			!(ai.Bit.Get(nil) && ao.Bit.Get(nil)) &&
			!(bi.Bit.Get(nil) && bo.Bit.Get(nil))
	}
}

// WithBus2 adds a 2-bit arithmetic logic unit with a communication bus.
func WithBus2(parent *group.Group, bus1, bus2, ai, ao, bi, bo, ri, ro, cin *wire.Wire) []*wire.Wire {
	group := parent.Group("ALU-BUS2")
	alu1 := WithBus(group, bus1, ai, ao, bi, bo, ri, ro, cin)
	last := len(alu1) - 1
	alu2 := WithBus(group, bus2, ai, ao, bi, bo, ri, ro, alu1[last])
	return slices.Concat(alu1[:last], alu2)
}

// WithBus4 adds a 4-bit arithmetic logic unit with a communication bus.
func WithBus4(parent *group.Group, bus [4]*wire.Wire, ai, ao, bi, bo, ri, ro, cin *wire.Wire) []*wire.Wire {
	group := parent.Group("ALU-BUS4")
	alu1 := WithBus2(group, bus[0], bus[1], ai, ao, bi, bo, ri, ro, cin)
	last := len(alu1) - 1
	alu2 := WithBus2(group, bus[2], bus[3], ai, ao, bi, bo, ri, ro, alu1[last])
	return slices.Concat(alu1[:last], alu2)
}

// WithBus8 adds an 8-bit arithmetic logic unit with a communication bus.
func WithBus8(parent *group.Group, bus [8]*wire.Wire, ai, ao, bi, bo, ri, ro, cin *wire.Wire) []*wire.Wire {
	group := parent.Group("ALU-BUS8")
	alu1 := WithBus4(group, wire.W4(bus[:4]), ai, ao, bi, bo, ri, ro, cin)
	last := len(alu1) - 1
	alu2 := WithBus4(group, wire.W4(bus[4:]), ai, ao, bi, bo, ri, ro, alu1[last])
	return slices.Concat(alu1[:last], alu2)
}
