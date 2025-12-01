// Package lib contains a library of circuits.
package lib

import (
	"slices"

	"github.com/kssilveira/circuit-engine/circuit"
	"github.com/kssilveira/circuit-engine/group"
	"github.com/kssilveira/circuit-engine/lib/gate"
	"github.com/kssilveira/circuit-engine/lib/latch"
	"github.com/kssilveira/circuit-engine/lib/sum"
	"github.com/kssilveira/circuit-engine/sfmt"
	"github.com/kssilveira/circuit-engine/wire"
)

// Register adds a register.
func Register(parent *group.Group, d, ei, eo *wire.Wire) []*wire.Wire {
	group := parent.Group(sfmt.Sprintf("REG(%s,%s,%s)", d.Name, ei.Name, eo.Name))
	q := &wire.Wire{}
	latch.DLatchRes(group, q, gate.Or(group, gate.And(group, q, gate.Not(group, ei)), gate.And(group, d, ei)), ei)
	q.Name = "reg" + group.Name[3:]
	res := gate.And(group, q, eo)
	res.Name = group.Name
	return []*wire.Wire{q, res}
}

// Register2 adds a 2-bit register.
func Register2(parent *group.Group, d1, d2, ei, eo *wire.Wire) []*wire.Wire {
	group := parent.Group("Register2")
	r1 := Register(group, d1, ei, eo)
	r2 := Register(group, d2, ei, eo)
	return append(r1, r2...)
}

// Register4 adds a 4-bit register.
func Register4(parent *group.Group, d1, d2, d3, d4, ei, eo *wire.Wire) []*wire.Wire {
	group := parent.Group("Register4")
	r1 := Register2(group, d1, d2, ei, eo)
	r2 := Register2(group, d3, d4, ei, eo)
	return append(r1, r2...)
}

// Register8 adds an 8-bit register.
func Register8(parent *group.Group, d1, d2, d3, d4, d5, d6, d7, d8, ei, eo *wire.Wire) []*wire.Wire {
	group := parent.Group("Register8")
	r1 := Register4(group, d1, d2, d3, d4, ei, eo)
	r2 := Register4(group, d5, d6, d7, d8, ei, eo)
	return append(r1, r2...)
}

