package transistor

import (
	"strings"

	"github.com/kssilveira/circuit-engine/config"
	"github.com/kssilveira/circuit-engine/draw"
	"github.com/kssilveira/circuit-engine/sfmt"
	"github.com/kssilveira/circuit-engine/wire"
)

type Transistor struct {
	Base         *wire.Wire
	Collector    *wire.Wire
	Emitter      *wire.Wire
	CollectorOut *wire.Wire
}

func (t *Transistor) Update() {
	t.Emitter.Bit.Set(t.Base.Bit.Get(t) && t.Collector.Bit.Get(t))
	if t.Collector.Bit.Get(t) {
		if t.Base.Bit.Get(t) && t.Emitter.Gnd.Get(t) {
			t.Collector.Gnd.Set(true)
			t.CollectorOut.Bit.Set(false)
		} else {
			t.Collector.Gnd.Set(false)
			t.CollectorOut.Bit.Set(true)
		}
	} else {
		t.Collector.Gnd.Set(false)
		t.CollectorOut.Bit.Set(false)
	}
}

func (t Transistor) String(depth int, cfg config.Config) string {
	var res []string
	for _, wire := range []*wire.Wire{t.Base, t.Collector, t.Emitter, t.CollectorOut} {
		one := sfmt.Sprintf("%v", *wire)
		if one == "" {
			continue
		}
		res = append(res, one)
	}
	return sfmt.Sprintf("%s%s", draw.StringPrefix(depth), strings.Join(res, "    "))
}

func (t Transistor) Graph(depth int, cfg config.Config) string {
	if !cfg.DrawNodes {
		return ""
	}
	prefix := draw.GraphPrefix(depth)
	var res []string
	res = append(res, sfmt.Sprintf(`"%p" [label="𓇲";shape=invtriangle];`, &t))
	for _, wire := range []*wire.Wire{t.Base, t.Collector} {
		if cfg.DrawShapePoint {
			res = append(res, sfmt.Sprintf(`%s"%v" [label= "";shape=point];`, prefix, *wire))
		}
		if cfg.DrawEdges {
			res = append(res, sfmt.Sprintf(`%s"%v" -> "%p" %s;`, prefix, *wire, &t, draw.EdgeColor(wire, wire)))
		}
	}
	for _, wire := range []*wire.Wire{t.Emitter, t.CollectorOut} {
		if wire.Name == "Unused" {
			continue
		}
		if cfg.DrawShapePoint {
			res = append(res, sfmt.Sprintf(`%s"%v" [label= "";shape=point];`, prefix, *wire))
		}
		if cfg.DrawEdges {
			res = append(res, sfmt.Sprintf(`%s"%p" -> "%v" %s;`, prefix, &t, *wire, draw.EdgeColor(wire, wire)))
		}
	}
	return strings.Join(res, "\n")
}
