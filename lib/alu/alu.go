// Package alu defines alithmetic and logic units.
package alu

import (
	"slices"

	"github.com/kssilveira/circuit-engine/group"
	"github.com/kssilveira/circuit-engine/lib/bus"
	"github.com/kssilveira/circuit-engine/lib/ram"
	"github.com/kssilveira/circuit-engine/lib/reg"
	"github.com/kssilveira/circuit-engine/lib/sum"
	"github.com/kssilveira/circuit-engine/sfmt"
	"github.com/kssilveira/circuit-engine/wire"
)

// Alu adds an artithmetic and logic unit.
func Alu(parent *group.Group, a, ai, b, bi, ri, ro, cin *wire.Wire) []*wire.Wire {
	group := parent.Group("ALU")
	ra := reg.Register(group, a, ai, group.True)
	rb := reg.Register(group, b, bi, group.True)
	rs := sum.Sum(group, ra, rb, cin)
	rs[1].Name = sfmt.Sprintf("C(%s,%s)", a.Name, b.Name)
	rr := reg.Register(group, rs[0], ri, ro)
	rr.Name = sfmt.Sprintf("R(S(%s,%s))", a.Name, b.Name)
	return []*wire.Wire{ra, rb, rr, rs[1]}
}

// Alu2 adds a 2-bit arithmetic and logic unit.
func Alu2(parent *group.Group, a1, a2, ai, b1, b2, bi, ri, ro, cin *wire.Wire) []*wire.Wire {
	group := parent.Group("ALU2")
	r1 := Alu(group, a1, ai, b1, bi, ri, ro, cin)
	last := len(r1) - 1
	r2 := Alu(group, a2, ai, b2, bi, ri, ro, r1[last])
	return append(r1[:last], r2...)
}

// N adds an N-bit arithmetic and logic unit.
func N(parent *group.Group, a []*wire.Wire, ai *wire.Wire, b []*wire.Wire, bi, ri, ro, cin *wire.Wire) []*wire.Wire {
	group := parent.Group(sfmt.Sprintf("ALU%d", len(a)))
	var res []*wire.Wire
	c := cin
	for j, aj := range a {
		ri := Alu(group, aj, ai, b[j], bi, ri, ro, c)
		last := len(ri) - 1
		res = append(res, ri[:last]...)
		c = ri[last]
	}
	return append(res, c)
}

// WithBus adds an arithmetic logic unit with a communication bus.
func WithBus(parent *group.Group, d, ai, bi, ri, ro, c *wire.Wire) []*wire.Wire {
	group := parent.Group("ALU-BUS")
	a := &wire.Wire{Name: sfmt.Sprintf("%sa", d.Name)}
	ra := reg.Register(group, a, ai, group.True)
	b := &wire.Wire{Name: sfmt.Sprintf("%sb", d.Name)}
	rb := reg.Register(group, b, bi, group.True)
	rs := sum.Sum(group, ra, rb, c)
	rr := reg.Register(group, rs[0], ri, ro)
	rbus := bus.Bus(group, d, rr, a, b)
	return append(rbus, ra, rb, rr, rs[1])
}

// WithBusInputValidation validates inputs with bus.
func WithBusInputValidation(ai, bi, ri, ro *wire.Wire) func() bool {
	return func() bool {
		return !(ri.Bit.Get(nil) && ro.Bit.Get(nil) && (ai.Bit.Get(nil) || bi.Bit.Get(nil)))
	}
}

// WithBus2 adds a 2-bit arithmetic logic unit with a communication bus.
func WithBus2(parent *group.Group, bus1, bus2, ai, bi, ri, ro, cin *wire.Wire) []*wire.Wire {
	group := parent.Group("ALU-BUS2")
	alu1 := WithBus(group, bus1, ai, bi, ri, ro, cin)
	last := len(alu1) - 1
	alu2 := WithBus(group, bus2, ai, bi, ri, ro, alu1[last])
	return slices.Concat(alu1[:last], alu2)
}

// WithBusN adds an N-bit arithmetic logic unit with a communication bus.
func WithBusN(parent *group.Group, d []*wire.Wire, ai, bi, ri, ro, c *wire.Wire) []*wire.Wire {
	group := parent.Group(sfmt.Sprintf("ALU-BUS%d", len(d)))
	prev := c
	var res []*wire.Wire
	for _, di := range d {
		alu := WithBus(group, di, ai, bi, ri, ro, prev)
		last := len(alu) - 1
		prev = alu[last]
		res = append(res, alu[:last]...)
	}
	return append(res, prev)
}

// WithRAM adds an arithmetic logic unit with RAM.
func WithRAM(parent *group.Group, d, ai, bi, ri, ro, c, mai, mi, mo *wire.Wire) []*wire.Wire {
	group := parent.Group("ALU-RAM")
	a := &wire.Wire{Name: sfmt.Sprintf("%sa", d.Name)}
	ra := reg.Register(group, a, ai, group.True)
	b := &wire.Wire{Name: sfmt.Sprintf("%sb", d.Name)}
	rb := reg.Register(group, b, bi, group.True)
	r := sum.Sum(group, ra, rb, c)
	rr := reg.Register(group, r[0], ri, ro)
	ma := &wire.Wire{Name: sfmt.Sprintf("%sma", d.Name)}
	rma := reg.Register(group, ma, mai, group.True)
	m := &wire.Wire{Name: sfmt.Sprintf("%sm", d.Name)}
	rm := ram.RAM(group, []*wire.Wire{rma}, []*wire.Wire{m}, mi, mo)
	rd := bus.IOn(group, append([]*wire.Wire{d, rr}, rm...), []*wire.Wire{a, b, m, ma})
	return slices.Concat(append(rd, r[1], ra, rb, rr, rma), rm)
}

// WithRAMInputValidation validates inputs with ram.
func WithRAMInputValidation(ai, bi, ri, ro, mai, mi, mo *wire.Wire) func() bool {
	return func() bool {
		return WithBusInputValidation(ai, bi, ri, ro)() &&
			!(mai.Bit.Get(nil) && mo.Bit.Get(nil))
	}
}
