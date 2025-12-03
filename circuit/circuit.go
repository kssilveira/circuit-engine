// Package circuit encapsulates circuits.
package circuit

import (
	"math/rand/v2"
	"strings"

	"github.com/kssilveira/circuit-engine/component"
	"github.com/kssilveira/circuit-engine/config"
	"github.com/kssilveira/circuit-engine/group"
	"github.com/kssilveira/circuit-engine/sfmt"
	"github.com/kssilveira/circuit-engine/wire"
)

// Circuit contains a single circuit.
type Circuit struct {
	Config           config.Config
	Inputs           []*wire.Wire
	Outputs          []*wire.Wire
	Components       []component.Component
	InputValidations []func() bool
}

// NewCircuit creates a new circuit.
func NewCircuit(config config.Config) *Circuit {
	return &Circuit{Config: config}
}

// In adds an input.
func (c *Circuit) In(name string) *wire.Wire {
	res := &wire.Wire{Name: name}
	c.Inputs = append(c.Inputs, res)
	return res
}

// Group adds a group.
func (c *Circuit) Group(name string) *group.Group {
	res := &group.Group{Name: name}
	c.Components = append(c.Components, res)
	return res
}

// Out adds an output.
func (c *Circuit) Out(res *wire.Wire) {
	c.Outputs = append(c.Outputs, res)
}

// Outs adds multiple outputs.
func (c *Circuit) Outs(outputs []*wire.Wire) {
	c.Outputs = append(c.Outputs, outputs...)
}

// Update updates the components.
func (c *Circuit) Update() {
	for _, component := range c.Components {
		component.Update(false /* updateReaders */)
	}
	for _, component := range c.Components {
		component.Update(true /* updateReaders */)
	}
	// fmt.Println(c.StringForUnitTest())
}

// AddInputValidation adds input validation.
func (c *Circuit) AddInputValidation(fn func() bool) {
	c.InputValidations = append(c.InputValidations, fn)
}

// Description returns the circuit description.
func (c *Circuit) Description() string {
	var res []string
	for _, input := range c.Inputs {
		res = append(res, input.Name)
	}
	res = append(res, "=>")
	for _, output := range c.Outputs {
		res = append(res, output.Name)
	}
	return strings.Join(res, " ")
}

func (c Circuit) String() string {
	var res []string
	var list []string
	for _, input := range c.Inputs {
		list = append(list, sfmt.Sprintf("  %v", *input))
	}
	res = append(res, sfmt.Sprintf("Inputs: %s", strings.Join(list, "")))
	res = append(res, sfmt.Sprintf("Outputs:"))
	for _, output := range c.Outputs {
		res = append(res, sfmt.Sprintf("  %v", *output))
	}
	res = append(res, "Components: ")
	for _, component := range c.Components {
		res = append(res, component.String(0, c.Config))
	}
	return strings.Join(res, "\n")
}

// StringForUnitTest returns the circuit string for unit tests.
func (c Circuit) StringForUnitTest() string {
	var res []string
	for _, input := range c.Inputs {
		res = append(res, wire.BoolToString(input.Bit.Get(nil)))
	}
	res = append(res, "=>")
	for _, output := range c.Outputs {
		res = append(res, wire.BoolToString(output.Bit.Get(nil)))
	}
	return strings.Join(res, "")
}

// Graph returns the graphviz graph.
func (c Circuit) Graph() string {
	res := []string{
		"digraph {",
		" rankdir=LR;",
	}
	for _, input := range c.Inputs {
		res = append(res, sfmt.Sprintf(` "%v"[shape=rarrow;fillcolor=black;style=filled;fontcolor=white;fontsize=30];`, *input))
	}
	for _, output := range c.Outputs {
		res = append(res, sfmt.Sprintf(` "%v"[shape=rarrow;fillcolor=black;style=filled;fontcolor=white;fontsize=30];`, *output))
	}
	for _, component := range c.Components {
		res = append(res, component.Graph(1, c.Config))
	}
	res = append(res, "}")
	return strings.Join(res, "\n")
}

// Simulate simulates the circuit.
func (c *Circuit) Simulate() []string {
	if len(c.Config.SimulateInputs) > 0 {
		return c.SimulateInputs(c.Config.SimulateInputs)
	}
	if !c.Config.DrawSingleGraph && len(c.Inputs) <= 7 {
		return c.simulate(0)
	}
	rand := rand.New(rand.NewPCG(42, 1024))
	var res []string
	for i := 0; i < 10; i++ {
		for _, input := range c.Inputs {
			input.Bit.SilentSet(rand.IntN(2) == 1)
		}
		res = append(res, c.simulate(len(c.Inputs))...)
		if c.Config.DrawSingleGraph {
			break
		}
	}
	return res
}

// SimulateInputs simulates the circuit for the given inputs.
func (c *Circuit) SimulateInputs(allInputs []string) []string {
	var res []string
	for _, inputs := range allInputs {
		for i, input := range inputs {
			c.Inputs[i].Bit.SilentSet(input == '1')
		}
		res = append(res, c.simulate(len(c.Inputs))...)
	}
	return res
}

func (c *Circuit) simulate(index int) []string {
	if index >= len(c.Inputs) {
		valid := true
		for _, fn := range c.InputValidations {
			if !fn() {
				valid = false
				break
			}
		}
		if !valid {
			return nil
		}
		c.Update()
		if c.Config.DrawGraph {
			return []string{c.Graph()}
		}
		if c.Config.IsUnitTest {
			return []string{c.StringForUnitTest()}
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
