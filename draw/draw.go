// Package draw contains the common code to print and draw circuits.
package draw

import (
	"strings"

	"github.com/kssilveira/circuit-engine/sfmt"
	"github.com/kssilveira/circuit-engine/wire"
)

// EdgeColor returns graphviz edge color.
func EdgeColor(a, b *wire.Wire) string {
	EdgeColor := "blue"
	if a.Bit.Get(nil) || b.Bit.Get(nil) {
		EdgeColor = "red"
	}
	return sfmt.Sprintf(`[color="%s"]`, EdgeColor)
}

// StringPrefix returns string prefix for print.
func StringPrefix(depth int) string {
	return strings.Repeat("|", depth)
}

// GraphPrefix returns string prefix for graph.
func GraphPrefix(depth int) string {
	return strings.Repeat(" ", depth)
}
