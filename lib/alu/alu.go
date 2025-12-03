// Package alu defines alithmetic and logic units.
package alu

import (
	"slices"

	"github.com/kssilveira/circuit-engine/group"
	"github.com/kssilveira/circuit-engine/lib/bus"
	"github.com/kssilveira/circuit-engine/lib/decode"
	"github.com/kssilveira/circuit-engine/lib/gate"
	"github.com/kssilveira/circuit-engine/lib/latch"
	"github.com/kssilveira/circuit-engine/lib/ram"
	"github.com/kssilveira/circuit-engine/lib/reg"
	"github.com/kssilveira/circuit-engine/lib/sum"
	"github.com/kssilveira/circuit-engine/sfmt"
	"github.com/kssilveira/circuit-engine/wire"
)

// Alu adds an artithmetic and logic unit.
func Alu(parent *group.Group, a, ai, b, bi, ri, ro, cin *wire.Wire) []*wire.Wire {
	group := parent.Group("ALU")
	ra := reg.Register(group, a, ai, group.True())
	rb := reg.Register(group, b, bi, group.True())
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
	ra := reg.Register(group, a, ai, group.True())
	b := &wire.Wire{Name: sfmt.Sprintf("%sb", d.Name)}
	rb := reg.Register(group, b, bi, group.True())
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
func WithRAM(parent *group.Group, d []*wire.Wire, ai, bi, ri, ro, c, mai, mi, mo *wire.Wire) []*wire.Wire {
	group := parent.Group("ALU-RAM")
	var a, b, ma, m []*wire.Wire
	for _, di := range d {
		a = append(a, &wire.Wire{Name: sfmt.Sprintf("%sa", di.Name)})
		b = append(b, &wire.Wire{Name: sfmt.Sprintf("%sb", di.Name)})
		ma = append(ma, &wire.Wire{Name: sfmt.Sprintf("%sma", di.Name)})
		m = append(m, &wire.Wire{Name: sfmt.Sprintf("%sm", di.Name)})
	}
	ra := reg.N(group, a, ai, group.True())
	rb := reg.N(group, b, bi, group.True())
	r := sum.N(group, ra, rb, c)
	last := len(r) - 1
	r[last].Name = sfmt.Sprintf("C(%s,%s)", a[last-1].Name, b[last-1].Name)
	rr := reg.N(group, r[:last], ri, ro)
	for i, ai := range a {
		rr[i].Name = sfmt.Sprintf("R(S(%s,%s))", ai.Name, b[i].Name)
	}
	rma := reg.N(group, ma, mai, group.True())
	rm := ram.RAM(group, rma, m, mi, mo)
	rd := bus.BnIOn(group, append([][]*wire.Wire{d, rr}, rm...), [][]*wire.Wire{a, b, m, ma})

	return slices.Concat(append(rd, r[last]), ra, rb, rr, rma, slices.Concat(rm...))
}

// WithRAMInputValidation validates inputs with ram.
func WithRAMInputValidation(ai, bi, ri, ro, mai, _, mo *wire.Wire) func() bool {
	return func() bool {
		return WithBusInputValidation(ai, bi, ri, ro)() &&
			!(mai.Bit.Get(nil) && mo.Bit.Get(nil))
	}
}

// WithCPU adds an arithmetic logic unit with CPU.
func WithCPU(parent *group.Group, e *wire.Wire, n int) []*wire.Wire {
	group := parent.Group("CPU")

	// step counter for microcode
	s := latch.CounterN(group, e, n)

	// current step
	sel := decode.Decode(group, s)
	for i := range sel {
		sel[i] = gate.And(group, sel[i], gate.Not(group, e))
	}

	// step 0: program counter out (co), memory address in (mi)
	co := gate.Or(group, sel[0], group.False())
	co.Name = "co"
	mi := gate.Or(group, sel[0], sel[2])
	mi.Name = "mi"

	// step 1: ram out (ro), instruction register in (ii), program counter increment (ce)
	ro := gate.Or(group, sel[1], sel[3])
	ro.Name = "ro"
	ii := gate.Or(group, sel[1], group.False())
	ii.Name = "ii"
	ce := gate.Or(group, sel[1], group.False())
	ce.Name = "ce"

	// step 2: program counter in (ci), instruction out (io)
	ci := gate.Or(group, sel[2], group.False())
	ci.Name = "ci"
	io := gate.Or(group, sel[2], group.False())
	io.Name = "io"

	// step 3: a register in (ai), ram memory out (ro, above)
	ai := gate.Or(group, sel[3], group.False())
	ai.Name = "ai"

	// b register in
	bi := &wire.Wire{Name: "bi"}
	// ram in
	ri := &wire.Wire{Name: "ri"}
	// total register in
	ti := &wire.Wire{Name: "ti"}
	// total register out
	to := &wire.Wire{Name: "to"}

	var a, b, d, i, m, r []*wire.Wire
	for bit := 0; bit < n; bit++ {
		// a register
		a = append(a, &wire.Wire{Name: "a"})
		// b register
		b = append(b, &wire.Wire{Name: "b"})
		// bus data
		d = append(d, &wire.Wire{Name: "d"})
		// instruction register
		i = append(i, &wire.Wire{Name: "i"})
		// memory address register
		m = append(m, &wire.Wire{Name: "m"})
		// ram output
		r = append(r, &wire.Wire{Name: "r"})
	}

	// a register
	ar := reg.N(group, a, ai, group.True())
	// b register
	br := reg.N(group, b, bi, group.True())
	// total
	t := sum.N(group, ar, br, group.False())
	// carry out
	last := len(t) - 1
	t[last].Name = sfmt.Sprintf("C(%s,%s)", a[last-1].Name, b[last-1].Name)
	// total register
	tr := reg.N(group, t[:last], ti, to)
	for i, ai := range a {
		tr[i].Name = sfmt.Sprintf("RS%s%s", ai.Name, b[i].Name)
	}

	// program counter register
	cr := reg.N(group, latch.CounterN(group, ce, n), ci, co)
	// instruction register
	ir := reg.N(group, i, ii, io)
	// memory address register
	mr := reg.N(group, m, mi, group.True())
	// ram output
	rr := ram.RAM(group, mr, r, ri, ro)
	// bus data
	dr := bus.BnIOn(group, append([][]*wire.Wire{d, cr, ir}, rr...), [][]*wire.Wire{a, b, i, m, r})

	return slices.Concat(ar, br, cr, dr, s, tr, ir, mr, slices.Concat(rr...))
}
