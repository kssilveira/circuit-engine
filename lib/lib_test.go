package lib

import (
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
		convert     bool
		want        []string
		isValidInt  func(inputs map[string]int) []int
		isValidBool func(inputs map[string]bool) []bool
	}{{
		name:    "TransistorEmitter",
		desc:    "b c => e",
		convert: true,
		want:    []string{"b(0) c(0) => e(0)", "b(0) c(1) => e(0)", "b(1) c(0) => e(0)", "b(1) c(1) => e(1)"},
		isValidBool: func(inputs map[string]bool) []bool {
			return []bool{inputs["b"] && inputs["c"]}
		},
	}, {
		name:    "TransistorGnd",
		desc:    "b c => co",
		convert: true,
		want:    []string{"b(0) c(0) => co(0)", "b(0) c(1) => co(1)", "b(1) c(0) => co(0)", "b(1) c(1) => co(0)"},
		isValidBool: func(inputs map[string]bool) []bool {
			return []bool{!inputs["b"] && inputs["c"]}
		},
	}, {
		name:    "Transistor",
		desc:    "b c => e co",
		convert: true,
		want:    []string{"b(0) c(0) => e(0) co(0)", "b(0) c(1) => e(0) co(1)", "b(1) c(0) => e(0) co(0)", "b(1) c(1) => e(1) co(1)"},
		isValidBool: func(inputs map[string]bool) []bool {
			return []bool{inputs["b"] && inputs["c"], inputs["c"]}
		},
	}, {
		name:    "Not",
		desc:    "a => NOT(a)",
		convert: true,
		want:    []string{"a(0) => NOT(a)(1)", "a(1) => NOT(a)(0)"},
		isValidBool: func(inputs map[string]bool) []bool {
			return []bool{!inputs["a"]}
		},
	}, {
		name:    "And",
		desc:    "a b => AND(a,b)",
		convert: true,
		want:    []string{"a(0) b(0) => AND(a,b)(0)", "a(0) b(1) => AND(a,b)(0)", "a(1) b(0) => AND(a,b)(0)", "a(1) b(1) => AND(a,b)(1)"},
		isValidBool: func(inputs map[string]bool) []bool {
			return []bool{inputs["a"] && inputs["b"]}
		},
	}, {
		name:    "Or",
		desc:    "a b => OR(a,b)",
		convert: true,
		want:    []string{"a(0) b(0) => OR(a,b)(0)", "a(0) b(1) => OR(a,b)(1)", "a(1) b(0) => OR(a,b)(1)", "a(1) b(1) => OR(a,b)(1)"},
		isValidBool: func(inputs map[string]bool) []bool {
			return []bool{inputs["a"] || inputs["b"]}
		},
	}, {
		name:    "OrRes",
		desc:    "a => OR(a,bOrRes)",
		convert: true,
		want:    []string{"a(0) => OR(a,bOrRes)(0)", "a(1) => OR(a,bOrRes)(1)"},
		isValidBool: func(inputs map[string]bool) []bool {
			return []bool{inputs["a"]}
		},
	}, {
		name:    "Nand",
		desc:    "a b => NAND(a,b)",
		convert: true,
		want:    []string{"a(0) b(0) => NAND(a,b)(1)", "a(0) b(1) => NAND(a,b)(1)", "a(1) b(0) => NAND(a,b)(1)", "a(1) b(1) => NAND(a,b)(0)"},
		isValidBool: func(inputs map[string]bool) []bool {
			return []bool{!(inputs["a"] && inputs["b"])}
		},
	}, {
		name:    "Nand(Nand)",
		desc:    "a b c => NAND(a,NAND(b,c))",
		convert: true,
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
		name:    "Xor",
		desc:    "a b => XOR(a,b)",
		convert: true,
		want:    []string{"a(0) b(0) => XOR(a,b)(0)", "a(0) b(1) => XOR(a,b)(1)", "a(1) b(0) => XOR(a,b)(1)", "a(1) b(1) => XOR(a,b)(0)"},
		isValidBool: func(inputs map[string]bool) []bool {
			return []bool{inputs["a"] != inputs["b"]}
		},
	}, {
		name:    "Nor",
		desc:    "a b => NOR(a,b)",
		convert: true,
		want:    []string{"a(0) b(0) => NOR(a,b)(1)", "a(0) b(1) => NOR(a,b)(0)", "a(1) b(0) => NOR(a,b)(0)", "a(1) b(1) => NOR(a,b)(0)"},
		isValidBool: func(inputs map[string]bool) []bool {
			return []bool{!(inputs["a"] || inputs["b"])}
		},
	}, {
		name:    "HalfSum",
		desc:    "a b => S(a,b) C(a,b)",
		convert: true,
		want: []string{
			"a(0) b(0) => S(a,b)(0) C(a,b)(0)", "a(0) b(1) => S(a,b)(1) C(a,b)(0)",
			"a(1) b(0) => S(a,b)(1) C(a,b)(0)", "a(1) b(1) => S(a,b)(0) C(a,b)(1)",
		},
		isValidInt: func(inputs map[string]int) []int {
			sum := inputs["a"] + inputs["b"]
			return []int{sum % 2, sum / 2}
		},
	}, {
		name:    "Sum",
		desc:    "a b c => S(a,b,c) C(a,b)",
		convert: true,
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
		name:    "Sum2",
		desc:    "a0 a1 b0 b1 c => S(a0,b0,c) S(a1,b1,C(a0,b0)) C(a1,b1)",
		convert: true,
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
		name:    "SumN",
		desc:    "a0 a1 b0 b1 c => S(a0,b0,c) S(a1,b1,C(a0,b0)) C(a1,b1)",
		convert: true,
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
		name:    "SRLatch",
		desc:    "s r => q nq",
		convert: true,
		want:    []string{"s(0) r(0) => q(1) nq(0)", "s(0) r(1) => q(0) nq(1)", "s(1) r(0) => q(1) nq(0)", "s(1) r(1) => q(0) nq(0)"},
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
		name:    "SRLatchWithEnable",
		desc:    "s r e => q nq",
		convert: true,
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
		name:    "DLatch",
		desc:    "d e => q nq",
		convert: true,
		want:    []string{"d(0) e(0) => q(1) nq(0)", "d(0) e(1) => q(0) nq(1)", "d(1) e(0) => q(0) nq(1)", "d(1) e(1) => q(1) nq(0)"},
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
		name:    "Register",
		desc:    "d i o => r(d,i,o) R(d,i,o)",
		convert: true,
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
		name:    "Register2",
		desc:    "d0 d1 i o => r(d0,i,o) R(d0,i,o) r(d1,i,o) R(d1,i,o)",
		convert: true,
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
		name:    "RegisterN",
		desc:    "d0 d1 i o => r(d0,i,o) R(d0,i,o) r(d1,i,o) R(d1,i,o)",
		convert: true,
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
		name:    "Alu",
		desc:    "a ai ao b bi bo ri ro c => R(a,ai,ao) R(b,bi,bo) R(S(a,b)) C(a,b)",
		convert: true,
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
		name:    "Alu2",
		desc:    "a0 a1 ai ao b0 b1 bi bo ri ro c => R(a0,ai,ao) R(b0,bi,bo) R(S(a0,b0)) R(a1,ai,ao) R(b1,bi,bo) R(S(a1,b1)) C(a1,b1)",
		convert: true,
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
		name:    "AluN",
		desc:    "a0 a1 ai ao b0 b1 bi bo ri ro c => R(a0,ai,ao) R(b0,bi,bo) R(S(a0,b0)) R(a1,ai,ao) R(b1,bi,bo) R(S(a1,b1)) C(a1,b1)",
		convert: true,
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
		name:    "Bus",
		desc:    "d ar br r => B(d) aw bw",
		convert: true,
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
		name:    "Bus2",
		desc:    "d0 d1 ar0 ar1 br0 br1 r0 r1 => B(d0) B(d1) aw0 aw1 bw0 bw1",
		convert: true,
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
		name:    "BusN",
		desc:    "d0 d1 ar0 ar1 br0 br1 r0 r1 => B(d0) B(d1) aw0 aw1 bw0 bw1",
		convert: true,
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
		name:    "BusIOn",
		desc:    "d ar br r => B(d) aw bw",
		convert: true,
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
		name:    "BUSbNioN",
		desc:    "d0 d1 ar0 ar1 br0 br1 r0 r1 => B(d0) B(d1) aw0 aw1 bw0 bw1",
		convert: true,
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
		desc: "bus ai ao bi bo ri ro cin" +
			" => B(bus) r(ALU-bus-a,ai,ao) R(ALU-bus-a,ai,ao) r(ALU-bus-b,bi,bo) R(ALU-bus-b,bi,bo)" +
			" r(S(r(ALU-bus-a,ai,ao),r(ALU-bus-b,bi,bo),cin),ri,ro)" +
			" R(S(r(ALU-bus-a,ai,ao),r(ALU-bus-b,bi,bo),cin),ri,ro)" +
			" C(r(ALU-bus-a,ai,ao),r(ALU-bus-b,bi,bo))",
		want: []string{
			"10000101=>11010101", "10100000=>11110101", "00101101=>11111101", "10001000=>11011101",
			"10000010=>11010111",
		},
		isValidInt: func() func(inputs map[string]int) []int {
			qa, qb, qr := 1, 1, 1
			return func(inputs map[string]int) []int {
				ra, rb, rr, bus, sum := 0, 0, 0, 0, 0
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
					bus = inputs["bus"] | ra | rb | rr
					if inputs["ai"] == 1 {
						qa = bus
					}
					if inputs["bi"] == 1 {
						qb = bus
					}
					sum = qa + qb + inputs["cin"]
					if inputs["ri"] == 1 {
						qr = sum % 2
					}
				}
				return []int{bus, qa, ra, qb, rb, qr, rr, sum / 2}
			}
		}(),
	}, {
		name: "AluWithBus2",
		desc: "bus1 bus2 ai ao bi bo ri ro cin" +
			" => B(bus1) r(ALU-bus1-a,ai,ao) R(ALU-bus1-a,ai,ao) r(ALU-bus1-b,bi,bo) R(ALU-bus1-b,bi,bo)" +
			" r(S(r(ALU-bus1-a,ai,ao),r(ALU-bus1-b,bi,bo),cin),ri,ro)" +
			" R(S(r(ALU-bus1-a,ai,ao),r(ALU-bus1-b,bi,bo),cin),ri,ro)" +
			" B(bus2) r(ALU-bus2-a,ai,ao) R(ALU-bus2-a,ai,ao) r(ALU-bus2-b,bi,bo) R(ALU-bus2-b,bi,bo)" +
			" r(S(r(ALU-bus2-a,ai,ao),r(ALU-bus2-b,bi,bo),C(r(ALU-bus1-a,ai,ao),r(ALU-bus1-b,bi,bo))),ri,ro)" +
			" R(S(r(ALU-bus2-a,ai,ao),r(ALU-bus2-b,bi,bo),C(r(ALU-bus1-a,ai,ao),r(ALU-bus1-b,bi,bo))),ri,ro)" +
			" C(r(ALU-bus2-a,ai,ao),r(ALU-bus2-b,bi,bo))",
		want: []string{
			"110110101=>111101011110101", "110000110=>110100011010111",
			"000101101=>111111011111101", "000000010=>110101111010111",
			"000100010=>111101111110111", "000101110=>111110011111111",
			"100000101=>110101001010101",
		},
		isValidInt: func() func(inputs map[string]int) []int {
			qa1, qb1, qr1 := 1, 1, 1
			qa2, qb2, qr2 := 1, 1, 1
			return func(inputs map[string]int) []int {
				ra1, rb1, rr1, bus1, sum1 := 0, 0, 0, 0, 0
				ra2, rb2, rr2, bus2, sum2 := 0, 0, 0, 0, 0
				for i := 0; i < 10; i++ {
					if inputs["ao"] == 1 {
						ra1, ra2 = qa1, qa2
					}
					if inputs["bo"] == 1 {
						rb1, rb2 = qb1, qb2
					}
					if inputs["ro"] == 1 {
						rr1, rr2 = qr1, qr2
					}
					bus1 = inputs["bus1"] | ra1 | rb1 | rr1
					bus2 = inputs["bus2"] | ra2 | rb2 | rr2
					if inputs["ai"] == 1 {
						qa1, qa2 = bus1, bus2
					}
					if inputs["bi"] == 1 {
						qb1, qb2 = bus1, bus2
					}
					sum1 = qa1 + qb1 + inputs["cin"]
					sum2 = qa2 + qb2 + sum1/2
					if inputs["ri"] == 1 {
						qr1, qr2 = sum1%2, sum2%2
					}
				}
				return []int{
					bus1, qa1, ra1, qb1, rb1, qr1, rr1,
					bus2, qa2, ra2, qb2, rb2, qr2, rr2,
					sum2 / 2}
			}
		}(),
	}, {
		name: "RAM",
		desc: "a d ei eo" +
			" => RAM RAM-s0 RAM-ei0 RAM-eo0 r(d,RAM-ei0,RAM-eo0) R(d,RAM-ei0,RAM-eo0)" +
			" RAM-s1 RAM-ei1 RAM-eo1 r(d,RAM-ei1,RAM-eo1) R(d,RAM-ei1,RAM-eo1)",
		want: []string{
			"0000=>01001000010", "0001=>11011100010", "0010=>01100000010", "0011=>01110000010",
			"0100=>01000000010", "0101=>01010000010", "0110=>01101000010", "0111=>11111100010",
			"1000=>00001010010", "1001=>10001010111", "1010=>00001011000", "1011=>00001011100",
			"1100=>00001010000", "1101=>00001010100", "1110=>00001011010", "1111=>10001011111",
		},
		isValidBool: func() func(inputs map[string]bool) []bool {
			q0, q1 := true, true
			return func(inputs map[string]bool) []bool {
				a, ei, eo := inputs["a"], inputs["ei"], inputs["eo"]
				s0, s1 := !a, a
				r0, r1 := false, false
				q, r := &q0, &r0
				if s1 {
					q, r = &q1, &r1
				}
				if ei {
					*q = inputs["d"]
				}
				if eo {
					*r = *q
				}
				return []bool{*r, s0, s0 && ei, s0 && eo, q0, r0, s1, s1 && ei, s1 && eo, q1, r1}
			}
		}(),
	}, {
		name: "RAMa2",
		desc: "a0 a1 d ei eo" +
			" => RAM RAM-s0 RAM-ei0 RAM-eo0 r(d,RAM-ei0,RAM-eo0) R(d,RAM-ei0,RAM-eo0)" +
			" RAM-s1 RAM-ei1 RAM-eo1 r(d,RAM-ei1,RAM-eo1) R(d,RAM-ei1,RAM-eo1)" +
			" RAM-s2 RAM-ei2 RAM-eo2 r(d,RAM-ei2,RAM-eo2) R(d,RAM-ei2,RAM-eo2)" +
			" RAM-s3 RAM-ei3 RAM-eo3 r(d,RAM-ei3,RAM-eo3) R(d,RAM-ei3,RAM-eo3)",
		want: []string{
			"00000=>010010000100001000010", "00001=>110111000100001000010",
			"00010=>011000000100001000010", "00011=>011100000100001000010",
			"00100=>010000000100001000010", "00101=>010100000100001000010",
			"00110=>011010000100001000010", "00111=>111111000100001000010",
			"01000=>000010000101001000010", "01001=>100010000101011100010",
			"01010=>000010000101100000010", "01011=>000010000101110000010",
			"01100=>000010000101000000010", "01101=>000010000101010000010",
			"01110=>000010000101101000010", "01111=>100010000101111100010",
			"10000=>000010100100001000010", "10001=>100010101110001000010",
			"10010=>000010110000001000010", "10011=>000010111000001000010",
			"10100=>000010100000001000010", "10101=>000010101000001000010",
			"10110=>000010110100001000010", "10111=>100010111110001000010",
			"11000=>000010000100001010010", "11001=>100010000100001010111",
			"11010=>000010000100001011000", "11011=>000010000100001011100",
			"11100=>000010000100001010000", "11101=>000010000100001010100",
			"11110=>000010000100001011010", "11111=>100010000100001011111",
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
				for i, si := range s {
					res = append(res, si, si && ei, si && eo, q[i], r[i])
				}
				return res
			}
		}(),
	}, {
		name: "RAMb2",
		desc: "a d0 d1 ei eo" +
			" => RAM RAM RAM-s0 RAM-ei0 RAM-eo0" +
			" r(d0,RAM-ei0,RAM-eo0) R(d0,RAM-ei0,RAM-eo0)" +
			" r(d1,RAM-ei0,RAM-eo0) R(d1,RAM-ei0,RAM-eo0)" +
			" RAM-s1 RAM-ei1 RAM-eo1" +
			" r(d0,RAM-ei1,RAM-eo1) R(d0,RAM-ei1,RAM-eo1)" +
			" r(d1,RAM-ei1,RAM-eo1) R(d1,RAM-ei1,RAM-eo1)",
		want: []string{
			"00000=>0010010100001010", "00001=>1110111110001010",
			"00010=>0011000000001010", "00011=>0011100000001010",
			"00100=>0010000000001010", "00101=>0010100000001010",
			"00110=>0011000100001010", "00111=>0111100110001010",
			"01000=>0010000100001010", "01001=>0110100110001010",
			"01010=>0011010000001010", "01011=>1011111000001010",
			"01100=>0010010000001010", "01101=>1010111000001010",
			"01110=>0011010100001010", "01111=>1111111110001010",
			"10000=>0000010101001010", "10001=>1100010101011111",
			"10010=>0000010101100000", "10011=>0000010101110000",
			"10100=>0000010101000000", "10101=>0000010101010000",
			"10110=>0000010101100010", "10111=>0100010101110011",
			"11000=>0000010101000010", "11001=>0100010101010011",
			"11010=>0000010101101000", "11011=>1000010101111100",
			"11100=>0000010101001000", "11101=>1000010101011100",
			"11110=>0000010101101010", "11111=>1100010101111111",
		},
		isValidBool: func() func(inputs map[string]bool) []bool {
			q0, q1 := []bool{true, true}, []bool{true, true}
			return func(inputs map[string]bool) []bool {
				a, ei, eo := inputs["a"], inputs["ei"], inputs["eo"]
				s0, s1 := !a, a
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
				return []bool{(*r)[0], (*r)[1], s0, s0 && ei, s0 && eo, q0[0], r0[0], q0[1], r0[1], s1, s1 && ei, s1 && eo, q1[0], r1[0], q1[1], r1[1]}
			}
		}(),
	}, {
		name: "RAMa2b2",
		desc: "a0 a1 d0 d1 ei eo" +
			" => RAM RAM RAM-s0 RAM-ei0 RAM-eo0" +
			" r(d0,RAM-ei0,RAM-eo0) R(d0,RAM-ei0,RAM-eo0)" +
			" r(d1,RAM-ei0,RAM-eo0) R(d1,RAM-ei0,RAM-eo0)" +
			" RAM-s1 RAM-ei1 RAM-eo1" +
			" r(d0,RAM-ei1,RAM-eo1) R(d0,RAM-ei1,RAM-eo1)" +
			" r(d1,RAM-ei1,RAM-eo1) R(d1,RAM-ei1,RAM-eo1)" +
			" RAM-s2 RAM-ei2 RAM-eo2" +
			" r(d0,RAM-ei2,RAM-eo2) R(d0,RAM-ei2,RAM-eo2)" +
			" r(d1,RAM-ei2,RAM-eo2) R(d1,RAM-ei2,RAM-eo2)" +
			" RAM-s3 RAM-ei3 RAM-eo3" +
			" r(d0,RAM-ei3,RAM-eo3) R(d0,RAM-ei3,RAM-eo3)" +
			" r(d1,RAM-ei3,RAM-eo3) R(d1,RAM-ei3,RAM-eo3)",
		want: []string{
			"000000=>001001010000101000010100001010", "000001=>111011111000101000010100001010",
			"000010=>001100000000101000010100001010", "000011=>001110000000101000010100001010",
			"000100=>001000000000101000010100001010", "000101=>001010000000101000010100001010",
			"000110=>001100010000101000010100001010", "000111=>011110011000101000010100001010",
			"001000=>001000010000101000010100001010", "001001=>011010011000101000010100001010",
			"001010=>001101000000101000010100001010", "001011=>101111100000101000010100001010",
			"001100=>001001000000101000010100001010", "001101=>101011100000101000010100001010",
			"001110=>001101010000101000010100001010", "001111=>111111111000101000010100001010",
			"010000=>000001010000101010010100001010", "010001=>110001010000101010111110001010",
			"010010=>000001010000101011000000001010", "010011=>000001010000101011100000001010",
			"010100=>000001010000101010000000001010", "010101=>000001010000101010100000001010",
			"010110=>000001010000101011000100001010", "010111=>010001010000101011100110001010",
			"011000=>000001010000101010000100001010", "011001=>010001010000101010100110001010",
			"011010=>000001010000101011010000001010", "011011=>100001010000101011111000001010",
			"011100=>000001010000101010010000001010", "011101=>100001010000101010111000001010",
			"011110=>000001010000101011010100001010", "011111=>110001010000101011111110001010",
			"100000=>000001010100101000010100001010", "100001=>110001010101111100010100001010",
			"100010=>000001010110000000010100001010", "100011=>000001010111000000010100001010",
			"100100=>000001010100000000010100001010", "100101=>000001010101000000010100001010",
			"100110=>000001010110001000010100001010", "100111=>010001010111001100010100001010",
			"101000=>000001010100001000010100001010", "101001=>010001010101001100010100001010",
			"101010=>000001010110100000010100001010", "101011=>100001010111110000010100001010",
			"101100=>000001010100100000010100001010", "101101=>100001010101110000010100001010",
			"101110=>000001010110101000010100001010", "101111=>110001010111111100010100001010",
			"110000=>000001010000101000010101001010", "110001=>110001010000101000010101011111",
			"110010=>000001010000101000010101100000", "110011=>000001010000101000010101110000",
			"110100=>000001010000101000010101000000", "110101=>000001010000101000010101010000",
			"110110=>000001010000101000010101100010", "110111=>010001010000101000010101110011",
			"111000=>000001010000101000010101000010", "111001=>010001010000101000010101010011",
			"111010=>000001010000101000010101101000", "111011=>100001010000101000010101111100",
			"111100=>000001010000101000010101001000", "111101=>100001010000101000010101011100",
			"111110=>000001010000101000010101101010", "111111=>110001010000101000010101111111",
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
				for i, si := range s {
					res = append(res, si, si && ei, si && eo, q[i][0], r[i][0], q[i][1], r[i][1])
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
		if in.convert {
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
		} else {
			converted = got
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
					got := out[len(c.Inputs)+len("=>")+i] == '1'
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
		desc:   "a => OR(a,bOrRes)",
		inputs: []string{"0", "1", "0"},
		want:   []string{"0=>0", "1=>1", "0=>1"},
	}, {
		name:   "SRLatchWithEnable",
		desc:   "s r e => q nq",
		inputs: []string{"000", "001", "010", "011", "000", "100", "101", "000"},
		want:   []string{"000=>10", "001=>10", "010=>10", "011=>01", "000=>01", "100=>01", "101=>10", "000=>10"},
	}, {
		name: "AluWithBus",
		desc: "bus ai ao bi bo ri ro cin" +
			" => B(bus) r(ALU-bus-a,ai,ao) R(ALU-bus-a,ai,ao) r(ALU-bus-b,bi,bo) R(ALU-bus-b,bi,bo)" +
			" r(S(r(ALU-bus-a,ai,ao),r(ALU-bus-b,bi,bo),cin),ri,ro)" +
			" R(S(r(ALU-bus-a,ai,ao),r(ALU-bus-b,bi,bo),cin),ri,ro)" +
			" C(r(ALU-bus-a,ai,ao),r(ALU-bus-b,bi,bo))",
		inputs: []string{"00000000", "10000000", "01000000", "00100000", "00001000", "01001000"},
		want: []string{
			// default to a=1 b=1 r=1 sum=0 cout=1
			"00000000=>01010101",
			// bus=1 writes to rbus
			"10000000=>11010101",
			// ai=1 sets a=0 from the bus
			"01000000=>00010100",
			// ao=1 writes a=0 to the bus
			"00100000=>00010100",
			// bo=1 writes b=1 to the bus
			"00001000=>10011100",
			// ai=bo=1 writes b=1 to ai
			"01001000=>11011101",
		},
	}, {
		name: "RAM",
		desc: "a d ei eo" +
			" => RAM RAM-s0 RAM-ei0 RAM-eo0" +
			" r(d,RAM-ei0,RAM-eo0) R(d,RAM-ei0,RAM-eo0)" +
			" RAM-s1 RAM-ei1 RAM-eo1" +
			" r(d,RAM-ei1,RAM-eo1) R(d,RAM-ei1,RAM-eo1)",
		inputs: []string{"0000", "0001", "0010", "0001", "1001"},
		want: []string{
			// default to s0=q0=q1=1
			"0000=>01001000010",
			// eo=1 writes q0=1 to res
			"0001=>11011100010",
			// ei=1 reads d=0 to q0
			"0010=>01100000010",
			// eo=1 writes q0=0 to res
			"0001=>01010000010",
			// a=eo=1 writes q1=1 to res
			"1001=>10000010111",
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
		if diff := cmp.Diff(in.want, got); diff != "" {
			t.Errorf("SimulateInputs(%q) want %#v,\ngot %#v,\ndiff -want +got:\n%s", in.name, in.want, got, diff)
		}
	}
}
