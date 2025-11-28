package lib

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/kssilveira/circuit-engine/circuit"
	"github.com/kssilveira/circuit-engine/config"
)

func TestOutputsCombinational(t *testing.T) {
	inputs := []struct {
		name    string
		want    []string
		isValid func(inputs map[string]int) []int
	}{{
		name: "TransistorEmitter",
		// base collector => emitter
		want: []string{"00=>0", "01=>0", "10=>0", "11=>1"},
	}, {
		name: "TransistorGnd",
		// base collector => collectorOut
		want: []string{"00=>0", "01=>1", "10=>0", "11=>0"},
	}, {
		name: "Transistor",
		// base collector => emitter collectorOut
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
		// a b => s carry
		want: []string{"00=>00", "01=>10", "10=>10", "11=>01"},
	}, {
		name: "Sum",
		// a b cin => s cout
		want: []string{"000=>00", "001=>10", "010=>10", "011=>01", "100=>10", "101=>01", "110=>01", "111=>11"},
	}, {
		name: "Sum2",
		// a1 a2 b1 b2 cin => s1 s2 cout
		want: []string{
			"00000=>000", "00001=>100", "00010=>010", "00011=>110",
			"00100=>100", "00101=>010", "00110=>110", "00111=>001",
			"01000=>010", "01001=>110", "01010=>001", "01011=>101",
			"01100=>110", "01101=>001", "01110=>101", "01111=>011",
			"10000=>100", "10001=>010", "10010=>110", "10011=>001",
			"10100=>010", "10101=>110", "10110=>001", "10111=>101",
			"11000=>110", "11001=>001", "11010=>101", "11011=>011",
			"11100=>001", "11101=>101", "11110=>011", "11111=>111",
		},
		isValid: func(inputs map[string]int) []int {
			sum1 := inputs["a1"] + inputs["b1"] + inputs["c"]
			sum2 := sum1/2 + inputs["a2"] + inputs["b2"]
			return []int{sum1 % 2, sum2 % 2, sum2 / 2}
		},
	}, {
		name: "Sum4",
		// a1 a2 a3 a4 b1 b2 b3 b4 cin => s1 s2 s3 s4 cout
		want: []string{
			"110110101=>10001", "110000110=>11110", "000101101=>11110", "000000010=>00010",
			"110111101=>11001", "000100010=>00001", "000101110=>01101", "110010110=>00001",
			"100000101=>01100", "001101111=>11011",
		},
		isValid: func(inputs map[string]int) []int {
			sum1 := inputs["a1"] + inputs["b1"] + inputs["c"]
			sum2 := sum1/2 + inputs["a2"] + inputs["b2"]
			sum3 := sum2/2 + inputs["a3"] + inputs["b3"]
			sum4 := sum3/2 + inputs["a4"] + inputs["b4"]
			return []int{sum1 % 2, sum2 % 2, sum3 % 2, sum4 % 2, sum4 / 2}
		},
	}, {
		name: "Sum8",
		// a1 a2 a3 a4 a5 a5 a7 a8 b1 b2 b3 b4 b5 b6 b7 b8 cin => s1 s2 s3 s4 s5 s6 s7 s8 cout
		want: []string{
			"11011010111000011=>110001110",
			"00001011010000000=>010010110",
			"10110111101000100=>010011001",
			"01000010111011001=>010111100",
			"01101000001010011=>110101010",
			"01111111011010010=>001010011",
			"00101000100111110=>101100001",
			"00010101010111000=>010001110",
			"11010011100110000=>001001110",
			"11100100110000110=>010101110",
		},
		isValid: func(inputs map[string]int) []int {
			sum1 := inputs["a1"] + inputs["b1"] + inputs["c"]
			sum2 := sum1/2 + inputs["a2"] + inputs["b2"]
			sum3 := sum2/2 + inputs["a3"] + inputs["b3"]
			sum4 := sum3/2 + inputs["a4"] + inputs["b4"]
			sum5 := sum4/2 + inputs["a5"] + inputs["b5"]
			sum6 := sum5/2 + inputs["a6"] + inputs["b6"]
			sum7 := sum6/2 + inputs["a7"] + inputs["b7"]
			sum8 := sum7/2 + inputs["a8"] + inputs["b8"]
			return []int{sum1 % 2, sum2 % 2, sum3 % 2, sum4 % 2, sum5 % 2, sum6 % 2, sum7 % 2, sum8 % 2, sum8 / 2}
		},
	}, {
		name: "SRLatch",
		// s r => q !q
		want: []string{"00=>10", "01=>01", "10=>10", "11=>00"},
	}, {
		name: "SRLatchWithEnable",
		// s r e => q !q
		want: []string{"000=>10", "001=>10", "010=>10", "011=>01", "100=>01", "101=>10", "110=>10", "111=>00"},
	}, {
		name: "DLatch",
		// d e => q !q
		want: []string{"00=>10", "01=>01", "10=>01", "11=>10"},
	}}
	for _, in := range inputs {
		c := circuit.NewCircuit(config.Config{IsUnitTest: true})
		c.Outs(Example(c, in.name))
		got := c.Simulate()
		if diff := cmp.Diff(in.want, got); diff != "" {
			t.Errorf("Simulate(%q) want %#v,\ngot %#v,\ndiff -want +got:\n%s", in.name, in.want, got, diff)
		}
		if in.isValid == nil {
			continue
		}
		for _, out := range got {
			inputs := map[string]int{}
			for i, input := range c.Inputs {
				inputs[input.Name] = int(out[i] - '0')
			}
			wants := in.isValid(inputs)
			for i, want := range wants {
				got := int(out[len(c.Inputs)+len("=>")+i] - '0')
				if want != got {
					t.Errorf("Simulate(%q) out %s output %s want %d got %#v", in.name, out, c.Outputs[i].Name, want, got)
				}
			}
		}
	}
}

func TestOutputsSequential(t *testing.T) {
	inputs := []struct {
		name   string
		inputs []string
		want   []string
	}{{
		name:   "OrRes",
		inputs: []string{"0", "1", "0"},
		want:   []string{"0=>0", "1=>1", "0=>1"},
	}, {
		name: "SRLatchWithEnable",
		// s r e => q !q
		inputs: []string{"000", "001", "010", "011", "000", "100", "101", "000"},
		want:   []string{"000=>10", "001=>10", "010=>10", "011=>01", "000=>01", "100=>01", "101=>10", "000=>10"},
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
