package draw

import (
	"strings"

	"github.com/kssilveira/circuit-engine/sfmt"
	"github.com/kssilveira/circuit-engine/wire"
)

func EdgeColor(a, b *wire.Wire) string {
	EdgeColor := "blue"
	if a.Bit.Get(nil) || b.Bit.Get(nil) {
		EdgeColor = "red"
	}
	return sfmt.Sprintf(`[color="%s"]`, EdgeColor)
}

func StringPrefix(depth int) string {
	return strings.Repeat("|", depth)
}

func GraphPrefix(depth int) string {
	return strings.Repeat(" ", depth)
}