// Alu adds an artithmetic and logic unit.
func Alu(parent *group.Group, a, ai, ao, b, bi, bo, ri, ro, cin *wire.Wire) []*wire.Wire {
	group := parent.Group("ALU")
	ra := Register(group, a, ai, ao)
	rb := Register(group, b, bi, bo)
	rs := sum.Adder(group, ra[0], rb[0], cin)
	rr := Register(group, rs[0], ri, ro)
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

func w4(w []*wire.Wire) [4]*wire.Wire {
	return [4]*wire.Wire(w)
}

// Alu8 adds an 8-bit arithmetic and logic unit.
func Alu8(parent *group.Group, a [8]*wire.Wire, ai, ao *wire.Wire, b [8]*wire.Wire, bi, bo, ri, ro, cin *wire.Wire) []*wire.Wire {
	group := parent.Group("ALU8")
	r1 := Alu4(group, w4(a[:4]), ai, ao, w4(b[:4]), bi, bo, ri, ro, cin)
	last := len(r1) - 1
	r2 := Alu4(group, w4(a[4:]), ai, ao, w4(b[4:]), bi, bo, ri, ro, r1[last])
	return append(r1[:last], r2...)
}

// Bus add a communication bus.
func Bus(parent *group.Group, bus, a, b, r, wa, wb *wire.Wire) []*wire.Wire {
	group := parent.Group(sfmt.Sprintf("BUS(%s)", bus.Name))
	res := &wire.Wire{Name: group.Name}
	wire1 := &wire.Wire{Name: sfmt.Sprintf("%s-wire1", res.Name)}
	wire2 := &wire.Wire{Name: sfmt.Sprintf("%s-wire2", res.Name)}
	group.JointWire(wire1, bus, a)
	group.JointWire(wire2, wire1, b)
	group.JointWire(res, wire2, r)
	group.JointWire(wa, res, res)
	group.JointWire(wb, res, res)
	return []*wire.Wire{res}
}

// Bus2 adds a 2-bit communication bus.
func Bus2(parent *group.Group, bus1, bus2, a1, a2, b1, b2, r1, r2, wa1, wa2, wb1, wb2 *wire.Wire) []*wire.Wire {
	group := parent.Group(sfmt.Sprintf("BUS2"))
	rbus1 := Bus(group, bus1, a1, b1, r1, wa1, wb1)
	rbus2 := Bus(group, bus2, a2, b2, r2, wa2, wb2)
	return slices.Concat(rbus1, rbus2)
}

// Bus4 adds a 4-bit communication bus.
func Bus4(parent *group.Group, bus, a, b, r, wa, wb [4]*wire.Wire) []*wire.Wire {
	group := parent.Group(sfmt.Sprintf("BUS2"))
	rbus1 := Bus2(group, bus[0], bus[1], a[0], a[1], b[0], b[1], r[0], r[1], wa[0], wa[1], wb[0], wb[1])
	rbus2 := Bus2(group, bus[2], bus[3], a[2], a[3], b[2], b[3], r[2], r[3], wa[2], wa[3], wb[2], wb[3])
	return slices.Concat(rbus1, rbus2)
}

// Bus8 adds an 8-bit communication bus.
func Bus8(parent *group.Group, bus, a, b, r, wa, wb [8]*wire.Wire) []*wire.Wire {
	group := parent.Group(sfmt.Sprintf("BUS2"))
	rbus1 := Bus4(group, w4(bus[:4]), w4(a[:4]), w4(b[:4]), w4(r[:4]), w4(wa[:4]), w4(wb[:4]))
	rbus2 := Bus4(group, w4(bus[4:]), w4(a[4:]), w4(b[4:]), w4(r[4:]), w4(wa[4:]), w4(wb[4:]))
	return slices.Concat(rbus1, rbus2)
}

// AluWithBus adds an arithmetic logic unit with a communication bus.
func AluWithBus(parent *group.Group, bus, ai, ao, bi, bo, ri, ro, cin *wire.Wire) []*wire.Wire {
	group := parent.Group("ALU-BUS")
	a := &wire.Wire{Name: sfmt.Sprintf("ALU-%s-a", bus.Name)}
	ra := Register(group, a, ai, ao)
	b := &wire.Wire{Name: sfmt.Sprintf("ALU-%s-b", bus.Name)}
	rb := Register(group, b, bi, bo)
	rs := sum.Adder(group, ra[0], rb[0], cin)
	rr := Register(group, rs[0], ri, ro)
	rbus := Bus(group, bus, ra[1], rb[1], rr[1], a, b)
	return slices.Concat(rbus, ra, rb, rr, []*wire.Wire{rs[1]})
}

func aluWithBusInputValidation(ai, ao, bi, bo, ri, ro *wire.Wire) func() bool {
	return func() bool {
		return !(ri.Bit.Get(nil) && ro.Bit.Get(nil) && (ai.Bit.Get(nil) || bi.Bit.Get(nil))) &&
			!(ai.Bit.Get(nil) && ao.Bit.Get(nil)) &&
			!(bi.Bit.Get(nil) && bo.Bit.Get(nil))
	}
}

// AluWithBus2 adds a 2-bit arithmetic logic unit with a communication bus.
func AluWithBus2(parent *group.Group, bus1, bus2, ai, ao, bi, bo, ri, ro, cin *wire.Wire) []*wire.Wire {
	group := parent.Group("ALU-BUS2")
	alu1 := AluWithBus(group, bus1, ai, ao, bi, bo, ri, ro, cin)
	last := len(alu1) - 1
	alu2 := AluWithBus(group, bus2, ai, ao, bi, bo, ri, ro, alu1[last])
	return slices.Concat(alu1[:last], alu2)
}

// AluWithBus4 adds a 4-bit arithmetic logic unit with a communication bus.
func AluWithBus4(parent *group.Group, bus [4]*wire.Wire, ai, ao, bi, bo, ri, ro, cin *wire.Wire) []*wire.Wire {
	group := parent.Group("ALU-BUS4")
	alu1 := AluWithBus2(group, bus[0], bus[1], ai, ao, bi, bo, ri, ro, cin)
	last := len(alu1) - 1
	alu2 := AluWithBus2(group, bus[2], bus[3], ai, ao, bi, bo, ri, ro, alu1[last])
	return slices.Concat(alu1[:last], alu2)
}

// AluWithBus8 adds an 8-bit arithmetic logic unit with a communication bus.
func AluWithBus8(parent *group.Group, bus [8]*wire.Wire, ai, ao, bi, bo, ri, ro, cin *wire.Wire) []*wire.Wire {
	group := parent.Group("ALU-BUS8")
	alu1 := AluWithBus4(group, w4(bus[:4]), ai, ao, bi, bo, ri, ro, cin)
	last := len(alu1) - 1
	alu2 := AluWithBus4(group, w4(bus[4:]), ai, ao, bi, bo, ri, ro, alu1[last])
	return slices.Concat(alu1[:last], alu2)
}

// RAM adds a random access memory.
func RAM(parent *group.Group, a, d, ei, eo *wire.Wire) []*wire.Wire {
	group := parent.Group(sfmt.Sprintf("RAM(%v,%v)", a.Name, d.Name))
	s := ramAddress(group, a)
	rei, reo := ramEnable(group, s, ei, eo)
	return ramRegisters(group, d, s, rei, reo)
}

func ramAddress(group *group.Group, a *wire.Wire) []*wire.Wire {
	s1 := gate.Not(group, a)
	s1.Name = group.Name + "-s1"
	s2 := gate.Or(group, a, a)
	s2.Name = group.Name + "-s2"
	return []*wire.Wire{s1, s2}
}

func ramEnable(group *group.Group, s []*wire.Wire, ei, eo *wire.Wire) ([]*wire.Wire, []*wire.Wire) {
	ei1 := gate.And(group, ei, s[0])
	ei1.Name = group.Name + "-ei1"
	ei2 := gate.And(group, ei, s[1])
	ei2.Name = group.Name + "-ei2"

	eo1 := gate.And(group, eo, s[0])
	eo1.Name = group.Name + "-eo1"
	eo2 := gate.And(group, eo, s[1])
	eo2.Name = group.Name + "-eo2"

	return []*wire.Wire{ei1, ei2}, []*wire.Wire{eo1, eo2}
}

func ramRegisters(group *group.Group, d *wire.Wire, s, ei, eo []*wire.Wire) []*wire.Wire {
	r1 := Register(group, d, ei[0], eo[0])
	r2 := Register(group, d, ei[1], eo[1])
	res := &wire.Wire{Name: group.Name}
	group.JointWire(res, r1[1], r2[1])
	return slices.Concat([]*wire.Wire{res, s[0], ei[0], eo[0]}, r1, []*wire.Wire{s[1], ei[1], eo[1]}, r2)
}

// Example returns the example with the given name.
func Example(c *circuit.Circuit, name string) []*wire.Wire {
	res, ok := examples[name]
	if !ok {
		return nil
	}
	return res(c)
}

// ExampleNames returns the available example names.
func ExampleNames() []string {
	var res []string
	for name := range examples {
		res = append(res, name)
	}
	return res
}

// W creates a wire.
func W(name string) *wire.Wire {
	return &wire.Wire{Name: name}
}

var (
	examples = map[string]func(*circuit.Circuit) []*wire.Wire{
		"TransistorEmitter": func(c *circuit.Circuit) []*wire.Wire {
			return gate.TransistorEmitter(c.Group(""), c.In("base"), c.In("collector"))
		},
		"TransistorGnd": func(c *circuit.Circuit) []*wire.Wire {
			return gate.TransistorGnd(c.Group(""), c.In("base"), c.In("collector"))
		},
		"Transistor": func(c *circuit.Circuit) []*wire.Wire {
			return gate.Transistor(c.Group(""), c.In("base"), c.In("collector"))
		},
		"Not": func(c *circuit.Circuit) []*wire.Wire {
			return []*wire.Wire{gate.Not(c.Group(""), c.In("a"))}
		},
		"And": func(c *circuit.Circuit) []*wire.Wire {
			return []*wire.Wire{gate.And(c.Group(""), c.In("a"), c.In("b"))}
		},
		"Or": func(c *circuit.Circuit) []*wire.Wire {
			return []*wire.Wire{gate.Or(c.Group(""), c.In("a"), c.In("b"))}
		},
		"OrRes": func(c *circuit.Circuit) []*wire.Wire {
			bOrRes := &wire.Wire{Name: "bOrRes"}
			return []*wire.Wire{gate.OrRes(c.Group(""), bOrRes, c.In("a"), bOrRes)}
		},
		"Nand": func(c *circuit.Circuit) []*wire.Wire {
			return []*wire.Wire{gate.Nand(c.Group(""), c.In("a"), c.In("b"))}
		},
		"Nand(Nand)": func(c *circuit.Circuit) []*wire.Wire {
			g := c.Group("")
			return []*wire.Wire{gate.Nand(g, c.In("a"), gate.Nand(g, c.In("b"), c.In("c")))}
		},
		"Xor": func(c *circuit.Circuit) []*wire.Wire {
			return []*wire.Wire{gate.Xor(c.Group(""), c.In("a"), c.In("b"))}
		},
		"Nor": func(c *circuit.Circuit) []*wire.Wire {
			return []*wire.Wire{gate.Nor(c.Group(""), c.In("a"), c.In("b"))}
		},
		"HalfAdder": func(c *circuit.Circuit) []*wire.Wire {
			return sum.HalfAdder(c.Group(""), c.In("a"), c.In("b"))
		},
		"Adder": func(c *circuit.Circuit) []*wire.Wire {
			return sum.Adder(c.Group(""), c.In("a"), c.In("b"), c.In("cin"))
		},
		"Adder2": func(c *circuit.Circuit) []*wire.Wire {
			return sum.Adder2(c.Group(""), c.In("a1"), c.In("a2"), c.In("b1"), c.In("b2"), c.In("cin"))
		},
		"Adder4": func(c *circuit.Circuit) []*wire.Wire {
			return sum.Adder4(
				c.Group(""),
				c.In("a1"), c.In("a2"), c.In("a3"), c.In("a4"),
				c.In("b1"), c.In("b2"), c.In("b3"), c.In("b4"),
				c.In("cin"))
		},
		"Adder8": func(c *circuit.Circuit) []*wire.Wire {
			return sum.Adder8(
				c.Group(""),
				[8]*wire.Wire{
					c.In("a1"), c.In("a2"), c.In("a3"), c.In("a4"), c.In("a5"), c.In("a6"), c.In("a7"), c.In("a8"),
				},
				[8]*wire.Wire{
					c.In("b1"), c.In("b2"), c.In("b3"), c.In("b4"), c.In("b5"), c.In("b6"), c.In("b7"), c.In("b8"),
				},
				c.In("cin"))
		},
		"AdderN": func(c *circuit.Circuit) []*wire.Wire {
			return sum.AdderN(
				c.Group(""),
				[]*wire.Wire{c.In("a1"), c.In("a2"), c.In("a3"), c.In("a4")},
				[]*wire.Wire{c.In("b1"), c.In("b2"), c.In("b3"), c.In("b4")},
				c.In("cin"))
		},
		"SRLatch": func(c *circuit.Circuit) []*wire.Wire {
			return latch.SRLatch(c.Group(""), c.In("s"), c.In("r"))
		},
		"SRLatchWithEnable": func(c *circuit.Circuit) []*wire.Wire {
			return latch.SRLatchWithEnable(c.Group(""), c.In("s"), c.In("r"), c.In("e"))
		},
		"DLatch": func(c *circuit.Circuit) []*wire.Wire {
			return latch.DLatch(c.Group(""), c.In("d"), c.In("e"))
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
				[4]*wire.Wire{c.In("a1"), c.In("a2"), c.In("a3"), c.In("a4")}, c.In("ai"), c.In("ao"),
				[4]*wire.Wire{c.In("b1"), c.In("b2"), c.In("b3"), c.In("b4")}, c.In("bi"), c.In("bo"),
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
			wa, wb := W("wa"), W("wb")
			return append(Bus(c.Group(""), c.In("bus"), c.In("a"), c.In("b"), c.In("r"), wa, wb), wa, wb)
		},
		"Bus2": func(c *circuit.Circuit) []*wire.Wire {
			wa1, wa2 := W("wa1"), W("wa2")
			wb1, wb2 := W("wb1"), W("wb2")
			return append(Bus2(
				c.Group(""), c.In("bus1"), c.In("bus2"),
				c.In("a1"), c.In("a2"), c.In("b1"), c.In("b2"), c.In("r1"), c.In("r2"),
				wa1, wa2, wb1, wb2),
				wa1, wa2, wb1, wb2)
		},
		"Bus4": func(c *circuit.Circuit) []*wire.Wire {
			wa1, wa2, wa3, wa4 := W("wa1"), W("wa2"), W("wa3"), W("wa4")
			wb1, wb2, wb3, wb4 := W("wb1"), W("wb2"), W("wb3"), W("wb4")
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
			wa1, wa2, wa3, wa4 := W("wa1"), W("wa2"), W("wa3"), W("wa4")
			wa5, wa6, wa7, wa8 := W("wa5"), W("wa6"), W("wa7"), W("wa8")
			wb1, wb2, wb3, wb4 := W("wb1"), W("wb2"), W("wb3"), W("wb4")
			wb5, wb6, wb7, wb8 := W("wb5"), W("wb6"), W("wb7"), W("wb8")
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
			c.AddInputValidation(aluWithBusInputValidation(ai, ao, bi, bo, ri, ro))
			return AluWithBus(c.Group(""), bus, ai, ao, bi, bo, ri, ro, cin)
		},
		"AluWithBus2": func(c *circuit.Circuit) []*wire.Wire {
			bus1, bus2 := c.In("bus1"), c.In("bus2")
			ai, ao := c.In("ai"), c.In("ao")
			bi, bo := c.In("bi"), c.In("bo")
			ri, ro := c.In("ri"), c.In("ro")
			cin := c.In("cin")
			c.AddInputValidation(aluWithBusInputValidation(ai, ao, bi, bo, ri, ro))
			return AluWithBus2(c.Group(""), bus1, bus2, ai, ao, bi, bo, ri, ro, cin)
		},
		"AluWithBus4": func(c *circuit.Circuit) []*wire.Wire {
			bus := [4]*wire.Wire{c.In("bus1"), c.In("bus2"), c.In("bus3"), c.In("bus4")}
			ai, ao := c.In("ai"), c.In("ao")
			bi, bo := c.In("bi"), c.In("bo")
			ri, ro := c.In("ri"), c.In("ro")
			cin := c.In("cin")
			c.AddInputValidation(aluWithBusInputValidation(ai, ao, bi, bo, ri, ro))
			return AluWithBus4(c.Group(""), bus, ai, ao, bi, bo, ri, ro, cin)
		},
		"AluWithBus8": func(c *circuit.Circuit) []*wire.Wire {
			bus := [8]*wire.Wire{
				c.In("bus1"), c.In("bus2"), c.In("bus3"), c.In("bus4"),
				c.In("bus5"), c.In("bus6"), c.In("bus7"), c.In("bus8"),
			}
			ai, ao := c.In("ai"), c.In("ao")
			bi, bo := c.In("bi"), c.In("bo")
			ri, ro := c.In("ri"), c.In("ro")
			cin := c.In("cin")
			c.AddInputValidation(aluWithBusInputValidation(ai, ao, bi, bo, ri, ro))
			return AluWithBus8(c.Group(""), bus, ai, ao, bi, bo, ri, ro, cin)
		},
		"RAM": func(c *circuit.Circuit) []*wire.Wire {
			return RAM(c.Group(""), c.In("a"), c.In("d"), c.In("ei"), c.In("eo"))
		},
		"": func(_ *circuit.Circuit) []*wire.Wire {
			return nil
		},
	}
)
