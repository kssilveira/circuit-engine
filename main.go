package main

import (
	"flag"
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/kssilveira/circuit-engine/circuit"
	"github.com/kssilveira/circuit-engine/config"
	"github.com/kssilveira/circuit-engine/group"
	"github.com/kssilveira/circuit-engine/lib"
	"github.com/kssilveira/circuit-engine/wire"
)

func HalfSum(parent *group.Group, a, b *wire.Wire) []*wire.Wire {
	group := parent.Group(fmt.Sprintf("SUM(%v,%v)", a.Name, b.Name))
	res := lib.Xor(group, a, b)
	res.Name = group.Name
	carry := lib.And(group, a, b)
	carry.Name = fmt.Sprintf("CARRY(%v,%v)", a.Name, b.Name)
	return []*wire.Wire{res, carry}
}

func Sum(parent *group.Group, a, b, cin *wire.Wire) []*wire.Wire {
	group := parent.Group(fmt.Sprintf("SUM(%v,%v,%v)", a.Name, b.Name, cin.Name))
	s1 := HalfSum(group, a, b)
	s2 := HalfSum(group, s1[0], cin)
	s2[0].Name = group.Name
	carry := lib.Or(group, s1[1], s2[1])
	carry.Name = fmt.Sprintf("CARRY(%v,%v)", a.Name, b.Name)
	return []*wire.Wire{s2[0], carry}
}

func Sum2(parent *group.Group, a1, a2, b1, b2, cin *wire.Wire) []*wire.Wire {
	group := parent.Group("SUM2")
	s1 := Sum(group, a1, b1, cin)
	s2 := Sum(group, a2, b2, s1[1])
	return []*wire.Wire{s1[0], s2[0], s2[1]}
}

func Sum4(parent *group.Group, a1, a2, a3, a4, b1, b2, b3, b4, cin *wire.Wire) []*wire.Wire {
	group := parent.Group("SUM4")
	s1 := Sum2(group, a1, a2, b1, b2, cin)
	s2 := Sum2(group, a3, a4, b3, b4, s1[2])
	return []*wire.Wire{s1[0], s1[1], s2[0], s2[1], s2[2]}
}

func Sum8(parent *group.Group, a1, a2, a3, a4, a5, a6, a7, a8, b1, b2, b3, b4, b5, b6, b7, b8, cin *wire.Wire) []*wire.Wire {
	group := parent.Group("SUM8")
	s1 := Sum4(group, a1, a2, a3, a4, b1, b2, b3, b4, cin)
	s2 := Sum4(group, a5, a6, a7, a8, b5, b6, b7, b8, s1[4])
	return []*wire.Wire{s1[0], s1[1], s1[2], s1[3], s2[0], s2[1], s2[2], s2[3], s2[4]}
}

func SRLatch(parent *group.Group, s, r *wire.Wire) []*wire.Wire {
	q := &wire.Wire{Name: "q"}
	return SRLatchRes(parent, q, s, r)
}

func SRLatchRes(parent *group.Group, q, s, r *wire.Wire) []*wire.Wire {
	group := parent.Group(fmt.Sprintf("SRLATCH(%v,%v)", s.Name, r.Name))
	nq := &wire.Wire{Name: "nq"}
	lib.NorRes(group, q, r, nq)
	lib.NorRes(group, nq, s, q)
	return []*wire.Wire{q, nq}
}

func SRLatchWithEnable(parent *group.Group, s, r, e *wire.Wire) []*wire.Wire {
	q := &wire.Wire{Name: "q"}
	return SRLatchResWithEnable(parent, q, s, r, e)
}

func SRLatchResWithEnable(parent *group.Group, q, s, r, e *wire.Wire) []*wire.Wire {
	group := parent.Group(fmt.Sprintf("SRLATCHEN(%v,%v,%v)", s.Name, r.Name, e.Name))
	return SRLatchRes(group, q, lib.And(group, s, e), lib.And(group, r, e))
}

func DLatch(parent *group.Group, d, e *wire.Wire) []*wire.Wire {
	q := &wire.Wire{Name: "q"}
	return DLatchRes(parent, q, d, e)
}

func DLatchRes(parent *group.Group, q, d, e *wire.Wire) []*wire.Wire {
	group := parent.Group(fmt.Sprintf("DLATCH(%v,%v)", d.Name, e.Name))
	return SRLatchResWithEnable(group, q, d, lib.Not(group, d), e)
}

func Register(parent *group.Group, d, ei, eo *wire.Wire) []*wire.Wire {
	group := parent.Group(fmt.Sprintf("Register(%v,%v,%v)", d.Name, ei.Name, eo.Name))
	q := &wire.Wire{}
	DLatchRes(group, q, lib.Or(group, lib.And(group, q, lib.Not(group, ei)), lib.And(group, d, ei)), ei)
	q.Name = group.Name + "-internal"
	res := lib.And(group, q, eo)
	res.Name = group.Name
	return []*wire.Wire{q, res}
}

