package lib

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/kssilveira/circuit-engine/circuit"
	"github.com/kssilveira/circuit-engine/config"
)

func TestOutputsCombinational(t *testing.T) {
	inputs := []struct {
		name string
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
	}, {
		name: "Not",
		want: []string{"0=>1", "1=>0"},
	}, {
		name: "And",
		want: []string{"00=>0", "01=>0", "10=>0", "11=>1"},
	}, {
		name: "Or",
		want: []string{"00=>0", "01=>1", "10=>1", "11=>1"},
	}, {
		name: "OrRes",
		want: []string{"0=>0", "1=>1"},
	}, {
		name: "Nand",
		want: []string{"00=>1", "01=>1", "10=>1", "11=>0"},
	}, {
		name: "Nand(Nand)",
		want: []string{"000=>1", "001=>1", "010=>1", "011=>1", "100=>0", "101=>0", "110=>0", "111=>1"},
	}}
	for _, in := range inputs {
		c := circuit.NewCircuit(config.Config{IsUnitTest: true})
		c.Outs(Example(c, in.name))
		got := c.Simulate()
		if diff := cmp.Diff(in.want, got); diff != "" {
			t.Errorf("Simulate(%q) want %#v,\ngot %#v,\ndiff -want +got:\n%s", in.name, in.want, got, diff)
		}
	}
}

func TestOutputsSequential(t *testing.T) {
	inputs := []struct {
		name string
		inputs []string
		want []string
	}{{
		name: "OrRes",
		inputs : []string{"0", "1", "0"},
		want: []string{"0=>0", "1=>1", "0=>1"},
	}}
	for _, in := range inputs {
		c := circuit.NewCircuit(config.Config{IsUnitTest: true})
		c.Outs(Example(c, in.name))
		got := c.SimulateInputs(in.inputs)
		if diff := cmp.Diff(in.want, got); diff != "" {
			t.Errorf("SimulateInputs(%q) want %#v,\ngot %#v,\ndiff -want +got:\n%s", in.name, in.want, got, diff)
		}
	}
}
