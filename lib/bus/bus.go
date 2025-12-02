// Package bus defines communication buses.
package bus

import (
	"slices"

	"github.com/kssilveira/circuit-engine/group"
	"github.com/kssilveira/circuit-engine/sfmt"
	"github.com/kssilveira/circuit-engine/wire"
)

// Bus add a communication bus.
func Bus(parent *group.Group, d, ar, br, r, aw, bw *wire.Wire) []*wire.Wire {
	group := parent.Group(sfmt.Sprintf("B(%s)", d.Name))
	res := &wire.Wire{Name: group.Name}
	wire1 := &wire.Wire{Name: sfmt.Sprintf("%s-wire1", res.Name)}
	wire2 := &wire.Wire{Name: sfmt.Sprintf("%s-wire2", res.Name)}
	group.JointWire(wire1, d, ar)
	group.JointWire(wire2, wire1, br)
	group.JointWire(res, wire2, r)
	group.JointWire(aw, res, res)
	group.JointWire(bw, res, res)
	return []*wire.Wire{res}
}

// Bus2 adds a 2-bit communication bus.
func Bus2(parent *group.Group, d0, d1, ar0, ar1, br0, br1, r0, r1, aw0, aw1, bw0, bw1 *wire.Wire) []*wire.Wire {
	group := parent.Group("BUS2")
	rbus0 := Bus(group, d0, ar0, br0, r0, aw0, bw0)
	rbus1 := Bus(group, d1, ar1, br1, r1, aw1, bw1)
	return slices.Concat(rbus0, rbus1)
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
