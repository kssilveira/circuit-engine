// Package lib contains a library of circuits.
package lib

import (
	"slices"

	"github.com/kssilveira/circuit-engine/circuit"
	"github.com/kssilveira/circuit-engine/group"
	"github.com/kssilveira/circuit-engine/lib/alu"
	"github.com/kssilveira/circuit-engine/lib/bus"
	"github.com/kssilveira/circuit-engine/lib/gate"
	"github.com/kssilveira/circuit-engine/lib/latch"
	"github.com/kssilveira/circuit-engine/lib/reg"
	"github.com/kssilveira/circuit-engine/lib/sum"
	"github.com/kssilveira/circuit-engine/sfmt"
	"github.com/kssilveira/circuit-engine/wire"
)

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
	r1 := reg.Register(group, d, ei[0], eo[0])
	r2 := reg.Register(group, d, ei[1], eo[1])
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
		"HalfSum": func(c *circuit.Circuit) []*wire.Wire {
			return sum.HalfSum(c.Group(""), c.In("a"), c.In("b"))
		},
		"Sum": func(c *circuit.Circuit) []*wire.Wire {
			return sum.Sum(c.Group(""), c.In("a"), c.In("b"), c.In("cin"))
		},
		"Sum2": func(c *circuit.Circuit) []*wire.Wire {
			return sum.Sum2(c.Group(""), c.In("a1"), c.In("a2"), c.In("b1"), c.In("b2"), c.In("cin"))
		},
		"Sum4": func(c *circuit.Circuit) []*wire.Wire {
			return sum.Sum4(
				c.Group(""),
				c.In("a1"), c.In("a2"), c.In("a3"), c.In("a4"),
				c.In("b1"), c.In("b2"), c.In("b3"), c.In("b4"),
				c.In("cin"))
		},
		"Sum8": func(c *circuit.Circuit) []*wire.Wire {
			return sum.Sum8(
				c.Group(""),
				[8]*wire.Wire{
					c.In("a1"), c.In("a2"), c.In("a3"), c.In("a4"), c.In("a5"), c.In("a6"), c.In("a7"), c.In("a8"),
				},
				[8]*wire.Wire{
					c.In("b1"), c.In("b2"), c.In("b3"), c.In("b4"), c.In("b5"), c.In("b6"), c.In("b7"), c.In("b8"),
				},
				c.In("cin"))
		},
		"SumN": func(c *circuit.Circuit) []*wire.Wire {
			return sum.N(
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
			return reg.Register(c.Group(""), c.In("d"), c.In("ei"), c.In("eo"))
		},
		"Register2": func(c *circuit.Circuit) []*wire.Wire {
			return reg.Register2(c.Group(""), c.In("d1"), c.In("d2"), c.In("ei"), c.In("eo"))
		},
		"Register4": func(c *circuit.Circuit) []*wire.Wire {
			return reg.Register4(c.Group(""), c.In("d1"), c.In("d2"), c.In("d3"), c.In("d4"), c.In("ei"), c.In("eo"))
		},
		"Register8": func(c *circuit.Circuit) []*wire.Wire {
			return reg.Register8(
				c.Group(""),
				c.In("d1"), c.In("d2"), c.In("d3"), c.In("d4"),
				c.In("d5"), c.In("d6"), c.In("d7"), c.In("d8"),
				c.In("ei"), c.In("eo"))
		},
		"Alu": func(c *circuit.Circuit) []*wire.Wire {
			return alu.Alu(
				c.Group(""),
				c.In("a"), c.In("ai"), c.In("ao"),
				c.In("b"), c.In("bi"), c.In("bo"),
				c.In("ri"), c.In("ro"), c.In("cin"))
		},
		"Alu2": func(c *circuit.Circuit) []*wire.Wire {
			return alu.Alu2(
				c.Group(""),
				c.In("a1"), c.In("a2"), c.In("ai"), c.In("ao"),
				c.In("b1"), c.In("b2"), c.In("bi"), c.In("bo"),
				c.In("ri"), c.In("ro"), c.In("cin"))
		},
		"Alu4": func(c *circuit.Circuit) []*wire.Wire {
			return alu.Alu4(
				c.Group(""),
				[4]*wire.Wire{c.In("a1"), c.In("a2"), c.In("a3"), c.In("a4")}, c.In("ai"), c.In("ao"),
				[4]*wire.Wire{c.In("b1"), c.In("b2"), c.In("b3"), c.In("b4")}, c.In("bi"), c.In("bo"),
				c.In("ri"), c.In("ro"), c.In("cin"))
		},
		"Alu8": func(c *circuit.Circuit) []*wire.Wire {
			return alu.Alu8(
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
			return append(bus.Bus(c.Group(""), c.In("bus"), c.In("a"), c.In("b"), c.In("r"), wa, wb), wa, wb)
		},
		"Bus2": func(c *circuit.Circuit) []*wire.Wire {
			wa1, wa2 := W("wa1"), W("wa2")
			wb1, wb2 := W("wb1"), W("wb2")
			return append(bus.Bus2(
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
			return slices.Concat(bus.Bus4(
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
			return slices.Concat(bus.Bus8(
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
			c.AddInputValidation(alu.WithBusInputValidation(ai, ao, bi, bo, ri, ro))
			return alu.WithBus(c.Group(""), bus, ai, ao, bi, bo, ri, ro, cin)
		},
		"AluWithBus2": func(c *circuit.Circuit) []*wire.Wire {
			bus1, bus2 := c.In("bus1"), c.In("bus2")
			ai, ao := c.In("ai"), c.In("ao")
			bi, bo := c.In("bi"), c.In("bo")
			ri, ro := c.In("ri"), c.In("ro")
			cin := c.In("cin")
			c.AddInputValidation(alu.WithBusInputValidation(ai, ao, bi, bo, ri, ro))
			return alu.WithBus2(c.Group(""), bus1, bus2, ai, ao, bi, bo, ri, ro, cin)
		},
		"AluWithBus4": func(c *circuit.Circuit) []*wire.Wire {
			bus := [4]*wire.Wire{c.In("bus1"), c.In("bus2"), c.In("bus3"), c.In("bus4")}
			ai, ao := c.In("ai"), c.In("ao")
			bi, bo := c.In("bi"), c.In("bo")
			ri, ro := c.In("ri"), c.In("ro")
			cin := c.In("cin")
			c.AddInputValidation(alu.WithBusInputValidation(ai, ao, bi, bo, ri, ro))
			return alu.WithBus4(c.Group(""), bus, ai, ao, bi, bo, ri, ro, cin)
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
			c.AddInputValidation(alu.WithBusInputValidation(ai, ao, bi, bo, ri, ro))
			return alu.WithBus8(c.Group(""), bus, ai, ao, bi, bo, ri, ro, cin)
		},
		"RAM": func(c *circuit.Circuit) []*wire.Wire {
			return RAM(c.Group(""), c.In("a"), c.In("d"), c.In("ei"), c.In("eo"))
		},
		"": func(_ *circuit.Circuit) []*wire.Wire {
			return nil
		},
	}
)
