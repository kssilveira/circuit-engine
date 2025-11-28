package circuit

import (
	"fmt"
	"math/rand/v2"
	"strings"

	"github.com/kssilveira/circuit-engine/bit"
	"github.com/kssilveira/circuit-engine/component"
	"github.com/kssilveira/circuit-engine/config"
	"github.com/kssilveira/circuit-engine/group"
	"github.com/kssilveira/circuit-engine/wire"
)

type Circuit struct {
	Config     config.Config
	Vcc        *wire.Wire
	Gnd        *wire.Wire
	Unused     *wire.Wire
	Inputs     []*wire.Wire
	Outputs    []*wire.Wire
	Components []component.Component
}

func NewCircuit(config config.Config) *Circuit {
	vcc := bit.Bit{}
	vcc.Set(true)
	gnd := bit.Bit{}
	gnd.Set(true)
	return &Circuit{Config: config, Vcc: &wire.Wire{Name: "Vcc", Bit: vcc}, Gnd: &wire.Wire{Name: "Gnd", Gnd: gnd}, Unused: &wire.Wire{Name: "Unused"}}
}

func (c *Circuit) In(name string) *wire.Wire {
	res := &wire.Wire{Name: name}
	c.Inputs = append(c.Inputs, res)
	return res
}

func (c *Circuit) Group(name string) *group.Group {
	res := &group.Group{Name: name, Vcc: c.Vcc, Gnd: c.Gnd, Unused: c.Unused}
	c.Components = append(c.Components, res)
	return res
}

func (c *Circuit) Out(res *wire.Wire) {
	c.Outputs = append(c.Outputs, res)
}

func (c *Circuit) Outs(outputs []*wire.Wire) {
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
		res = append(res, component.String(0, c.Config))
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
		res = append(res, component.Graph(1, c.Config))
	}
	res = append(res, "}")
	return strings.Join(res, "\n")
}

func (c *Circuit) Simulate() []string {
	if !c.Config.DrawSingleGraph && len(c.Inputs) <= 9 {
		return c.simulate(0)
	}
	rand := rand.New(rand.NewPCG(42, 1024))
	for _, input := range c.Inputs {
		input.Bit.SilentSet(rand.IntN(2) == 1)
	}
	return c.simulate(len(c.Inputs))
}

func (c *Circuit) simulate(index int) []string {
	if index >= len(c.Inputs) {
		c.Update()
		if c.Config.DrawGraph {
			return []string{c.Graph()}
		}
		return []string{c.String()}
	}
	var res []string
	for _, value := range []bool{false, true} {
		c.Inputs[index].Bit.SilentSet(value)
		res = append(res, c.simulate(index+1)...)
	}
	return res
}
