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
	group := parent.Group(sfmt.Sprintf("SRLATCH(%s,%s,%s)", s.Name, r.Name, q.Name))
	nqname := sfmt.Sprintf("n%s", q.Name)
	nq := &wire.Wire{Name: nqname}
	qname := q.Name
	gate.NorRes(group, q, r, nq)
	gate.NorRes(group, nq, s, q)
	q.Name = qname
	nq.Name = nqname
	return []*wire.Wire{q, nq}
}

// SRLatchWithEnable adds a set-reset latch with enable wire.
func SRLatchWithEnable(parent *group.Group, s, r, e *wire.Wire) []*wire.Wire {
	q := &wire.Wire{Name: "q"}
	return SRLatchResWithEnable(parent, q, s, r, e)
}

// SRLatchResWithEnable adds a set-reset latch with enable wire using the result parameter.
func SRLatchResWithEnable(parent *group.Group, q, s, r, e *wire.Wire) []*wire.Wire {
	group := parent.Group(sfmt.Sprintf("SRLATCHEN(%s,%s,%s,%s)", s.Name, r.Name, e.Name, q.Name))
	return SRLatchRes(group, q, gate.And(group, s, e), gate.And(group, r, e))
}

// DLatch adds a data latch.
func DLatch(parent *group.Group, d, e *wire.Wire) []*wire.Wire {
	q := &wire.Wire{Name: "q"}
	return DLatchRes(parent, q, d, e)
}

// DLatchRes adds a data latch using the result parameter.
func DLatchRes(parent *group.Group, q, d, e *wire.Wire) []*wire.Wire {
	group := parent.Group(sfmt.Sprintf("DLATCH(%s,%s,%s)", d.Name, e.Name, q.Name))
	return SRLatchResWithEnable(group, q, d, gate.Not(group, d), e)
}

// MSJKLatch adds a master-slave JK latch.
func MSJKLatch(parent *group.Group, j, k, e *wire.Wire) []*wire.Wire {
	mq := &wire.Wire{Name: "mq"}
	return MSJKLatchRes(parent, mq, j, k, e)
}

// MSJKLatchRes adds a master-slave JK latch using the result parameter.
func MSJKLatchRes(parent *group.Group, mq, j, k, e *wire.Wire) []*wire.Wire {
	group := parent.Group(sfmt.Sprintf("MSJKLATCH(%s,%s,%s,%s)", j.Name, k.Name, e.Name, mq.Name))
	sq := &wire.Wire{Name: "sq"}
	sqs := SRLatchResWithEnable(group, sq, gate.And(group, j, gate.Not(group, mq)), gate.And(group, k, mq), e)
	return SRLatchResWithEnable(group, mq, sq, sqs[1], gate.Not(group, e))
}
