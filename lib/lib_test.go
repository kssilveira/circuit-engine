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
	}, {
		name: "Xor",
		want: []string{"00=>0", "01=>1", "10=>1", "11=>0"},
	}, {
		name: "Nor",
		want: []string{"00=>1", "01=>0", "10=>0", "11=>0"},
	}, {
		name: "HalfSum",
		want: []string{"00=>00", "01=>10", "10=>10", "11=>01"},
	}, {
		name: "Sum",
		want: []string{"000=>00", "001=>10", "010=>10", "011=>01", "100=>10", "101=>01", "110=>01", "111=>11"},
	}, {
		name: "Sum2",
		want: []string{
			// a1 a2 b1 b2 cin => s1 s2 cout
			"00000=>000", "00001=>100", "00010=>010", "00011=>110", "00100=>100",
			"00101=>010", "00110=>110", "00111=>001", "01000=>010", "01001=>110",
			"01010=>001", "01011=>101", "01100=>110", "01101=>001", "01110=>101",
			"01111=>011", "10000=>100", "10001=>010", "10010=>110", "10011=>001",
			"10100=>010", "10101=>110", "10110=>001", "10111=>101", "11000=>110",
			"11001=>001", "11010=>101", "11011=>011", "11100=>001", "11101=>101",
			"11110=>011", "11111=>111",
		},
	}, {
		name: "",
		want: []string{"=>"},
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
