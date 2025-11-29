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
	group := parent.Group(fmt.Sprintf("NOT(%s)", a.Name))
	res := &wire.Wire{Name: group.Name}
	group.Transistor(a, group.Vcc, group.Gnd, res)
	return res
}

func And(parent *group.Group, a, b *wire.Wire) *wire.Wire {
	group := parent.Group(fmt.Sprintf("AND(%s,%s)", a.Name, b.Name))
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
	group := parent.Group(fmt.Sprintf("OR(%s,%s)", a.Name, b.Name))
	res.Name = group.Name
	wire1 := &wire.Wire{Name: fmt.Sprintf("%s-wire1", res.Name)}
	wire2 := &wire.Wire{Name: fmt.Sprintf("%s-wire2", res.Name)}
	group.AddTransistors([]*transistor.Transistor{
		{Base: a, Collector: group.Vcc, Emitter: wire1},
		{Base: b, Collector: group.Vcc, Emitter: wire2},
	})
	group.JointWire(res, wire1, wire2)
	return res
}

func Nand(parent *group.Group, a, b *wire.Wire) *wire.Wire {
	group := parent.Group(fmt.Sprintf("NAND(%s,%s)", a.Name, b.Name))
	res := &wire.Wire{Name: group.Name}
	wire := &wire.Wire{Name: fmt.Sprintf("%s-wire", res.Name)}
	group.AddTransistors([]*transistor.Transistor{
		{Base: a, Collector: group.Vcc, Emitter: wire, CollectorOut: res},
		{Base: b, Collector: wire, Emitter: group.Gnd},
	})
	return res
}

func Xor(parent *group.Group, a, b *wire.Wire) *wire.Wire {
	group := parent.Group(fmt.Sprintf("XOR(%s,%s)", a.Name, b.Name))
	res := And(group, Or(group, a, b), Nand(group, a, b))
	res.Name = group.Name
	return res
}

func Nor(parent *group.Group, a, b *wire.Wire) *wire.Wire {
	res := &wire.Wire{}
	return NorRes(parent, res, a, b)
}

func NorRes(parent *group.Group, res, a, b *wire.Wire) *wire.Wire {
	group := parent.Group(fmt.Sprintf("NOR(%s,%s)", a.Name, b.Name))
	res.Name = group.Name
	wire1 := &wire.Wire{Name: fmt.Sprintf("%s-wire1", res.Name)}
	wire2 := &wire.Wire{Name: fmt.Sprintf("%s-wire2", res.Name)}
	group.AddTransistors([]*transistor.Transistor{
		{Base: a, Collector: group.Vcc, Emitter: group.Gnd, CollectorOut: wire1},
		{Base: b, Collector: group.Vcc, Emitter: group.Gnd, CollectorOut: wire2},
	})
	group.JointWireIsAnd(res, wire1, wire2)
	return res
}

func HalfSum(parent *group.Group, a, b *wire.Wire) []*wire.Wire {
	group := parent.Group(fmt.Sprintf("SUM(%s,%s)", a.Name, b.Name))
	res := Xor(group, a, b)
	res.Name = group.Name
	carry := And(group, a, b)
	carry.Name = fmt.Sprintf("CARRY(%s,%s)", a.Name, b.Name)
	return []*wire.Wire{res, carry}
}

func Sum(parent *group.Group, a, b, cin *wire.Wire) []*wire.Wire {
	group := parent.Group(fmt.Sprintf("SUM(%s,%s,%s)", a.Name, b.Name, cin.Name))
	s1 := HalfSum(group, a, b)
	s2 := HalfSum(group, s1[0], cin)
	s2[0].Name = group.Name
	carry := Or(group, s1[1], s2[1])
	carry.Name = fmt.Sprintf("CARRY(%s,%s)", a.Name, b.Name)
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
	group := parent.Group(fmt.Sprintf("SRLATCH(%s,%s)", s.Name, r.Name))
	nq := &wire.Wire{Name: "nq"}
	name := q.Name
	NorRes(group, q, r, nq)
	NorRes(group, nq, s, q)
	q.Name = name
	nq.Name = "nq"
	return []*wire.Wire{q, nq}
}

