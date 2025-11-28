package lib

import (
	"fmt"
	"slices"

	"github.com/kssilveira/circuit-engine/circuit"
	"github.com/kssilveira/circuit-engine/group"
	"github.com/kssilveira/circuit-engine/transistor"
	"github.com/kssilveira/circuit-engine/wire"
)

func TransistorEmitter(parent *group.Group, base, collector *wire.Wire) []*wire.Wire {
	group := parent.Group("TransistorEmitter")
	emitter := &wire.Wire{Name: "emitter"}
	collectorOut := &wire.Wire{Name: "collector_out"}
	group.Transistor(base, collector, emitter, collectorOut)
	return []*wire.Wire{emitter}
}

func TransistorGnd(parent *group.Group, base, collector *wire.Wire) []*wire.Wire {
	group := parent.Group("TransistorGnd")
	collectorOut := &wire.Wire{Name: "collector_out"}
	group.Transistor(base, collector, group.Gnd, collectorOut)
	return []*wire.Wire{collectorOut}
}

func Transistor(parent *group.Group, base, collector *wire.Wire) []*wire.Wire {
	group := parent.Group("Transistor")
	emitter := &wire.Wire{Name: "emitter"}
	collectorOut := &wire.Wire{Name: "collector_out"}
	group.Transistor(base, collector, emitter, collectorOut)
	return []*wire.Wire{emitter, collectorOut}
}

func Not(parent *group.Group, a *wire.Wire) *wire.Wire {
	group := parent.Group(fmt.Sprintf("NOT(%v)", a.Name))
	res := &wire.Wire{Name: group.Name}
	group.Transistor(a, group.Vcc, group.Gnd, res)
	return res
}

func And(parent *group.Group, a, b *wire.Wire) *wire.Wire {
	group := parent.Group(fmt.Sprintf("AND(%v,%v)", a.Name, b.Name))
	res := &wire.Wire{Name: group.Name}
	wire := &wire.Wire{Name: fmt.Sprintf("%s-wire", res.Name)}
	group.AddTransistors([]*transistor.Transistor{
		{Base: a, Collector: group.Vcc, Emitter: wire},
		{Base: b, Collector: wire, Emitter: res},
	})
	return res
}

func Or(parent *group.Group, a, b *wire.Wire) *wire.Wire {
	res := &wire.Wire{}
	return OrRes(parent, res, a, b)
}

func OrRes(parent *group.Group, res, a, b *wire.Wire) *wire.Wire {
	group := parent.Group(fmt.Sprintf("OR(%v,%v)", a.Name, b.Name))
	res.Name = group.Name
	wire1 := &wire.Wire{Name: fmt.Sprintf("%s-wire1", res.Name)}
	wire2 := &wire.Wire{Name: fmt.Sprintf("%s-wire2", res.Name)}
	group.AddTransistors([]*transistor.Transistor{
		{Base: a, Collector: group.Vcc, Emitter: wire1},
		{Base: b, Collector: group.Vcc, Emitter: wire2},
	})
	group.JointWire(res, wire1, wire2, false /* isAnd */)
	return res
}

func Nand(parent *group.Group, a, b *wire.Wire) *wire.Wire {
	group := parent.Group(fmt.Sprintf("NAND(%v,%v)", a.Name, b.Name))
	res := &wire.Wire{Name: group.Name}
	wire := &wire.Wire{Name: fmt.Sprintf("%s-wire", res.Name)}
	group.AddTransistors([]*transistor.Transistor{
		{Base: a, Collector: group.Vcc, Emitter: wire, CollectorOut: res},
		{Base: b, Collector: wire, Emitter: group.Gnd},
	})
	return res
}

func Xor(parent *group.Group, a, b *wire.Wire) *wire.Wire {
	group := parent.Group(fmt.Sprintf("XOR(%v,%v)", a.Name, b.Name))
	res := And(group, Or(group, a, b), Nand(group, a, b))
	res.Name = group.Name
	return res
}

func Nor(parent *group.Group, a, b *wire.Wire) *wire.Wire {
	res := &wire.Wire{}
	return NorRes(parent, res, a, b)
}

