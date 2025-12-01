// Package latch defines latches.
package latch

import (
	"github.com/kssilveira/circuit-engine/group"
	"github.com/kssilveira/circuit-engine/lib/gate"
	"github.com/kssilveira/circuit-engine/sfmt"
	"github.com/kssilveira/circuit-engine/wire"
)

// SRLatch adds a set-reset latch.
func SRLatch(parent *group.Group, s, r *wire.Wire) []*wire.Wire {
	q := &wire.Wire{Name: "q"}
	return SRLatchRes(parent, q, s, r)
}

// SRLatchRes adds a set-reset latch using the result parameter.
func SRLatchRes(parent *group.Group, q, s, r *wire.Wire) []*wire.Wire {
	group := parent.Group(sfmt.Sprintf("SRLATCH(%s,%s)", s.Name, r.Name))
	nq := &wire.Wire{Name: "nq"}
	name := q.Name
	gate.NorRes(group, q, r, nq)
	gate.NorRes(group, nq, s, q)
	q.Name = name
	nq.Name = "nq"
	return []*wire.Wire{q, nq}
}

// SRLatchWithEnable adds a set-reset latch with enable wire.
func SRLatchWithEnable(parent *group.Group, s, r, e *wire.Wire) []*wire.Wire {
	q := &wire.Wire{Name: "q"}
	return SRLatchResWithEnable(parent, q, s, r, e)
}

// SRLatchResWithEnable adds a set-reset latch with enable wrie using the result parameter.
func SRLatchResWithEnable(parent *group.Group, q, s, r, e *wire.Wire) []*wire.Wire {
	group := parent.Group(sfmt.Sprintf("SRLATCHEN(%s,%s,%s)", s.Name, r.Name, e.Name))
	return SRLatchRes(group, q, gate.And(group, s, e), gate.And(group, r, e))
}

// DLatch adds a data latch.
func DLatch(parent *group.Group, d, e *wire.Wire) []*wire.Wire {
	q := &wire.Wire{Name: "q"}
	return DLatchRes(parent, q, d, e)
}

// DLatchRes adds a data latch using the result parameter.
func DLatchRes(parent *group.Group, q, d, e *wire.Wire) []*wire.Wire {
	group := parent.Group(sfmt.Sprintf("DLATCH(%s,%s)", d.Name, e.Name))
	return SRLatchResWithEnable(group, q, d, gate.Not(group, d), e)
}
