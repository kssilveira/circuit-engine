package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/kssilveira/circuit-engine/circuit"
	"github.com/kssilveira/circuit-engine/config"
	"github.com/kssilveira/circuit-engine/lib"
)

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
	return nil
}

func main() {
	if err := all(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