func NorRes(parent *group.Group, res, a, b *wire.Wire) *wire.Wire {
	group := parent.Group(fmt.Sprintf("NOR(%v,%v)", a.Name, b.Name))
	res.Name = group.Name
	wire1 := &wire.Wire{Name: fmt.Sprintf("%s-wire1", res.Name)}
	wire2 := &wire.Wire{Name: fmt.Sprintf("%s-wire2", res.Name)}
	group.AddTransistors([]*transistor.Transistor{
		{Base: a, Collector: group.Vcc, Emitter: group.Gnd, CollectorOut: wire1},
		{Base: b, Collector: group.Vcc, Emitter: group.Gnd, CollectorOut: wire2},
	})
	group.JointWire(res, wire1, wire2, true /* IsAnd */)
	return res
}

func HalfSum(parent *group.Group, a, b *wire.Wire) []*wire.Wire {
	group := parent.Group(fmt.Sprintf("SUM(%v,%v)", a.Name, b.Name))
	res := Xor(group, a, b)
	res.Name = group.Name
	carry := And(group, a, b)
	carry.Name = fmt.Sprintf("CARRY(%v,%v)", a.Name, b.Name)
	return []*wire.Wire{res, carry}
}

func Sum(parent *group.Group, a, b, cin *wire.Wire) []*wire.Wire {
	group := parent.Group(fmt.Sprintf("SUM(%v,%v,%v)", a.Name, b.Name, cin.Name))
	s1 := HalfSum(group, a, b)
	s2 := HalfSum(group, s1[0], cin)
	s2[0].Name = group.Name
	carry := Or(group, s1[1], s2[1])
	carry.Name = fmt.Sprintf("CARRY(%v,%v)", a.Name, b.Name)
	return []*wire.Wire{s2[0], carry}
}

func Sum2(parent *group.Group, a1, a2, b1, b2, cin *wire.Wire) []*wire.Wire {
	group := parent.Group("SUM2")
	s1 := Sum(group, a1, b1, cin)
	s2 := Sum(group, a2, b2, s1[1])
	return []*wire.Wire{s1[0], s2[0], s2[1]}
}

func Sum4(parent *group.Group, a1, a2, a3, a4, b1, b2, b3, b4, cin *wire.Wire) []*wire.Wire {
	group := parent.Group("SUM4")
	s1 := Sum2(group, a1, a2, b1, b2, cin)
	s2 := Sum2(group, a3, a4, b3, b4, s1[2])
	return []*wire.Wire{s1[0], s1[1], s2[0], s2[1], s2[2]}
}

func Sum8(parent *group.Group, a [8]*wire.Wire, b [8]*wire.Wire, cin *wire.Wire) []*wire.Wire {
	group := parent.Group("SUM8")
	s1 := Sum4(group, a[0], a[1], a[2], a[3], b[0], b[1], b[2], b[3], cin)
	s2 := Sum4(group, a[4], a[5], a[6], a[7], b[4], b[5], b[6], b[7], s1[4])
	return []*wire.Wire{s1[0], s1[1], s1[2], s1[3], s2[0], s2[1], s2[2], s2[3], s2[4]}
}

func SRLatch(parent *group.Group, s, r *wire.Wire) []*wire.Wire {
	q := &wire.Wire{Name: "q"}
	return SRLatchRes(parent, q, s, r)
}

func SRLatchRes(parent *group.Group, q, s, r *wire.Wire) []*wire.Wire {
	group := parent.Group(fmt.Sprintf("SRLATCH(%v,%v)", s.Name, r.Name))
	nq := &wire.Wire{Name: "nq"}
	NorRes(group, q, r, nq)
	NorRes(group, nq, s, q)
	return []*wire.Wire{q, nq}
}

func SRLatchWithEnable(parent *group.Group, s, r, e *wire.Wire) []*wire.Wire {
	q := &wire.Wire{Name: "q"}
	return SRLatchResWithEnable(parent, q, s, r, e)
}

func SRLatchResWithEnable(parent *group.Group, q, s, r, e *wire.Wire) []*wire.Wire {
	group := parent.Group(fmt.Sprintf("SRLATCHEN(%v,%v,%v)", s.Name, r.Name, e.Name))
	return SRLatchRes(group, q, And(group, s, e), And(group, r, e))
}

func DLatch(parent *group.Group, d, e *wire.Wire) []*wire.Wire {
	q := &wire.Wire{Name: "q"}
	return DLatchRes(parent, q, d, e)
}

func DLatchRes(parent *group.Group, q, d, e *wire.Wire) []*wire.Wire {
	group := parent.Group(fmt.Sprintf("DLATCH(%v,%v)", d.Name, e.Name))
	return SRLatchResWithEnable(group, q, d, Not(group, d), e)
}

