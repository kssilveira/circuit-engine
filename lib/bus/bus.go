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

// Bion adds a communication bus with multiple inputs and outputs.
func Bion(parent *group.Group, r, w []*wire.Wire) []*wire.Wire {
	group := parent.Group(sfmt.Sprintf("B(%s)", r[0].Name))
	prev := &wire.Wire{Name: sfmt.Sprintf("%s-wire", group.Name)}
	for _, ri := range r {
		next := &wire.Wire{Name: sfmt.Sprintf("%s-wire", group.Name)}
		group.JointWire(next, prev, ri)
		prev = next
	}
	res := prev
	res.Name = group.Name
	for _, wi := range w {
		group.JointWire(wi, res, res)
	}
	return []*wire.Wire{res}
}

// BnIOn adds an N-bit communication bus with multiple inputs and outputs
func BnIOn(parent *group.Group, r, w [][]*wire.Wire) []*wire.Wire {
	group := parent.Group(sfmt.Sprintf("B(%d,%d,%d)", len(r[0]), len(r), len(w)))
	var res []*wire.Wire
	for j := range r[0] {
		var rj []*wire.Wire
		for _, ri := range r {
			rj = append(rj, ri[j])
		}
		var wj []*wire.Wire
		for _, wi := range w {
			wj = append(wj, wi[j])
		}
		res = append(res, Bion(group, rj, wj)...)
	}
	return res
}
