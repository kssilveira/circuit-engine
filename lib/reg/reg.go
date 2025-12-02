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
func Register(parent *group.Group, d, ei, eo *wire.Wire) *wire.Wire {
	group := parent.Group(sfmt.Sprintf("R(%s,%s,%s)", d.Name, ei.Name, eo.Name))
	q := &wire.Wire{}
	latch.DLatchRes(group, q, gate.Or(group, gate.And(group, q, gate.Not(group, ei)), gate.And(group, d, ei)), ei)
	q.Name = "r" + group.Name[1:]
	res := gate.And(group, q, eo)
	res.Name = group.Name
	return res
}

// Register2 adds a 2-bit register.
func Register2(parent *group.Group, d1, d2, ei, eo *wire.Wire) []*wire.Wire {
	group := parent.Group("Register2")
	r1 := Register(group, d1, ei, eo)
	r2 := Register(group, d2, ei, eo)
	return []*wire.Wire{r1, r2}
}

// N adds a N-bit register.
func N(parent *group.Group, d []*wire.Wire, ei, eo *wire.Wire) []*wire.Wire {
	group := parent.Group(sfmt.Sprintf("Register%d", len(d)))
	var res []*wire.Wire
	for _, di := range d {
		res = append(res, Register(group, di, ei, eo))
	}
	return res
}
