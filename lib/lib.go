// Package lib contains a library of circuits.
package lib

import (
	"github.com/kssilveira/circuit-engine/circuit"
	"github.com/kssilveira/circuit-engine/lib/alu"
	"github.com/kssilveira/circuit-engine/lib/bus"
	"github.com/kssilveira/circuit-engine/lib/gate"
	"github.com/kssilveira/circuit-engine/lib/latch"
	"github.com/kssilveira/circuit-engine/lib/ram"
	"github.com/kssilveira/circuit-engine/lib/reg"
	"github.com/kssilveira/circuit-engine/lib/sum"
	"github.com/kssilveira/circuit-engine/wire"
)

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
			return gate.TransistorEmitter(c.Group(""), c.In("b"), c.In("c"))
		},
		"TransistorGnd": func(c *circuit.Circuit) []*wire.Wire {
			return gate.TransistorGnd(c.Group(""), c.In("b"), c.In("c"))
		},
		"Transistor": func(c *circuit.Circuit) []*wire.Wire {
			return gate.Transistor(c.Group(""), c.In("b"), c.In("c"))
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
			return sum.Sum(c.Group(""), c.In("a"), c.In("b"), c.In("c"))
		},
		"Sum2": func(c *circuit.Circuit) []*wire.Wire {
			return sum.Sum2(c.Group(""), c.In("a0"), c.In("a1"), c.In("b0"), c.In("b1"), c.In("c"))
		},
		"SumN": func(c *circuit.Circuit) []*wire.Wire {
			return sum.N(c.Group(""), []*wire.Wire{c.In("a0"), c.In("a1")}, []*wire.Wire{c.In("b0"), c.In("b1")}, c.In("c"))
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
			return reg.Register(c.Group(""), c.In("d"), c.In("i"), c.In("o"))
		},
		"Register2": func(c *circuit.Circuit) []*wire.Wire {
			return reg.Register2(c.Group(""), c.In("d0"), c.In("d1"), c.In("i"), c.In("o"))
		},
		"RegisterN": func(c *circuit.Circuit) []*wire.Wire {
			return reg.N(c.Group(""), []*wire.Wire{c.In("d0"), c.In("d1")}, c.In("i"), c.In("o"))
		},
		"Alu": func(c *circuit.Circuit) []*wire.Wire {
			return alu.Alu(
				c.Group(""),
				c.In("a"), c.In("ai"), c.In("ao"),
				c.In("b"), c.In("bi"), c.In("bo"),
				c.In("ri"), c.In("ro"), c.In("c"))
		},
		"Alu2": func(c *circuit.Circuit) []*wire.Wire {
			return alu.Alu2(
				c.Group(""),
				c.In("a0"), c.In("a1"), c.In("ai"), c.In("ao"),
				c.In("b0"), c.In("b1"), c.In("bi"), c.In("bo"),
				c.In("ri"), c.In("ro"), c.In("c"))
		},
		"AluN": func(c *circuit.Circuit) []*wire.Wire {
			return alu.N(
				c.Group(""),
				[]*wire.Wire{c.In("a0"), c.In("a1")}, c.In("ai"), c.In("ao"),
				[]*wire.Wire{c.In("b0"), c.In("b1")}, c.In("bi"), c.In("bo"),
				c.In("ri"), c.In("ro"), c.In("c"))
		},
		"Bus": func(c *circuit.Circuit) []*wire.Wire {
			aw, bw := W("aw"), W("bw")
			return append(bus.Bus(c.Group(""), c.In("d"), c.In("ar"), c.In("br"), c.In("r"), aw, bw), aw, bw)
		},
		"Bus2": func(c *circuit.Circuit) []*wire.Wire {
			aw0, aw1 := W("aw0"), W("aw1")
			bw0, bw1 := W("bw0"), W("bw1")
			return append(bus.Bus2(
				c.Group(""), c.In("d0"), c.In("d1"),
				c.In("ar0"), c.In("ar1"), c.In("br0"), c.In("br1"), c.In("r0"), c.In("r1"),
				aw0, aw1, bw0, bw1),
				aw0, aw1, bw0, bw1)
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
			return ram.RAM(
				c.Group(""), []*wire.Wire{c.In("a")}, []*wire.Wire{c.In("d")}, c.In("ei"), c.In("eo"))
		},
		"RAMa2": func(c *circuit.Circuit) []*wire.Wire {
			return ram.RAM(
				c.Group(""), []*wire.Wire{c.In("a0"), c.In("a1")}, []*wire.Wire{c.In("d")}, c.In("ei"), c.In("eo"))
		},
		"RAMb2": func(c *circuit.Circuit) []*wire.Wire {
			return ram.RAM(
				c.Group(""), []*wire.Wire{c.In("a")}, []*wire.Wire{c.In("d0"), c.In("d1")}, c.In("ei"), c.In("eo"))
		},
		"RAMa2b2": func(c *circuit.Circuit) []*wire.Wire {
			return ram.RAM(
				c.Group(""), []*wire.Wire{c.In("a0"), c.In("a1")}, []*wire.Wire{c.In("d0"), c.In("d1")},
				c.In("ei"), c.In("eo"))
		},
		"": func(_ *circuit.Circuit) []*wire.Wire {
			return nil
		},
	}
)