func Register(parent *group.Group, d, ei, eo *wire.Wire) []*wire.Wire {
	group := parent.Group(fmt.Sprintf("Register(%v,%v,%v)", d.Name, ei.Name, eo.Name))
	q := &wire.Wire{}
	DLatchRes(group, q, Or(group, And(group, q, Not(group, ei)), And(group, d, ei)), ei)
	q.Name = group.Name + "-internal"
	res := And(group, q, eo)
	res.Name = group.Name
	return []*wire.Wire{q, res}
}

func Register2(parent *group.Group, d1, d2, ei, eo *wire.Wire) []*wire.Wire {
	group := parent.Group("Register2")
	r1 := Register(group, d1, ei, eo)
	r2 := Register(group, d2, ei, eo)
	return append(r1, r2...)
}

func Register4(parent *group.Group, d1, d2, d3, d4, ei, eo *wire.Wire) []*wire.Wire {
	group := parent.Group("Register4")
	r1 := Register2(group, d1, d2, ei, eo)
	r2 := Register2(group, d3, d4, ei, eo)
	return append(r1, r2...)
}

func Register8(parent *group.Group, d1, d2, d3, d4, d5, d6, d7, d8, ei, eo *wire.Wire) []*wire.Wire {
	group := parent.Group("Register8")
	r1 := Register4(group, d1, d2, d3, d4, ei, eo)
	r2 := Register4(group, d5, d6, d7, d8, ei, eo)
	return append(r1, r2...)
}

func Alu(parent *group.Group, a, ai, ao, b, bi, bo, ri, ro, cin *wire.Wire) []*wire.Wire {
	group := parent.Group("ALU")
	ra := Register(group, a, ai, ao)
	rb := Register(group, b, bi, bo)
	rs := Sum(group, ra[0], rb[0], cin)
	rr := Register(group, rs[0], ri, ro)
	return slices.Concat(ra, rb, rr, []*wire.Wire{rs[1]})
}

func Alu2(parent *group.Group, a1, a2, ai, ao, b1, b2, bi, bo, ri, ro, cin *wire.Wire) []*wire.Wire {
	group := parent.Group("ALU2")
	r1 := Alu(group, a1, ai, ao, b1, bi, bo, ri, ro, cin)
	last := len(r1) - 1
	r2 := Alu(group, a2, ai, ao, b2, bi, bo, ri, ro, r1[last])
	return append(r1[:last], r2...)
}

func Alu4(parent *group.Group, a [4]*wire.Wire, ai, ao *wire.Wire, b [4]*wire.Wire, bi, bo, ri, ro, cin *wire.Wire) []*wire.Wire {
	group := parent.Group("ALU4")
	r1 := Alu2(group, a[0], a[1], ai, ao, b[0], b[1], bi, bo, ri, ro, cin)
	last := len(r1) - 1
	r2 := Alu2(group, a[2], a[3], ai, ao, b[2], b[3], bi, bo, ri, ro, r1[last])
	return append(r1[:last], r2...)
}

func Alu8(parent *group.Group, a [8]*wire.Wire, ai, ao *wire.Wire, b [8]*wire.Wire, bi, bo, ri, ro, cin *wire.Wire) []*wire.Wire {
	group := parent.Group("ALU8")
	r1 := Alu4(group, [4]*wire.Wire(a[:4]), ai, ao, [4]*wire.Wire(b[:4]), bi, bo, ri, ro, cin)
	last := len(r1) - 1
	r2 := Alu4(group, [4]*wire.Wire(a[4:]), ai, ao, [4]*wire.Wire(b[4:]), bi, bo, ri, ro, r1[last])
	return append(r1[:last], r2...)
}

func Bus(parent *group.Group, bus, a, b, r, wa, wb *wire.Wire) []*wire.Wire {
	group := parent.Group(fmt.Sprintf("BUS"))
	res := &wire.Wire{Name: group.Name}
	wire1 := &wire.Wire{Name: fmt.Sprintf("%s-wire1", res.Name)}
	wire2 := &wire.Wire{Name: fmt.Sprintf("%s-wire2", res.Name)}
	group.JointWire(wire1, bus, a, false /* isAnd */)
	group.JointWire(wire2, wire1, b, false /* isAnd */)
	group.JointWire(res, wire2, r, false /* isAnd */)
	group.JointWire(wa, res, res, false /* isAnd */)
	group.JointWire(wb, res, res, false /* isAnd */)
	return []*wire.Wire{res}
}

