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
func RAM(parent *group.Group, a, d []*wire.Wire, ei, eo *wire.Wire) []*wire.Wire {
	group := parent.Group("RAM")
	s := ramAddress(group, a)
	rei, reo := ramEnable(group, s, ei, eo)
	return ramRegisters(group, d, s, rei, reo)
}

func ramAddress(group *group.Group, a []*wire.Wire) []*wire.Wire {
	var s []*wire.Wire
	for address := 0; address < 1<<len(a); address++ {
		si := &wire.Wire{}
		si.Bit.Set(true)
		for i, ai := range a {
			if address>>i&1 == 1 {
				si = gate.And(group, si, ai)
			} else {
				si = gate.And(group, si, gate.Not(group, ai))
			}
		}
		si.Name = sfmt.Sprintf("%s-s%d", group.Name, address)
		s = append(s, si)
	}
	return s
}

func ramEnable(group *group.Group, s []*wire.Wire, ei, eo *wire.Wire) ([]*wire.Wire, []*wire.Wire) {
	var rei, reo []*wire.Wire
	for i, si := range s {
		reii := gate.And(group, ei, si)
		reii.Name = sfmt.Sprintf("%s-ei%d", group.Name, i)
		rei = append(rei, reii)

		reoi := gate.And(group, eo, si)
		reoi.Name = sfmt.Sprintf("%s-eo%d", group.Name, i)
		reo = append(reo, reoi)
	}
	return rei, reo
}

func ramRegisters(group *group.Group, d, s, ei, eo []*wire.Wire) []*wire.Wire {
	var prev []*wire.Wire
	for range d {
		one := &wire.Wire{}
		one.Bit.Set(false)
		prev = append(prev, one)
	}
	var all []*wire.Wire
	for i, eii := range ei {
		ri := reg.N(group, d, eii, eo[i])
		var next []*wire.Wire
		for i := range d {
			res := &wire.Wire{}
			group.JointWire(res, prev[i], ri[2*i+1])
			next = append(next, res)
		}
		prev = next
		all = slices.Concat(all, []*wire.Wire{s[i], eii, eo[i]}, ri)
	}
	for _, res := range prev {
		res.Name = group.Name
	}
	return append(prev, all...)
}
