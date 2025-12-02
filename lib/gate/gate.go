// Package gate contains circuits implementing logic gates.
package gate

import (
	"github.com/kssilveira/circuit-engine/group"
	"github.com/kssilveira/circuit-engine/sfmt"
	"github.com/kssilveira/circuit-engine/transistor"
	"github.com/kssilveira/circuit-engine/wire"
)

// TransistorEmitter adds a transitor-emitter.
func TransistorEmitter(parent *group.Group, base, collector *wire.Wire) []*wire.Wire {
	group := parent.Group("TransistorEmitter")
	emitter := &wire.Wire{Name: "e"}
	collectorOut := &wire.Wire{Name: "co"}
	group.Transistor(base, collector, emitter, collectorOut)
	return []*wire.Wire{emitter}
}

// TransistorGnd adds a transistor-ground.
func TransistorGnd(parent *group.Group, base, collector *wire.Wire) []*wire.Wire {
	group := parent.Group("TransistorGnd")
	collectorOut := &wire.Wire{Name: "co"}
	group.Transistor(base, collector, group.Gnd, collectorOut)
	return []*wire.Wire{collectorOut}
}

// Transistor adds a transistor.
func Transistor(parent *group.Group, base, collector *wire.Wire) []*wire.Wire {
	group := parent.Group("Transistor")
	emitter := &wire.Wire{Name: "e"}
	collectorOut := &wire.Wire{Name: "co"}
	group.Transistor(base, collector, emitter, collectorOut)
	return []*wire.Wire{emitter, collectorOut}
}

// Not adds a NOT gate.
func Not(parent *group.Group, a *wire.Wire) *wire.Wire {
	group := parent.Group(sfmt.Sprintf("NOT(%s)", a.Name))
	res := &wire.Wire{Name: group.Name}
	group.Transistor(a, group.Vcc, group.Gnd, res)
	return res
}

// And adds an AND gate.
func And(parent *group.Group, a, b *wire.Wire) *wire.Wire {
	group := parent.Group(sfmt.Sprintf("AND(%s,%s)", a.Name, b.Name))
	res := &wire.Wire{Name: group.Name}
	wire := &wire.Wire{Name: sfmt.Sprintf("%s-wire", res.Name)}
	group.AddTransistors([]*transistor.Transistor{
		{Base: a, Collector: group.Vcc, Emitter: wire},
		{Base: b, Collector: wire, Emitter: res},
	})
	return res
}

// Or adds an OR gate.
func Or(parent *group.Group, a, b *wire.Wire) *wire.Wire {
	res := &wire.Wire{}
	return OrRes(parent, res, a, b)
}

// OrRes adds an OR gate using the given result wire.
func OrRes(parent *group.Group, res, a, b *wire.Wire) *wire.Wire {
	group := parent.Group(sfmt.Sprintf("OR(%s,%s)", a.Name, b.Name))
	res.Name = group.Name
	wire1 := &wire.Wire{Name: sfmt.Sprintf("%s-wire1", res.Name)}
	wire2 := &wire.Wire{Name: sfmt.Sprintf("%s-wire2", res.Name)}
	group.AddTransistors([]*transistor.Transistor{
		{Base: a, Collector: group.Vcc, Emitter: wire1},
		{Base: b, Collector: group.Vcc, Emitter: wire2},
	})
	group.JointWire(res, wire1, wire2)
	return res
}

// Nand adds a NAND gate.
func Nand(parent *group.Group, a, b *wire.Wire) *wire.Wire {
	group := parent.Group(sfmt.Sprintf("NAND(%s,%s)", a.Name, b.Name))
	res := &wire.Wire{Name: group.Name}
	wire := &wire.Wire{Name: sfmt.Sprintf("%s-wire", res.Name)}
	group.AddTransistors([]*transistor.Transistor{
		{Base: a, Collector: group.Vcc, Emitter: wire, CollectorOut: res},
		{Base: b, Collector: wire, Emitter: group.Gnd},
	})
	return res
}

// Xor adds a XOR gate.
func Xor(parent *group.Group, a, b *wire.Wire) *wire.Wire {
	group := parent.Group(sfmt.Sprintf("XOR(%s,%s)", a.Name, b.Name))
	res := And(group, Or(group, a, b), Nand(group, a, b))
	res.Name = group.Name
	return res
}

// Nor adds a NOR gate.
func Nor(parent *group.Group, a, b *wire.Wire) *wire.Wire {
	res := &wire.Wire{}
	return NorRes(parent, res, a, b)
}

// NorRes adds a NOR gate with the given result parameter.
func NorRes(parent *group.Group, res, a, b *wire.Wire) *wire.Wire {
	group := parent.Group(sfmt.Sprintf("NOR(%s,%s)", a.Name, b.Name))
	res.Name = group.Name
	wire1 := &wire.Wire{Name: sfmt.Sprintf("%s-wire1", res.Name)}
	wire2 := &wire.Wire{Name: sfmt.Sprintf("%s-wire2", res.Name)}
	group.AddTransistors([]*transistor.Transistor{
		{Base: a, Collector: group.Vcc, Emitter: group.Gnd, CollectorOut: wire1},
		{Base: b, Collector: group.Vcc, Emitter: group.Gnd, CollectorOut: wire2},
	})
	group.JointWireIsAnd(res, wire1, wire2)
	return res
}
