// Package ram defines random access memory components.
package ram

import (
	"slices"

	"github.com/kssilveira/circuit-engine/group"
	"github.com/kssilveira/circuit-engine/lib/gate"
	"github.com/kssilveira/circuit-engine/lib/reg"
	"github.com/kssilveira/circuit-engine/sfmt"
	"github.com/kssilveira/circuit-engine/wire"
)

// RAM adds a random access memory.
func RAM(parent *group.Group, a, d, ei, eo *wire.Wire) []*wire.Wire {
	group := parent.Group(sfmt.Sprintf("RAM(%v,%v)", a.Name, d.Name))
	s := ramAddress(group, a)
	rei, reo := ramEnable(group, s, ei, eo)
	return ramRegisters(group, d, s, rei, reo)
}

func ramAddress(group *group.Group, a *wire.Wire) []*wire.Wire {
	s1 := gate.Not(group, a)
	s1.Name = group.Name + "-s1"
	s2 := gate.Or(group, a, a)
	s2.Name = group.Name + "-s2"
	return []*wire.Wire{s1, s2}
}

func ramEnable(group *group.Group, s []*wire.Wire, ei, eo *wire.Wire) ([]*wire.Wire, []*wire.Wire) {
	ei1 := gate.And(group, ei, s[0])
	ei1.Name = group.Name + "-ei1"
	ei2 := gate.And(group, ei, s[1])
	ei2.Name = group.Name + "-ei2"

	eo1 := gate.And(group, eo, s[0])
	eo1.Name = group.Name + "-eo1"
	eo2 := gate.And(group, eo, s[1])
	eo2.Name = group.Name + "-eo2"

	return []*wire.Wire{ei1, ei2}, []*wire.Wire{eo1, eo2}
}

func ramRegisters(group *group.Group, d *wire.Wire, s, ei, eo []*wire.Wire) []*wire.Wire {
	r1 := reg.Register(group, d, ei[0], eo[0])
	r2 := reg.Register(group, d, ei[1], eo[1])
	res := &wire.Wire{Name: group.Name}
	group.JointWire(res, r1[1], r2[1])
	return slices.Concat([]*wire.Wire{res, s[0], ei[0], eo[0]}, r1, []*wire.Wire{s[1], ei[1], eo[1]}, r2)
}