func Register2(parent *group.Group, d1, d2, ei, eo *wire.Wire) []*wire.Wire {
	group := parent.Group("Register2")
	r1 := Register(group, d1, ei, eo)
	r2 := Register(group, d2, ei, eo)
	return append(r1, r2...)
}

func Register4(parent *group.Group, d1, d2, d3, d4, ei, eo *wire.Wire) []*wire.Wire {
	group := parent.Group("Register4")
	r1 := Register2(group, d1, d2, ei, eo)
	r2 := Register2(group, d3, d4, ei, eo)
	return append(r1, r2...)
}

func Register8(parent *group.Group, d1, d2, d3, d4, d5, d6, d7, d8, ei, eo *wire.Wire) []*wire.Wire {
	group := parent.Group("Register8")
	r1 := Register4(group, d1, d2, d3, d4, ei, eo)
	r2 := Register4(group, d5, d6, d7, d8, ei, eo)
	return append(r1, r2...)
}

func Alu(parent *group.Group, a, ai, ao, b, bi, bo, ri, ro, carry *wire.Wire) []*wire.Wire {
	group := parent.Group("ALU")
	ra := Register(group, a, ai, ao)
	rb := Register(group, b, bi, bo)
	rs := Sum(group, ra[0], rb[0], carry)
	rr := Register(group, rs[0], ri, ro)
	return slices.Concat(ra, rb, rr, []*wire.Wire{rs[1]})
}

func Alu2(parent *group.Group, a1, a2, ai, ao, b1, b2, bi, bo, ri, ro, carry *wire.Wire) []*wire.Wire {
	group := parent.Group("ALU2")
	r1 := Alu(group, a1, ai, ao, b1, bi, bo, ri, ro, carry)
	r2 := Alu(group, a2, ai, ao, b2, bi, bo, ri, ro, r1[6])
	return append(r1[:6], r2...)
}

func examples(c *circuit.Circuit, g *group.Group) {
	c.Outs(HalfSum(g, c.In("a"), c.In("b")))
	c.Outs(Sum(g, c.In("a"), c.In("b"), c.In("c")))
	c.Outs(Sum2(g, c.In("a1"), c.In("a2"), c.In("b1"), c.In("b2"), c.In("c")))
	c.Outs(Sum4(g, c.In("a1"), c.In("a2"), c.In("a3"), c.In("a4"), c.In("b1"), c.In("b2"), c.In("b3"), c.In("b4"), c.In("c")))
	c.Outs(Sum8(g, c.In("a1"), c.In("a2"), c.In("a3"), c.In("a4"), c.In("a5"), c.In("a6"), c.In("a7"), c.In("a8"), c.In("b1"), c.In("b2"), c.In("b3"), c.In("b4"), c.In("b5"), c.In("b6"), c.In("b7"), c.In("b8"), c.In("c")))
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

func all() error {
	maxPrintDepth := flag.Int("max_print_depth", -1, "max print depth")
	drawGraph := flag.Bool("draw_graph", false, "draw graph")
	drawSingleGraph := flag.Bool("draw_single_graph", false, "draw single graph")
	drawNodes := flag.Bool("draw_nodes", true, "draw nodes")
	drawEdges := flag.Bool("draw_edges", true, "draw edges")
	drawShapePoint := flag.Bool("draw_shape_point", false, "draw shape point")
	isUnitTest := flag.Bool("is_unit_test", false, "is unit test")
	exampleName := flag.String("example_name", "TransistorEmitter", "example name")

	flag.Parse()

	c := circuit.NewCircuit(config.Config{
		MaxPrintDepth:   *maxPrintDepth,
		DrawGraph:       *drawGraph,
		DrawSingleGraph: *drawSingleGraph,
		DrawNodes:       *drawNodes,
		DrawEdges:       *drawEdges,
		DrawShapePoint:  *drawShapePoint,
		IsUnitTest:      *isUnitTest,
	})

	outs := lib.Example(c, *exampleName)
	if len(outs) == 0 {
		return fmt.Errorf("invalid --example_name %q, valid names are %q", *exampleName, lib.ExampleNames())
	}
	c.Outs(outs)

	res := c.Simulate()
	fmt.Println(strings.Join(res, "\n\n"))
	// fmt.Println(strings.Join(c.Simulate(), "\n"))

	if !*drawGraph {
		return nil
	}
	for i, graph := range res {
		if i >= 4 || (*drawSingleGraph && i >= 1) {
			break
		}
		if err := os.WriteFile(fmt.Sprintf("%d.dot", i), []byte(graph), 0644); err != nil {
			return fmt.Errorf("WriteFile got err %v", err)
		}
	}
	// $ for file in *.dot; do dot -Tsvg "${file}" > "${file}".svg; done
	// $ google-chrome *.svg
	return nil
}

func main() {
	if err := all(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
