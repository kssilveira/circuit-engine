// Package group encapsulates a group of components.
package group

import (
	"strings"

	"github.com/kssilveira/circuit-engine/component"
	"github.com/kssilveira/circuit-engine/config"
	"github.com/kssilveira/circuit-engine/draw"
	"github.com/kssilveira/circuit-engine/jointwire"
	"github.com/kssilveira/circuit-engine/sfmt"
	"github.com/kssilveira/circuit-engine/transistor"
	"github.com/kssilveira/circuit-engine/wire"
)

// Group contains a group of components.
type Group struct {
	Name       string
	Components []component.Component
}

// Group creates a new group.
func (g *Group) Group(name string) *Group {
	res := &Group{Name: name}
	g.Components = append(g.Components, res)
	return res
}

// Update updates all components.
func (g *Group) Update(updateReaders bool) {
	for _, component := range g.Components {
		component.Update(updateReaders)
	}
}

// Vcc creates a Vcc.
func (g *Group) Vcc() *wire.Wire {
	res := &wire.Wire{Name: "Vcc"}
	res.Bit.SilentSet(true)
	return res
}

// Gnd creates a Gnd.
func (g *Group) Gnd() *wire.Wire {
	res := &wire.Wire{Name: "Gnd"}
	res.Gnd.SilentSet(true)
	return res
}

// True creates a True.
func (g *Group) True() *wire.Wire {
	res := &wire.Wire{Name: "T"}
	res.Bit.SilentSet(true)
	return res
}

// False creates a False.
func (g *Group) False() *wire.Wire {
	return &wire.Wire{Name: "F"}
}

// Unused creates an Unused.
func (g *Group) Unused() *wire.Wire {
	return &wire.Wire{Name: "Unused"}
}

// JointWire adds a joint wire.
func (g *Group) JointWire(res, a, b *wire.Wire) {
	g.jointWire(res, a, b, false /* isAnd */)
}

// JointWireIsAnd adds an AND joint wire.
func (g *Group) JointWireIsAnd(res, a, b *wire.Wire) {
	g.jointWire(res, a, b, true /* isAnd */)
}

func (g *Group) jointWire(res, a, b *wire.Wire, isAnd bool) {
	g.Components = append(g.Components, &jointwire.JointWire{Res: res, A: a, B: b, IsAnd: isAnd})
}

// Transistor adds a transistor.
func (g *Group) Transistor(base, collector, emitter, collectorOut *wire.Wire) {
	g.AddTransistor(&transistor.Transistor{Base: base, Collector: collector, Emitter: emitter, CollectorOut: collectorOut})
}

// AddTransistor adds a transistor.
func (g *Group) AddTransistor(transistor *transistor.Transistor) {
	if transistor.CollectorOut == nil {
		transistor.CollectorOut = g.Unused()
	}
	g.Components = append(g.Components, transistor)
}

// AddTransistors adds multiple transistors.
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

// Graph returns a graphviz graph.
func (g Group) Graph(depth int, cfg config.Config) string {
	if cfg.MaxPrintDepth >= 0 && depth >= cfg.MaxPrintDepth {
		return ""
	}
	prefix := draw.GraphPrefix(depth)
	nextPrefix := draw.GraphPrefix(depth + 1)
	res := []string{
		sfmt.Sprintf("%ssubgraph cluster_%p {", prefix, &g),
		sfmt.Sprintf(`%slabel="%s";`, nextPrefix, g.Name),
		sfmt.Sprintf(`%sgraph[style=dotted];`, nextPrefix),
		sfmt.Sprintf(`%s"%p"[style=invis,shape=point];`, nextPrefix, &g),
	}
	for _, component := range g.Components {
		one := component.Graph(depth+1, cfg)
		if one == "" {
			continue
		}
		res = append(res, one)
	}
	res = append(res, sfmt.Sprintf("%s}", prefix))
	return strings.Join(res, "\n")
}
