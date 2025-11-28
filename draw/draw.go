package draw

import (
	"fmt"
	"strings"

	"github.com/kssilveira/circuit-engine/wire"
)

func EdgeColor(a, b *wire.Wire) string {
	EdgeColor := "blue"
	if a.Bit.Get(nil) || b.Bit.Get(nil) {
		EdgeColor = "red"
	}
	return fmt.Sprintf(`[color="%s"]`, EdgeColor)
}

func StringPrefix(depth int) string {
	return strings.Repeat("|", depth)
}

func GraphPrefix(depth int) string {
	return strings.Repeat(" ", depth)
}
