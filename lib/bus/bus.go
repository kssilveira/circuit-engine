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
	rbus1 := Bus4(
		group, wire.W4(bus[:4]), wire.W4(a[:4]), wire.W4(b[:4]), wire.W4(r[:4]),
		wire.W4(wa[:4]), wire.W4(wb[:4]))
	rbus2 := Bus4(
		group, wire.W4(bus[4:]), wire.W4(a[4:]), wire.W4(b[4:]), wire.W4(r[4:]),
		wire.W4(wa[4:]), wire.W4(wb[4:]))
	return slices.Concat(rbus1, rbus2)
}
