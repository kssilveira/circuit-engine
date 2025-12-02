package lib

import (
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
		desc        string
		want        []string
		isValidInt  func(inputs map[string]int) []int
		isValidBool func(inputs map[string]bool) []bool
	}{{
		name: "TransistorEmitter",
		desc: "b c => e",
		want: []string{"b(0) c(0) => e(0)", "b(0) c(1) => e(0)", "b(1) c(0) => e(0)", "b(1) c(1) => e(1)"},
		isValidBool: func(inputs map[string]bool) []bool {
			return []bool{inputs["b"] && inputs["c"]}
		},
	}, {
		name: "TransistorGnd",
		desc: "b c => co",
		want: []string{"b(0) c(0) => co(0)", "b(0) c(1) => co(1)", "b(1) c(0) => co(0)", "b(1) c(1) => co(0)"},
		isValidBool: func(inputs map[string]bool) []bool {
			return []bool{!inputs["b"] && inputs["c"]}
		},
	}, {
		name: "Transistor",
		desc: "b c => e co",
		want: []string{"b(0) c(0) => e(0) co(0)", "b(0) c(1) => e(0) co(1)", "b(1) c(0) => e(0) co(0)", "b(1) c(1) => e(1) co(1)"},
		isValidBool: func(inputs map[string]bool) []bool {
			return []bool{inputs["b"] && inputs["c"], inputs["c"]}
		},
	}, {
		name: "Not",
		desc: "a => NOT(a)",
		want: []string{"a(0) => NOT(a)(1)", "a(1) => NOT(a)(0)"},
		isValidBool: func(inputs map[string]bool) []bool {
			return []bool{!inputs["a"]}
		},
	}, {
		name: "And",
		desc: "a b => AND(a,b)",
		want: []string{"a(0) b(0) => AND(a,b)(0)", "a(0) b(1) => AND(a,b)(0)", "a(1) b(0) => AND(a,b)(0)", "a(1) b(1) => AND(a,b)(1)"},
		isValidBool: func(inputs map[string]bool) []bool {
			return []bool{inputs["a"] && inputs["b"]}
		},
	}, {
		name: "Or",
		desc: "a b => OR(a,b)",
		want: []string{"a(0) b(0) => OR(a,b)(0)", "a(0) b(1) => OR(a,b)(1)", "a(1) b(0) => OR(a,b)(1)", "a(1) b(1) => OR(a,b)(1)"},
		isValidBool: func(inputs map[string]bool) []bool {
			return []bool{inputs["a"] || inputs["b"]}
		},
	}, {
		name: "OrRes",
		desc: "a => OR(a,res)",
		want: []string{"a(0) => OR(a,res)(0)", "a(1) => OR(a,res)(1)"},
		isValidBool: func(inputs map[string]bool) []bool {
			return []bool{inputs["a"]}
		},
	}, {
		name: "Nand",
		desc: "a b => NAND(a,b)",
		want: []string{"a(0) b(0) => NAND(a,b)(1)", "a(0) b(1) => NAND(a,b)(1)", "a(1) b(0) => NAND(a,b)(1)", "a(1) b(1) => NAND(a,b)(0)"},
		isValidBool: func(inputs map[string]bool) []bool {
			return []bool{!(inputs["a"] && inputs["b"])}
		},
	}, {
		name: "Nand(Nand)",
		desc: "a b c => NAND(a,NAND(b,c))",
		want: []string{
			"a(0) b(0) c(0) => NAND(a,NAND(b,c))(1)", "a(0) b(0) c(1) => NAND(a,NAND(b,c))(1)",
			"a(0) b(1) c(0) => NAND(a,NAND(b,c))(1)", "a(0) b(1) c(1) => NAND(a,NAND(b,c))(1)",
			"a(1) b(0) c(0) => NAND(a,NAND(b,c))(0)", "a(1) b(0) c(1) => NAND(a,NAND(b,c))(0)",
			"a(1) b(1) c(0) => NAND(a,NAND(b,c))(0)", "a(1) b(1) c(1) => NAND(a,NAND(b,c))(1)",
		},
		isValidBool: func(inputs map[string]bool) []bool {
			return []bool{!(inputs["a"] && !(inputs["b"] && inputs["c"]))}
		},
	}, {
		name: "Xor",
		desc: "a b => XOR(a,b)",
		want: []string{"a(0) b(0) => XOR(a,b)(0)", "a(0) b(1) => XOR(a,b)(1)", "a(1) b(0) => XOR(a,b)(1)", "a(1) b(1) => XOR(a,b)(0)"},
		isValidBool: func(inputs map[string]bool) []bool {
			return []bool{inputs["a"] != inputs["b"]}
		},
	}, {
		name: "Nor",
		desc: "a b => NOR(a,b)",
		want: []string{"a(0) b(0) => NOR(a,b)(1)", "a(0) b(1) => NOR(a,b)(0)", "a(1) b(0) => NOR(a,b)(0)", "a(1) b(1) => NOR(a,b)(0)"},
		isValidBool: func(inputs map[string]bool) []bool {
			return []bool{!(inputs["a"] || inputs["b"])}
		},
	}, {
		name: "HalfSum",
		desc: "a b => S(a,b) C(a,b)",
		want: []string{
			"a(0) b(0) => S(a,b)(0) C(a,b)(0)", "a(0) b(1) => S(a,b)(1) C(a,b)(0)",
			"a(1) b(0) => S(a,b)(1) C(a,b)(0)", "a(1) b(1) => S(a,b)(0) C(a,b)(1)",
		},
		isValidInt: func(inputs map[string]int) []int {
			sum := inputs["a"] + inputs["b"]
			return []int{sum % 2, sum / 2}
		},
	}, {
		name: "Sum",
		desc: "a b c => S(a,b,c) C(a,b)",
		want: []string{
			"a(0) b(0) c(0) => S(a,b,c)(0) C(a,b)(0)", "a(0) b(0) c(1) => S(a,b,c)(1) C(a,b)(0)",
			"a(0) b(1) c(0) => S(a,b,c)(1) C(a,b)(0)", "a(0) b(1) c(1) => S(a,b,c)(0) C(a,b)(1)",
			"a(1) b(0) c(0) => S(a,b,c)(1) C(a,b)(0)", "a(1) b(0) c(1) => S(a,b,c)(0) C(a,b)(1)",
			"a(1) b(1) c(0) => S(a,b,c)(0) C(a,b)(1)", "a(1) b(1) c(1) => S(a,b,c)(1) C(a,b)(1)",
		},
		isValidInt: func(inputs map[string]int) []int {
			sum := inputs["a"] + inputs["b"] + inputs["c"]
			return []int{sum % 2, sum / 2}
		},
	}, {
		name: "Sum2",
		desc: "a0 a1 b0 b1 c => S(a0,b0,c) S(a1,b1,C(a0,b0)) C(a1,b1)",
		want: []string{
			"a0(0) a1(0) b0(0) b1(0) c(0) => S(a0,b0,c)(0) S(a1,b1,C(a0,b0))(0) C(a1,b1)(0)",
			"a0(0) a1(0) b0(0) b1(0) c(1) => S(a0,b0,c)(1) S(a1,b1,C(a0,b0))(0) C(a1,b1)(0)",
			"a0(0) a1(0) b0(0) b1(1) c(0) => S(a0,b0,c)(0) S(a1,b1,C(a0,b0))(1) C(a1,b1)(0)",
			"a0(0) a1(0) b0(0) b1(1) c(1) => S(a0,b0,c)(1) S(a1,b1,C(a0,b0))(1) C(a1,b1)(0)",
			"a0(0) a1(0) b0(1) b1(0) c(0) => S(a0,b0,c)(1) S(a1,b1,C(a0,b0))(0) C(a1,b1)(0)",
			"a0(0) a1(0) b0(1) b1(0) c(1) => S(a0,b0,c)(0) S(a1,b1,C(a0,b0))(1) C(a1,b1)(0)",
			"a0(0) a1(0) b0(1) b1(1) c(0) => S(a0,b0,c)(1) S(a1,b1,C(a0,b0))(1) C(a1,b1)(0)",
			"a0(0) a1(0) b0(1) b1(1) c(1) => S(a0,b0,c)(0) S(a1,b1,C(a0,b0))(0) C(a1,b1)(1)",
			"a0(0) a1(1) b0(0) b1(0) c(0) => S(a0,b0,c)(0) S(a1,b1,C(a0,b0))(1) C(a1,b1)(0)",
			"a0(0) a1(1) b0(0) b1(0) c(1) => S(a0,b0,c)(1) S(a1,b1,C(a0,b0))(1) C(a1,b1)(0)",
			"a0(0) a1(1) b0(0) b1(1) c(0) => S(a0,b0,c)(0) S(a1,b1,C(a0,b0))(0) C(a1,b1)(1)",
			"a0(0) a1(1) b0(0) b1(1) c(1) => S(a0,b0,c)(1) S(a1,b1,C(a0,b0))(0) C(a1,b1)(1)",
			"a0(0) a1(1) b0(1) b1(0) c(0) => S(a0,b0,c)(1) S(a1,b1,C(a0,b0))(1) C(a1,b1)(0)",
			"a0(0) a1(1) b0(1) b1(0) c(1) => S(a0,b0,c)(0) S(a1,b1,C(a0,b0))(0) C(a1,b1)(1)",
			"a0(0) a1(1) b0(1) b1(1) c(0) => S(a0,b0,c)(1) S(a1,b1,C(a0,b0))(0) C(a1,b1)(1)",
			"a0(0) a1(1) b0(1) b1(1) c(1) => S(a0,b0,c)(0) S(a1,b1,C(a0,b0))(1) C(a1,b1)(1)",
			"a0(1) a1(0) b0(0) b1(0) c(0) => S(a0,b0,c)(1) S(a1,b1,C(a0,b0))(0) C(a1,b1)(0)",
			"a0(1) a1(0) b0(0) b1(0) c(1) => S(a0,b0,c)(0) S(a1,b1,C(a0,b0))(1) C(a1,b1)(0)",
			"a0(1) a1(0) b0(0) b1(1) c(0) => S(a0,b0,c)(1) S(a1,b1,C(a0,b0))(1) C(a1,b1)(0)",
			"a0(1) a1(0) b0(0) b1(1) c(1) => S(a0,b0,c)(0) S(a1,b1,C(a0,b0))(0) C(a1,b1)(1)",
			"a0(1) a1(0) b0(1) b1(0) c(0) => S(a0,b0,c)(0) S(a1,b1,C(a0,b0))(1) C(a1,b1)(0)",
			"a0(1) a1(0) b0(1) b1(0) c(1) => S(a0,b0,c)(1) S(a1,b1,C(a0,b0))(1) C(a1,b1)(0)",
			"a0(1) a1(0) b0(1) b1(1) c(0) => S(a0,b0,c)(0) S(a1,b1,C(a0,b0))(0) C(a1,b1)(1)",
			"a0(1) a1(0) b0(1) b1(1) c(1) => S(a0,b0,c)(1) S(a1,b1,C(a0,b0))(0) C(a1,b1)(1)",
			"a0(1) a1(1) b0(0) b1(0) c(0) => S(a0,b0,c)(1) S(a1,b1,C(a0,b0))(1) C(a1,b1)(0)",
			"a0(1) a1(1) b0(0) b1(0) c(1) => S(a0,b0,c)(0) S(a1,b1,C(a0,b0))(0) C(a1,b1)(1)",
			"a0(1) a1(1) b0(0) b1(1) c(0) => S(a0,b0,c)(1) S(a1,b1,C(a0,b0))(0) C(a1,b1)(1)",
			"a0(1) a1(1) b0(0) b1(1) c(1) => S(a0,b0,c)(0) S(a1,b1,C(a0,b0))(1) C(a1,b1)(1)",
			"a0(1) a1(1) b0(1) b1(0) c(0) => S(a0,b0,c)(0) S(a1,b1,C(a0,b0))(0) C(a1,b1)(1)",
			"a0(1) a1(1) b0(1) b1(0) c(1) => S(a0,b0,c)(1) S(a1,b1,C(a0,b0))(0) C(a1,b1)(1)",
			"a0(1) a1(1) b0(1) b1(1) c(0) => S(a0,b0,c)(0) S(a1,b1,C(a0,b0))(1) C(a1,b1)(1)",
			"a0(1) a1(1) b0(1) b1(1) c(1) => S(a0,b0,c)(1) S(a1,b1,C(a0,b0))(1) C(a1,b1)(1)",
		},
		isValidInt: func(inputs map[string]int) []int {
			sum0 := inputs["a0"] + inputs["b0"] + inputs["c"]
			sum1 := sum0/2 + inputs["a1"] + inputs["b1"]
			return []int{sum0 % 2, sum1 % 2, sum1 / 2}
		},
	}, {
		name: "SumN",
		desc: "a0 a1 b0 b1 c => S(a0,b0,c) S(a1,b1,C(a0,b0)) C(a1,b1)",
		want: []string{
			"a0(0) a1(0) b0(0) b1(0) c(0) => S(a0,b0,c)(0) S(a1,b1,C(a0,b0))(0) C(a1,b1)(0)",
			"a0(0) a1(0) b0(0) b1(0) c(1) => S(a0,b0,c)(1) S(a1,b1,C(a0,b0))(0) C(a1,b1)(0)",
			"a0(0) a1(0) b0(0) b1(1) c(0) => S(a0,b0,c)(0) S(a1,b1,C(a0,b0))(1) C(a1,b1)(0)",
			"a0(0) a1(0) b0(0) b1(1) c(1) => S(a0,b0,c)(1) S(a1,b1,C(a0,b0))(1) C(a1,b1)(0)",
			"a0(0) a1(0) b0(1) b1(0) c(0) => S(a0,b0,c)(1) S(a1,b1,C(a0,b0))(0) C(a1,b1)(0)",
			"a0(0) a1(0) b0(1) b1(0) c(1) => S(a0,b0,c)(0) S(a1,b1,C(a0,b0))(1) C(a1,b1)(0)",
			"a0(0) a1(0) b0(1) b1(1) c(0) => S(a0,b0,c)(1) S(a1,b1,C(a0,b0))(1) C(a1,b1)(0)",
			"a0(0) a1(0) b0(1) b1(1) c(1) => S(a0,b0,c)(0) S(a1,b1,C(a0,b0))(0) C(a1,b1)(1)",
			"a0(0) a1(1) b0(0) b1(0) c(0) => S(a0,b0,c)(0) S(a1,b1,C(a0,b0))(1) C(a1,b1)(0)",
			"a0(0) a1(1) b0(0) b1(0) c(1) => S(a0,b0,c)(1) S(a1,b1,C(a0,b0))(1) C(a1,b1)(0)",
			"a0(0) a1(1) b0(0) b1(1) c(0) => S(a0,b0,c)(0) S(a1,b1,C(a0,b0))(0) C(a1,b1)(1)",
			"a0(0) a1(1) b0(0) b1(1) c(1) => S(a0,b0,c)(1) S(a1,b1,C(a0,b0))(0) C(a1,b1)(1)",
			"a0(0) a1(1) b0(1) b1(0) c(0) => S(a0,b0,c)(1) S(a1,b1,C(a0,b0))(1) C(a1,b1)(0)",
			"a0(0) a1(1) b0(1) b1(0) c(1) => S(a0,b0,c)(0) S(a1,b1,C(a0,b0))(0) C(a1,b1)(1)",
			"a0(0) a1(1) b0(1) b1(1) c(0) => S(a0,b0,c)(1) S(a1,b1,C(a0,b0))(0) C(a1,b1)(1)",
			"a0(0) a1(1) b0(1) b1(1) c(1) => S(a0,b0,c)(0) S(a1,b1,C(a0,b0))(1) C(a1,b1)(1)",
			"a0(1) a1(0) b0(0) b1(0) c(0) => S(a0,b0,c)(1) S(a1,b1,C(a0,b0))(0) C(a1,b1)(0)",
			"a0(1) a1(0) b0(0) b1(0) c(1) => S(a0,b0,c)(0) S(a1,b1,C(a0,b0))(1) C(a1,b1)(0)",
			"a0(1) a1(0) b0(0) b1(1) c(0) => S(a0,b0,c)(1) S(a1,b1,C(a0,b0))(1) C(a1,b1)(0)",
			"a0(1) a1(0) b0(0) b1(1) c(1) => S(a0,b0,c)(0) S(a1,b1,C(a0,b0))(0) C(a1,b1)(1)",
			"a0(1) a1(0) b0(1) b1(0) c(0) => S(a0,b0,c)(0) S(a1,b1,C(a0,b0))(1) C(a1,b1)(0)",
			"a0(1) a1(0) b0(1) b1(0) c(1) => S(a0,b0,c)(1) S(a1,b1,C(a0,b0))(1) C(a1,b1)(0)",
			"a0(1) a1(0) b0(1) b1(1) c(0) => S(a0,b0,c)(0) S(a1,b1,C(a0,b0))(0) C(a1,b1)(1)",
			"a0(1) a1(0) b0(1) b1(1) c(1) => S(a0,b0,c)(1) S(a1,b1,C(a0,b0))(0) C(a1,b1)(1)",
			"a0(1) a1(1) b0(0) b1(0) c(0) => S(a0,b0,c)(1) S(a1,b1,C(a0,b0))(1) C(a1,b1)(0)",
			"a0(1) a1(1) b0(0) b1(0) c(1) => S(a0,b0,c)(0) S(a1,b1,C(a0,b0))(0) C(a1,b1)(1)",
			"a0(1) a1(1) b0(0) b1(1) c(0) => S(a0,b0,c)(1) S(a1,b1,C(a0,b0))(0) C(a1,b1)(1)",
			"a0(1) a1(1) b0(0) b1(1) c(1) => S(a0,b0,c)(0) S(a1,b1,C(a0,b0))(1) C(a1,b1)(1)",
			"a0(1) a1(1) b0(1) b1(0) c(0) => S(a0,b0,c)(0) S(a1,b1,C(a0,b0))(0) C(a1,b1)(1)",
			"a0(1) a1(1) b0(1) b1(0) c(1) => S(a0,b0,c)(1) S(a1,b1,C(a0,b0))(0) C(a1,b1)(1)",
			"a0(1) a1(1) b0(1) b1(1) c(0) => S(a0,b0,c)(0) S(a1,b1,C(a0,b0))(1) C(a1,b1)(1)",
			"a0(1) a1(1) b0(1) b1(1) c(1) => S(a0,b0,c)(1) S(a1,b1,C(a0,b0))(1) C(a1,b1)(1)",
		},
		isValidInt: func(inputs map[string]int) []int {
			sum0 := inputs["a0"] + inputs["b0"] + inputs["c"]
			sum1 := sum0/2 + inputs["a1"] + inputs["b1"]
			return []int{sum0 % 2, sum1 % 2, sum1 / 2}
		},
	}, {
		name: "SRLatch",
		desc: "s r => q nq",
		want: []string{"s(0) r(0) => q(1) nq(0)", "s(0) r(1) => q(0) nq(1)", "s(1) r(0) => q(1) nq(0)", "s(1) r(1) => q(0) nq(0)"},
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
		desc: "s r e => q nq",
		want: []string{
			"s(0) r(0) e(0) => q(1) nq(0)", "s(0) r(0) e(1) => q(1) nq(0)", "s(0) r(1) e(0) => q(1) nq(0)", "s(0) r(1) e(1) => q(0) nq(1)",
			"s(1) r(0) e(0) => q(0) nq(1)", "s(1) r(0) e(1) => q(1) nq(0)", "s(1) r(1) e(0) => q(1) nq(0)", "s(1) r(1) e(1) => q(0) nq(0)",
		},
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
		desc: "d e => q nq",
		want: []string{"d(0) e(0) => q(1) nq(0)", "d(0) e(1) => q(0) nq(1)", "d(1) e(0) => q(0) nq(1)", "d(1) e(1) => q(1) nq(0)"},
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
		name: "Register",
		desc: "d i o => r(d,i,o) R(d,i,o)",
		want: []string{
			"d(0) i(0) o(0) => r(d,i,o)(1) R(d,i,o)(0)", "d(0) i(0) o(1) => r(d,i,o)(1) R(d,i,o)(1)",
			"d(0) i(1) o(0) => r(d,i,o)(0) R(d,i,o)(0)", "d(0) i(1) o(1) => r(d,i,o)(0) R(d,i,o)(0)",
			"d(1) i(0) o(0) => r(d,i,o)(0) R(d,i,o)(0)", "d(1) i(0) o(1) => r(d,i,o)(0) R(d,i,o)(0)",
			"d(1) i(1) o(0) => r(d,i,o)(1) R(d,i,o)(0)", "d(1) i(1) o(1) => r(d,i,o)(1) R(d,i,o)(1)",
		},
		isValidBool: func() func(inputs map[string]bool) []bool {
			q := true
			return func(inputs map[string]bool) []bool {
				if inputs["i"] {
					q = inputs["d"]
				}
				return []bool{q, inputs["o"] && q}
			}
		}(),
	}, {
		name: "Register2",
		desc: "d0 d1 i o => r(d0,i,o) R(d0,i,o) r(d1,i,o) R(d1,i,o)",
		want: []string{
			"d0(0) d1(0) i(0) o(0) => r(d0,i,o)(1) R(d0,i,o)(0) r(d1,i,o)(1) R(d1,i,o)(0)",
			"d0(0) d1(0) i(0) o(1) => r(d0,i,o)(1) R(d0,i,o)(1) r(d1,i,o)(1) R(d1,i,o)(1)",
			"d0(0) d1(0) i(1) o(0) => r(d0,i,o)(0) R(d0,i,o)(0) r(d1,i,o)(0) R(d1,i,o)(0)",
			"d0(0) d1(0) i(1) o(1) => r(d0,i,o)(0) R(d0,i,o)(0) r(d1,i,o)(0) R(d1,i,o)(0)",
			"d0(0) d1(1) i(0) o(0) => r(d0,i,o)(0) R(d0,i,o)(0) r(d1,i,o)(0) R(d1,i,o)(0)",
			"d0(0) d1(1) i(0) o(1) => r(d0,i,o)(0) R(d0,i,o)(0) r(d1,i,o)(0) R(d1,i,o)(0)",
			"d0(0) d1(1) i(1) o(0) => r(d0,i,o)(0) R(d0,i,o)(0) r(d1,i,o)(1) R(d1,i,o)(0)",
			"d0(0) d1(1) i(1) o(1) => r(d0,i,o)(0) R(d0,i,o)(0) r(d1,i,o)(1) R(d1,i,o)(1)",
			"d0(1) d1(0) i(0) o(0) => r(d0,i,o)(0) R(d0,i,o)(0) r(d1,i,o)(1) R(d1,i,o)(0)",
			"d0(1) d1(0) i(0) o(1) => r(d0,i,o)(0) R(d0,i,o)(0) r(d1,i,o)(1) R(d1,i,o)(1)",
			"d0(1) d1(0) i(1) o(0) => r(d0,i,o)(1) R(d0,i,o)(0) r(d1,i,o)(0) R(d1,i,o)(0)",
			"d0(1) d1(0) i(1) o(1) => r(d0,i,o)(1) R(d0,i,o)(1) r(d1,i,o)(0) R(d1,i,o)(0)",
			"d0(1) d1(1) i(0) o(0) => r(d0,i,o)(1) R(d0,i,o)(0) r(d1,i,o)(0) R(d1,i,o)(0)",
			"d0(1) d1(1) i(0) o(1) => r(d0,i,o)(1) R(d0,i,o)(1) r(d1,i,o)(0) R(d1,i,o)(0)",
			"d0(1) d1(1) i(1) o(0) => r(d0,i,o)(1) R(d0,i,o)(0) r(d1,i,o)(1) R(d1,i,o)(0)",
			"d0(1) d1(1) i(1) o(1) => r(d0,i,o)(1) R(d0,i,o)(1) r(d1,i,o)(1) R(d1,i,o)(1)",
		},
		isValidBool: func() func(inputs map[string]bool) []bool {
			q0, q1 := true, true
			return func(inputs map[string]bool) []bool {
				if inputs["i"] {
					q0, q1 = inputs["d0"], inputs["d1"]
				}
				eo := inputs["o"]
				return []bool{q0, eo && q0, q1, eo && q1}
			}
		}(),
	}, {
		name: "RegisterN",
		desc: "d0 d1 i o => r(d0,i,o) R(d0,i,o) r(d1,i,o) R(d1,i,o)",
		want: []string{
			"d0(0) d1(0) i(0) o(0) => r(d0,i,o)(1) R(d0,i,o)(0) r(d1,i,o)(1) R(d1,i,o)(0)",
			"d0(0) d1(0) i(0) o(1) => r(d0,i,o)(1) R(d0,i,o)(1) r(d1,i,o)(1) R(d1,i,o)(1)",
			"d0(0) d1(0) i(1) o(0) => r(d0,i,o)(0) R(d0,i,o)(0) r(d1,i,o)(0) R(d1,i,o)(0)",
			"d0(0) d1(0) i(1) o(1) => r(d0,i,o)(0) R(d0,i,o)(0) r(d1,i,o)(0) R(d1,i,o)(0)",
			"d0(0) d1(1) i(0) o(0) => r(d0,i,o)(0) R(d0,i,o)(0) r(d1,i,o)(0) R(d1,i,o)(0)",
			"d0(0) d1(1) i(0) o(1) => r(d0,i,o)(0) R(d0,i,o)(0) r(d1,i,o)(0) R(d1,i,o)(0)",
			"d0(0) d1(1) i(1) o(0) => r(d0,i,o)(0) R(d0,i,o)(0) r(d1,i,o)(1) R(d1,i,o)(0)",
			"d0(0) d1(1) i(1) o(1) => r(d0,i,o)(0) R(d0,i,o)(0) r(d1,i,o)(1) R(d1,i,o)(1)",
			"d0(1) d1(0) i(0) o(0) => r(d0,i,o)(0) R(d0,i,o)(0) r(d1,i,o)(1) R(d1,i,o)(0)",
			"d0(1) d1(0) i(0) o(1) => r(d0,i,o)(0) R(d0,i,o)(0) r(d1,i,o)(1) R(d1,i,o)(1)",
			"d0(1) d1(0) i(1) o(0) => r(d0,i,o)(1) R(d0,i,o)(0) r(d1,i,o)(0) R(d1,i,o)(0)",
			"d0(1) d1(0) i(1) o(1) => r(d0,i,o)(1) R(d0,i,o)(1) r(d1,i,o)(0) R(d1,i,o)(0)",
			"d0(1) d1(1) i(0) o(0) => r(d0,i,o)(1) R(d0,i,o)(0) r(d1,i,o)(0) R(d1,i,o)(0)",
			"d0(1) d1(1) i(0) o(1) => r(d0,i,o)(1) R(d0,i,o)(1) r(d1,i,o)(0) R(d1,i,o)(0)",
			"d0(1) d1(1) i(1) o(0) => r(d0,i,o)(1) R(d0,i,o)(0) r(d1,i,o)(1) R(d1,i,o)(0)",
			"d0(1) d1(1) i(1) o(1) => r(d0,i,o)(1) R(d0,i,o)(1) r(d1,i,o)(1) R(d1,i,o)(1)",
		},
		isValidBool: func() func(inputs map[string]bool) []bool {
			q0, q1 := true, true
			return func(inputs map[string]bool) []bool {
				if inputs["i"] {
					q0, q1 = inputs["d0"], inputs["d1"]
				}
				o := inputs["o"]
				return []bool{q0, o && q0, q1, o && q1}
			}
		}(),
	}, {
		name: "Alu",
		desc: "a ai ao b bi bo ri ro c => R(a,ai,ao) R(b,bi,bo) R(S(a,b)) C(a,b)",
		want: []string{
			"a(1) ai(1) ao(0) b(1) bi(1) bo(0) ri(1) ro(0) c(1) => R(a,ai,ao)(0) R(b,bi,bo)(0) R(S(a,b))(0) C(a,b)(1)",
			"a(1) ai(1) ao(0) b(0) bi(0) bo(0) ri(1) ro(1) c(0) => R(a,ai,ao)(0) R(b,bi,bo)(0) R(S(a,b))(0) C(a,b)(1)",
			"a(0) ai(0) ao(0) b(1) bi(0) bo(1) ri(1) ro(0) c(1) => R(a,ai,ao)(0) R(b,bi,bo)(1) R(S(a,b))(0) C(a,b)(1)",
			"a(0) ai(0) ao(0) b(0) bi(0) bo(0) ri(0) ro(1) c(0) => R(a,ai,ao)(0) R(b,bi,bo)(0) R(S(a,b))(1) C(a,b)(1)",
			"a(1) ai(1) ao(0) b(1) bi(1) bo(1) ri(1) ro(0) c(1) => R(a,ai,ao)(0) R(b,bi,bo)(1) R(S(a,b))(0) C(a,b)(1)",
			"a(0) ai(0) ao(0) b(1) bi(0) bo(0) ri(0) ro(1) c(0) => R(a,ai,ao)(0) R(b,bi,bo)(0) R(S(a,b))(1) C(a,b)(1)",
			"a(0) ai(0) ao(0) b(1) bi(0) bo(1) ri(1) ro(1) c(0) => R(a,ai,ao)(0) R(b,bi,bo)(1) R(S(a,b))(0) C(a,b)(1)",
			"a(1) ai(1) ao(0) b(0) bi(1) bo(0) ri(1) ro(1) c(0) => R(a,ai,ao)(0) R(b,bi,bo)(0) R(S(a,b))(1) C(a,b)(0)",
			"a(1) ai(0) ao(0) b(0) bi(0) bo(0) ri(1) ro(0) c(1) => R(a,ai,ao)(0) R(b,bi,bo)(0) R(S(a,b))(0) C(a,b)(1)",
			"a(0) ai(0) ao(1) b(1) bi(0) bo(1) ri(1) ro(1) c(1) => R(a,ai,ao)(1) R(b,bi,bo)(0) R(S(a,b))(0) C(a,b)(1)",
		},
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
				return []int{inputs["ao"] & qa, inputs["bo"] & qb, inputs["ro"] & qr, sum / 2}
			}
		}(),
	}, {
		name: "Alu2",
		desc: "a0 a1 ai ao b0 b1 bi bo ri ro c => R(a0,ai,ao) R(b0,bi,bo) R(S(a0,b0)) R(a1,ai,ao) R(b1,bi,bo) R(S(a1,b1)) C(a1,b1)",
		want: []string{
			"a0(1) a1(1) ai(0) ao(1) b0(1) b1(0) bi(1) bo(0) ri(1) ro(1) c(1) => R(a0,ai,ao)(1) R(b0,bi,bo)(0) R(S(a0,b0))(1) R(a1,ai,ao)(1) R(b1,bi,bo)(0) R(S(a1,b1))(0) C(a1,b1)(1)",
			"a0(0) a1(0) ai(0) ao(0) b0(1) b1(1) bi(0) bo(0) ri(0) ro(0) c(1) => R(a0,ai,ao)(0) R(b0,bi,bo)(0) R(S(a0,b0))(0) R(a1,ai,ao)(0) R(b1,bi,bo)(0) R(S(a1,b1))(0) C(a1,b1)(1)",
			"a0(0) a1(1) ai(1) ao(0) b0(1) b1(0) bi(0) bo(0) ri(0) ro(0) c(0) => R(a0,ai,ao)(0) R(b0,bi,bo)(0) R(S(a0,b0))(0) R(a1,ai,ao)(0) R(b1,bi,bo)(0) R(S(a1,b1))(0) C(a1,b1)(0)",
			"a0(0) a1(1) ai(0) ao(1) b0(1) b1(0) bi(1) bo(1) ri(1) ro(1) c(0) => R(a0,ai,ao)(0) R(b0,bi,bo)(1) R(S(a0,b0))(1) R(a1,ai,ao)(1) R(b1,bi,bo)(0) R(S(a1,b1))(1) C(a1,b1)(0)",
			"a0(1) a1(0) ai(0) ao(0) b0(1) b1(0) bi(0) bo(0) ri(1) ro(0) c(0) => R(a0,ai,ao)(0) R(b0,bi,bo)(0) R(S(a0,b0))(0) R(a1,ai,ao)(0) R(b1,bi,bo)(0) R(S(a1,b1))(0) C(a1,b1)(0)",
			"a0(0) a1(0) ai(1) ao(0) b0(1) b1(1) bi(1) bo(0) ri(1) ro(1) c(0) => R(a0,ai,ao)(0) R(b0,bi,bo)(0) R(S(a0,b0))(1) R(a1,ai,ao)(0) R(b1,bi,bo)(0) R(S(a1,b1))(1) C(a1,b1)(0)",
			"a0(0) a1(1) ai(0) ao(1) b0(1) b1(0) bi(1) bo(0) ri(0) ro(0) c(0) => R(a0,ai,ao)(0) R(b0,bi,bo)(0) R(S(a0,b0))(0) R(a1,ai,ao)(0) R(b1,bi,bo)(0) R(S(a1,b1))(0) C(a1,b1)(0)",
			"a0(0) a1(1) ai(0) ao(1) b0(0) b1(0) bi(1) bo(1) ri(0) ro(1) c(1) => R(a0,ai,ao)(0) R(b0,bi,bo)(0) R(S(a0,b0))(1) R(a1,ai,ao)(0) R(b1,bi,bo)(0) R(S(a1,b1))(1) C(a1,b1)(0)",
			"a0(1) a1(1) ai(1) ao(1) b0(1) b1(0) bi(1) bo(1) ri(0) ro(1) c(0) => R(a0,ai,ao)(1) R(b0,bi,bo)(1) R(S(a0,b0))(1) R(a1,ai,ao)(1) R(b1,bi,bo)(0) R(S(a1,b1))(1) C(a1,b1)(1)",
			"a0(0) a1(1) ai(0) ao(0) b0(0) b1(1) bi(0) bo(1) ri(0) ro(0) c(0) => R(a0,ai,ao)(0) R(b0,bi,bo)(1) R(S(a0,b0))(0) R(a1,ai,ao)(0) R(b1,bi,bo)(0) R(S(a1,b1))(0) C(a1,b1)(1)",
		},
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
					inputs["ao"] & qa0, inputs["bo"] & qb0, inputs["ro"] & qr0,
					inputs["ao"] & qa1, inputs["bo"] & qb1, inputs["ro"] & qr1,
					sum1 / 2,
				}
			}
		}(),
	}, {
		name: "AluN",
		desc: "a0 a1 ai ao b0 b1 bi bo ri ro c => R(a0,ai,ao) R(b0,bi,bo) R(S(a0,b0)) R(a1,ai,ao) R(b1,bi,bo) R(S(a1,b1)) C(a1,b1)",
		want: []string{
			"a0(1) a1(1) ai(0) ao(1) b0(1) b1(0) bi(1) bo(0) ri(1) ro(1) c(1) => R(a0,ai,ao)(1) R(b0,bi,bo)(0) R(S(a0,b0))(1) R(a1,ai,ao)(1) R(b1,bi,bo)(0) R(S(a1,b1))(0) C(a1,b1)(1)",
			"a0(0) a1(0) ai(0) ao(0) b0(1) b1(1) bi(0) bo(0) ri(0) ro(0) c(1) => R(a0,ai,ao)(0) R(b0,bi,bo)(0) R(S(a0,b0))(0) R(a1,ai,ao)(0) R(b1,bi,bo)(0) R(S(a1,b1))(0) C(a1,b1)(1)",
			"a0(0) a1(1) ai(1) ao(0) b0(1) b1(0) bi(0) bo(0) ri(0) ro(0) c(0) => R(a0,ai,ao)(0) R(b0,bi,bo)(0) R(S(a0,b0))(0) R(a1,ai,ao)(0) R(b1,bi,bo)(0) R(S(a1,b1))(0) C(a1,b1)(0)",
			"a0(0) a1(1) ai(0) ao(1) b0(1) b1(0) bi(1) bo(1) ri(1) ro(1) c(0) => R(a0,ai,ao)(0) R(b0,bi,bo)(1) R(S(a0,b0))(1) R(a1,ai,ao)(1) R(b1,bi,bo)(0) R(S(a1,b1))(1) C(a1,b1)(0)",
			"a0(1) a1(0) ai(0) ao(0) b0(1) b1(0) bi(0) bo(0) ri(1) ro(0) c(0) => R(a0,ai,ao)(0) R(b0,bi,bo)(0) R(S(a0,b0))(0) R(a1,ai,ao)(0) R(b1,bi,bo)(0) R(S(a1,b1))(0) C(a1,b1)(0)",
			"a0(0) a1(0) ai(1) ao(0) b0(1) b1(1) bi(1) bo(0) ri(1) ro(1) c(0) => R(a0,ai,ao)(0) R(b0,bi,bo)(0) R(S(a0,b0))(1) R(a1,ai,ao)(0) R(b1,bi,bo)(0) R(S(a1,b1))(1) C(a1,b1)(0)",
			"a0(0) a1(1) ai(0) ao(1) b0(1) b1(0) bi(1) bo(0) ri(0) ro(0) c(0) => R(a0,ai,ao)(0) R(b0,bi,bo)(0) R(S(a0,b0))(0) R(a1,ai,ao)(0) R(b1,bi,bo)(0) R(S(a1,b1))(0) C(a1,b1)(0)",
			"a0(0) a1(1) ai(0) ao(1) b0(0) b1(0) bi(1) bo(1) ri(0) ro(1) c(1) => R(a0,ai,ao)(0) R(b0,bi,bo)(0) R(S(a0,b0))(1) R(a1,ai,ao)(0) R(b1,bi,bo)(0) R(S(a1,b1))(1) C(a1,b1)(0)",
			"a0(1) a1(1) ai(1) ao(1) b0(1) b1(0) bi(1) bo(1) ri(0) ro(1) c(0) => R(a0,ai,ao)(1) R(b0,bi,bo)(1) R(S(a0,b0))(1) R(a1,ai,ao)(1) R(b1,bi,bo)(0) R(S(a1,b1))(1) C(a1,b1)(1)",
			"a0(0) a1(1) ai(0) ao(0) b0(0) b1(1) bi(0) bo(1) ri(0) ro(0) c(0) => R(a0,ai,ao)(0) R(b0,bi,bo)(1) R(S(a0,b0))(0) R(a1,ai,ao)(0) R(b1,bi,bo)(0) R(S(a1,b1))(0) C(a1,b1)(1)",
		},
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
					inputs["ao"] & qa0, inputs["bo"] & qb0, inputs["ro"] & qr0,
					inputs["ao"] & qa1, inputs["bo"] & qb1, inputs["ro"] & qr1,
					sum1 / 2,
				}
			}
		}(),
	}, {
		name: "Bus",
		desc: "d ar br r => B(d) aw bw",
		want: []string{
			"d(0) ar(0) br(0) r(0) => B(d)(0) aw(0) bw(0)",
			"d(0) ar(0) br(0) r(1) => B(d)(1) aw(1) bw(1)",
			"d(0) ar(0) br(1) r(0) => B(d)(1) aw(1) bw(1)",
			"d(0) ar(0) br(1) r(1) => B(d)(1) aw(1) bw(1)",
			"d(0) ar(1) br(0) r(0) => B(d)(1) aw(1) bw(1)",
			"d(0) ar(1) br(0) r(1) => B(d)(1) aw(1) bw(1)",
			"d(0) ar(1) br(1) r(0) => B(d)(1) aw(1) bw(1)",
			"d(0) ar(1) br(1) r(1) => B(d)(1) aw(1) bw(1)",
			"d(1) ar(0) br(0) r(0) => B(d)(1) aw(1) bw(1)",
			"d(1) ar(0) br(0) r(1) => B(d)(1) aw(1) bw(1)",
			"d(1) ar(0) br(1) r(0) => B(d)(1) aw(1) bw(1)",
			"d(1) ar(0) br(1) r(1) => B(d)(1) aw(1) bw(1)",
			"d(1) ar(1) br(0) r(0) => B(d)(1) aw(1) bw(1)",
			"d(1) ar(1) br(0) r(1) => B(d)(1) aw(1) bw(1)",
			"d(1) ar(1) br(1) r(0) => B(d)(1) aw(1) bw(1)",
			"d(1) ar(1) br(1) r(1) => B(d)(1) aw(1) bw(1)",
		},
		isValidBool: func(inputs map[string]bool) []bool {
			bus := inputs["d"] || inputs["ar"] || inputs["br"] || inputs["r"]
			return []bool{bus, bus, bus}
		},
	}, {
		name: "Bus2",
		desc: "d0 d1 ar0 ar1 br0 br1 r0 r1 => B(d0) B(d1) aw0 aw1 bw0 bw1",
		want: []string{
			"d0(1) d1(1) ar0(0) ar1(1) br0(1) br1(0) r0(1) r1(0) => B(d0)(1) B(d1)(1) aw0(1) aw1(1) bw0(1) bw1(1)",
			"d0(1) d1(1) ar0(1) ar1(0) br0(0) br1(0) r0(0) r1(1) => B(d0)(1) B(d1)(1) aw0(1) aw1(1) bw0(1) bw1(1)",
			"d0(1) d1(0) ar0(0) ar1(0) br0(0) br1(1) r0(0) r1(1) => B(d0)(1) B(d1)(1) aw0(1) aw1(1) bw0(1) bw1(1)",
			"d0(1) d1(0) ar0(1) ar1(0) br0(0) br1(0) r0(0) r1(0) => B(d0)(1) B(d1)(0) aw0(1) aw1(0) bw0(1) bw1(0)",
			"d0(0) d1(0) ar0(1) ar1(0) br0(1) br1(1) r0(0) r1(1) => B(d0)(1) B(d1)(1) aw0(1) aw1(1) bw0(1) bw1(1)",
			"d0(1) d1(1) ar0(1) ar1(0) br0(1) br1(0) r0(0) r1(0) => B(d0)(1) B(d1)(1) aw0(1) aw1(1) bw0(1) bw1(1)",
			"d0(1) d1(0) ar0(0) ar1(0) br0(1) br1(0) r0(0) r1(0) => B(d0)(1) B(d1)(0) aw0(1) aw1(0) bw0(1) bw1(0)",
			"d0(0) d1(1) ar0(0) ar1(1) br0(1) br1(1) r0(0) r1(1) => B(d0)(1) B(d1)(1) aw0(1) aw1(1) bw0(1) bw1(1)",
			"d0(1) d1(0) ar0(0) ar1(1) br0(0) br1(1) r0(1) r1(0) => B(d0)(1) B(d1)(1) aw0(1) aw1(1) bw0(1) bw1(1)",
			"d0(1) d1(0) ar0(0) ar1(0) br0(0) br1(0) r0(1) r1(0) => B(d0)(1) B(d1)(0) aw0(1) aw1(0) bw0(1) bw1(0)",
		},
		isValidBool: func(inputs map[string]bool) []bool {
			bus0 := inputs["d0"] || inputs["ar0"] || inputs["br0"] || inputs["r0"]
			bus1 := inputs["d1"] || inputs["ar1"] || inputs["br1"] || inputs["r1"]
			return []bool{bus0, bus1, bus0, bus1, bus0, bus1}
		},
	}, {
		name: "BusN",
		desc: "d0 d1 ar0 ar1 br0 br1 r0 r1 => B(d0) B(d1) aw0 aw1 bw0 bw1",
		want: []string{
			"d0(1) d1(1) ar0(0) ar1(1) br0(1) br1(0) r0(1) r1(0) => B(d0)(1) B(d1)(1) aw0(1) aw1(1) bw0(1) bw1(1)",
			"d0(1) d1(1) ar0(1) ar1(0) br0(0) br1(0) r0(0) r1(1) => B(d0)(1) B(d1)(1) aw0(1) aw1(1) bw0(1) bw1(1)",
			"d0(1) d1(0) ar0(0) ar1(0) br0(0) br1(1) r0(0) r1(1) => B(d0)(1) B(d1)(1) aw0(1) aw1(1) bw0(1) bw1(1)",
			"d0(1) d1(0) ar0(1) ar1(0) br0(0) br1(0) r0(0) r1(0) => B(d0)(1) B(d1)(0) aw0(1) aw1(0) bw0(1) bw1(0)",
			"d0(0) d1(0) ar0(1) ar1(0) br0(1) br1(1) r0(0) r1(1) => B(d0)(1) B(d1)(1) aw0(1) aw1(1) bw0(1) bw1(1)",
			"d0(1) d1(1) ar0(1) ar1(0) br0(1) br1(0) r0(0) r1(0) => B(d0)(1) B(d1)(1) aw0(1) aw1(1) bw0(1) bw1(1)",
			"d0(1) d1(0) ar0(0) ar1(0) br0(1) br1(0) r0(0) r1(0) => B(d0)(1) B(d1)(0) aw0(1) aw1(0) bw0(1) bw1(0)",
			"d0(0) d1(1) ar0(0) ar1(1) br0(1) br1(1) r0(0) r1(1) => B(d0)(1) B(d1)(1) aw0(1) aw1(1) bw0(1) bw1(1)",
			"d0(1) d1(0) ar0(0) ar1(1) br0(0) br1(1) r0(1) r1(0) => B(d0)(1) B(d1)(1) aw0(1) aw1(1) bw0(1) bw1(1)",
			"d0(1) d1(0) ar0(0) ar1(0) br0(0) br1(0) r0(1) r1(0) => B(d0)(1) B(d1)(0) aw0(1) aw1(0) bw0(1) bw1(0)",
		},
		isValidBool: func(inputs map[string]bool) []bool {
			bus0 := inputs["d0"] || inputs["ar0"] || inputs["br0"] || inputs["r0"]
			bus1 := inputs["d1"] || inputs["ar1"] || inputs["br1"] || inputs["r1"]
			return []bool{bus0, bus1, bus0, bus1, bus0, bus1}
		},
	}, {
		name: "BusIOn",
		desc: "d ar br r => B(d) aw bw",
		want: []string{
			"d(0) ar(0) br(0) r(0) => B(d)(0) aw(0) bw(0)",
			"d(0) ar(0) br(0) r(1) => B(d)(1) aw(1) bw(1)",
			"d(0) ar(0) br(1) r(0) => B(d)(1) aw(1) bw(1)",
			"d(0) ar(0) br(1) r(1) => B(d)(1) aw(1) bw(1)",
			"d(0) ar(1) br(0) r(0) => B(d)(1) aw(1) bw(1)",
			"d(0) ar(1) br(0) r(1) => B(d)(1) aw(1) bw(1)",
			"d(0) ar(1) br(1) r(0) => B(d)(1) aw(1) bw(1)",
			"d(0) ar(1) br(1) r(1) => B(d)(1) aw(1) bw(1)",
			"d(1) ar(0) br(0) r(0) => B(d)(1) aw(1) bw(1)",
			"d(1) ar(0) br(0) r(1) => B(d)(1) aw(1) bw(1)",
			"d(1) ar(0) br(1) r(0) => B(d)(1) aw(1) bw(1)",
			"d(1) ar(0) br(1) r(1) => B(d)(1) aw(1) bw(1)",
			"d(1) ar(1) br(0) r(0) => B(d)(1) aw(1) bw(1)",
			"d(1) ar(1) br(0) r(1) => B(d)(1) aw(1) bw(1)",
			"d(1) ar(1) br(1) r(0) => B(d)(1) aw(1) bw(1)",
			"d(1) ar(1) br(1) r(1) => B(d)(1) aw(1) bw(1)",
		},
		isValidBool: func(inputs map[string]bool) []bool {
			bus := inputs["d"] || inputs["ar"] || inputs["br"] || inputs["r"]
			return []bool{bus, bus, bus}
		},
	}, {
		name: "BusBnIOn",
		desc: "d0 d1 ar0 ar1 br0 br1 r0 r1 => B(d0) B(d1) aw0 aw1 bw0 bw1",
		want: []string{
			"d0(1) d1(1) ar0(0) ar1(1) br0(1) br1(0) r0(1) r1(0) => B(d0)(1) B(d1)(1) aw0(1) aw1(1) bw0(1) bw1(1)",
			"d0(1) d1(1) ar0(1) ar1(0) br0(0) br1(0) r0(0) r1(1) => B(d0)(1) B(d1)(1) aw0(1) aw1(1) bw0(1) bw1(1)",
			"d0(1) d1(0) ar0(0) ar1(0) br0(0) br1(1) r0(0) r1(1) => B(d0)(1) B(d1)(1) aw0(1) aw1(1) bw0(1) bw1(1)",
			"d0(1) d1(0) ar0(1) ar1(0) br0(0) br1(0) r0(0) r1(0) => B(d0)(1) B(d1)(0) aw0(1) aw1(0) bw0(1) bw1(0)",
			"d0(0) d1(0) ar0(1) ar1(0) br0(1) br1(1) r0(0) r1(1) => B(d0)(1) B(d1)(1) aw0(1) aw1(1) bw0(1) bw1(1)",
			"d0(1) d1(1) ar0(1) ar1(0) br0(1) br1(0) r0(0) r1(0) => B(d0)(1) B(d1)(1) aw0(1) aw1(1) bw0(1) bw1(1)",
			"d0(1) d1(0) ar0(0) ar1(0) br0(1) br1(0) r0(0) r1(0) => B(d0)(1) B(d1)(0) aw0(1) aw1(0) bw0(1) bw1(0)",
			"d0(0) d1(1) ar0(0) ar1(1) br0(1) br1(1) r0(0) r1(1) => B(d0)(1) B(d1)(1) aw0(1) aw1(1) bw0(1) bw1(1)",
			"d0(1) d1(0) ar0(0) ar1(1) br0(0) br1(1) r0(1) r1(0) => B(d0)(1) B(d1)(1) aw0(1) aw1(1) bw0(1) bw1(1)",
			"d0(1) d1(0) ar0(0) ar1(0) br0(0) br1(0) r0(1) r1(0) => B(d0)(1) B(d1)(0) aw0(1) aw1(0) bw0(1) bw1(0)",
		},
		isValidBool: func(inputs map[string]bool) []bool {
			bus0 := inputs["d0"] || inputs["ar0"] || inputs["br0"] || inputs["r0"]
			bus1 := inputs["d1"] || inputs["ar1"] || inputs["br1"] || inputs["r1"]
			return []bool{bus0, bus1, bus0, bus1, bus0, bus1}
		},
	}, {
		name: "AluWithBus",
		desc: "d ai ao bi bo ri ro c => B(d) R(da,ai,ao) R(db,bi,bo) R(S(r(da,ai,ao),r(db,bi,bo),c),ri,ro) C(r(da,ai,ao),r(db,bi,bo))",
		want: []string{
			"d(1) ai(0) ao(0) bi(0) bo(0) ri(1) ro(0) c(1) => B(d)(1) R(da,ai,ao)(0) R(db,bi,bo)(0) R(S(r(da,ai,ao),r(db,bi,bo),c),ri,ro)(0) C(r(da,ai,ao),r(db,bi,bo))(1)",
			"d(1) ai(0) ao(1) bi(0) bo(0) ri(0) ro(0) c(0) => B(d)(1) R(da,ai,ao)(1) R(db,bi,bo)(0) R(S(r(da,ai,ao),r(db,bi,bo),c),ri,ro)(0) C(r(da,ai,ao),r(db,bi,bo))(1)",
			"d(0) ai(0) ao(1) bi(0) bo(1) ri(1) ro(0) c(1) => B(d)(1) R(da,ai,ao)(1) R(db,bi,bo)(1) R(S(r(da,ai,ao),r(db,bi,bo),c),ri,ro)(0) C(r(da,ai,ao),r(db,bi,bo))(1)",
			"d(1) ai(0) ao(0) bi(0) bo(1) ri(0) ro(0) c(0) => B(d)(1) R(da,ai,ao)(0) R(db,bi,bo)(1) R(S(r(da,ai,ao),r(db,bi,bo),c),ri,ro)(0) C(r(da,ai,ao),r(db,bi,bo))(1)",
			"d(1) ai(0) ao(0) bi(0) bo(0) ri(0) ro(1) c(0) => B(d)(1) R(da,ai,ao)(0) R(db,bi,bo)(0) R(S(r(da,ai,ao),r(db,bi,bo),c),ri,ro)(1) C(r(da,ai,ao),r(db,bi,bo))(1)",
		},
		isValidInt: func() func(inputs map[string]int) []int {
			qa, qb, qr := 1, 1, 1
			return func(inputs map[string]int) []int {
				ra, rb, rr, d, sum := 0, 0, 0, 0, 0
				for i := 0; i < 10; i++ {
					if inputs["ao"] == 1 {
						ra = qa
					}
					if inputs["bo"] == 1 {
						rb = qb
					}
					if inputs["ro"] == 1 {
						rr = qr
					}
					d = inputs["d"] | ra | rb | rr
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
				return []int{d, ra, rb, rr, sum / 2}
			}
		}(),
	}, {
		name: "AluWithBus2",
		desc: "d0 d1 ai ao bi bo ri ro c" +
			" => B(d0) R(d0a,ai,ao) R(d0b,bi,bo) R(S(r(d0a,ai,ao),r(d0b,bi,bo),c),ri,ro)" +
			" B(d1) R(d1a,ai,ao) R(d1b,bi,bo) R(S(r(d1a,ai,ao),r(d1b,bi,bo),C(r(d0a,ai,ao),r(d0b,bi,bo))),ri,ro)" +
			" C(r(d1a,ai,ao),r(d1b,bi,bo))",
		want: []string{
			"d0(1) d1(1) ai(0) ao(1) bi(1) bo(0) ri(1) ro(0) c(1) => B(d0)(1) R(d0a,ai,ao)(1) R(d0b,bi,bo)(0) R(S(r(d0a,ai,ao),r(d0b,bi,bo),c),ri,ro)(0) B(d1)(1) R(d1a,ai,ao)(1) R(d1b,bi,bo)(0) R(S(r(d1a,ai,ao),r(d1b,bi,bo),C(r(d0a,ai,ao),r(d0b,bi,bo))),ri,ro)(0) C(r(d1a,ai,ao),r(d1b,bi,bo))(1)",
			"d0(1) d1(1) ai(0) ao(0) bi(0) bo(0) ri(1) ro(1) c(0) => B(d0)(1) R(d0a,ai,ao)(0) R(d0b,bi,bo)(0) R(S(r(d0a,ai,ao),r(d0b,bi,bo),c),ri,ro)(0) B(d1)(1) R(d1a,ai,ao)(0) R(d1b,bi,bo)(0) R(S(r(d1a,ai,ao),r(d1b,bi,bo),C(r(d0a,ai,ao),r(d0b,bi,bo))),ri,ro)(1) C(r(d1a,ai,ao),r(d1b,bi,bo))(1)",
			"d0(0) d1(0) ai(0) ao(1) bi(0) bo(1) ri(1) ro(0) c(1) => B(d0)(1) R(d0a,ai,ao)(1) R(d0b,bi,bo)(1) R(S(r(d0a,ai,ao),r(d0b,bi,bo),c),ri,ro)(0) B(d1)(1) R(d1a,ai,ao)(1) R(d1b,bi,bo)(1) R(S(r(d1a,ai,ao),r(d1b,bi,bo),C(r(d0a,ai,ao),r(d0b,bi,bo))),ri,ro)(0) C(r(d1a,ai,ao),r(d1b,bi,bo))(1)",
			"d0(0) d1(0) ai(0) ao(0) bi(0) bo(0) ri(0) ro(1) c(0) => B(d0)(1) R(d0a,ai,ao)(0) R(d0b,bi,bo)(0) R(S(r(d0a,ai,ao),r(d0b,bi,bo),c),ri,ro)(1) B(d1)(1) R(d1a,ai,ao)(0) R(d1b,bi,bo)(0) R(S(r(d1a,ai,ao),r(d1b,bi,bo),C(r(d0a,ai,ao),r(d0b,bi,bo))),ri,ro)(1) C(r(d1a,ai,ao),r(d1b,bi,bo))(1)",
			"d0(0) d1(0) ai(0) ao(1) bi(0) bo(0) ri(0) ro(1) c(0) => B(d0)(1) R(d0a,ai,ao)(1) R(d0b,bi,bo)(0) R(S(r(d0a,ai,ao),r(d0b,bi,bo),c),ri,ro)(1) B(d1)(1) R(d1a,ai,ao)(1) R(d1b,bi,bo)(0) R(S(r(d1a,ai,ao),r(d1b,bi,bo),C(r(d0a,ai,ao),r(d0b,bi,bo))),ri,ro)(1) C(r(d1a,ai,ao),r(d1b,bi,bo))(1)",
			"d0(0) d1(0) ai(0) ao(1) bi(0) bo(1) ri(1) ro(1) c(0) => B(d0)(1) R(d0a,ai,ao)(1) R(d0b,bi,bo)(1) R(S(r(d0a,ai,ao),r(d0b,bi,bo),c),ri,ro)(0) B(d1)(1) R(d1a,ai,ao)(1) R(d1b,bi,bo)(1) R(S(r(d1a,ai,ao),r(d1b,bi,bo),C(r(d0a,ai,ao),r(d0b,bi,bo))),ri,ro)(1) C(r(d1a,ai,ao),r(d1b,bi,bo))(1)",
			"d0(1) d1(0) ai(0) ao(0) bi(0) bo(0) ri(1) ro(0) c(1) => B(d0)(1) R(d0a,ai,ao)(0) R(d0b,bi,bo)(0) R(S(r(d0a,ai,ao),r(d0b,bi,bo),c),ri,ro)(0) B(d1)(0) R(d1a,ai,ao)(0) R(d1b,bi,bo)(0) R(S(r(d1a,ai,ao),r(d1b,bi,bo),C(r(d0a,ai,ao),r(d0b,bi,bo))),ri,ro)(0) C(r(d1a,ai,ao),r(d1b,bi,bo))(1)",
		},
		isValidInt: func() func(inputs map[string]int) []int {
			qa0, qb0, qr0 := 1, 1, 1
			qa1, qb1, qr1 := 1, 1, 1
			return func(inputs map[string]int) []int {
				ra0, rb0, rr0, d0, sum0 := 0, 0, 0, 0, 0
				ra1, rb1, rr1, d1, sum1 := 0, 0, 0, 0, 0
				for i := 0; i < 10; i++ {
					if inputs["ao"] == 1 {
						ra0, ra1 = qa0, qa1
					}
					if inputs["bo"] == 1 {
						rb0, rb1 = qb0, qb1
					}
					if inputs["ro"] == 1 {
						rr0, rr1 = qr0, qr1
					}
					d0 = inputs["d0"] | ra0 | rb0 | rr0
					d1 = inputs["d1"] | ra1 | rb1 | rr1
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
					d0, ra0, rb0, rr0,
					d1, ra1, rb1, rr1,
					sum1 / 2}
			}
		}(),
	}, {
		name: "AluWithBusN",
		desc: "d0 d1 ai ao bi bo ri ro c" +
			" => B(d0) R(d0a,ai,ao) R(d0b,bi,bo) R(S(r(d0a,ai,ao),r(d0b,bi,bo),c),ri,ro)" +
			" B(d1) R(d1a,ai,ao) R(d1b,bi,bo) R(S(r(d1a,ai,ao),r(d1b,bi,bo),C(r(d0a,ai,ao),r(d0b,bi,bo))),ri,ro)" +
			" C(r(d1a,ai,ao),r(d1b,bi,bo))",
		want: []string{
			"d0(1) d1(1) ai(0) ao(1) bi(1) bo(0) ri(1) ro(0) c(1) => B(d0)(1) R(d0a,ai,ao)(1) R(d0b,bi,bo)(0) R(S(r(d0a,ai,ao),r(d0b,bi,bo),c),ri,ro)(0) B(d1)(1) R(d1a,ai,ao)(1) R(d1b,bi,bo)(0) R(S(r(d1a,ai,ao),r(d1b,bi,bo),C(r(d0a,ai,ao),r(d0b,bi,bo))),ri,ro)(0) C(r(d1a,ai,ao),r(d1b,bi,bo))(1)",
			"d0(1) d1(1) ai(0) ao(0) bi(0) bo(0) ri(1) ro(1) c(0) => B(d0)(1) R(d0a,ai,ao)(0) R(d0b,bi,bo)(0) R(S(r(d0a,ai,ao),r(d0b,bi,bo),c),ri,ro)(0) B(d1)(1) R(d1a,ai,ao)(0) R(d1b,bi,bo)(0) R(S(r(d1a,ai,ao),r(d1b,bi,bo),C(r(d0a,ai,ao),r(d0b,bi,bo))),ri,ro)(1) C(r(d1a,ai,ao),r(d1b,bi,bo))(1)",
			"d0(0) d1(0) ai(0) ao(1) bi(0) bo(1) ri(1) ro(0) c(1) => B(d0)(1) R(d0a,ai,ao)(1) R(d0b,bi,bo)(1) R(S(r(d0a,ai,ao),r(d0b,bi,bo),c),ri,ro)(0) B(d1)(1) R(d1a,ai,ao)(1) R(d1b,bi,bo)(1) R(S(r(d1a,ai,ao),r(d1b,bi,bo),C(r(d0a,ai,ao),r(d0b,bi,bo))),ri,ro)(0) C(r(d1a,ai,ao),r(d1b,bi,bo))(1)",
			"d0(0) d1(0) ai(0) ao(0) bi(0) bo(0) ri(0) ro(1) c(0) => B(d0)(1) R(d0a,ai,ao)(0) R(d0b,bi,bo)(0) R(S(r(d0a,ai,ao),r(d0b,bi,bo),c),ri,ro)(1) B(d1)(1) R(d1a,ai,ao)(0) R(d1b,bi,bo)(0) R(S(r(d1a,ai,ao),r(d1b,bi,bo),C(r(d0a,ai,ao),r(d0b,bi,bo))),ri,ro)(1) C(r(d1a,ai,ao),r(d1b,bi,bo))(1)",
			"d0(0) d1(0) ai(0) ao(1) bi(0) bo(0) ri(0) ro(1) c(0) => B(d0)(1) R(d0a,ai,ao)(1) R(d0b,bi,bo)(0) R(S(r(d0a,ai,ao),r(d0b,bi,bo),c),ri,ro)(1) B(d1)(1) R(d1a,ai,ao)(1) R(d1b,bi,bo)(0) R(S(r(d1a,ai,ao),r(d1b,bi,bo),C(r(d0a,ai,ao),r(d0b,bi,bo))),ri,ro)(1) C(r(d1a,ai,ao),r(d1b,bi,bo))(1)",
			"d0(0) d1(0) ai(0) ao(1) bi(0) bo(1) ri(1) ro(1) c(0) => B(d0)(1) R(d0a,ai,ao)(1) R(d0b,bi,bo)(1) R(S(r(d0a,ai,ao),r(d0b,bi,bo),c),ri,ro)(0) B(d1)(1) R(d1a,ai,ao)(1) R(d1b,bi,bo)(1) R(S(r(d1a,ai,ao),r(d1b,bi,bo),C(r(d0a,ai,ao),r(d0b,bi,bo))),ri,ro)(1) C(r(d1a,ai,ao),r(d1b,bi,bo))(1)",
			"d0(1) d1(0) ai(0) ao(0) bi(0) bo(0) ri(1) ro(0) c(1) => B(d0)(1) R(d0a,ai,ao)(0) R(d0b,bi,bo)(0) R(S(r(d0a,ai,ao),r(d0b,bi,bo),c),ri,ro)(0) B(d1)(0) R(d1a,ai,ao)(0) R(d1b,bi,bo)(0) R(S(r(d1a,ai,ao),r(d1b,bi,bo),C(r(d0a,ai,ao),r(d0b,bi,bo))),ri,ro)(0) C(r(d1a,ai,ao),r(d1b,bi,bo))(1)",
		},
		isValidInt: func() func(inputs map[string]int) []int {
			qa0, qb0, qr0 := 1, 1, 1
			qa1, qb1, qr1 := 1, 1, 1
			return func(inputs map[string]int) []int {
				ra0, rb0, rr0, d0, sum0 := 0, 0, 0, 0, 0
				ra1, rb1, rr1, d1, sum1 := 0, 0, 0, 0, 0
				for i := 0; i < 10; i++ {
					if inputs["ao"] == 1 {
						ra0, ra1 = qa0, qa1
					}
					if inputs["bo"] == 1 {
						rb0, rb1 = qb0, qb1
					}
					if inputs["ro"] == 1 {
						rr0, rr1 = qr0, qr1
					}
					d0 = inputs["d0"] | ra0 | rb0 | rr0
					d1 = inputs["d1"] | ra1 | rb1 | rr1
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
					d0, ra0, rb0, rr0,
					d1, ra1, rb1, rr1,
					sum1 / 2}
			}
		}(),
	}, {
		name: "RAM",
		desc: "a d i o => RAM0 R(d,i0,o0) R(d,i1,o1)",
		want: []string{
			"a(0) d(0) i(0) o(0) => RAM0(0) R(d,i0,o0)(0) R(d,i1,o1)(0)",
			"a(0) d(0) i(0) o(1) => RAM0(1) R(d,i0,o0)(1) R(d,i1,o1)(0)",
			"a(0) d(0) i(1) o(0) => RAM0(0) R(d,i0,o0)(0) R(d,i1,o1)(0)",
			"a(0) d(0) i(1) o(1) => RAM0(0) R(d,i0,o0)(0) R(d,i1,o1)(0)",
			"a(0) d(1) i(0) o(0) => RAM0(0) R(d,i0,o0)(0) R(d,i1,o1)(0)",
			"a(0) d(1) i(0) o(1) => RAM0(0) R(d,i0,o0)(0) R(d,i1,o1)(0)",
			"a(0) d(1) i(1) o(0) => RAM0(0) R(d,i0,o0)(0) R(d,i1,o1)(0)",
			"a(0) d(1) i(1) o(1) => RAM0(1) R(d,i0,o0)(1) R(d,i1,o1)(0)",
			"a(1) d(0) i(0) o(0) => RAM0(0) R(d,i0,o0)(0) R(d,i1,o1)(0)",
			"a(1) d(0) i(0) o(1) => RAM0(1) R(d,i0,o0)(0) R(d,i1,o1)(1)",
			"a(1) d(0) i(1) o(0) => RAM0(0) R(d,i0,o0)(0) R(d,i1,o1)(0)",
			"a(1) d(0) i(1) o(1) => RAM0(0) R(d,i0,o0)(0) R(d,i1,o1)(0)",
			"a(1) d(1) i(0) o(0) => RAM0(0) R(d,i0,o0)(0) R(d,i1,o1)(0)",
			"a(1) d(1) i(0) o(1) => RAM0(0) R(d,i0,o0)(0) R(d,i1,o1)(0)",
			"a(1) d(1) i(1) o(0) => RAM0(0) R(d,i0,o0)(0) R(d,i1,o1)(0)",
			"a(1) d(1) i(1) o(1) => RAM0(1) R(d,i0,o0)(0) R(d,i1,o1)(1)",
		},
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
				return []bool{*r, r0, r1}
			}
		}(),
	}, {
		name: "RAMa2",
		desc: "a0 a1 d ei eo => RAM0 R(d,i0,o0) R(d,i1,o1) R(d,i2,o2) R(d,i3,o3)",
		want: []string{
			"a0(0) a1(0) d(0) ei(0) eo(0) => RAM0(0) R(d,i0,o0)(0) R(d,i1,o1)(0) R(d,i2,o2)(0) R(d,i3,o3)(0)",
			"a0(0) a1(0) d(0) ei(0) eo(1) => RAM0(1) R(d,i0,o0)(1) R(d,i1,o1)(0) R(d,i2,o2)(0) R(d,i3,o3)(0)",
			"a0(0) a1(0) d(0) ei(1) eo(0) => RAM0(0) R(d,i0,o0)(0) R(d,i1,o1)(0) R(d,i2,o2)(0) R(d,i3,o3)(0)",
			"a0(0) a1(0) d(0) ei(1) eo(1) => RAM0(0) R(d,i0,o0)(0) R(d,i1,o1)(0) R(d,i2,o2)(0) R(d,i3,o3)(0)",
			"a0(0) a1(0) d(1) ei(0) eo(0) => RAM0(0) R(d,i0,o0)(0) R(d,i1,o1)(0) R(d,i2,o2)(0) R(d,i3,o3)(0)",
			"a0(0) a1(0) d(1) ei(0) eo(1) => RAM0(0) R(d,i0,o0)(0) R(d,i1,o1)(0) R(d,i2,o2)(0) R(d,i3,o3)(0)",
			"a0(0) a1(0) d(1) ei(1) eo(0) => RAM0(0) R(d,i0,o0)(0) R(d,i1,o1)(0) R(d,i2,o2)(0) R(d,i3,o3)(0)",
			"a0(0) a1(0) d(1) ei(1) eo(1) => RAM0(1) R(d,i0,o0)(1) R(d,i1,o1)(0) R(d,i2,o2)(0) R(d,i3,o3)(0)",
			"a0(0) a1(1) d(0) ei(0) eo(0) => RAM0(0) R(d,i0,o0)(0) R(d,i1,o1)(0) R(d,i2,o2)(0) R(d,i3,o3)(0)",
			"a0(0) a1(1) d(0) ei(0) eo(1) => RAM0(1) R(d,i0,o0)(0) R(d,i1,o1)(0) R(d,i2,o2)(1) R(d,i3,o3)(0)",
			"a0(0) a1(1) d(0) ei(1) eo(0) => RAM0(0) R(d,i0,o0)(0) R(d,i1,o1)(0) R(d,i2,o2)(0) R(d,i3,o3)(0)",
			"a0(0) a1(1) d(0) ei(1) eo(1) => RAM0(0) R(d,i0,o0)(0) R(d,i1,o1)(0) R(d,i2,o2)(0) R(d,i3,o3)(0)",
			"a0(0) a1(1) d(1) ei(0) eo(0) => RAM0(0) R(d,i0,o0)(0) R(d,i1,o1)(0) R(d,i2,o2)(0) R(d,i3,o3)(0)",
			"a0(0) a1(1) d(1) ei(0) eo(1) => RAM0(0) R(d,i0,o0)(0) R(d,i1,o1)(0) R(d,i2,o2)(0) R(d,i3,o3)(0)",
			"a0(0) a1(1) d(1) ei(1) eo(0) => RAM0(0) R(d,i0,o0)(0) R(d,i1,o1)(0) R(d,i2,o2)(0) R(d,i3,o3)(0)",
			"a0(0) a1(1) d(1) ei(1) eo(1) => RAM0(1) R(d,i0,o0)(0) R(d,i1,o1)(0) R(d,i2,o2)(1) R(d,i3,o3)(0)",
			"a0(1) a1(0) d(0) ei(0) eo(0) => RAM0(0) R(d,i0,o0)(0) R(d,i1,o1)(0) R(d,i2,o2)(0) R(d,i3,o3)(0)",
			"a0(1) a1(0) d(0) ei(0) eo(1) => RAM0(1) R(d,i0,o0)(0) R(d,i1,o1)(1) R(d,i2,o2)(0) R(d,i3,o3)(0)",
			"a0(1) a1(0) d(0) ei(1) eo(0) => RAM0(0) R(d,i0,o0)(0) R(d,i1,o1)(0) R(d,i2,o2)(0) R(d,i3,o3)(0)",
			"a0(1) a1(0) d(0) ei(1) eo(1) => RAM0(0) R(d,i0,o0)(0) R(d,i1,o1)(0) R(d,i2,o2)(0) R(d,i3,o3)(0)",
			"a0(1) a1(0) d(1) ei(0) eo(0) => RAM0(0) R(d,i0,o0)(0) R(d,i1,o1)(0) R(d,i2,o2)(0) R(d,i3,o3)(0)",
			"a0(1) a1(0) d(1) ei(0) eo(1) => RAM0(0) R(d,i0,o0)(0) R(d,i1,o1)(0) R(d,i2,o2)(0) R(d,i3,o3)(0)",
			"a0(1) a1(0) d(1) ei(1) eo(0) => RAM0(0) R(d,i0,o0)(0) R(d,i1,o1)(0) R(d,i2,o2)(0) R(d,i3,o3)(0)",
			"a0(1) a1(0) d(1) ei(1) eo(1) => RAM0(1) R(d,i0,o0)(0) R(d,i1,o1)(1) R(d,i2,o2)(0) R(d,i3,o3)(0)",
			"a0(1) a1(1) d(0) ei(0) eo(0) => RAM0(0) R(d,i0,o0)(0) R(d,i1,o1)(0) R(d,i2,o2)(0) R(d,i3,o3)(0)",
			"a0(1) a1(1) d(0) ei(0) eo(1) => RAM0(1) R(d,i0,o0)(0) R(d,i1,o1)(0) R(d,i2,o2)(0) R(d,i3,o3)(1)",
			"a0(1) a1(1) d(0) ei(1) eo(0) => RAM0(0) R(d,i0,o0)(0) R(d,i1,o1)(0) R(d,i2,o2)(0) R(d,i3,o3)(0)",
			"a0(1) a1(1) d(0) ei(1) eo(1) => RAM0(0) R(d,i0,o0)(0) R(d,i1,o1)(0) R(d,i2,o2)(0) R(d,i3,o3)(0)",
			"a0(1) a1(1) d(1) ei(0) eo(0) => RAM0(0) R(d,i0,o0)(0) R(d,i1,o1)(0) R(d,i2,o2)(0) R(d,i3,o3)(0)",
			"a0(1) a1(1) d(1) ei(0) eo(1) => RAM0(0) R(d,i0,o0)(0) R(d,i1,o1)(0) R(d,i2,o2)(0) R(d,i3,o3)(0)",
			"a0(1) a1(1) d(1) ei(1) eo(0) => RAM0(0) R(d,i0,o0)(0) R(d,i1,o1)(0) R(d,i2,o2)(0) R(d,i3,o3)(0)",
			"a0(1) a1(1) d(1) ei(1) eo(1) => RAM0(1) R(d,i0,o0)(0) R(d,i1,o1)(0) R(d,i2,o2)(0) R(d,i3,o3)(1)",
		},
		isValidBool: func() func(inputs map[string]bool) []bool {
			q := []bool{true, true, true, true}
			return func(inputs map[string]bool) []bool {
				a0, a1, ei, eo := inputs["a0"], inputs["a1"], inputs["ei"], inputs["eo"]
				s := []bool{!a0 && !a1, a0 && !a1, !a0 && a1, a0 && a1}
				r := []bool{false, false, false, false}
				index := 0
				for i, si := range s {
					if si {
						index = i
						break
					}
				}
				if ei {
					q[index] = inputs["d"]
				}
				if eo {
					r[index] = q[index]
				}
				res := []bool{r[index]}
				for i := range s {
					res = append(res, r[i])
				}
				return res
			}
		}(),
	}, {
		name: "RAMb2",
		desc: "a d0 d1 ei eo => RAM0 RAM1 R(d0,i0,o0) R(d1,i0,o0) R(d0,i1,o1) R(d1,i1,o1)",
		want: []string{
			"a(0) d0(0) d1(0) ei(0) eo(0) => RAM0(0) RAM1(0) R(d0,i0,o0)(0) R(d1,i0,o0)(0) R(d0,i1,o1)(0) R(d1,i1,o1)(0)",
			"a(0) d0(0) d1(0) ei(0) eo(1) => RAM0(1) RAM1(1) R(d0,i0,o0)(1) R(d1,i0,o0)(1) R(d0,i1,o1)(0) R(d1,i1,o1)(0)",
			"a(0) d0(0) d1(0) ei(1) eo(0) => RAM0(0) RAM1(0) R(d0,i0,o0)(0) R(d1,i0,o0)(0) R(d0,i1,o1)(0) R(d1,i1,o1)(0)",
			"a(0) d0(0) d1(0) ei(1) eo(1) => RAM0(0) RAM1(0) R(d0,i0,o0)(0) R(d1,i0,o0)(0) R(d0,i1,o1)(0) R(d1,i1,o1)(0)",
			"a(0) d0(0) d1(1) ei(0) eo(0) => RAM0(0) RAM1(0) R(d0,i0,o0)(0) R(d1,i0,o0)(0) R(d0,i1,o1)(0) R(d1,i1,o1)(0)",
			"a(0) d0(0) d1(1) ei(0) eo(1) => RAM0(0) RAM1(0) R(d0,i0,o0)(0) R(d1,i0,o0)(0) R(d0,i1,o1)(0) R(d1,i1,o1)(0)",
			"a(0) d0(0) d1(1) ei(1) eo(0) => RAM0(0) RAM1(0) R(d0,i0,o0)(0) R(d1,i0,o0)(0) R(d0,i1,o1)(0) R(d1,i1,o1)(0)",
			"a(0) d0(0) d1(1) ei(1) eo(1) => RAM0(0) RAM1(1) R(d0,i0,o0)(0) R(d1,i0,o0)(1) R(d0,i1,o1)(0) R(d1,i1,o1)(0)",
			"a(0) d0(1) d1(0) ei(0) eo(0) => RAM0(0) RAM1(0) R(d0,i0,o0)(0) R(d1,i0,o0)(0) R(d0,i1,o1)(0) R(d1,i1,o1)(0)",
			"a(0) d0(1) d1(0) ei(0) eo(1) => RAM0(0) RAM1(1) R(d0,i0,o0)(0) R(d1,i0,o0)(1) R(d0,i1,o1)(0) R(d1,i1,o1)(0)",
			"a(0) d0(1) d1(0) ei(1) eo(0) => RAM0(0) RAM1(0) R(d0,i0,o0)(0) R(d1,i0,o0)(0) R(d0,i1,o1)(0) R(d1,i1,o1)(0)",
			"a(0) d0(1) d1(0) ei(1) eo(1) => RAM0(1) RAM1(0) R(d0,i0,o0)(1) R(d1,i0,o0)(0) R(d0,i1,o1)(0) R(d1,i1,o1)(0)",
			"a(0) d0(1) d1(1) ei(0) eo(0) => RAM0(0) RAM1(0) R(d0,i0,o0)(0) R(d1,i0,o0)(0) R(d0,i1,o1)(0) R(d1,i1,o1)(0)",
			"a(0) d0(1) d1(1) ei(0) eo(1) => RAM0(1) RAM1(0) R(d0,i0,o0)(1) R(d1,i0,o0)(0) R(d0,i1,o1)(0) R(d1,i1,o1)(0)",
			"a(0) d0(1) d1(1) ei(1) eo(0) => RAM0(0) RAM1(0) R(d0,i0,o0)(0) R(d1,i0,o0)(0) R(d0,i1,o1)(0) R(d1,i1,o1)(0)",
			"a(0) d0(1) d1(1) ei(1) eo(1) => RAM0(1) RAM1(1) R(d0,i0,o0)(1) R(d1,i0,o0)(1) R(d0,i1,o1)(0) R(d1,i1,o1)(0)",
			"a(1) d0(0) d1(0) ei(0) eo(0) => RAM0(0) RAM1(0) R(d0,i0,o0)(0) R(d1,i0,o0)(0) R(d0,i1,o1)(0) R(d1,i1,o1)(0)",
			"a(1) d0(0) d1(0) ei(0) eo(1) => RAM0(1) RAM1(1) R(d0,i0,o0)(0) R(d1,i0,o0)(0) R(d0,i1,o1)(1) R(d1,i1,o1)(1)",
			"a(1) d0(0) d1(0) ei(1) eo(0) => RAM0(0) RAM1(0) R(d0,i0,o0)(0) R(d1,i0,o0)(0) R(d0,i1,o1)(0) R(d1,i1,o1)(0)",
			"a(1) d0(0) d1(0) ei(1) eo(1) => RAM0(0) RAM1(0) R(d0,i0,o0)(0) R(d1,i0,o0)(0) R(d0,i1,o1)(0) R(d1,i1,o1)(0)",
			"a(1) d0(0) d1(1) ei(0) eo(0) => RAM0(0) RAM1(0) R(d0,i0,o0)(0) R(d1,i0,o0)(0) R(d0,i1,o1)(0) R(d1,i1,o1)(0)",
			"a(1) d0(0) d1(1) ei(0) eo(1) => RAM0(0) RAM1(0) R(d0,i0,o0)(0) R(d1,i0,o0)(0) R(d0,i1,o1)(0) R(d1,i1,o1)(0)",
			"a(1) d0(0) d1(1) ei(1) eo(0) => RAM0(0) RAM1(0) R(d0,i0,o0)(0) R(d1,i0,o0)(0) R(d0,i1,o1)(0) R(d1,i1,o1)(0)",
			"a(1) d0(0) d1(1) ei(1) eo(1) => RAM0(0) RAM1(1) R(d0,i0,o0)(0) R(d1,i0,o0)(0) R(d0,i1,o1)(0) R(d1,i1,o1)(1)",
			"a(1) d0(1) d1(0) ei(0) eo(0) => RAM0(0) RAM1(0) R(d0,i0,o0)(0) R(d1,i0,o0)(0) R(d0,i1,o1)(0) R(d1,i1,o1)(0)",
			"a(1) d0(1) d1(0) ei(0) eo(1) => RAM0(0) RAM1(1) R(d0,i0,o0)(0) R(d1,i0,o0)(0) R(d0,i1,o1)(0) R(d1,i1,o1)(1)",
			"a(1) d0(1) d1(0) ei(1) eo(0) => RAM0(0) RAM1(0) R(d0,i0,o0)(0) R(d1,i0,o0)(0) R(d0,i1,o1)(0) R(d1,i1,o1)(0)",
			"a(1) d0(1) d1(0) ei(1) eo(1) => RAM0(1) RAM1(0) R(d0,i0,o0)(0) R(d1,i0,o0)(0) R(d0,i1,o1)(1) R(d1,i1,o1)(0)",
			"a(1) d0(1) d1(1) ei(0) eo(0) => RAM0(0) RAM1(0) R(d0,i0,o0)(0) R(d1,i0,o0)(0) R(d0,i1,o1)(0) R(d1,i1,o1)(0)",
			"a(1) d0(1) d1(1) ei(0) eo(1) => RAM0(1) RAM1(0) R(d0,i0,o0)(0) R(d1,i0,o0)(0) R(d0,i1,o1)(1) R(d1,i1,o1)(0)",
			"a(1) d0(1) d1(1) ei(1) eo(0) => RAM0(0) RAM1(0) R(d0,i0,o0)(0) R(d1,i0,o0)(0) R(d0,i1,o1)(0) R(d1,i1,o1)(0)",
			"a(1) d0(1) d1(1) ei(1) eo(1) => RAM0(1) RAM1(1) R(d0,i0,o0)(0) R(d1,i0,o0)(0) R(d0,i1,o1)(1) R(d1,i1,o1)(1)",
		},
		isValidBool: func() func(inputs map[string]bool) []bool {
			q0, q1 := []bool{true, true}, []bool{true, true}
			return func(inputs map[string]bool) []bool {
				a, ei, eo := inputs["a"], inputs["ei"], inputs["eo"]
				s1 := a
				r0, r1 := []bool{false, false}, []bool{false, false}
				q, r := &q0, &r0
				if s1 {
					q, r = &q1, &r1
				}
				if ei {
					*q = []bool{inputs["d0"], inputs["d1"]}
				}
				if eo {
					*r = *q
				}
				return slices.Concat(*r, r0, r1)
			}
		}(),
	}, {
		name: "RAMa2b2",
		desc: "a0 a1 d0 d1 ei eo => RAM0 RAM1 R(d0,i0,o0) R(d1,i0,o0) R(d0,i1,o1) R(d1,i1,o1) R(d0,i2,o2) R(d1,i2,o2) R(d0,i3,o3) R(d1,i3,o3)",
		want: []string{
			"a0(0) a1(0) d0(0) d1(0) ei(0) eo(0) => RAM0(0) RAM1(0) R(d0,i0,o0)(0) R(d1,i0,o0)(0) R(d0,i1,o1)(0) R(d1,i1,o1)(0) R(d0,i2,o2)(0) R(d1,i2,o2)(0) R(d0,i3,o3)(0) R(d1,i3,o3)(0)",
			"a0(0) a1(0) d0(0) d1(0) ei(0) eo(1) => RAM0(1) RAM1(1) R(d0,i0,o0)(1) R(d1,i0,o0)(1) R(d0,i1,o1)(0) R(d1,i1,o1)(0) R(d0,i2,o2)(0) R(d1,i2,o2)(0) R(d0,i3,o3)(0) R(d1,i3,o3)(0)",
			"a0(0) a1(0) d0(0) d1(0) ei(1) eo(0) => RAM0(0) RAM1(0) R(d0,i0,o0)(0) R(d1,i0,o0)(0) R(d0,i1,o1)(0) R(d1,i1,o1)(0) R(d0,i2,o2)(0) R(d1,i2,o2)(0) R(d0,i3,o3)(0) R(d1,i3,o3)(0)",
			"a0(0) a1(0) d0(0) d1(0) ei(1) eo(1) => RAM0(0) RAM1(0) R(d0,i0,o0)(0) R(d1,i0,o0)(0) R(d0,i1,o1)(0) R(d1,i1,o1)(0) R(d0,i2,o2)(0) R(d1,i2,o2)(0) R(d0,i3,o3)(0) R(d1,i3,o3)(0)",
			"a0(0) a1(0) d0(0) d1(1) ei(0) eo(0) => RAM0(0) RAM1(0) R(d0,i0,o0)(0) R(d1,i0,o0)(0) R(d0,i1,o1)(0) R(d1,i1,o1)(0) R(d0,i2,o2)(0) R(d1,i2,o2)(0) R(d0,i3,o3)(0) R(d1,i3,o3)(0)",
			"a0(0) a1(0) d0(0) d1(1) ei(0) eo(1) => RAM0(0) RAM1(0) R(d0,i0,o0)(0) R(d1,i0,o0)(0) R(d0,i1,o1)(0) R(d1,i1,o1)(0) R(d0,i2,o2)(0) R(d1,i2,o2)(0) R(d0,i3,o3)(0) R(d1,i3,o3)(0)",
			"a0(0) a1(0) d0(0) d1(1) ei(1) eo(0) => RAM0(0) RAM1(0) R(d0,i0,o0)(0) R(d1,i0,o0)(0) R(d0,i1,o1)(0) R(d1,i1,o1)(0) R(d0,i2,o2)(0) R(d1,i2,o2)(0) R(d0,i3,o3)(0) R(d1,i3,o3)(0)",
			"a0(0) a1(0) d0(0) d1(1) ei(1) eo(1) => RAM0(0) RAM1(1) R(d0,i0,o0)(0) R(d1,i0,o0)(1) R(d0,i1,o1)(0) R(d1,i1,o1)(0) R(d0,i2,o2)(0) R(d1,i2,o2)(0) R(d0,i3,o3)(0) R(d1,i3,o3)(0)",
			"a0(0) a1(0) d0(1) d1(0) ei(0) eo(0) => RAM0(0) RAM1(0) R(d0,i0,o0)(0) R(d1,i0,o0)(0) R(d0,i1,o1)(0) R(d1,i1,o1)(0) R(d0,i2,o2)(0) R(d1,i2,o2)(0) R(d0,i3,o3)(0) R(d1,i3,o3)(0)",
			"a0(0) a1(0) d0(1) d1(0) ei(0) eo(1) => RAM0(0) RAM1(1) R(d0,i0,o0)(0) R(d1,i0,o0)(1) R(d0,i1,o1)(0) R(d1,i1,o1)(0) R(d0,i2,o2)(0) R(d1,i2,o2)(0) R(d0,i3,o3)(0) R(d1,i3,o3)(0)",
			"a0(0) a1(0) d0(1) d1(0) ei(1) eo(0) => RAM0(0) RAM1(0) R(d0,i0,o0)(0) R(d1,i0,o0)(0) R(d0,i1,o1)(0) R(d1,i1,o1)(0) R(d0,i2,o2)(0) R(d1,i2,o2)(0) R(d0,i3,o3)(0) R(d1,i3,o3)(0)",
			"a0(0) a1(0) d0(1) d1(0) ei(1) eo(1) => RAM0(1) RAM1(0) R(d0,i0,o0)(1) R(d1,i0,o0)(0) R(d0,i1,o1)(0) R(d1,i1,o1)(0) R(d0,i2,o2)(0) R(d1,i2,o2)(0) R(d0,i3,o3)(0) R(d1,i3,o3)(0)",
			"a0(0) a1(0) d0(1) d1(1) ei(0) eo(0) => RAM0(0) RAM1(0) R(d0,i0,o0)(0) R(d1,i0,o0)(0) R(d0,i1,o1)(0) R(d1,i1,o1)(0) R(d0,i2,o2)(0) R(d1,i2,o2)(0) R(d0,i3,o3)(0) R(d1,i3,o3)(0)",
			"a0(0) a1(0) d0(1) d1(1) ei(0) eo(1) => RAM0(1) RAM1(0) R(d0,i0,o0)(1) R(d1,i0,o0)(0) R(d0,i1,o1)(0) R(d1,i1,o1)(0) R(d0,i2,o2)(0) R(d1,i2,o2)(0) R(d0,i3,o3)(0) R(d1,i3,o3)(0)",
			"a0(0) a1(0) d0(1) d1(1) ei(1) eo(0) => RAM0(0) RAM1(0) R(d0,i0,o0)(0) R(d1,i0,o0)(0) R(d0,i1,o1)(0) R(d1,i1,o1)(0) R(d0,i2,o2)(0) R(d1,i2,o2)(0) R(d0,i3,o3)(0) R(d1,i3,o3)(0)",
			"a0(0) a1(0) d0(1) d1(1) ei(1) eo(1) => RAM0(1) RAM1(1) R(d0,i0,o0)(1) R(d1,i0,o0)(1) R(d0,i1,o1)(0) R(d1,i1,o1)(0) R(d0,i2,o2)(0) R(d1,i2,o2)(0) R(d0,i3,o3)(0) R(d1,i3,o3)(0)",
			"a0(0) a1(1) d0(0) d1(0) ei(0) eo(0) => RAM0(0) RAM1(0) R(d0,i0,o0)(0) R(d1,i0,o0)(0) R(d0,i1,o1)(0) R(d1,i1,o1)(0) R(d0,i2,o2)(0) R(d1,i2,o2)(0) R(d0,i3,o3)(0) R(d1,i3,o3)(0)",
			"a0(0) a1(1) d0(0) d1(0) ei(0) eo(1) => RAM0(1) RAM1(1) R(d0,i0,o0)(0) R(d1,i0,o0)(0) R(d0,i1,o1)(0) R(d1,i1,o1)(0) R(d0,i2,o2)(1) R(d1,i2,o2)(1) R(d0,i3,o3)(0) R(d1,i3,o3)(0)",
			"a0(0) a1(1) d0(0) d1(0) ei(1) eo(0) => RAM0(0) RAM1(0) R(d0,i0,o0)(0) R(d1,i0,o0)(0) R(d0,i1,o1)(0) R(d1,i1,o1)(0) R(d0,i2,o2)(0) R(d1,i2,o2)(0) R(d0,i3,o3)(0) R(d1,i3,o3)(0)",
			"a0(0) a1(1) d0(0) d1(0) ei(1) eo(1) => RAM0(0) RAM1(0) R(d0,i0,o0)(0) R(d1,i0,o0)(0) R(d0,i1,o1)(0) R(d1,i1,o1)(0) R(d0,i2,o2)(0) R(d1,i2,o2)(0) R(d0,i3,o3)(0) R(d1,i3,o3)(0)",
			"a0(0) a1(1) d0(0) d1(1) ei(0) eo(0) => RAM0(0) RAM1(0) R(d0,i0,o0)(0) R(d1,i0,o0)(0) R(d0,i1,o1)(0) R(d1,i1,o1)(0) R(d0,i2,o2)(0) R(d1,i2,o2)(0) R(d0,i3,o3)(0) R(d1,i3,o3)(0)",
			"a0(0) a1(1) d0(0) d1(1) ei(0) eo(1) => RAM0(0) RAM1(0) R(d0,i0,o0)(0) R(d1,i0,o0)(0) R(d0,i1,o1)(0) R(d1,i1,o1)(0) R(d0,i2,o2)(0) R(d1,i2,o2)(0) R(d0,i3,o3)(0) R(d1,i3,o3)(0)",
			"a0(0) a1(1) d0(0) d1(1) ei(1) eo(0) => RAM0(0) RAM1(0) R(d0,i0,o0)(0) R(d1,i0,o0)(0) R(d0,i1,o1)(0) R(d1,i1,o1)(0) R(d0,i2,o2)(0) R(d1,i2,o2)(0) R(d0,i3,o3)(0) R(d1,i3,o3)(0)",
			"a0(0) a1(1) d0(0) d1(1) ei(1) eo(1) => RAM0(0) RAM1(1) R(d0,i0,o0)(0) R(d1,i0,o0)(0) R(d0,i1,o1)(0) R(d1,i1,o1)(0) R(d0,i2,o2)(0) R(d1,i2,o2)(1) R(d0,i3,o3)(0) R(d1,i3,o3)(0)",
			"a0(0) a1(1) d0(1) d1(0) ei(0) eo(0) => RAM0(0) RAM1(0) R(d0,i0,o0)(0) R(d1,i0,o0)(0) R(d0,i1,o1)(0) R(d1,i1,o1)(0) R(d0,i2,o2)(0) R(d1,i2,o2)(0) R(d0,i3,o3)(0) R(d1,i3,o3)(0)",
			"a0(0) a1(1) d0(1) d1(0) ei(0) eo(1) => RAM0(0) RAM1(1) R(d0,i0,o0)(0) R(d1,i0,o0)(0) R(d0,i1,o1)(0) R(d1,i1,o1)(0) R(d0,i2,o2)(0) R(d1,i2,o2)(1) R(d0,i3,o3)(0) R(d1,i3,o3)(0)",
			"a0(0) a1(1) d0(1) d1(0) ei(1) eo(0) => RAM0(0) RAM1(0) R(d0,i0,o0)(0) R(d1,i0,o0)(0) R(d0,i1,o1)(0) R(d1,i1,o1)(0) R(d0,i2,o2)(0) R(d1,i2,o2)(0) R(d0,i3,o3)(0) R(d1,i3,o3)(0)",
			"a0(0) a1(1) d0(1) d1(0) ei(1) eo(1) => RAM0(1) RAM1(0) R(d0,i0,o0)(0) R(d1,i0,o0)(0) R(d0,i1,o1)(0) R(d1,i1,o1)(0) R(d0,i2,o2)(1) R(d1,i2,o2)(0) R(d0,i3,o3)(0) R(d1,i3,o3)(0)",
			"a0(0) a1(1) d0(1) d1(1) ei(0) eo(0) => RAM0(0) RAM1(0) R(d0,i0,o0)(0) R(d1,i0,o0)(0) R(d0,i1,o1)(0) R(d1,i1,o1)(0) R(d0,i2,o2)(0) R(d1,i2,o2)(0) R(d0,i3,o3)(0) R(d1,i3,o3)(0)",
			"a0(0) a1(1) d0(1) d1(1) ei(0) eo(1) => RAM0(1) RAM1(0) R(d0,i0,o0)(0) R(d1,i0,o0)(0) R(d0,i1,o1)(0) R(d1,i1,o1)(0) R(d0,i2,o2)(1) R(d1,i2,o2)(0) R(d0,i3,o3)(0) R(d1,i3,o3)(0)",
			"a0(0) a1(1) d0(1) d1(1) ei(1) eo(0) => RAM0(0) RAM1(0) R(d0,i0,o0)(0) R(d1,i0,o0)(0) R(d0,i1,o1)(0) R(d1,i1,o1)(0) R(d0,i2,o2)(0) R(d1,i2,o2)(0) R(d0,i3,o3)(0) R(d1,i3,o3)(0)",
			"a0(0) a1(1) d0(1) d1(1) ei(1) eo(1) => RAM0(1) RAM1(1) R(d0,i0,o0)(0) R(d1,i0,o0)(0) R(d0,i1,o1)(0) R(d1,i1,o1)(0) R(d0,i2,o2)(1) R(d1,i2,o2)(1) R(d0,i3,o3)(0) R(d1,i3,o3)(0)",
			"a0(1) a1(0) d0(0) d1(0) ei(0) eo(0) => RAM0(0) RAM1(0) R(d0,i0,o0)(0) R(d1,i0,o0)(0) R(d0,i1,o1)(0) R(d1,i1,o1)(0) R(d0,i2,o2)(0) R(d1,i2,o2)(0) R(d0,i3,o3)(0) R(d1,i3,o3)(0)",
			"a0(1) a1(0) d0(0) d1(0) ei(0) eo(1) => RAM0(1) RAM1(1) R(d0,i0,o0)(0) R(d1,i0,o0)(0) R(d0,i1,o1)(1) R(d1,i1,o1)(1) R(d0,i2,o2)(0) R(d1,i2,o2)(0) R(d0,i3,o3)(0) R(d1,i3,o3)(0)",
			"a0(1) a1(0) d0(0) d1(0) ei(1) eo(0) => RAM0(0) RAM1(0) R(d0,i0,o0)(0) R(d1,i0,o0)(0) R(d0,i1,o1)(0) R(d1,i1,o1)(0) R(d0,i2,o2)(0) R(d1,i2,o2)(0) R(d0,i3,o3)(0) R(d1,i3,o3)(0)",
			"a0(1) a1(0) d0(0) d1(0) ei(1) eo(1) => RAM0(0) RAM1(0) R(d0,i0,o0)(0) R(d1,i0,o0)(0) R(d0,i1,o1)(0) R(d1,i1,o1)(0) R(d0,i2,o2)(0) R(d1,i2,o2)(0) R(d0,i3,o3)(0) R(d1,i3,o3)(0)",
			"a0(1) a1(0) d0(0) d1(1) ei(0) eo(0) => RAM0(0) RAM1(0) R(d0,i0,o0)(0) R(d1,i0,o0)(0) R(d0,i1,o1)(0) R(d1,i1,o1)(0) R(d0,i2,o2)(0) R(d1,i2,o2)(0) R(d0,i3,o3)(0) R(d1,i3,o3)(0)",
			"a0(1) a1(0) d0(0) d1(1) ei(0) eo(1) => RAM0(0) RAM1(0) R(d0,i0,o0)(0) R(d1,i0,o0)(0) R(d0,i1,o1)(0) R(d1,i1,o1)(0) R(d0,i2,o2)(0) R(d1,i2,o2)(0) R(d0,i3,o3)(0) R(d1,i3,o3)(0)",
			"a0(1) a1(0) d0(0) d1(1) ei(1) eo(0) => RAM0(0) RAM1(0) R(d0,i0,o0)(0) R(d1,i0,o0)(0) R(d0,i1,o1)(0) R(d1,i1,o1)(0) R(d0,i2,o2)(0) R(d1,i2,o2)(0) R(d0,i3,o3)(0) R(d1,i3,o3)(0)",
			"a0(1) a1(0) d0(0) d1(1) ei(1) eo(1) => RAM0(0) RAM1(1) R(d0,i0,o0)(0) R(d1,i0,o0)(0) R(d0,i1,o1)(0) R(d1,i1,o1)(1) R(d0,i2,o2)(0) R(d1,i2,o2)(0) R(d0,i3,o3)(0) R(d1,i3,o3)(0)",
			"a0(1) a1(0) d0(1) d1(0) ei(0) eo(0) => RAM0(0) RAM1(0) R(d0,i0,o0)(0) R(d1,i0,o0)(0) R(d0,i1,o1)(0) R(d1,i1,o1)(0) R(d0,i2,o2)(0) R(d1,i2,o2)(0) R(d0,i3,o3)(0) R(d1,i3,o3)(0)",
			"a0(1) a1(0) d0(1) d1(0) ei(0) eo(1) => RAM0(0) RAM1(1) R(d0,i0,o0)(0) R(d1,i0,o0)(0) R(d0,i1,o1)(0) R(d1,i1,o1)(1) R(d0,i2,o2)(0) R(d1,i2,o2)(0) R(d0,i3,o3)(0) R(d1,i3,o3)(0)",
			"a0(1) a1(0) d0(1) d1(0) ei(1) eo(0) => RAM0(0) RAM1(0) R(d0,i0,o0)(0) R(d1,i0,o0)(0) R(d0,i1,o1)(0) R(d1,i1,o1)(0) R(d0,i2,o2)(0) R(d1,i2,o2)(0) R(d0,i3,o3)(0) R(d1,i3,o3)(0)",
			"a0(1) a1(0) d0(1) d1(0) ei(1) eo(1) => RAM0(1) RAM1(0) R(d0,i0,o0)(0) R(d1,i0,o0)(0) R(d0,i1,o1)(1) R(d1,i1,o1)(0) R(d0,i2,o2)(0) R(d1,i2,o2)(0) R(d0,i3,o3)(0) R(d1,i3,o3)(0)",
			"a0(1) a1(0) d0(1) d1(1) ei(0) eo(0) => RAM0(0) RAM1(0) R(d0,i0,o0)(0) R(d1,i0,o0)(0) R(d0,i1,o1)(0) R(d1,i1,o1)(0) R(d0,i2,o2)(0) R(d1,i2,o2)(0) R(d0,i3,o3)(0) R(d1,i3,o3)(0)",
			"a0(1) a1(0) d0(1) d1(1) ei(0) eo(1) => RAM0(1) RAM1(0) R(d0,i0,o0)(0) R(d1,i0,o0)(0) R(d0,i1,o1)(1) R(d1,i1,o1)(0) R(d0,i2,o2)(0) R(d1,i2,o2)(0) R(d0,i3,o3)(0) R(d1,i3,o3)(0)",
			"a0(1) a1(0) d0(1) d1(1) ei(1) eo(0) => RAM0(0) RAM1(0) R(d0,i0,o0)(0) R(d1,i0,o0)(0) R(d0,i1,o1)(0) R(d1,i1,o1)(0) R(d0,i2,o2)(0) R(d1,i2,o2)(0) R(d0,i3,o3)(0) R(d1,i3,o3)(0)",
			"a0(1) a1(0) d0(1) d1(1) ei(1) eo(1) => RAM0(1) RAM1(1) R(d0,i0,o0)(0) R(d1,i0,o0)(0) R(d0,i1,o1)(1) R(d1,i1,o1)(1) R(d0,i2,o2)(0) R(d1,i2,o2)(0) R(d0,i3,o3)(0) R(d1,i3,o3)(0)",
			"a0(1) a1(1) d0(0) d1(0) ei(0) eo(0) => RAM0(0) RAM1(0) R(d0,i0,o0)(0) R(d1,i0,o0)(0) R(d0,i1,o1)(0) R(d1,i1,o1)(0) R(d0,i2,o2)(0) R(d1,i2,o2)(0) R(d0,i3,o3)(0) R(d1,i3,o3)(0)",
			"a0(1) a1(1) d0(0) d1(0) ei(0) eo(1) => RAM0(1) RAM1(1) R(d0,i0,o0)(0) R(d1,i0,o0)(0) R(d0,i1,o1)(0) R(d1,i1,o1)(0) R(d0,i2,o2)(0) R(d1,i2,o2)(0) R(d0,i3,o3)(1) R(d1,i3,o3)(1)",
			"a0(1) a1(1) d0(0) d1(0) ei(1) eo(0) => RAM0(0) RAM1(0) R(d0,i0,o0)(0) R(d1,i0,o0)(0) R(d0,i1,o1)(0) R(d1,i1,o1)(0) R(d0,i2,o2)(0) R(d1,i2,o2)(0) R(d0,i3,o3)(0) R(d1,i3,o3)(0)",
			"a0(1) a1(1) d0(0) d1(0) ei(1) eo(1) => RAM0(0) RAM1(0) R(d0,i0,o0)(0) R(d1,i0,o0)(0) R(d0,i1,o1)(0) R(d1,i1,o1)(0) R(d0,i2,o2)(0) R(d1,i2,o2)(0) R(d0,i3,o3)(0) R(d1,i3,o3)(0)",
			"a0(1) a1(1) d0(0) d1(1) ei(0) eo(0) => RAM0(0) RAM1(0) R(d0,i0,o0)(0) R(d1,i0,o0)(0) R(d0,i1,o1)(0) R(d1,i1,o1)(0) R(d0,i2,o2)(0) R(d1,i2,o2)(0) R(d0,i3,o3)(0) R(d1,i3,o3)(0)",
			"a0(1) a1(1) d0(0) d1(1) ei(0) eo(1) => RAM0(0) RAM1(0) R(d0,i0,o0)(0) R(d1,i0,o0)(0) R(d0,i1,o1)(0) R(d1,i1,o1)(0) R(d0,i2,o2)(0) R(d1,i2,o2)(0) R(d0,i3,o3)(0) R(d1,i3,o3)(0)",
			"a0(1) a1(1) d0(0) d1(1) ei(1) eo(0) => RAM0(0) RAM1(0) R(d0,i0,o0)(0) R(d1,i0,o0)(0) R(d0,i1,o1)(0) R(d1,i1,o1)(0) R(d0,i2,o2)(0) R(d1,i2,o2)(0) R(d0,i3,o3)(0) R(d1,i3,o3)(0)",
			"a0(1) a1(1) d0(0) d1(1) ei(1) eo(1) => RAM0(0) RAM1(1) R(d0,i0,o0)(0) R(d1,i0,o0)(0) R(d0,i1,o1)(0) R(d1,i1,o1)(0) R(d0,i2,o2)(0) R(d1,i2,o2)(0) R(d0,i3,o3)(0) R(d1,i3,o3)(1)",
			"a0(1) a1(1) d0(1) d1(0) ei(0) eo(0) => RAM0(0) RAM1(0) R(d0,i0,o0)(0) R(d1,i0,o0)(0) R(d0,i1,o1)(0) R(d1,i1,o1)(0) R(d0,i2,o2)(0) R(d1,i2,o2)(0) R(d0,i3,o3)(0) R(d1,i3,o3)(0)",
			"a0(1) a1(1) d0(1) d1(0) ei(0) eo(1) => RAM0(0) RAM1(1) R(d0,i0,o0)(0) R(d1,i0,o0)(0) R(d0,i1,o1)(0) R(d1,i1,o1)(0) R(d0,i2,o2)(0) R(d1,i2,o2)(0) R(d0,i3,o3)(0) R(d1,i3,o3)(1)",
			"a0(1) a1(1) d0(1) d1(0) ei(1) eo(0) => RAM0(0) RAM1(0) R(d0,i0,o0)(0) R(d1,i0,o0)(0) R(d0,i1,o1)(0) R(d1,i1,o1)(0) R(d0,i2,o2)(0) R(d1,i2,o2)(0) R(d0,i3,o3)(0) R(d1,i3,o3)(0)",
			"a0(1) a1(1) d0(1) d1(0) ei(1) eo(1) => RAM0(1) RAM1(0) R(d0,i0,o0)(0) R(d1,i0,o0)(0) R(d0,i1,o1)(0) R(d1,i1,o1)(0) R(d0,i2,o2)(0) R(d1,i2,o2)(0) R(d0,i3,o3)(1) R(d1,i3,o3)(0)",
			"a0(1) a1(1) d0(1) d1(1) ei(0) eo(0) => RAM0(0) RAM1(0) R(d0,i0,o0)(0) R(d1,i0,o0)(0) R(d0,i1,o1)(0) R(d1,i1,o1)(0) R(d0,i2,o2)(0) R(d1,i2,o2)(0) R(d0,i3,o3)(0) R(d1,i3,o3)(0)",
			"a0(1) a1(1) d0(1) d1(1) ei(0) eo(1) => RAM0(1) RAM1(0) R(d0,i0,o0)(0) R(d1,i0,o0)(0) R(d0,i1,o1)(0) R(d1,i1,o1)(0) R(d0,i2,o2)(0) R(d1,i2,o2)(0) R(d0,i3,o3)(1) R(d1,i3,o3)(0)",
			"a0(1) a1(1) d0(1) d1(1) ei(1) eo(0) => RAM0(0) RAM1(0) R(d0,i0,o0)(0) R(d1,i0,o0)(0) R(d0,i1,o1)(0) R(d1,i1,o1)(0) R(d0,i2,o2)(0) R(d1,i2,o2)(0) R(d0,i3,o3)(0) R(d1,i3,o3)(0)",
			"a0(1) a1(1) d0(1) d1(1) ei(1) eo(1) => RAM0(1) RAM1(1) R(d0,i0,o0)(0) R(d1,i0,o0)(0) R(d0,i1,o1)(0) R(d1,i1,o1)(0) R(d0,i2,o2)(0) R(d1,i2,o2)(0) R(d0,i3,o3)(1) R(d1,i3,o3)(1)",
		},
		isValidBool: func() func(inputs map[string]bool) []bool {
			q := [][]bool{{true, true}, {true, true}, {true, true}, {true, true}}
			return func(inputs map[string]bool) []bool {
				a0, a1, ei, eo := inputs["a0"], inputs["a1"], inputs["ei"], inputs["eo"]
				s := []bool{!a0 && !a1, a0 && !a1, !a0 && a1, a0 && a1}
				r := [][]bool{{false, false}, {false, false}, {false, false}, {false, false}}
				index := 0
				for i, si := range s {
					if si {
						index = i
						break
					}
				}
				if ei {
					q[index][0] = inputs["d0"]
					q[index][1] = inputs["d1"]
				}
				if eo {
					r[index] = q[index]
				}
				var res = r[index]
				for i := range s {
					res = append(res, r[i]...)
				}
				return res
			}
		}(),
	}}
	for _, in := range inputs {
		c := circuit.NewCircuit(config.Config{IsUnitTest: true})
		c.Outs(Example(c, in.name))
		gotDesc := c.Description()
		if diff := cmp.Diff(in.desc, gotDesc); diff != "" {
			t.Errorf("Simulate(%q) want %#v,\ngot %#v,\ndiff -want +got:\n%s", in.name, in.desc, gotDesc, diff)
		}
		got := c.Simulate()
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
			t.Errorf("Simulate(%q) want %#v,\ngot %#v,\ndiff -want +got:\n%s", in.name, in.want, converted, diff)
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
					got := int(out[len(c.Inputs)+len("=>")+i] - '0')
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
				wants := in.isValidBool(inputs)
				for i, want := range wants {
					oi := len(c.Inputs) + len("=>") + i
					if oi >= len(out) {
						t.Errorf("Simulate(%q) out %s want %t got <nothing>", in.name, out, want)
						continue
					}
					got := out[oi] == '1'
					if want != got {
						t.Errorf("Simulate(%q) out %s output %s want %t got %#v", in.name, out, c.Outputs[i].Name, want, got)
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
		name:   "AluWithBus",
		desc:   "d ai ao bi bo ri ro c => B(d) R(da,ai,ao) R(db,bi,bo) R(S(r(da,ai,ao),r(db,bi,bo),c),ri,ro) C(r(da,ai,ao),r(db,bi,bo))",
		inputs: []string{"00000000", "10000000", "01000000", "00100000", "00001000", "01001000"},
		want: []string{
			// default to a=b=1 sum=0 cout=1
			"d(0) ai(0) ao(0) bi(0) bo(0) ri(0) ro(0) c(0) => B(d)(0) R(da,ai,ao)(0) R(db,bi,bo)(0) R(S(r(da,ai,ao),r(db,bi,bo),c),ri,ro)(0) C(r(da,ai,ao),r(db,bi,bo))(1)",
			// bus=1 writes to the bus
			"d(1) ai(0) ao(0) bi(0) bo(0) ri(0) ro(0) c(0) => B(d)(1) R(da,ai,ao)(0) R(db,bi,bo)(0) R(S(r(da,ai,ao),r(db,bi,bo),c),ri,ro)(0) C(r(da,ai,ao),r(db,bi,bo))(1)",
			// ai=1 sets a=0 from the bus
			"d(0) ai(1) ao(0) bi(0) bo(0) ri(0) ro(0) c(0) => B(d)(0) R(da,ai,ao)(0) R(db,bi,bo)(0) R(S(r(da,ai,ao),r(db,bi,bo),c),ri,ro)(0) C(r(da,ai,ao),r(db,bi,bo))(0)",
			// ao=1 writes a=0 to the bus
			"d(0) ai(0) ao(1) bi(0) bo(0) ri(0) ro(0) c(0) => B(d)(0) R(da,ai,ao)(0) R(db,bi,bo)(0) R(S(r(da,ai,ao),r(db,bi,bo),c),ri,ro)(0) C(r(da,ai,ao),r(db,bi,bo))(0)",
			// bo=1 writes b=1 to the bus
			"d(0) ai(0) ao(0) bi(0) bo(1) ri(0) ro(0) c(0) => B(d)(1) R(da,ai,ao)(0) R(db,bi,bo)(1) R(S(r(da,ai,ao),r(db,bi,bo),c),ri,ro)(0) C(r(da,ai,ao),r(db,bi,bo))(0)",
			// ai=bo=1 writes b=1 to ai
			"d(0) ai(1) ao(0) bi(0) bo(1) ri(0) ro(0) c(0) => B(d)(1) R(da,ai,ao)(0) R(db,bi,bo)(1) R(S(r(da,ai,ao),r(db,bi,bo),c),ri,ro)(0) C(r(da,ai,ao),r(db,bi,bo))(1)",
		},
	}, {
		name:   "RAM",
		desc:   "a d i o => RAM0 R(d,i0,o0) R(d,i1,o1)",
		inputs: []string{"0000", "0001", "0010", "0001", "1001"},
		want: []string{
			// default to s0=q0=q1=1
			"a(0) d(0) i(0) o(0) => RAM0(0) R(d,i0,o0)(0) R(d,i1,o1)(0)",
			// o=1 writes q0=1 to res
			"a(0) d(0) i(0) o(1) => RAM0(1) R(d,i0,o0)(1) R(d,i1,o1)(0)",
			// i=1 reads d=0 to q0
			"a(0) d(0) i(1) o(0) => RAM0(0) R(d,i0,o0)(0) R(d,i1,o1)(0)",
			// o=1 writes q0=0 to res
			"a(0) d(0) i(0) o(1) => RAM0(0) R(d,i0,o0)(0) R(d,i1,o1)(0)",
			// a=o=1 writes q1=1 to res
			"a(1) d(0) i(0) o(1) => RAM0(1) R(d,i0,o0)(0) R(d,i1,o1)(1)",
		},
	}}
	for _, in := range inputs {
		c := circuit.NewCircuit(config.Config{IsUnitTest: true})
		c.Outs(Example(c, in.name))
		gotDesc := c.Description()
		if diff := cmp.Diff(in.desc, gotDesc); diff != "" {
			t.Errorf("Simulate(%q) want %#v,\ngot %#v,\ndiff -want +got:\n%s", in.name, in.desc, gotDesc, diff)
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
