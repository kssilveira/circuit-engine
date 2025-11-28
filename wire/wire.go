package wire

import (
	"fmt"
	"strings"

	"github.com/kssilveira/circuit-engine/bit"
)

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
		BoolToString(w.Bit.Get(nil)),
	}
	if w.Gnd.Get(nil) {
		list = append(list, "Gnd")
	}
	res := []string{
		fmt.Sprintf("%v=", w.Name),
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

func BoolToString(a bool) string {
	if a {
		return "1"
	}
	return "0"
}
