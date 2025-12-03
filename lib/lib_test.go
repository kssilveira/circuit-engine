package lib

import (
	"fmt"
	"os"
	"slices"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/kssilveira/circuit-engine/circuit"
	"github.com/kssilveira/circuit-engine/config"
	"github.com/kssilveira/circuit-engine/sfmt"
)

func TestOutputsCombinational(t *testing.T) {
	inputs := []struct {
		name        string
		isValidInt  func(inputs map[string]int) []int
		isValidBool func(inputs map[string]bool) []bool
	}{{
		name: "TransistorEmitter",
		isValidBool: func(inputs map[string]bool) []bool {
			return []bool{inputs["b"] && inputs["c"]}
		},
	}, {
		name: "TransistorGnd",
		isValidBool: func(inputs map[string]bool) []bool {
			return []bool{!inputs["b"] && inputs["c"]}
		},
	}, {
		name: "Transistor",
		isValidBool: func(inputs map[string]bool) []bool {
			return []bool{inputs["b"] && inputs["c"], inputs["c"]}
		},
	}, {
		name: "Not",
		isValidBool: func(inputs map[string]bool) []bool {
			return []bool{!inputs["a"]}
		},
	}, {
		name: "And",
		isValidBool: func(inputs map[string]bool) []bool {
			return []bool{inputs["a"] && inputs["b"]}
		},
	}, {
		name: "Or",
		isValidBool: func(inputs map[string]bool) []bool {
			return []bool{inputs["a"] || inputs["b"]}
		},
	}, {
		name: "OrRes",
		isValidBool: func(inputs map[string]bool) []bool {
			return []bool{inputs["a"]}
		},
	}, {
		name: "Nand",
		isValidBool: func(inputs map[string]bool) []bool {
			return []bool{!(inputs["a"] && inputs["b"])}
		},
	}, {
		name: "Nand(Nand)",
		isValidBool: func(inputs map[string]bool) []bool {
			return []bool{!(inputs["a"] && !(inputs["b"] && inputs["c"]))}
		},
	}, {
		name: "Xor",
		isValidBool: func(inputs map[string]bool) []bool {
			return []bool{inputs["a"] != inputs["b"]}
		},
	}, {
		name: "Nor",
		isValidBool: func(inputs map[string]bool) []bool {
			return []bool{!(inputs["a"] || inputs["b"])}
		},
	}, {
		name: "HalfSum",
		isValidInt: func(inputs map[string]int) []int {
			sum := inputs["a"] + inputs["b"]
			return []int{sum % 2, sum / 2}
		},
	}, {
		name: "Sum",
		isValidInt: func(inputs map[string]int) []int {
			sum := inputs["a"] + inputs["b"] + inputs["c"]
			return []int{sum % 2, sum / 2}
		},
	}, {
		name: "Sum2",
		isValidInt: func(inputs map[string]int) []int {
			sum0 := inputs["a0"] + inputs["b0"] + inputs["c"]
			sum1 := sum0/2 + inputs["a1"] + inputs["b1"]
			return []int{sum0 % 2, sum1 % 2, sum1 / 2}
		},
	}, {
		name: "SumN",
		isValidInt: func(inputs map[string]int) []int {
			sum0 := inputs["a0"] + inputs["b0"] + inputs["c"]
			sum1 := sum0/2 + inputs["a1"] + inputs["b1"]
			return []int{sum0 % 2, sum1 % 2, sum1 / 2}
		},
	}, {
		name: "SRLatch",
		isValidBool: func() func(inputs map[string]bool) []bool {
			q := true
			return func(inputs map[string]bool) []bool {
				if inputs["s"] && inputs["r"] {
					return []bool{false, false}
				}
				if inputs["s"] {
					q = true
				}
				if inputs["r"] {
					q = false
				}
				return []bool{q, !q}
			}
		}(),
	}, {
		name: "SRLatchWithEnable",
		isValidBool: func() func(inputs map[string]bool) []bool {
			q := true
			return func(inputs map[string]bool) []bool {
				if !inputs["e"] {
					return []bool{q, !q}
				}
				if inputs["s"] && inputs["r"] {
					return []bool{false, false}
				}
				if inputs["s"] {
					q = true
				}
				if inputs["r"] {
					q = false
				}
				return []bool{q, !q}
			}
		}(),
	}, {
		name: "DLatch",
		isValidBool: func() func(inputs map[string]bool) []bool {
			q := true
			return func(inputs map[string]bool) []bool {
				if !inputs["e"] {
					return []bool{q, !q}
				}
				q = inputs["d"]
				return []bool{q, !q}
			}
		}(),
	}, {
		name: "MSJKLatch",
		isValidBool: func() func(inputs map[string]bool) []bool {
			mq := true
			sq := true
			return func(inputs map[string]bool) []bool {
				if inputs["e"] {
					if inputs["j"] && inputs["k"] {
						sq = !sq
					}
					if inputs["j"] {
						sq = true
					}
					if inputs["k"] {
						sq = false
					}
				} else {
					mq = sq
				}
				return []bool{mq, !mq}
			}
		}(),
	}, {
		name: "Register",
		isValidBool: func() func(inputs map[string]bool) []bool {
			q := true
			return func(inputs map[string]bool) []bool {
				if inputs["i"] {
					q = inputs["d"]
				}
				return []bool{inputs["o"] && q}
			}
		}(),
	}, {
		name: "Register2",
		isValidBool: func() func(inputs map[string]bool) []bool {
			q0, q1 := true, true
			return func(inputs map[string]bool) []bool {
				if inputs["i"] {
					q0, q1 = inputs["d0"], inputs["d1"]
				}
				o := inputs["o"]
				return []bool{o && q0, o && q1}
			}
		}(),
	}, {
		name: "RegisterN",
		isValidBool: func() func(inputs map[string]bool) []bool {
			q0, q1 := true, true
			return func(inputs map[string]bool) []bool {
				if inputs["i"] {
					q0, q1 = inputs["d0"], inputs["d1"]
				}
				o := inputs["o"]
				return []bool{o && q0, o && q1}
			}
		}(),
	}, {
		name: "Alu",
		isValidInt: func() func(inputs map[string]int) []int {
			qa, qb, qr := 1, 1, 1
			return func(inputs map[string]int) []int {
				if inputs["ai"] == 1 {
					qa = inputs["a"]
				}
				if inputs["bi"] == 1 {
					qb = inputs["b"]
				}
				sum := qa + qb + inputs["c"]
				if inputs["ri"] == 1 {
					qr = sum % 2
				}
				return []int{qa, qb, inputs["ro"] & qr, sum / 2}
			}
		}(),
	}, {
		name: "Alu2",
		isValidInt: func() func(inputs map[string]int) []int {
			qa0, qb0, qr0 := 1, 1, 1
			qa1, qb1, qr1 := 1, 1, 1
			return func(inputs map[string]int) []int {
				if inputs["ai"] == 1 {
					qa0, qa1 = inputs["a0"], inputs["a1"]
				}
				if inputs["bi"] == 1 {
					qb0, qb1 = inputs["b0"], inputs["b1"]
				}
				sum0 := qa0 + qb0 + inputs["c"]
				sum1 := qa1 + qb1 + sum0/2
				if inputs["ri"] == 1 {
					qr0, qr1 = sum0%2, sum1%2
				}
				return []int{
					qa0, qb0, inputs["ro"] & qr0,
					qa1, qb1, inputs["ro"] & qr1,
					sum1 / 2,
				}
			}
		}(),
	}, {
		name: "AluN",
		isValidInt: func() func(inputs map[string]int) []int {
			qa0, qb0, qr0 := 1, 1, 1
			qa1, qb1, qr1 := 1, 1, 1
			return func(inputs map[string]int) []int {
				if inputs["ai"] == 1 {
					qa0, qa1 = inputs["a0"], inputs["a1"]
				}
				if inputs["bi"] == 1 {
					qb0, qb1 = inputs["b0"], inputs["b1"]
				}
				sum0 := qa0 + qb0 + inputs["c"]
				sum1 := qa1 + qb1 + sum0/2
				if inputs["ri"] == 1 {
					qr0, qr1 = sum0%2, sum1%2
				}
				return []int{
					qa0, qb0, inputs["ro"] & qr0,
					qa1, qb1, inputs["ro"] & qr1,
					sum1 / 2,
				}
			}
		}(),
	}, {
		name: "Bus",
		isValidBool: func(inputs map[string]bool) []bool {
			bus := inputs["d"] || inputs["r"]
			return []bool{bus, bus, bus}
		},
	}, {
		name: "Bus2",
		isValidBool: func(inputs map[string]bool) []bool {
			bus0 := inputs["d0"] || inputs["r0"]
			bus1 := inputs["d1"] || inputs["r1"]
			return []bool{bus0, bus1, bus0, bus1, bus0, bus1}
		},
	}, {
		name: "BusN",
		isValidBool: func(inputs map[string]bool) []bool {
			bus0 := inputs["d0"] || inputs["r0"]
			bus1 := inputs["d1"] || inputs["r1"]
			return []bool{bus0, bus1, bus0, bus1, bus0, bus1}
		},
	}, {
		name: "BusIOn",
		isValidBool: func(inputs map[string]bool) []bool {
			bus := inputs["d"] || inputs["ar"] || inputs["br"] || inputs["r"]
			return []bool{bus, bus, bus}
		},
	}, {
		name: "BusBnIOn",
		isValidBool: func(inputs map[string]bool) []bool {
			bus0 := inputs["d0"] || inputs["ar0"] || inputs["br0"] || inputs["r0"]
			bus1 := inputs["d1"] || inputs["ar1"] || inputs["br1"] || inputs["r1"]
			return []bool{bus0, bus1, bus0, bus1, bus0, bus1}
		},
	}, {
		name: "AluWithBus",
		isValidInt: func() func(inputs map[string]int) []int {
			qa, qb, qr := 1, 1, 1
			return func(inputs map[string]int) []int {
				rr, d, sum := 0, 0, 0
				for i := 0; i < 10; i++ {
					if inputs["ro"] == 1 {
						rr = qr
					}
					d = inputs["d"] | rr
					if inputs["ai"] == 1 {
						qa = d
					}
					if inputs["bi"] == 1 {
						qb = d
					}
					sum = qa + qb + inputs["c"]
					if inputs["ri"] == 1 {
						qr = sum % 2
					}
				}
				return []int{d, qa, qb, rr, sum / 2}
			}
		}(),
	}, {
		name: "AluWithBus2",
		isValidInt: func() func(inputs map[string]int) []int {
			qa0, qb0, qr0 := 1, 1, 1
			qa1, qb1, qr1 := 1, 1, 1
			return func(inputs map[string]int) []int {
				rr0, d0, sum0 := 0, 0, 0
				rr1, d1, sum1 := 0, 0, 0
				for i := 0; i < 10; i++ {
					if inputs["ro"] == 1 {
						rr0, rr1 = qr0, qr1
					}
					d0 = inputs["d0"] | rr0
					d1 = inputs["d1"] | rr1
					if inputs["ai"] == 1 {
						qa0, qa1 = d0, d1
					}
					if inputs["bi"] == 1 {
						qb0, qb1 = d0, d1
					}
					sum0 = qa0 + qb0 + inputs["c"]
					sum1 = qa1 + qb1 + sum0/2
					if inputs["ri"] == 1 {
						qr0, qr1 = sum0%2, sum1%2
					}
				}
				return []int{
					d0, qa0, qb0, rr0,
					d1, qa1, qb1, rr1,
					sum1 / 2}
			}
		}(),
	}, {
		name: "AluWithBusN",
		isValidInt: func() func(inputs map[string]int) []int {
			qa0, qb0, qr0 := 1, 1, 1
			qa1, qb1, qr1 := 1, 1, 1
			return func(inputs map[string]int) []int {
				rr0, d0, sum0 := 0, 0, 0
				rr1, d1, sum1 := 0, 0, 0
				for i := 0; i < 10; i++ {
					if inputs["ro"] == 1 {
						rr0, rr1 = qr0, qr1
					}
					d0 = inputs["d0"] | rr0
					d1 = inputs["d1"] | rr1
					if inputs["ai"] == 1 {
						qa0, qa1 = d0, d1
					}
					if inputs["bi"] == 1 {
						qb0, qb1 = d0, d1
					}
					sum0 = qa0 + qb0 + inputs["c"]
					sum1 = qa1 + qb1 + sum0/2
					if inputs["ri"] == 1 {
						qr0, qr1 = sum0%2, sum1%2
					}
				}
				return []int{
					d0, qa0, qb0, rr0,
					d1, qa1, qb1, rr1,
					sum1 / 2}
			}
		}(),
	}, {
		name: "RAM",
		isValidBool: func() func(inputs map[string]bool) []bool {
			q0, q1 := true, true
			return func(inputs map[string]bool) []bool {
				a, i, o := inputs["a"], inputs["i"], inputs["o"]
				s1 := a
				r0, r1 := false, false
				q, r := &q0, &r0
				if s1 {
					q, r = &q1, &r1
				}
				if i {
					*q = inputs["d"]
				}
				if o {
					*r = *q
				}
				return []bool{r0, r1}
			}
		}(),
	}, {
		name: "RAMa2",
		isValidBool: func() func(inputs map[string]bool) []bool {
			q := []bool{true, true, true, true}
			return func(inputs map[string]bool) []bool {
				a0, a1, i, o := inputs["a0"], inputs["a1"], inputs["i"], inputs["o"]
				s := []bool{!a0 && !a1, a0 && !a1, !a0 && a1, a0 && a1}
				r := []bool{false, false, false, false}
				index := 0
				for i, si := range s {
					if si {
						index = i
						break
					}
				}
				if i {
					q[index] = inputs["d"]
				}
				if o {
					r[index] = q[index]
				}
				var res []bool
				for i := range s {
					res = append(res, r[i])
				}
				return res
			}
		}(),
	}, {
		name: "RAMb2",
		isValidBool: func() func(inputs map[string]bool) []bool {
			q0, q1 := []bool{true, true}, []bool{true, true}
			return func(inputs map[string]bool) []bool {
				a, i, o := inputs["a"], inputs["i"], inputs["o"]
				s1 := a
				r0, r1 := []bool{false, false}, []bool{false, false}
				q, r := &q0, &r0
				if s1 {
					q, r = &q1, &r1
				}
				if i {
					*q = []bool{inputs["d0"], inputs["d1"]}
				}
				if o {
					*r = *q
				}
				return slices.Concat(r0, r1)
			}
		}(),
	}, {
		name: "RAMa2b2",
		isValidBool: func() func(inputs map[string]bool) []bool {
			q := [][]bool{{true, true}, {true, true}, {true, true}, {true, true}}
			return func(inputs map[string]bool) []bool {
				a0, a1, i, o := inputs["a0"], inputs["a1"], inputs["i"], inputs["o"]
				s := []bool{!a0 && !a1, a0 && !a1, !a0 && a1, a0 && a1}
				r := [][]bool{{false, false}, {false, false}, {false, false}, {false, false}}
				index := 0
				for i, si := range s {
					if si {
						index = i
						break
					}
				}
				if i {
					q[index][0] = inputs["d0"]
					q[index][1] = inputs["d1"]
				}
				if o {
					r[index] = q[index]
				}
				var res []bool
				for i := range s {
					res = append(res, r[i]...)
				}
				return res
			}
		}(),
	}, {
		name: "AluWithRAM",
		isValidInt: func() func(inputs map[string]int) []int {
			qa, qb, qr, qma, qm0, qm1 := 1, 1, 1, 1, 1, 1
			return func(inputs map[string]int) []int {
				rr, d, sum := 0, 0, 0
				rm0, rm1 := 0, 0
				qm, rm := &qm0, &rm0
				s1 := qma
				if s1 == 1 {
					qm, rm = &qm1, &rm1
				}
				for i := 0; i < 10; i++ {
					if inputs["ro"] == 1 {
						rr = qr
					}
					if inputs["mo"] == 1 {
						*rm = *qm
					}
					if inputs["d"] == 1 || rr == 1 || *rm == 1 {
						d = 1
					}
					if inputs["ai"] == 1 {
						qa = d
					}
					if inputs["bi"] == 1 {
						qb = d
					}
					if inputs["mai"] == 1 {
						qma = d
					}
					if inputs["mi"] == 1 {
						*qm = d
					}
					sum = qa + qb + inputs["c"]
					if inputs["ri"] == 1 {
						qr = sum % 2
					}
				}
				return []int{d, sum / 2, qa, qb, rr, qma, rm0, rm1}
			}
		}(),
	}, {
		name: "AluWithRAM2",
		isValidInt: func() func(inputs map[string]int) []int {
			qa, qb, qr, qma := []int{1, 1}, []int{1, 1}, []int{1, 1}, []int{1, 1}
			qm := [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}
			return func(inputs map[string]int) []int {
				rr, d, sum := []int{0, 0}, []int{0, 0}, []int{0, 0}
				rm := [][]int{{0, 0}, {0, 0}, {0, 0}, {0, 0}}
				qmr, rmr := &qm[0], &rm[0]
				s := []bool{
					qma[0] == 0 && qma[1] == 0, qma[0] == 1 && qma[1] == 0,
					qma[0] == 0 && qma[1] == 1, qma[0] == 1 && qma[1] == 1,
				}
				for i, si := range s {
					if si {
						qmr, rmr = &qm[i], &rm[i]
						break
					}
				}
				for i := 0; i < 10; i++ {
					if inputs["ro"] == 1 {
						rr = qr
					}
					if inputs["mo"] == 1 {
						*rmr = *qmr
					}
					ind := []int{inputs["d0"], inputs["d1"]}
					for i := range d {
						if ind[i] == 1 || rr[i] == 1 || (*rmr)[i] == 1 {
							d[i] = 1
						}
					}
					if inputs["ai"] == 1 {
						qa = d
					}
					if inputs["bi"] == 1 {
						qb = d
					}
					if inputs["mai"] == 1 {
						qma = d
					}
					if inputs["mi"] == 1 {
						*qmr = d
					}
					sum[0] = qa[0] + qb[0] + inputs["c"]
					sum[1] = qa[1] + qb[1] + sum[0]/2
					if inputs["ri"] == 1 {
						qr = []int{sum[0] % 2, sum[1] % 2}
					}
				}
				return slices.Concat(append(d, sum[1]/2), qa, qb, rr, qma, slices.Concat(rm...))
			}
		}(),
	}}
	for _, in := range inputs {
		c := circuit.NewCircuit(config.Config{IsUnitTest: true})
		c.Outs(Example(c, in.name))
		gotDesc := c.Description()
		got := c.Simulate()
		converted := []string{gotDesc, ""}
		for _, out := range got {
			var one []string
			for i, input := range c.Inputs {
				one = append(one, sfmt.Sprintf("%s(%s)", input.Name, string(out[i])))
			}
			one = append(one, "=>")
			for i, output := range c.Outputs {
				one = append(one, sfmt.Sprintf("%s(%s)", output.Name, string(out[i+len(c.Inputs)+len("=>")])))
			}
			converted = append(converted, strings.Join(one, " "))
		}
		converted = append(converted, "")
		if err := os.WriteFile(fmt.Sprintf("testdata/%s.txt", in.name), []byte(strings.Join(converted, "\n")), 0644); err != nil {
			t.Errorf("WriteFile got err %v", err)
		}

		if in.isValidInt != nil {
			for _, out := range got {
				inputs := map[string]int{}
				for i, input := range c.Inputs {
					inputs[input.Name] = int(out[i] - '0')
				}
				outputs := map[string]int{}
				for i, output := range c.Outputs {
					outputs[output.Name] = int(out[len(c.Inputs)+len("=>")+i] - '0')
				}
				wants := in.isValidInt(inputs)
				for i, want := range wants {
					oi := len(c.Inputs) + len("=>") + i
					if oi >= len(out) {
						t.Errorf("Simulate(%q) out %s want %d got <nothing>", in.name, out, want)
						continue
					}
					got := int(out[oi] - '0')
					if want != got {
						t.Errorf("Simulate(%q) out %s output %s want %d got %#v\n\ninputs\n\n%#v\n\noutputs\n\n%#v\n\n", in.name, out, c.Outputs[i].Name, want, got, inputs, outputs)
					}
				}
			}
		}
		if in.isValidBool != nil {
			for _, out := range got {
				inputs := map[string]bool{}
				for i, input := range c.Inputs {
					inputs[input.Name] = out[i] == '1'
				}
				outputs := map[string]bool{}
				for i, output := range c.Outputs {
					outputs[output.Name] = out[len(c.Inputs)+len("=>")+i] == '1'
				}
				wants := in.isValidBool(inputs)
				for i, want := range wants {
					oi := len(c.Inputs) + len("=>") + i
					if oi >= len(out) {
						t.Errorf("Simulate(%q) out %s want %t got <nothing>", in.name, out, want)
						continue
					}
					got := out[oi] == '1'
					if want != got {
						t.Errorf("Simulate(%q) out %s output %s want %t got %#v\n\ninputs\n\n%#v\n\noutputs\n\n%#v\n\n", in.name, out, c.Outputs[i].Name, want, got, inputs, outputs)
					}
				}
			}
		}
	}
}