func AluWithBus(parent *group.Group, bus, ai, ao, bi, bo, ri, ro, cin *wire.Wire) []*wire.Wire {
	group := parent.Group("ALU-BUS")
	a := &wire.Wire{Name: fmt.Sprintf("ALU-%s-a", bus.Name)}
	ra := Register(group, a, ai, ao)
	b := &wire.Wire{Name: fmt.Sprintf("ALU-%s-b", bus.Name)}
	rb := Register(group, b, bi, bo)
	rs := Sum(group, ra[0], rb[0], cin)
	rr := Register(group, rs[0], ri, ro)
	rbus := Bus(group, bus, ra[1], rb[1], rr[1], a, b)
	return slices.Concat(rbus, ra, rb, rr, []*wire.Wire{rs[1]})
}

func Example(c *circuit.Circuit, name string) []*wire.Wire {
	res, ok := examples[name]
	if !ok {
		return nil
	}
	return res(c)
}

func ExampleNames() []string {
	var res []string
	for name := range examples {
		res = append(res, name)
	}
	return res
}

var (
	examples = map[string]func(*circuit.Circuit) []*wire.Wire{
		"TransistorEmitter": func(c *circuit.Circuit) []*wire.Wire {
			return TransistorEmitter(c.Group(""), c.In("base"), c.In("collector"))
		},
		"TransistorGnd": func(c *circuit.Circuit) []*wire.Wire {
			return TransistorGnd(c.Group(""), c.In("base"), c.In("collector"))
		},
		"Transistor": func(c *circuit.Circuit) []*wire.Wire {
			return Transistor(c.Group(""), c.In("base"), c.In("collector"))
		},
		"Not": func(c *circuit.Circuit) []*wire.Wire {
			return []*wire.Wire{Not(c.Group(""), c.In("a"))}
		},
		"And": func(c *circuit.Circuit) []*wire.Wire {
			return []*wire.Wire{And(c.Group(""), c.In("a"), c.In("b"))}
		},
		"Or": func(c *circuit.Circuit) []*wire.Wire {
			return []*wire.Wire{Or(c.Group(""), c.In("a"), c.In("b"))}
		},
		"OrRes": func(c *circuit.Circuit) []*wire.Wire {
			bOrRes := &wire.Wire{Name: "b"}
			return []*wire.Wire{OrRes(c.Group(""), bOrRes, c.In("a"), bOrRes)}
		},
		"Nand": func(c *circuit.Circuit) []*wire.Wire {
			return []*wire.Wire{Nand(c.Group(""), c.In("a"), c.In("b"))}
		},
		"Nand(Nand)": func(c *circuit.Circuit) []*wire.Wire {
			g := c.Group("")
			return []*wire.Wire{Nand(g, c.In("a"), Nand(g, c.In("b"), c.In("c")))}
		},
		"Xor": func(c *circuit.Circuit) []*wire.Wire {
			return []*wire.Wire{Xor(c.Group(""), c.In("a"), c.In("b"))}
		},
		"Nor": func(c *circuit.Circuit) []*wire.Wire {
			return []*wire.Wire{Nor(c.Group(""), c.In("a"), c.In("b"))}
		},
		"HalfSum": func(c *circuit.Circuit) []*wire.Wire {
			return HalfSum(c.Group(""), c.In("a"), c.In("b"))
		},
		"Sum": func(c *circuit.Circuit) []*wire.Wire {
			return Sum(c.Group(""), c.In("a"), c.In("b"), c.In("c"))
		},
		"Sum2": func(c *circuit.Circuit) []*wire.Wire {
			return Sum2(c.Group(""), c.In("a1"), c.In("a2"), c.In("b1"), c.In("b2"), c.In("c"))
		},
		"Sum4": func(c *circuit.Circuit) []*wire.Wire {
			return Sum4(
				c.Group(""),
				c.In("a1"), c.In("a2"), c.In("a3"), c.In("a4"),
				c.In("b1"), c.In("b2"), c.In("b3"), c.In("b4"),
				c.In("c"))
		},
		"Sum8": func(c *circuit.Circuit) []*wire.Wire {
			return Sum8(
				c.Group(""),
				[8]*wire.Wire{
					c.In("a1"), c.In("a2"), c.In("a3"), c.In("a4"), c.In("a5"), c.In("a6"), c.In("a7"), c.In("a8"),
				},
				[8]*wire.Wire{
					c.In("b1"), c.In("b2"), c.In("b3"), c.In("b4"), c.In("b5"), c.In("b6"), c.In("b7"), c.In("b8"),
				},
				c.In("c"))
		},
		"SRLatch": func(c *circuit.Circuit) []*wire.Wire {
			return SRLatch(c.Group(""), c.In("s"), c.In("r"))
		},
		"SRLatchWithEnable": func(c *circuit.Circuit) []*wire.Wire {
			return SRLatchWithEnable(c.Group(""), c.In("s"), c.In("r"), c.In("e"))
		},
		"DLatch": func(c *circuit.Circuit) []*wire.Wire {
			return DLatch(c.Group(""), c.In("d"), c.In("e"))
		},
		"Register": func(c *circuit.Circuit) []*wire.Wire {
			return Register(c.Group(""), c.In("d"), c.In("ei"), c.In("eo"))
		},
		"Register2": func(c *circuit.Circuit) []*wire.Wire {
			return Register2(c.Group(""), c.In("d1"), c.In("d2"), c.In("ei"), c.In("eo"))
		},
		"Register4": func(c *circuit.Circuit) []*wire.Wire {
			return Register4(c.Group(""), c.In("d1"), c.In("d2"), c.In("d3"), c.In("d4"), c.In("ei"), c.In("eo"))
		},
		"Register8": func(c *circuit.Circuit) []*wire.Wire {
			return Register8(
				c.Group(""),
				c.In("d1"), c.In("d2"), c.In("d3"), c.In("d4"),
				c.In("d5"), c.In("d6"), c.In("d7"), c.In("d8"),
				c.In("ei"), c.In("eo"))
		},
		"Alu": func(c *circuit.Circuit) []*wire.Wire {
			return Alu(
				c.Group(""),
				c.In("a"), c.In("ai"), c.In("ao"),
				c.In("b"), c.In("bi"), c.In("bo"),
				c.In("ri"), c.In("ro"), c.In("cin"))
		},
		"Alu2": func(c *circuit.Circuit) []*wire.Wire {
			return Alu2(
				c.Group(""),
				c.In("a1"), c.In("a2"), c.In("ai"), c.In("ao"),
				c.In("b1"), c.In("b2"), c.In("bi"), c.In("bo"),
				c.In("ri"), c.In("ro"), c.In("cin"))
		},
		"Alu4": func(c *circuit.Circuit) []*wire.Wire {
			return Alu4(
				c.Group(""),
				[4]*wire.Wire{
					c.In("a1"), c.In("a2"), c.In("a3"), c.In("a4"),
				}, c.In("ai"), c.In("ao"),
				[4]*wire.Wire{
					c.In("b1"), c.In("b2"), c.In("b3"), c.In("b4"),
				}, c.In("bi"), c.In("bo"),
				c.In("ri"), c.In("ro"), c.In("cin"))
		},
		"Alu8": func(c *circuit.Circuit) []*wire.Wire {
			return Alu8(
				c.Group(""),
				[8]*wire.Wire{
					c.In("a1"), c.In("a2"), c.In("a3"), c.In("a4"), c.In("a5"), c.In("a6"), c.In("a7"), c.In("a8"),
				}, c.In("ai"), c.In("ao"),
				[8]*wire.Wire{
					c.In("b1"), c.In("b2"), c.In("b3"), c.In("b4"), c.In("b5"), c.In("b6"), c.In("b7"), c.In("b8"),
				}, c.In("bi"), c.In("bo"),
				c.In("ri"), c.In("ro"), c.In("cin"))
		},
		"Bus": func(c *circuit.Circuit) []*wire.Wire {
			wa := &wire.Wire{Name: "wa"}
			wb := &wire.Wire{Name: "wb"}
			return append(Bus(c.Group(""), c.In("bus"), c.In("a"), c.In("b"), c.In("r"), wa, wb), wa, wb)
		},
		"AluWithBus": func(c *circuit.Circuit) []*wire.Wire {
			bus := c.In("bus")
			ai, ao := c.In("ai"), c.In("ao")
			bi, bo := c.In("bi"), c.In("bo")
			ri, ro := c.In("ri"), c.In("ro")
			cin := c.In("cin")
			c.AddInputValidation(func() bool {
				res := !(ri.Bit.Get(nil) && ro.Bit.Get(nil) && (ai.Bit.Get(nil) || bi.Bit.Get(nil)))
				return res
			})
			return AluWithBus(c.Group(""), bus, ai, ao, bi, bo, ri, ro, cin)
		},
		"": func(c *circuit.Circuit) []*wire.Wire {
			return nil
		},
	}
)
