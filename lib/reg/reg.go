// Package reg defines registers.
package reg

import (
	"github.com/kssilveira/circuit-engine/group"
	"github.com/kssilveira/circuit-engine/lib/gate"
	"github.com/kssilveira/circuit-engine/lib/latch"
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
