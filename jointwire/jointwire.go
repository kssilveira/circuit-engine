package jointwire

import (
	"fmt"
	"strings"

	"github.com/kssilveira/circuit-engine/config"
	"github.com/kssilveira/circuit-engine/draw"
	"github.com/kssilveira/circuit-engine/wire"
)

type JointWire struct {
	Res   *wire.Wire
	A     *wire.Wire
	B     *wire.Wire
	IsAnd bool
}

func (w *JointWire) Update() {
	if w.IsAnd {
		w.Res.Bit.Set(w.A.Bit.Get(w) && w.B.Bit.Get(w))
		return
	}
	w.Res.Bit.Set(w.A.Bit.Get(w) || w.B.Bit.Get(w))
}

func (w JointWire) String(depth int, cfg config.Config) string {
	var res []string
	for _, wire := range []*wire.Wire{w.A, w.B, w.Res} {
		one := fmt.Sprintf("%v", *wire)
		if one == "" {
			continue
		}
		res = append(res, one)
	}
	name := "OR"
	if w.IsAnd {
		name = "AND"
	}
	return fmt.Sprintf("%s%s %s", draw.StringPrefix(depth), name, strings.Join(res, "    "))
}

func (w JointWire) Graph(depth int, cfg config.Config) string {
	if !cfg.DrawNodes {
		return ""
	}
	prefix := draw.GraphPrefix(depth)
	var res []string
	if cfg.DrawShapePoint {
		res = append(res, fmt.Sprintf(`%s"%v" [label= "";shape=point];`, prefix, *w.Res))
	}
	for _, wire := range []*wire.Wire{w.A, w.B} {
		if cfg.DrawShapePoint {
			res = append(res, fmt.Sprintf(`%s"%v" [label= "";shape=point];`, prefix, *wire))
		}
		if cfg.DrawEdges {
			res = append(res, fmt.Sprintf(`%s"%v" -> "%v" %s;`, prefix, *wire, *w.Res, draw.EdgeColor(wire, w.Res)))
		}
	}
	return strings.Join(res, "\n")
}
