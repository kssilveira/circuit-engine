// Package sum contains adder circuits.
package sum

import (
	"github.com/kssilveira/circuit-engine/group"
	"github.com/kssilveira/circuit-engine/lib/gate"
	"github.com/kssilveira/circuit-engine/sfmt"
	"github.com/kssilveira/circuit-engine/wire"
)

// HalfSum adds a half adder.
func HalfSum(parent *group.Group, a, b *wire.Wire) []*wire.Wire {
	group := parent.Group(sfmt.Sprintf("S(%s,%s)", a.Name, b.Name))
	res := gate.Xor(group, a, b)
	res.Name = group.Name
	carry := gate.And(group, a, b)
	carry.Name = sfmt.Sprintf("C(%s,%s)", a.Name, b.Name)
	return []*wire.Wire{res, carry}
}

// Sum adds an adder.
func Sum(parent *group.Group, a, b, cin *wire.Wire) []*wire.Wire {
	group := parent.Group(sfmt.Sprintf("S(%s,%s,%s)", a.Name, b.Name, cin.Name))
	s1 := HalfSum(group, a, b)
	s2 := HalfSum(group, s1[0], cin)
	s2[0].Name = group.Name
	carry := gate.Or(group, s1[1], s2[1])
	carry.Name = sfmt.Sprintf("C(%s,%s)", a.Name, b.Name)
	return []*wire.Wire{s2[0], carry}
}

// Sum2 adds a 2-bit adder.
func Sum2(parent *group.Group, a1, a2, b1, b2, cin *wire.Wire) []*wire.Wire {
	group := parent.Group("SUM2")
	s1 := Sum(group, a1, b1, cin)
	s2 := Sum(group, a2, b2, s1[1])
	return []*wire.Wire{s1[0], s2[0], s2[1]}
}

// N adds an N-bit adder.
func N(parent *group.Group, an, bn []*wire.Wire, cin *wire.Wire) []*wire.Wire {
	group := parent.Group(sfmt.Sprintf("SUM%d", len(an)))
	var s []*wire.Wire
	carry := cin
	for i, a := range an {
		si := Sum(group, a, bn[i], carry)
		s = append(s, si[0])
		carry = si[1]
	}
	return append(s, carry)
}
