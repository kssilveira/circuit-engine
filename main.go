package main

import (
	"flag"
	"fmt"
	"math/rand/v2"
	"os"
	"slices"
	"strings"
)

var (
	maxPrintDepth   = flag.Int("max_print_depth", -1, "max print depth")
	drawGraph       = flag.Bool("draw_graph", false, "draw graph")
	drawSingleGraph = flag.Bool("draw_single_graph", false, "draw single graph")
	drawShapePoint  = flag.Bool("draw_shape_point", false, "draw shape point")
	drawEdges       = flag.Bool("draw_edges", true, "draw edges")
	drawNodes       = flag.Bool("draw_nodes", true, "draw nodes")
)

func boolToString(a bool) string {
	if a {
		return "1"
	}
	return "0"
}

func depthToString(depth int) string {
	if *drawGraph {
		return strings.Repeat(" ", depth)
	}
	return strings.Repeat("|", depth)
}

type Component interface {
	Update()
	String(int) string
	Graph(int) string
}

type Bool struct {
	bit     bool
	Readers []Component
}

func (b *Bool) Get(reader Component) bool {
	if reader == nil {
		return b.bit
	}
	found := false
	for _, r := range b.Readers {
		if r == reader {
			found = true
			break
		}
	}
	if !found {
		b.Readers = append(b.Readers, reader)
	}
	return b.bit
}

func (b *Bool) Set(v bool) {
	if v != b.bit {
		b.bit = v
		for _, reader := range b.Readers {
			reader.Update()
		}
	}
}

type Wire struct {
	Name string
	Bit  Bool
	Gnd  Bool
}

