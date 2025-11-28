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

func Register8(parent *group.Group, d1, d2, d3, d4, d5, d6, d7, d8, ei, eo *wire.Wire) []*wire.Wire {
	group := parent.Group("Register8")
	r1 := lib.Register4(group, d1, d2, d3, d4, ei, eo)
	r2 := lib.Register4(group, d5, d6, d7, d8, ei, eo)
	return append(r1, r2...)
}

func Alu(parent *group.Group, a, ai, ao, b, bi, bo, ri, ro, carry *wire.Wire) []*wire.Wire {
	group := parent.Group("ALU")
	ra := lib.Register(group, a, ai, ao)
	rb := lib.Register(group, b, bi, bo)
	rs := lib.Sum(group, ra[0], rb[0], carry)
	rr := lib.Register(group, rs[0], ri, ro)
	return slices.Concat(ra, rb, rr, []*wire.Wire{rs[1]})
}

func Alu2(parent *group.Group, a1, a2, ai, ao, b1, b2, bi, bo, ri, ro, carry *wire.Wire) []*wire.Wire {
	group := parent.Group("ALU2")
	r1 := Alu(group, a1, ai, ao, b1, bi, bo, ri, ro, carry)
	r2 := Alu(group, a2, ai, ao, b2, bi, bo, ri, ro, r1[6])
	return append(r1[:6], r2...)
}

func examples(c *circuit.Circuit, g *group.Group) {
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
