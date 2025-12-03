// Package wire encapsulates wires.
package wire

import (
	"strings"

	"github.com/kssilveira/circuit-engine/bit"
	"github.com/kssilveira/circuit-engine/sfmt"
)

// Wire contains a single wire.
type Wire struct {
	Name string
	Bit  bit.Bit
	Gnd  bit.Bit
}

func (w Wire) String() string {
	if w.Name == "Vcc" || w.Name == "Gnd" {
		return w.Name
	}
	if w.Name == "Unused" {
		return ""
	}
	list := []string{
		BoolToString(w.Bit.SilentGet()),
	}
	if w.Gnd.Get(nil) {
		list = append(list, "Gnd")
	}
	res := []string{
		sfmt.Sprintf("%v=", w.Name),
	}
	if len(list) > 1 {
		res = append(res, "{")
	}
	res = append(res, strings.Join(list, ", "))
	if len(list) > 1 {
		res = append(res, "}")
	}
	return strings.Join(res, "")
}

// BoolToString converts a bool to string.
func BoolToString(a bool) string {
	if a {
		return "1"
	}
	return "0"
}

// W4 creates an array with 4 wires.
func W4(w []*Wire) [4]*Wire {
	return [4]*Wire(w)
}