func (w Wire) String() string {
	if w.Name == "Vcc" || w.Name == "Gnd" {
		return w.Name
	}
	if w.Name == "Unused" {
		return ""
	}
	list := []string{
		boolToString(w.Bit.Get(nil)),
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

type JointWire struct {
	Res   *Wire
	A     *Wire
	B     *Wire
	IsAnd bool
}

func (w JointWire) String(depth int) string {
	var res []string
	for _, wire := range []*Wire{w.A, w.B, w.Res} {
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
	return fmt.Sprintf("%s%s %s", depthToString(depth), name, strings.Join(res, "    "))
}

func color(a, b *Wire) string {
	color := "blue"
	if a.Bit.Get(nil) || b.Bit.Get(nil) {
		color = "red"
	}
	return fmt.Sprintf(`[color="%s"]`, color)
}

func (w JointWire) Graph(depth int) string {
	if !*drawNodes {
		return ""
	}
	prefix := depthToString(depth)
	var res []string
	if *drawShapePoint {
		res = append(res, fmt.Sprintf(`%s"%v" [label= "";shape=point];`, prefix, *w.Res))
	}
	for _, wire := range []*Wire{w.A, w.B} {
		if *drawShapePoint {
			res = append(res, fmt.Sprintf(`%s"%v" [label= "";shape=point];`, prefix, *wire))
		}
		if *drawEdges {
			res = append(res, fmt.Sprintf(`%s"%v" -> "%v" %s;`, prefix, *wire, *w.Res, color(wire, w.Res)))
		}
	}
	return strings.Join(res, "\n")
}

func (w *JointWire) Update() {
	if w.IsAnd {
		w.Res.Bit.Set(w.A.Bit.Get(w) && w.B.Bit.Get(w))
		return
	}
	w.Res.Bit.Set(w.A.Bit.Get(w) || w.B.Bit.Get(w))
}

type Transistor struct {
	Base         *Wire
	Collector    *Wire
	Emitter      *Wire
	CollectorOut *Wire
}

func (t Transistor) String(depth int) string {
	var res []string
	for _, wire := range []*Wire{t.Base, t.Collector, t.Emitter, t.CollectorOut} {
		one := fmt.Sprintf("%v", *wire)
		if one == "" {
			continue
		}
		res = append(res, one)
	}
	return fmt.Sprintf("%s%s", depthToString(depth), strings.Join(res, "    "))
}

func (t Transistor) Graph(depth int) string {
	if !*drawNodes {
		return ""
	}
	prefix := depthToString(depth)
	var res []string
	res = append(res, fmt.Sprintf(`"%p" [label="𓇲";shape=invtriangle];`, &t))
	for _, wire := range []*Wire{t.Base, t.Collector} {
		if *drawShapePoint {
			res = append(res, fmt.Sprintf(`%s"%v" [label= "";shape=point];`, prefix, *wire))
		}
		if *drawEdges {
			res = append(res, fmt.Sprintf(`%s"%v" -> "%p" %s;`, prefix, *wire, &t, color(wire, wire)))
		}
	}
	for _, wire := range []*Wire{t.Emitter, t.CollectorOut} {
		if wire.Name == "Unused" {
			continue
		}
		if *drawShapePoint {
			res = append(res, fmt.Sprintf(`%s"%v" [label= "";shape=point];`, prefix, *wire))
		}
		if *drawEdges {
			res = append(res, fmt.Sprintf(`%s"%p" -> "%v" %s;`, prefix, &t, *wire, color(wire, wire)))
		}
	}
	return strings.Join(res, "\n")
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

type Group struct {
	Name       string
	Vcc        *Wire
	Gnd        *Wire
	Unused     *Wire
	Components []Component
}

func (g *Group) Group(name string) *Group {
	res := &Group{Name: name, Vcc: g.Vcc, Gnd: g.Gnd, Unused: g.Unused}
	g.Components = append(g.Components, res)
	return res
}

var (
	horizontalLine = strings.Repeat("-", 10)
)

func (g Group) String(depth int) string {
	if *maxPrintDepth >= 0 && depth >= *maxPrintDepth {
		return ""
	}
	prefix := depthToString(depth)
	res := []string{
		prefix + g.Name,
		prefix + horizontalLine,
	}
	for _, component := range g.Components {
		one := component.String(depth + 1)
		if one == "" {
			continue
		}
		res = append(res, one)
	}
	res = append(res, prefix+horizontalLine)
	return strings.Join(res, "\n")
}

func (g Group) Graph(depth int) string {
	if *maxPrintDepth >= 0 && depth >= *maxPrintDepth {
		return ""
	}
	prefix := depthToString(depth)
	nextPrefix := depthToString(depth + 1)
	res := []string{
		fmt.Sprintf("%ssubgraph cluster_%p {", prefix, &g),
		fmt.Sprintf(`%slabel="%s";`, nextPrefix, g.Name),
		fmt.Sprintf(`%sgraph[style=dotted];`, nextPrefix),
		fmt.Sprintf(`%s"%p"[style=invis,shape=point];`, nextPrefix, &g),
	}
	for _, component := range g.Components {
		one := component.Graph(depth + 1)
		if one == "" {
			continue
		}
		res = append(res, one)
	}
	res = append(res, fmt.Sprintf("%s}", prefix))
	return strings.Join(res, "\n")
}

func (g *Group) JointWire(res, a, b *Wire, isAnd bool) {
	g.Components = append(g.Components, &JointWire{Res: res, A: a, B: b, IsAnd: isAnd})
}

func (g *Group) AddTransistor(transistor *Transistor) {
	if transistor.CollectorOut == nil {
		transistor.CollectorOut = g.Unused
	}
	g.Components = append(g.Components, transistor)
}

func (g *Group) Transistor(base, collector, emitter, collectorOut *Wire) {
	g.AddTransistor(&Transistor{Base: base, Collector: collector, Emitter: emitter, CollectorOut: collectorOut})
}

func (g *Group) AddTransistors(transistors []*Transistor) {
	for _, transistor := range transistors {
		g.AddTransistor(transistor)
	}
}

func (g *Group) Update() {
	for _, component := range g.Components {
		component.Update()
	}
}

type Circuit struct {
	Vcc        *Wire
	Gnd        *Wire
	Unused     *Wire
	Inputs     []*Wire
	Outputs    []*Wire
	Components []Component
}

func NewCircuit() *Circuit {
	return &Circuit{Vcc: &Wire{Name: "Vcc", Bit: Bool{bit: true}}, Gnd: &Wire{Name: "Gnd", Gnd: Bool{bit: true}}, Unused: &Wire{Name: "Unused"}}
}

func (c *Circuit) In(name string) *Wire {
	res := &Wire{Name: name}
	c.Inputs = append(c.Inputs, res)
	return res
}

func (c *Circuit) Group(name string) *Group {
	res := &Group{Name: name, Vcc: c.Vcc, Gnd: c.Gnd, Unused: c.Unused}
	c.Components = append(c.Components, res)
	return res
}

func (c *Circuit) Out(res *Wire) {
	c.Outputs = append(c.Outputs, res)
}

func (c *Circuit) Outs(outputs []*Wire) {
	c.Outputs = append(c.Outputs, outputs...)
}

func (c *Circuit) Update() {
	for _, component := range c.Components {
		component.Update()
	}
}

func (c Circuit) String() string {
	var res []string
	var list []string
	for _, input := range c.Inputs {
		list = append(list, fmt.Sprintf("  %v", *input))
	}
	res = append(res, fmt.Sprintf("Inputs: %s", strings.Join(list, "")))
	res = append(res, fmt.Sprintf("Outputs:"))
	for _, output := range c.Outputs {
		res = append(res, fmt.Sprintf("  %v", *output))
	}
	res = append(res, "Components: ")
	for _, component := range c.Components {
		res = append(res, component.String(0))
	}
	return strings.Join(res, "\n")
}

func (c Circuit) Graph() string {
	res := []string{
		"digraph {",
		" rankdir=LR;",
	}
	for _, input := range c.Inputs {
		res = append(res, fmt.Sprintf(` "%v"[shape=rarrow;fillcolor=black;style=filled;fontcolor=white;fontsize=30];`, *input))
	}
	for _, output := range c.Outputs {
		res = append(res, fmt.Sprintf(` "%v"[shape=rarrow;fillcolor=black;style=filled;fontcolor=white;fontsize=30];`, *output))
	}
	for _, component := range c.Components {
		res = append(res, component.Graph(1))
	}
	res = append(res, "}")
	return strings.Join(res, "\n")
}

func (c *Circuit) Simulate() []string {
	if !*drawSingleGraph && len(c.Inputs) <= 9 {
		return c.simulate(0)
	}
	rand := rand.New(rand.NewPCG(42, 1024))
	for _, input := range c.Inputs {
		input.Bit.bit = rand.IntN(2) == 1
	}
	return c.simulate(len(c.Inputs))
}

func (c *Circuit) simulate(index int) []string {
	if index >= len(c.Inputs) {
		c.Update()
		if *drawGraph {
			return []string{c.Graph()}
		}
		return []string{c.String()}
	}
	var res []string
	for _, value := range []bool{false, true} {
		c.Inputs[index].Bit.bit = value
		res = append(res, c.simulate(index+1)...)
	}
	return res
}

func transistor(parent *Group, base, collector *Wire) []*Wire {
	group := parent.Group("transistor")
	emitter := &Wire{Name: "emitter"}
	collectorOut := &Wire{Name: "collector_out"}
	group.Transistor(base, collector, emitter, collectorOut)
	return []*Wire{emitter, collectorOut}
}

func transistorGnd(parent *Group, base, collector *Wire) []*Wire {
	group := parent.Group("transistorGnd")
	collectorOut := &Wire{Name: "collector_out"}
	group.Transistor(base, collector, group.Gnd, collectorOut)
	return []*Wire{collectorOut}
}

func Not(parent *Group, a *Wire) *Wire {
	group := parent.Group(fmt.Sprintf("NOT(%v)", a.Name))
	res := &Wire{Name: group.Name}
	group.Transistor(a, group.Vcc, group.Gnd, res)
	return res
}

func And(parent *Group, a, b *Wire) *Wire {
	group := parent.Group(fmt.Sprintf("AND(%v,%v)", a.Name, b.Name))
	res := &Wire{Name: group.Name}
	wire := &Wire{Name: fmt.Sprintf("%s-wire", res.Name)}
	group.AddTransistors([]*Transistor{
		{Base: a, Collector: group.Vcc, Emitter: wire},
		{Base: b, Collector: wire, Emitter: res},
	})
	return res
}

func Or(parent *Group, a, b *Wire) *Wire {
	res := &Wire{}
	return OrRes(parent, res, a, b)
}

func OrRes(parent *Group, res, a, b *Wire) *Wire {
	group := parent.Group(fmt.Sprintf("OR(%v,%v)", a.Name, b.Name))
	res.Name = group.Name
	wire1 := &Wire{Name: fmt.Sprintf("%s-wire1", res.Name)}
	wire2 := &Wire{Name: fmt.Sprintf("%s-wire2", res.Name)}
	group.AddTransistors([]*Transistor{
		{Base: a, Collector: group.Vcc, Emitter: wire1},
		{Base: b, Collector: group.Vcc, Emitter: wire2},
	})
	group.JointWire(res, wire1, wire2, false /* isAnd */)
	return res
}

func Nand(parent *Group, a, b *Wire) *Wire {
	group := parent.Group(fmt.Sprintf("NAND(%v,%v)", a.Name, b.Name))
	res := &Wire{Name: group.Name}
	wire := &Wire{Name: fmt.Sprintf("%s-wire", res.Name)}
	group.AddTransistors([]*Transistor{
		{Base: a, Collector: group.Vcc, Emitter: wire, CollectorOut: res},
		{Base: b, Collector: wire, Emitter: group.Gnd},
	})
	return res
}

func Xor(parent *Group, a, b *Wire) *Wire {
	group := parent.Group(fmt.Sprintf("XOR(%v,%v)", a.Name, b.Name))
	res := And(group, Or(group, a, b), Nand(group, a, b))
	res.Name = group.Name
	return res
}

func Nor(parent *Group, a, b *Wire) *Wire {
	res := &Wire{}
	return NorRes(parent, res, a, b)
}

func NorRes(parent *Group, res, a, b *Wire) *Wire {
	group := parent.Group(fmt.Sprintf("NOR(%v,%v)", a.Name, b.Name))
	res.Name = group.Name
	wire1 := &Wire{Name: fmt.Sprintf("%s-wire1", res.Name)}
	wire2 := &Wire{Name: fmt.Sprintf("%s-wire2", res.Name)}
	group.AddTransistors([]*Transistor{
		{Base: a, Collector: group.Vcc, Emitter: group.Gnd, CollectorOut: wire1},
		{Base: b, Collector: group.Vcc, Emitter: group.Gnd, CollectorOut: wire2},
	})
	group.JointWire(res, wire1, wire2, true /* IsAnd */)
	return res
}

func HalfSum(parent *Group, a, b *Wire) []*Wire {
	group := parent.Group(fmt.Sprintf("SUM(%v,%v)", a.Name, b.Name))
	res := Xor(group, a, b)
	res.Name = group.Name
	carry := And(group, a, b)
	carry.Name = fmt.Sprintf("CARRY(%v,%v)", a.Name, b.Name)
	return []*Wire{res, carry}
}

func Sum(parent *Group, a, b, cin *Wire) []*Wire {
	group := parent.Group(fmt.Sprintf("SUM(%v,%v,%v)", a.Name, b.Name, cin.Name))
	s1 := HalfSum(group, a, b)
	s2 := HalfSum(group, s1[0], cin)
	s2[0].Name = group.Name
	carry := Or(group, s1[1], s2[1])
	carry.Name = fmt.Sprintf("CARRY(%v,%v)", a.Name, b.Name)
	return []*Wire{s2[0], carry}
}

func Sum2(parent *Group, a1, a2, b1, b2, cin *Wire) []*Wire {
	group := parent.Group("SUM2")
	s1 := Sum(group, a1, b1, cin)
	s2 := Sum(group, a2, b2, s1[1])
	return []*Wire{s1[0], s2[0], s2[1]}
}

func Sum4(parent *Group, a1, a2, a3, a4, b1, b2, b3, b4, cin *Wire) []*Wire {
	group := parent.Group("SUM4")
	s1 := Sum2(group, a1, a2, b1, b2, cin)
	s2 := Sum2(group, a3, a4, b3, b4, s1[2])
	return []*Wire{s1[0], s1[1], s2[0], s2[1], s2[2]}
}

func Sum8(parent *Group, a1, a2, a3, a4, a5, a6, a7, a8, b1, b2, b3, b4, b5, b6, b7, b8, cin *Wire) []*Wire {
	group := parent.Group("SUM8")
	s1 := Sum4(group, a1, a2, a3, a4, b1, b2, b3, b4, cin)
	s2 := Sum4(group, a5, a6, a7, a8, b5, b6, b7, b8, s1[4])
	return []*Wire{s1[0], s1[1], s1[2], s1[3], s2[0], s2[1], s2[2], s2[3], s2[4]}
}

func SRLatch(parent *Group, s, r *Wire) []*Wire {
	q := &Wire{Name: "q"}
	return SRLatchRes(parent, q, s, r)
}

func SRLatchRes(parent *Group, q, s, r *Wire) []*Wire {
	group := parent.Group(fmt.Sprintf("SRLATCH(%v,%v)", s.Name, r.Name))
	nq := &Wire{Name: "nq"}
	NorRes(group, q, r, nq)
	NorRes(group, nq, s, q)
	return []*Wire{q, nq}
}

func SRLatchWithEnable(parent *Group, s, r, e *Wire) []*Wire {
	q := &Wire{Name: "q"}
	return SRLatchResWithEnable(parent, q, s, r, e)
}

func SRLatchResWithEnable(parent *Group, q, s, r, e *Wire) []*Wire {
	group := parent.Group(fmt.Sprintf("SRLATCHEN(%v,%v,%v)", s.Name, r.Name, e.Name))
	return SRLatchRes(group, q, And(group, s, e), And(group, r, e))
}

func DLatch(parent *Group, d, e *Wire) []*Wire {
	q := &Wire{Name: "q"}
	return DLatchRes(parent, q, d, e)
}

func DLatchRes(parent *Group, q, d, e *Wire) []*Wire {
	group := parent.Group(fmt.Sprintf("DLATCH(%v,%v)", d.Name, e.Name))
	return SRLatchResWithEnable(group, q, d, Not(group, d), e)
}

func Register(parent *Group, d, ei, eo *Wire) []*Wire {
	group := parent.Group(fmt.Sprintf("Register(%v,%v,%v)", d.Name, ei.Name, eo.Name))
	q := &Wire{}
	DLatchRes(group, q, Or(group, And(group, q, Not(group, ei)), And(group, d, ei)), ei)
	q.Name = group.Name + "-internal"
	res := And(group, q, eo)
	res.Name = group.Name
	return []*Wire{q, res}
}

func Register2(parent *Group, d1, d2, ei, eo *Wire) []*Wire {
	group := parent.Group("Register2")
	r1 := Register(group, d1, ei, eo)
	r2 := Register(group, d2, ei, eo)
	return append(r1, r2...)
}

func Register4(parent *Group, d1, d2, d3, d4, ei, eo *Wire) []*Wire {
	group := parent.Group("Register4")
	r1 := Register2(group, d1, d2, ei, eo)
	r2 := Register2(group, d3, d4, ei, eo)
	return append(r1, r2...)
}

func Register8(parent *Group, d1, d2, d3, d4, d5, d6, d7, d8, ei, eo *Wire) []*Wire {
	group := parent.Group("Register8")
	r1 := Register4(group, d1, d2, d3, d4, ei, eo)
	r2 := Register4(group, d5, d6, d7, d8, ei, eo)
	return append(r1, r2...)
}

func Alu(parent *Group, a, ai, ao, b, bi, bo, ri, ro, carry *Wire) []*Wire {
	group := parent.Group("ALU")
	ra := Register(group, a, ai, ao)
	rb := Register(group, b, bi, bo)
	rs := Sum(group, ra[0], rb[0], carry)
	rr := Register(group, rs[0], ri, ro)
	return slices.Concat(ra, rb, rr, []*Wire{rs[1]})
}

func Alu2(parent *Group, a1, a2, ai, ao, b1, b2, bi, bo, ri, ro, carry *Wire) []*Wire {
	group := parent.Group("ALU2")
	r1 := Alu(group, a1, ai, ao, b1, bi, bo, ri, ro, carry)
	r2 := Alu(group, a2, ai, ao, b2, bi, bo, ri, ro, r1[6])
	return append(r1[:6], r2...)
}

func all() {
	c := NewCircuit()
	g := c.Group("")

	c.Outs(transistor(g, c.In("base"), c.In("collector")))

	c.Out(Not(g, c.In("a")))
	c.Out(And(g, c.In("a"), c.In("b")))
	c.Out(Or(g, c.In("a"), c.In("b")))
	c.Out(Nand(g, c.In("a"), c.In("b")))
	c.Out(Nand(g, c.In("a"), Nand(g, c.In("b"), c.In("c"))))
	c.Out(Xor(g, c.In("a"), c.In("b")))
	c.Outs(HalfSum(g, c.In("a"), c.In("b")))
	c.Outs(Sum(g, c.In("a"), c.In("b"), c.In("c")))
	c.Outs(Sum2(g, c.In("a1"), c.In("a2"), c.In("b1"), c.In("b2"), c.In("c")))
	c.Outs(Sum4(g, c.In("a1"), c.In("a2"), c.In("a3"), c.In("a4"), c.In("b1"), c.In("b2"), c.In("b3"), c.In("b4"), c.In("c")))
	c.Outs(Sum8(g, c.In("a1"), c.In("a2"), c.In("a3"), c.In("a4"), c.In("a5"), c.In("a6"), c.In("a7"), c.In("a8"), c.In("b1"), c.In("b2"), c.In("b3"), c.In("b4"), c.In("b5"), c.In("b6"), c.In("b7"), c.In("b8"), c.In("c")))
	bOrRes := &Wire{Name: "b"}
	c.Out(OrRes(g, bOrRes, c.In("a"), bOrRes))
	c.Out(Nor(g, c.In("a"), c.In("b")))
	c.Outs(SRLatch(g, c.In("s"), c.In("r")))
	c.Outs(SRLatchWithEnable(g, c.In("s"), c.In("r"), c.In("e")))
	c.Outs(DLatch(g, c.In("d"), c.In("e")))
	c.Outs(Register(g, c.In("d"), c.In("ei"), c.In("eo")))
	c.Outs(Register2(g, c.In("d1"), c.In("d2"), c.In("ei"), c.In("eo")))
	c.Outs(Register4(g, c.In("d1"), c.In("d2"), c.In("d3"), c.In("d4"), c.In("ei"), c.In("eo")))
	c.Outs(Register8(g, c.In("d1"), c.In("d2"), c.In("d3"), c.In("d4"), c.In("d5"), c.In("d6"), c.In("d7"), c.In("d8"), c.In("ei"), c.In("eo")))
	c.Outs(Alu(g, c.In("a"), c.In("ai"), c.In("ao"), c.In("b"), c.In("bi"), c.In("bo"), c.In("ri"), c.In("ro"), c.In("carry")))
	c.Outs(Alu2(g, c.In("a1"), c.In("a2"), c.In("ai"), c.In("ao"), c.In("b1"), c.In("b2"), c.In("bi"), c.In("bo"), c.In("ri"), c.In("ro"), c.In("carry")))
}

func main() {
	flag.Parse()
	c := NewCircuit()
	g := c.Group("")

	c.Outs(transistorGnd(g, c.In("base"), c.In("collector")))

	res := c.Simulate()
	fmt.Println(strings.Join(res, "\n"))
	// fmt.Println(strings.Join(c.Simulate(), "\n"))

	if !*drawGraph {
		return
	}
	for i, graph := range res {
		if i >= 4 || (*drawSingleGraph && i >= 1) {
			break
		}
		if err := os.WriteFile(fmt.Sprintf("%d.dot", i), []byte(graph), 0644); err != nil {
			panic(fmt.Errorf("WriteFile got err %v", err))
		}
	}
	// $ for file in *.dot; do dot -Tsvg "${file}" > "${file}".svg; done
	// $ google-chrome *.svg
}
