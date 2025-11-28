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
		name: "TransistorEmitter",
		want: []string{"00=>0", "01=>0", "10=>0", "11=>1"},
	}, {
		name: "TransistorGnd",
		want: []string{"00=>0", "01=>1", "10=>0", "11=>0"},
	}, {
		name: "Transistor",
		want: []string{"00=>00", "01=>01", "10=>00", "11=>11"},
	}}
	for _, in := range inputs {
		c := circuit.NewCircuit(config.Config{IsUnitTest: true})
		c.Outs(Example(c, in.name))
		got := c.Simulate()
		if diff := cmp.Diff(in.want, got); diff != "" {
			t.Errorf("Simulate(%q) want\n\t%#v\ngot\n\t%#v\ndiff -want +got:\n%s", in.name, in.want, got, diff)
		}
	}
}