func SRLatchWithEnable(parent *group.Group, s, r, e *wire.Wire) []*wire.Wire {
	q := &wire.Wire{Name: "q"}
	return SRLatchResWithEnable(parent, q, s, r, e)
}

func SRLatchResWithEnable(parent *group.Group, q, s, r, e *wire.Wire) []*wire.Wire {
	group := parent.Group(fmt.Sprintf("SRLATCHEN(%s,%s,%s)", s.Name, r.Name, e.Name))
	return SRLatchRes(group, q, And(group, s, e), And(group, r, e))
}

func DLatch(parent *group.Group, d, e *wire.Wire) []*wire.Wire {
	q := &wire.Wire{Name: "q"}
	return DLatchRes(parent, q, d, e)
}

func DLatchRes(parent *group.Group, q, d, e *wire.Wire) []*wire.Wire {
	group := parent.Group(fmt.Sprintf("DLATCH(%s,%s)", d.Name, e.Name))
	return SRLatchResWithEnable(group, q, d, Not(group, d), e)
}

func Register(parent *group.Group, d, ei, eo *wire.Wire) []*wire.Wire {
	group := parent.Group(fmt.Sprintf("REG(%s,%s,%s)", d.Name, ei.Name, eo.Name))
	q := &wire.Wire{}
	DLatchRes(group, q, Or(group, And(group, q, Not(group, ei)), And(group, d, ei)), ei)
	q.Name = "reg" + group.Name[3:]
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

func w4(w []*wire.Wire) [4]*wire.Wire {
	return [4]*wire.Wire(w)
}

func Alu8(parent *group.Group, a [8]*wire.Wire, ai, ao *wire.Wire, b [8]*wire.Wire, bi, bo, ri, ro, cin *wire.Wire) []*wire.Wire {
	group := parent.Group("ALU8")
	r1 := Alu4(group, w4(a[:4]), ai, ao, w4(b[:4]), bi, bo, ri, ro, cin)
	last := len(r1) - 1
	r2 := Alu4(group, w4(a[4:]), ai, ao, w4(b[4:]), bi, bo, ri, ro, r1[last])
	return append(r1[:last], r2...)
}

func Bus(parent *group.Group, bus, a, b, r, wa, wb *wire.Wire) []*wire.Wire {
	group := parent.Group(fmt.Sprintf("BUS(%s)", bus.Name))
	res := &wire.Wire{Name: group.Name}
	wire1 := &wire.Wire{Name: fmt.Sprintf("%s-wire1", res.Name)}
	wire2 := &wire.Wire{Name: fmt.Sprintf("%s-wire2", res.Name)}
	group.JointWire(wire1, bus, a)
	group.JointWire(wire2, wire1, b)
	group.JointWire(res, wire2, r)
	group.JointWire(wa, res, res)
	group.JointWire(wb, res, res)
	return []*wire.Wire{res}
}

func Bus2(parent *group.Group, bus1, bus2, a1, a2, b1, b2, r1, r2, wa1, wa2, wb1, wb2 *wire.Wire) []*wire.Wire {
	group := parent.Group(fmt.Sprintf("BUS2"))
	rbus1 := Bus(group, bus1, a1, b1, r1, wa1, wb1)
	rbus2 := Bus(group, bus2, a2, b2, r2, wa2, wb2)
	return slices.Concat(rbus1, rbus2)
}

func Bus4(parent *group.Group, bus, a, b, r, wa, wb [4]*wire.Wire) []*wire.Wire {
	group := parent.Group(fmt.Sprintf("BUS2"))
	rbus1 := Bus2(group, bus[0], bus[1], a[0], a[1], b[0], b[1], r[0], r[1], wa[0], wa[1], wb[0], wb[1])
	rbus2 := Bus2(group, bus[2], bus[3], a[2], a[3], b[2], b[3], r[2], r[3], wa[2], wa[3], wb[2], wb[3])
	return slices.Concat(rbus1, rbus2)
}

func Bus8(parent *group.Group, bus, a, b, r, wa, wb [8]*wire.Wire) []*wire.Wire {
	group := parent.Group(fmt.Sprintf("BUS2"))
	rbus1 := Bus4(group, w4(bus[:4]), w4(a[:4]), w4(b[:4]), w4(r[:4]), w4(wa[:4]), w4(wb[:4]))
	rbus2 := Bus4(group, w4(bus[4:]), w4(a[4:]), w4(b[4:]), w4(r[4:]), w4(wa[4:]), w4(wb[4:]))
	return slices.Concat(rbus1, rbus2)
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

func aluWithBusInputValidation(ai, bi, ri, ro *wire.Wire) func() bool {
	return func() bool {
		return !(ri.Bit.Get(nil) && ro.Bit.Get(nil) && (ai.Bit.Get(nil) || bi.Bit.Get(nil)))
	}
}

func AluWithBus2(parent *group.Group, bus1, bus2, ai, ao, bi, bo, ri, ro, cin *wire.Wire) []*wire.Wire {
	group := parent.Group("ALU-BUS2")
	alu1 := AluWithBus(group, bus1, ai, ao, bi, bo, ri, ro, cin)
	last := len(alu1) - 1
	alu2 := AluWithBus(group, bus2, ai, ao, bi, bo, ri, ro, alu1[last])
	return slices.Concat(alu1[:last], alu2)
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
			bOrRes := &wire.Wire{Name: "bOrRes"}
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
			return Sum(c.Group(""), c.In("a"), c.In("b"), c.In("cin"))
		},
		"Sum2": func(c *circuit.Circuit) []*wire.Wire {
			return Sum2(c.Group(""), c.In("a1"), c.In("a2"), c.In("b1"), c.In("b2"), c.In("cin"))
		},
		"Sum4": func(c *circuit.Circuit) []*wire.Wire {
			return Sum4(
				c.Group(""),
				c.In("a1"), c.In("a2"), c.In("a3"), c.In("a4"),
				c.In("b1"), c.In("b2"), c.In("b3"), c.In("b4"),
				c.In("cin"))
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
				c.In("cin"))
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
		"Bus2": func(c *circuit.Circuit) []*wire.Wire {
			wa1, wa2 := &wire.Wire{Name: "wa1"}, &wire.Wire{Name: "wa2"}
			wb1, wb2 := &wire.Wire{Name: "wb1"}, &wire.Wire{Name: "wb2"}
			return append(Bus2(
				c.Group(""), c.In("bus1"), c.In("bus2"),
				c.In("a1"), c.In("a2"), c.In("b1"), c.In("b2"), c.In("r1"), c.In("r2"),
				wa1, wa2, wb1, wb2),
				wa1, wa2, wb1, wb2)
		},
		"Bus4": func(c *circuit.Circuit) []*wire.Wire {
			wa1, wa2 := &wire.Wire{Name: "wa1"}, &wire.Wire{Name: "wa2"}
			wa3, wa4 := &wire.Wire{Name: "wa3"}, &wire.Wire{Name: "wa4"}
			wb1, wb2 := &wire.Wire{Name: "wb1"}, &wire.Wire{Name: "wb2"}
			wb3, wb4 := &wire.Wire{Name: "wb3"}, &wire.Wire{Name: "wb4"}
			wa := [4]*wire.Wire{wa1, wa2, wa3, wa4}
			wb := [4]*wire.Wire{wb1, wb2, wb3, wb4}
			return slices.Concat(Bus4(
				c.Group(""), [4]*wire.Wire{c.In("bus1"), c.In("bus2"), c.In("bus3"), c.In("bus4")},
				[4]*wire.Wire{c.In("a1"), c.In("a2"), c.In("a3"), c.In("a4")},
				[4]*wire.Wire{c.In("b1"), c.In("b2"), c.In("b3"), c.In("b4")},
				[4]*wire.Wire{c.In("r1"), c.In("r2"), c.In("r3"), c.In("r4")},
				wa, wb), wa[:], wb[:])
		},
		"Bus8": func(c *circuit.Circuit) []*wire.Wire {
			wa1, wa2 := &wire.Wire{Name: "wa1"}, &wire.Wire{Name: "wa2"}
			wa3, wa4 := &wire.Wire{Name: "wa3"}, &wire.Wire{Name: "wa4"}
			wa5, wa6 := &wire.Wire{Name: "wa5"}, &wire.Wire{Name: "wa6"}
			wa7, wa8 := &wire.Wire{Name: "wa7"}, &wire.Wire{Name: "wa8"}
			wb1, wb2 := &wire.Wire{Name: "wb1"}, &wire.Wire{Name: "wb2"}
			wb3, wb4 := &wire.Wire{Name: "wb3"}, &wire.Wire{Name: "wb4"}
			wb5, wb6 := &wire.Wire{Name: "wb5"}, &wire.Wire{Name: "wb6"}
			wb7, wb8 := &wire.Wire{Name: "wb7"}, &wire.Wire{Name: "wb8"}
			wa := [8]*wire.Wire{wa1, wa2, wa3, wa4, wa5, wa6, wa7, wa8}
			wb := [8]*wire.Wire{wb1, wb2, wb3, wb4, wb5, wb6, wb7, wb8}
			return slices.Concat(Bus8(
				c.Group(""),
				[8]*wire.Wire{c.In("bus1"), c.In("bus2"), c.In("bus3"), c.In("bus4"),
					c.In("bus5"), c.In("bus6"), c.In("bus7"), c.In("bus8")},
				[8]*wire.Wire{c.In("a1"), c.In("a2"), c.In("a3"), c.In("a4"),
					c.In("a5"), c.In("a6"), c.In("a7"), c.In("a8")},
				[8]*wire.Wire{c.In("b1"), c.In("b2"), c.In("b3"), c.In("b4"),
					c.In("b5"), c.In("b6"), c.In("b7"), c.In("b8")},
				[8]*wire.Wire{c.In("r1"), c.In("r2"), c.In("r3"), c.In("r4"),
					c.In("r5"), c.In("r6"), c.In("r7"), c.In("r8")},
				wa, wb), wa[:], wb[:])
		},
		"AluWithBus": func(c *circuit.Circuit) []*wire.Wire {
			bus := c.In("bus")
			ai, ao := c.In("ai"), c.In("ao")
			bi, bo := c.In("bi"), c.In("bo")
			ri, ro := c.In("ri"), c.In("ro")
			cin := c.In("cin")
			c.AddInputValidation(aluWithBusInputValidation(ai, bi, ri, ro))
			return AluWithBus(c.Group(""), bus, ai, ao, bi, bo, ri, ro, cin)
		},
		"AluWithBus2": func(c *circuit.Circuit) []*wire.Wire {
			bus1, bus2 := c.In("bus1"), c.In("bus2")
			ai, ao := c.In("ai"), c.In("ao")
			bi, bo := c.In("bi"), c.In("bo")
			ri, ro := c.In("ri"), c.In("ro")
			cin := c.In("cin")
			c.AddInputValidation(aluWithBusInputValidation(ai, bi, ri, ro))
			return AluWithBus2(c.Group(""), bus1, bus2, ai, ao, bi, bo, ri, ro, cin)
		},
		"": func(c *circuit.Circuit) []*wire.Wire {
			return nil
		},
	}
)
