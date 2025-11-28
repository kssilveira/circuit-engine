package group

import (
	"fmt"
	"strings"

	"github.com/kssilveira/circuit-engine/component"
	"github.com/kssilveira/circuit-engine/config"
	"github.com/kssilveira/circuit-engine/draw"
	"github.com/kssilveira/circuit-engine/jointwire"
	"github.com/kssilveira/circuit-engine/transistor"
	"github.com/kssilveira/circuit-engine/wire"
)

type Group struct {
	Name       string
	Vcc        *wire.Wire
	Gnd        *wire.Wire
	Unused     *wire.Wire
	Components []component.Component
}

func (g *Group) Group(name string) *Group {
	res := &Group{Name: name, Vcc: g.Vcc, Gnd: g.Gnd, Unused: g.Unused}
	g.Components = append(g.Components, res)
	return res
}

func (g *Group) Update() {
	for _, component := range g.Components {
		component.Update()
	}
}

func (g *Group) JointWire(res, a, b *wire.Wire) {
	g.jointWire(res, a, b, false /* isAnd */)
}

func (g *Group) JointWireIsAnd(res, a, b *wire.Wire) {
	g.jointWire(res, a, b, true /* isAnd */)
}

func (g *Group) jointWire(res, a, b *wire.Wire, isAnd bool) {
	g.Components = append(g.Components, &jointwire.JointWire{Res: res, A: a, B: b, IsAnd: isAnd})
}

func (g *Group) Transistor(base, collector, emitter, collectorOut *wire.Wire) {
	g.AddTransistor(&transistor.Transistor{Base: base, Collector: collector, Emitter: emitter, CollectorOut: collectorOut})
}

func (g *Group) AddTransistor(transistor *transistor.Transistor) {
	if transistor.CollectorOut == nil {
		transistor.CollectorOut = g.Unused
	}
	g.Components = append(g.Components, transistor)
}

func (g *Group) AddTransistors(transistors []*transistor.Transistor) {
	for _, transistor := range transistors {
		g.AddTransistor(transistor)
	}
}

var (
	horizontalLine = strings.Repeat("-", 10)
)

func (g Group) String(depth int, cfg config.Config) string {
	if cfg.MaxPrintDepth >= 0 && depth >= cfg.MaxPrintDepth {
		return ""
	}
	prefix := draw.StringPrefix(depth)
	res := []string{
		prefix + g.Name,
		prefix + horizontalLine,
	}
	for _, component := range g.Components {
		one := component.String(depth+1, cfg)
		if one == "" {
			continue
		}
		res = append(res, one)
	}
	res = append(res, prefix+horizontalLine)
	return strings.Join(res, "\n")
}

func (g Group) Graph(depth int, cfg config.Config) string {
	if cfg.MaxPrintDepth >= 0 && depth >= cfg.MaxPrintDepth {
		return ""
	}
	prefix := draw.GraphPrefix(depth)
	nextPrefix := draw.GraphPrefix(depth + 1)
	res := []string{
		fmt.Sprintf("%ssubgraph cluster_%p {", prefix, &g),
		fmt.Sprintf(`%slabel="%s";`, nextPrefix, g.Name),
		fmt.Sprintf(`%sgraph[style=dotted];`, nextPrefix),
		fmt.Sprintf(`%s"%p"[style=invis,shape=point];`, nextPrefix, &g),
	}
	for _, component := range g.Components {
		one := component.Graph(depth+1, cfg)
		if one == "" {
			continue
		}
		res = append(res, one)
	}
	res = append(res, fmt.Sprintf("%s}", prefix))
	return strings.Join(res, "\n")
}