func TestOutputsSequential(t *testing.T) {
	inputs := []struct {
		name   string
		desc   string
		inputs []string
		want   []string
	}{{
		name:   "OrRes",
		desc:   "a => OR(a,res)",
		inputs: []string{"0", "1", "0"},
		want:   []string{"a(0) => OR(a,res)(0)", "a(1) => OR(a,res)(1)", "a(0) => OR(a,res)(1)"},
	}, {
		name:   "SRLatchWithEnable",
		desc:   "s r e => q nq",
		inputs: []string{"000", "001", "010", "011", "000", "100", "101", "000"},
		want: []string{
			"s(0) r(0) e(0) => q(1) nq(0)", "s(0) r(0) e(1) => q(1) nq(0)", "s(0) r(1) e(0) => q(1) nq(0)", "s(0) r(1) e(1) => q(0) nq(1)",
			"s(0) r(0) e(0) => q(0) nq(1)", "s(1) r(0) e(0) => q(0) nq(1)", "s(1) r(0) e(1) => q(1) nq(0)", "s(0) r(0) e(0) => q(1) nq(0)",
		},
	}, {
		name:   "MSJKLatch",
		desc:   "j k e => mq nmq",
		inputs: []string{"000", "011", "000", "101", "000", "111", "000", "111", "000"},
		want: []string{
			"j(0) k(0) e(0) => mq(1) nmq(0)", "j(0) k(1) e(1) => mq(1) nmq(0)",
			"j(0) k(0) e(0) => mq(0) nmq(1)", "j(1) k(0) e(1) => mq(0) nmq(1)",
			"j(0) k(0) e(0) => mq(1) nmq(0)", "j(1) k(1) e(1) => mq(1) nmq(0)",
			"j(0) k(0) e(0) => mq(0) nmq(1)", "j(1) k(1) e(1) => mq(0) nmq(1)",
			"j(0) k(0) e(0) => mq(1) nmq(0)",
		},
	}, {
		name:   "AluWithBus",
		desc:   "d ai bi ri ro c => B(d) R(da,ai,T) R(db,bi,T) R(S(R(da,ai,T),R(db,bi,T),c),ri,ro) C(R(da,ai,T),R(db,bi,T))",
		inputs: []string{"000000", "100000", "010000", "000110", "000111", "001010", "000111"},
		want: []string{
			// default to a=b=1 sum=0 cout=1
			"d(0) ai(0) bi(0) ri(0) ro(0) c(0) => B(d)(0) R(da,ai,T)(1) R(db,bi,T)(1) R(S(R(da,ai,T),R(db,bi,T),c),ri,ro)(0) C(R(da,ai,T),R(db,bi,T))(1)",
			// d=1 writes to the bus
			"d(1) ai(0) bi(0) ri(0) ro(0) c(0) => B(d)(1) R(da,ai,T)(1) R(db,bi,T)(1) R(S(R(da,ai,T),R(db,bi,T),c),ri,ro)(0) C(R(da,ai,T),R(db,bi,T))(1)",
			// ai=1 sets a=0 from the bus
			"d(0) ai(1) bi(0) ri(0) ro(0) c(0) => B(d)(0) R(da,ai,T)(0) R(db,bi,T)(1) R(S(R(da,ai,T),R(db,bi,T),c),ri,ro)(0) C(R(da,ai,T),R(db,bi,T))(0)",
			// ri=ro=1 writes sum=1 to r and bus
			"d(0) ai(0) bi(0) ri(1) ro(1) c(0) => B(d)(1) R(da,ai,T)(0) R(db,bi,T)(1) R(S(R(da,ai,T),R(db,bi,T),c),ri,ro)(1) C(R(da,ai,T),R(db,bi,T))(0)",
			// ri=ro=c=1 writes sum=0 to r and bus
			"d(0) ai(0) bi(0) ri(1) ro(1) c(1) => B(d)(0) R(da,ai,T)(0) R(db,bi,T)(1) R(S(R(da,ai,T),R(db,bi,T),c),ri,ro)(0) C(R(da,ai,T),R(db,bi,T))(1)",
			// bi=ro=1 writes sum=0 to b and bus
			"d(0) ai(0) bi(1) ri(0) ro(1) c(0) => B(d)(0) R(da,ai,T)(0) R(db,bi,T)(0) R(S(R(da,ai,T),R(db,bi,T),c),ri,ro)(0) C(R(da,ai,T),R(db,bi,T))(0)",
			// ri=ro=c=1 writes sum=1 to r and bus
			"d(0) ai(0) bi(0) ri(1) ro(1) c(1) => B(d)(1) R(da,ai,T)(0) R(db,bi,T)(0) R(S(R(da,ai,T),R(db,bi,T),c),ri,ro)(1) C(R(da,ai,T),R(db,bi,T))(0)",
		},
	}, {
		name:   "RAM",
		desc:   "a d i o => R(d,i0,o0) R(d,i1,o1)",
		inputs: []string{"0000", "0001", "0010", "0001", "1001"},
		want: []string{
			// default to s0=q0=q1=1
			"a(0) d(0) i(0) o(0) => R(d,i0,o0)(0) R(d,i1,o1)(0)",
			// o=1 writes q0=1 to res
			"a(0) d(0) i(0) o(1) => R(d,i0,o0)(1) R(d,i1,o1)(0)",
			// i=1 reads d=0 to q0
			"a(0) d(0) i(1) o(0) => R(d,i0,o0)(0) R(d,i1,o1)(0)",
			// o=1 writes q0=0 to res
			"a(0) d(0) i(0) o(1) => R(d,i0,o0)(0) R(d,i1,o1)(0)",
			// a=o=1 writes q1=1 to res
			"a(1) d(0) i(0) o(1) => R(d,i0,o0)(0) R(d,i1,o1)(1)",
		},
	}}
	for _, in := range inputs {
		c := circuit.NewCircuit(config.Config{IsUnitTest: true})
		c.Outs(Example(c, in.name))
		gotDesc := c.Description()
		if diff := cmp.Diff(in.desc, gotDesc); diff != "" {
			t.Errorf("SimulateInputs(%q) want %#v,\ngot %#v,\ndiff -want +got:\n%s", in.name, in.desc, gotDesc, diff)
		}
		for _, inputs := range in.inputs {
			if len(inputs) != len(c.Inputs) {
				t.Errorf("SimulateInputs(%q) inputs want %d got %d", in.name, len(inputs), len(c.Inputs))
			}
		}
		got := c.SimulateInputs(in.inputs)
		var converted []string
		for _, out := range got {
			var one []string
			for i, input := range c.Inputs {
				one = append(one, sfmt.Sprintf("%s(%s)", input.Name, string(out[i])))
			}
			one = append(one, "=>")
			for i, output := range c.Outputs {
				one = append(one, sfmt.Sprintf("%s(%s)", output.Name, string(out[i+len(c.Inputs)+len("=>")])))
			}
			converted = append(converted, strings.Join(one, " "))
		}
		if diff := cmp.Diff(in.want, converted); diff != "" {
			t.Errorf("SimulateInputs(%q) want %#v,\ngot %#v,\ndiff -want +got:\n%s", in.name, in.want, converted, diff)
		}
	}
}
