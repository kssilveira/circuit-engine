// Package bus defines communication buses.
package bus

import (
	"slices"

	"github.com/kssilveira/circuit-engine/group"
	"github.com/kssilveira/circuit-engine/sfmt"
	"github.com/kssilveira/circuit-engine/wire"
)

// Bus add a communication bus.
func Bus(parent *group.Group, bus, a, b, r, wa, wb *wire.Wire) []*wire.Wire {
	group := parent.Group(sfmt.Sprintf("B(%s)", bus.Name))
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
	group := parent.Group("BUS2")
	rbus1 := Bus(group, bus1, a1, b1, r1, wa1, wb1)
	rbus2 := Bus(group, bus2, a2, b2, r2, wa2, wb2)
	return slices.Concat(rbus1, rbus2)
}

// N adds an N-bit communication bus.
func N(parent *group.Group, d, ar, br, r, aw, bw []*wire.Wire) []*wire.Wire {
	group := parent.Group(sfmt.Sprintf("BUS%d", len(d)))
	var res []*wire.Wire
	for i, di := range d {
		res = append(res, Bus(group, di, ar[i], br[i], r[i], aw[i], bw[i])...)
	}
	return res
}
