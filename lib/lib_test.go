package lib

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/kssilveira/circuit-engine/circuit"
	"github.com/kssilveira/circuit-engine/config"
	"github.com/kssilveira/circuit-engine/wire"
)

func TestOutputs(t *testing.T) {
	inputs := []struct {
		name string
		in   func(c *circuit.Circuit) []*wire.Wire
		want []string
	}{{
		name: "transistor",
		in: func(c *circuit.Circuit) []*wire.Wire {
			return Transistor(c.Group(""), c.In("base"), c.In("collector"))
		},
		want: []string{"00=>00", "01=>01", "10=>00", "11=>11"},
	}}
	for _, in := range inputs {
		c := circuit.NewCircuit(config.Config{IsUnitTest: true})
		c.Outs(in.in(c))
		got := c.Simulate()
		if diff := cmp.Diff(in.want, got); diff != "" {
			t.Errorf("Simulate(%q) want\n\t%#v\ngot\n\t%#v\ndiff -want +got:\n%s", in.name, in.want, got, diff)
		}
	}
}
