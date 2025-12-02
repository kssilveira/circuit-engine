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
		want:    []string{"b(0)c(0) => e(0)", "b(0)c(1) => e(0)", "b(1)c(0) => e(0)", "b(1)c(1) => e(1)"},
		isValidBool: func(inputs map[string]bool) []bool {
			return []bool{inputs["b"] && inputs["c"]}
		},
	}, {
		name:    "TransistorGnd",
		desc:    "b c => co",
		convert: true,
		want:    []string{"b(0)c(0) => co(0)", "b(0)c(1) => co(1)", "b(1)c(0) => co(0)", "b(1)c(1) => co(0)"},
		isValidBool: func(inputs map[string]bool) []bool {
			return []bool{!inputs["b"] && inputs["c"]}
		},
	}, {
		name:    "Transistor",
		desc:    "b c => e co",
		convert: true,
		want:    []string{"b(0)c(0) => e(0)co(0)", "b(0)c(1) => e(0)co(1)", "b(1)c(0) => e(0)co(0)", "b(1)c(1) => e(1)co(1)"},
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
		want:    []string{"a(0)b(0) => AND(a,b)(0)", "a(0)b(1) => AND(a,b)(0)", "a(1)b(0) => AND(a,b)(0)", "a(1)b(1) => AND(a,b)(1)"},
		isValidBool: func(inputs map[string]bool) []bool {
			return []bool{inputs["a"] && inputs["b"]}
		},
	}, {
		name:    "Or",
		desc:    "a b => OR(a,b)",
		convert: true,
		want:    []string{"a(0)b(0) => OR(a,b)(0)", "a(0)b(1) => OR(a,b)(1)", "a(1)b(0) => OR(a,b)(1)", "a(1)b(1) => OR(a,b)(1)"},
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
		want:    []string{"a(0)b(0) => NAND(a,b)(1)", "a(0)b(1) => NAND(a,b)(1)", "a(1)b(0) => NAND(a,b)(1)", "a(1)b(1) => NAND(a,b)(0)"},
		isValidBool: func(inputs map[string]bool) []bool {
			return []bool{!(inputs["a"] && inputs["b"])}
		},
	}, {
		name:    "Nand(Nand)",
		desc:    "a b c => NAND(a,NAND(b,c))",
		convert: true,
		want: []string{
			"a(0)b(0)c(0) => NAND(a,NAND(b,c))(1)", "a(0)b(0)c(1) => NAND(a,NAND(b,c))(1)",
			"a(0)b(1)c(0) => NAND(a,NAND(b,c))(1)", "a(0)b(1)c(1) => NAND(a,NAND(b,c))(1)",
			"a(1)b(0)c(0) => NAND(a,NAND(b,c))(0)", "a(1)b(0)c(1) => NAND(a,NAND(b,c))(0)",
			"a(1)b(1)c(0) => NAND(a,NAND(b,c))(0)", "a(1)b(1)c(1) => NAND(a,NAND(b,c))(1)",
		},
		isValidBool: func(inputs map[string]bool) []bool {
			return []bool{!(inputs["a"] && !(inputs["b"] && inputs["c"]))}
		},
	}, {
		name:    "Xor",
		desc:    "a b => XOR(a,b)",
		convert: true,
		want:    []string{"a(0)b(0) => XOR(a,b)(0)", "a(0)b(1) => XOR(a,b)(1)", "a(1)b(0) => XOR(a,b)(1)", "a(1)b(1) => XOR(a,b)(0)"},
		isValidBool: func(inputs map[string]bool) []bool {
			return []bool{inputs["a"] != inputs["b"]}
		},
	}, {
		name:    "Nor",
		desc:    "a b => NOR(a,b)",
		convert: true,
		want:    []string{"a(0)b(0) => NOR(a,b)(1)", "a(0)b(1) => NOR(a,b)(0)", "a(1)b(0) => NOR(a,b)(0)", "a(1)b(1) => NOR(a,b)(0)"},
		isValidBool: func(inputs map[string]bool) []bool {
			return []bool{!(inputs["a"] || inputs["b"])}
		},
	}, {
		name: "HalfSum",
		desc: "a b => SUM(a,b) CARRY(a,b)",
		want: []string{"00=>00", "01=>10", "10=>10", "11=>01"},
		isValidInt: func(inputs map[string]int) []int {
			sum := inputs["a"] + inputs["b"]
			return []int{sum % 2, sum / 2}
		},
	}, {
		name: "Sum",
		desc: "a b cin => SUM(a,b,cin) CARRY(a,b)",
		want: []string{"000=>00", "001=>10", "010=>10", "011=>01", "100=>10", "101=>01", "110=>01", "111=>11"},
		isValidInt: func(inputs map[string]int) []int {
			sum := inputs["a"] + inputs["b"] + inputs["cin"]
			return []int{sum % 2, sum / 2}
		},
	}, {
		name: "Sum2",
		desc: "a1 a2 b1 b2 cin => SUM(a1,b1,cin) SUM(a2,b2,CARRY(a1,b1)) CARRY(a2,b2)",
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
		isValidInt: func(inputs map[string]int) []int {
			sum1 := inputs["a1"] + inputs["b1"] + inputs["cin"]
			sum2 := sum1/2 + inputs["a2"] + inputs["b2"]
			return []int{sum1 % 2, sum2 % 2, sum2 / 2}
		},
	}, {
		name: "Sum4",
		desc: "a1 a2 a3 a4 b1 b2 b3 b4 cin" +
			" => SUM(a1,b1,cin) SUM(a2,b2,CARRY(a1,b1)) SUM(a3,b3,CARRY(a2,b2)) SUM(a4,b4,CARRY(a3,b3))" +
			" CARRY(a4,b4)",
		want: []string{
			"110110101=>10001", "110000110=>11110", "000101101=>11110", "000000010=>00010",
			"110111101=>11001", "000100010=>00001", "000101110=>01101", "110010110=>00001",
			"100000101=>01100", "001101111=>11011",
		},
		isValidInt: func(inputs map[string]int) []int {
			sum1 := inputs["a1"] + inputs["b1"] + inputs["cin"]
			sum2 := sum1/2 + inputs["a2"] + inputs["b2"]
			sum3 := sum2/2 + inputs["a3"] + inputs["b3"]
			sum4 := sum3/2 + inputs["a4"] + inputs["b4"]
			return []int{sum1 % 2, sum2 % 2, sum3 % 2, sum4 % 2, sum4 / 2}
		},
	}, {
		name: "Sum8",
		desc: "a1 a2 a3 a4 a5 a6 a7 a8" +
			" b1 b2 b3 b4 b5 b6 b7 b8 cin" +
			" => SUM(a1,b1,cin) SUM(a2,b2,CARRY(a1,b1)) SUM(a3,b3,CARRY(a2,b2)) SUM(a4,b4,CARRY(a3,b3))" +
			" SUM(a5,b5,CARRY(a4,b4)) SUM(a6,b6,CARRY(a5,b5)) SUM(a7,b7,CARRY(a6,b6)) SUM(a8,b8,CARRY(a7,b7))" +
			" CARRY(a8,b8)",
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
		isValidInt: func(inputs map[string]int) []int {
			sum1 := inputs["a1"] + inputs["b1"] + inputs["cin"]
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
		name: "SumN",
		desc: "a1 a2 a3 a4 b1 b2 b3 b4 cin" +
			" => SUM(a1,b1,cin) SUM(a2,b2,CARRY(a1,b1)) SUM(a3,b3,CARRY(a2,b2)) SUM(a4,b4,CARRY(a3,b3))" +
			" CARRY(a4,b4)",
		want: []string{
			"110110101=>10001", "110000110=>11110", "000101101=>11110", "000000010=>00010",
			"110111101=>11001", "000100010=>00001", "000101110=>01101", "110010110=>00001",
			"100000101=>01100", "001101111=>11011",
		},
		isValidInt: func(inputs map[string]int) []int {
			sum1 := inputs["a1"] + inputs["b1"] + inputs["cin"]
			sum2 := sum1/2 + inputs["a2"] + inputs["b2"]
			sum3 := sum2/2 + inputs["a3"] + inputs["b3"]
			sum4 := sum3/2 + inputs["a4"] + inputs["b4"]
			return []int{sum1 % 2, sum2 % 2, sum3 % 2, sum4 % 2, sum4 / 2}
		},
	}, {
		name: "SRLatch",
		desc: "s r => q nq",
		want: []string{"00=>10", "01=>01", "10=>10", "11=>00"},
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
		want: []string{"000=>10", "001=>10", "010=>10", "011=>01", "100=>01", "101=>10", "110=>10", "111=>00"},
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
		want: []string{"00=>10", "01=>01", "10=>01", "11=>10"},
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
		desc: "d ei eo => reg(d,ei,eo) REG(d,ei,eo)",
		want: []string{"000=>10", "001=>11", "010=>00", "011=>00", "100=>00", "101=>00", "110=>10", "111=>11"},
		isValidBool: func() func(inputs map[string]bool) []bool {
			q := true
			return func(inputs map[string]bool) []bool {
				if inputs["ei"] {
					q = inputs["d"]
				}
				return []bool{q, inputs["eo"] && q}
			}
		}(),
	}, {
		name: "Register2",
		desc: "d1 d2 ei eo => reg(d1,ei,eo) REG(d1,ei,eo) reg(d2,ei,eo) REG(d2,ei,eo)",
		want: []string{
			"0000=>1010", "0001=>1111", "0010=>0000", "0011=>0000",
			"0100=>0000", "0101=>0000", "0110=>0010", "0111=>0011",
			"1000=>0010", "1001=>0011", "1010=>1000", "1011=>1100",
			"1100=>1000", "1101=>1100", "1110=>1010", "1111=>1111",
		},
		isValidBool: func() func(inputs map[string]bool) []bool {
			q1, q2 := true, true
			return func(inputs map[string]bool) []bool {
				if inputs["ei"] {
					q1, q2 = inputs["d1"], inputs["d2"]
				}
				eo := inputs["eo"]
				return []bool{q1, eo && q1, q2, eo && q2}
			}
		}(),
	}, {
		name: "Register4",
		desc: "d1 d2 d3 d4 ei eo" +
			" => reg(d1,ei,eo) REG(d1,ei,eo) reg(d2,ei,eo) REG(d2,ei,eo)" +
			" reg(d3,ei,eo) REG(d3,ei,eo) reg(d4,ei,eo) REG(d4,ei,eo)",
		want: []string{
			"000000=>10101010", "000001=>11111111", "000010=>00000000", "000011=>00000000",
			"000100=>00000000", "000101=>00000000", "000110=>00000010", "000111=>00000011",
			"001000=>00000010", "001001=>00000011", "001010=>00001000", "001011=>00001100",
			"001100=>00001000", "001101=>00001100", "001110=>00001010", "001111=>00001111",
			"010000=>00001010", "010001=>00001111", "010010=>00100000", "010011=>00110000",
			"010100=>00100000", "010101=>00110000", "010110=>00100010", "010111=>00110011",
			"011000=>00100010", "011001=>00110011", "011010=>00101000", "011011=>00111100",
			"011100=>00101000", "011101=>00111100", "011110=>00101010", "011111=>00111111",
			"100000=>00101010", "100001=>00111111", "100010=>10000000", "100011=>11000000",
			"100100=>10000000", "100101=>11000000", "100110=>10000010", "100111=>11000011",
			"101000=>10000010", "101001=>11000011", "101010=>10001000", "101011=>11001100",
			"101100=>10001000", "101101=>11001100", "101110=>10001010", "101111=>11001111",
			"110000=>10001010", "110001=>11001111", "110010=>10100000", "110011=>11110000",
			"110100=>10100000", "110101=>11110000", "110110=>10100010", "110111=>11110011",
			"111000=>10100010", "111001=>11110011", "111010=>10101000", "111011=>11111100",
			"111100=>10101000", "111101=>11111100", "111110=>10101010", "111111=>11111111",
		},
		isValidBool: func() func(inputs map[string]bool) []bool {
			q1, q2, q3, q4 := true, true, true, true
			return func(inputs map[string]bool) []bool {
				if inputs["ei"] {
					q1, q2, q3, q4 = inputs["d1"], inputs["d2"], inputs["d3"], inputs["d4"]
				}
				eo := inputs["eo"]
				return []bool{q1, eo && q1, q2, eo && q2, q3, eo && q3, q4, eo && q4}
			}
		}(),
	}, {
		name: "Register8",
		desc: "d1 d2 d3 d4 d5 d6 d7 d8 ei eo" +
			" => reg(d1,ei,eo) REG(d1,ei,eo) reg(d2,ei,eo) REG(d2,ei,eo)" +
			" reg(d3,ei,eo) REG(d3,ei,eo) reg(d4,ei,eo) REG(d4,ei,eo)" +
			" reg(d5,ei,eo) REG(d5,ei,eo) reg(d6,ei,eo) REG(d6,ei,eo)" +
			" reg(d7,ei,eo) REG(d7,ei,eo) reg(d8,ei,eo) REG(d8,ei,eo)",
		want: []string{
			"1101101011=>1111001111001100", "1000011000=>1010001010001000",
			"0101101000=>1010001010001000", "0000101101=>1111001111001100",
			"1110100010=>1010100010000000", "0010000101=>1111110011000000",
			"1101100101=>1111110011000000", "1010000010=>1000100000000000",
			"1001101111=>1100001111001111", "1110110100=>1000001010001010",
		},
		isValidBool: func() func(inputs map[string]bool) []bool {
			q1, q2, q3, q4 := true, true, true, true
			q5, q6, q7, q8 := true, true, true, true
			return func(inputs map[string]bool) []bool {
				if inputs["ei"] {
					q1, q2, q3, q4 = inputs["d1"], inputs["d2"], inputs["d3"], inputs["d4"]
					q5, q6, q7, q8 = inputs["d5"], inputs["d6"], inputs["d7"], inputs["d8"]
				}
				eo := inputs["eo"]
				return []bool{q1, eo && q1, q2, eo && q2, q3, eo && q3, q4, eo && q4, q5, eo && q5, q6, eo && q6, q7, eo && q7, q8, eo && q8}
			}
		}(),
	}, {
		name: "Alu",
		desc: "a ai ao b bi bo ri ro cin" +
			" => reg(a,ai,ao) REG(a,ai,ao) reg(b,bi,bo) REG(b,bi,bo)" +
			" reg(SUM(reg(a,ai,ao),reg(b,bi,bo),cin),ri,ro) REG(SUM(reg(a,ai,ao),reg(b,bi,bo),cin),ri,ro)" +
			" CARRY(reg(a,ai,ao),reg(b,bi,bo))",
		want: []string{
			"110110101=>1010101", "110000110=>1010001", "000101101=>1011101", "000000010=>1010111",
			"110111101=>1011101", "000100010=>1010111", "000101110=>1011001", "110010110=>1000110",
			"100000101=>1000001", "001101111=>1100001",
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
				sum := qa + qb + inputs["cin"]
				if inputs["ri"] == 1 {
					qr = sum % 2
				}
				return []int{qa, inputs["ao"] & qa, qb, inputs["bo"] & qb, qr, inputs["ro"] & qr, sum / 2}
			}
		}(),
	}, {
		name: "Alu2",
		desc: "a1 a2 ai ao b1 b2 bi bo ri ro cin" +
			" => reg(a1,ai,ao) REG(a1,ai,ao) reg(b1,bi,bo) REG(b1,bi,bo)" +
			" reg(SUM(reg(a1,ai,ao),reg(b1,bi,bo),cin),ri,ro) REG(SUM(reg(a1,ai,ao),reg(b1,bi,bo),cin),ri,ro)" +
			" reg(a2,ai,ao) REG(a2,ai,ao) reg(b2,bi,bo) REG(b2,bi,bo)" +
			" reg(SUM(reg(a2,ai,ao),reg(b2,bi,bo),CARRY(reg(a1,ai,ao),reg(b1,bi,bo))),ri,ro)" +
			" REG(SUM(reg(a2,ai,ao),reg(b2,bi,bo),CARRY(reg(a1,ai,ao),reg(b1,bi,bo))),ri,ro)" +
			" CARRY(reg(a2,ai,ao),reg(b2,bi,bo))",
		want: []string{
			"11011010111=>1110111100001", "00001100001=>1010101000001",
			"01101000000=>0010101000000", "01011011110=>0011111100110",
			"10001000100=>0010101000100", "00101110110=>0010110010110",
			"01011010000=>0010100000100", "01010011011=>0000110000110",
			"11111011010=>1111111100111", "01000101000=>1011101000101",
		},
		isValidInt: func() func(inputs map[string]int) []int {
			qa1, qb1, qr1 := 1, 1, 1
			qa2, qb2, qr2 := 1, 1, 1
			return func(inputs map[string]int) []int {
				if inputs["ai"] == 1 {
					qa1, qa2 = inputs["a1"], inputs["a2"]
				}
				if inputs["bi"] == 1 {
					qb1, qb2 = inputs["b1"], inputs["b2"]
				}
				sum1 := qa1 + qb1 + inputs["cin"]
				sum2 := qa2 + qb2 + sum1/2
				if inputs["ri"] == 1 {
					qr1, qr2 = sum1%2, sum2%2
				}
				return []int{
					qa1, inputs["ao"] & qa1, qb1, inputs["bo"] & qb1, qr1, inputs["ro"] & qr1,
					qa2, inputs["ao"] & qa2, qb2, inputs["bo"] & qb2, qr2, inputs["ro"] & qr2,
					sum2 / 2,
				}
			}
		}(),
	}, {
		name: "Alu4",
		desc: "a1 a2 a3 a4 ai ao b1 b2 b3 b4 bi bo ri ro cin" +
			" => reg(a1,ai,ao) REG(a1,ai,ao) reg(b1,bi,bo) REG(b1,bi,bo)" +
			" reg(SUM(reg(a1,ai,ao),reg(b1,bi,bo),cin),ri,ro) REG(SUM(reg(a1,ai,ao),reg(b1,bi,bo),cin),ri,ro)" +
			" reg(a2,ai,ao) REG(a2,ai,ao) reg(b2,bi,bo) REG(b2,bi,bo)" +
			" reg(SUM(reg(a2,ai,ao),reg(b2,bi,bo),CARRY(reg(a1,ai,ao),reg(b1,bi,bo))),ri,ro)" +
			" REG(SUM(reg(a2,ai,ao),reg(b2,bi,bo),CARRY(reg(a1,ai,ao),reg(b1,bi,bo))),ri,ro)" +
			" reg(a3,ai,ao) REG(a3,ai,ao) reg(b3,bi,bo) REG(b3,bi,bo)" +
			" reg(SUM(reg(a3,ai,ao),reg(b3,bi,bo),CARRY(reg(a2,ai,ao),reg(b2,bi,bo))),ri,ro)" +
			" REG(SUM(reg(a3,ai,ao),reg(b3,bi,bo),CARRY(reg(a2,ai,ao),reg(b2,bi,bo))),ri,ro)" +
			" reg(a4,ai,ao) REG(a4,ai,ao) reg(b4,bi,bo) REG(b4,bi,bo)" +
			" reg(SUM(reg(a4,ai,ao),reg(b4,bi,bo),CARRY(reg(a3,ai,ao),reg(b3,bi,bo))),ri,ro)" +
			" REG(SUM(reg(a4,ai,ao),reg(b4,bi,bo),CARRY(reg(a3,ai,ao),reg(b3,bi,bo))),ri,ro)" +
			" CARRY(reg(a4,ai,ao),reg(b4,bi,bo))",
		want: []string{
			"110110101110000=>1010101000100010101010101",
			"110000101101000=>1011101000100011101011101",
			"000010110111101=>0011000011000000100011100",
			"000100010000101=>0010000010000000100010100",
			"110110010110100=>1000101010000000101010001",
			"000101001101111=>1100001111110000111111001",
			"111011010010001=>1100001110101100100000000",
			"010001001111100=>1100101100101111000011001",
			"001010101011100=>0011100000001011000000100",
			"011010011100110=>0010111000111010000000110",
		},
		isValidInt: func() func(inputs map[string]int) []int {
			qa1, qb1, qr1 := 1, 1, 1
			qa2, qb2, qr2 := 1, 1, 1
			qa3, qb3, qr3 := 1, 1, 1
			qa4, qb4, qr4 := 1, 1, 1
			return func(inputs map[string]int) []int {
				if inputs["ai"] == 1 {
					qa1, qa2, qa3, qa4 = inputs["a1"], inputs["a2"], inputs["a3"], inputs["a4"]
				}
				if inputs["bi"] == 1 {
					qb1, qb2, qb3, qb4 = inputs["b1"], inputs["b2"], inputs["b3"], inputs["b4"]
				}
				sum1 := qa1 + qb1 + inputs["cin"]
				sum2 := qa2 + qb2 + sum1/2
				sum3 := qa3 + qb3 + sum2/2
				sum4 := qa4 + qb4 + sum3/2
				if inputs["ri"] == 1 {
					qr1, qr2, qr3, qr4 = sum1%2, sum2%2, sum3%2, sum4%2
				}
				return []int{
					qa1, inputs["ao"] & qa1, qb1, inputs["bo"] & qb1, qr1, inputs["ro"] & qr1,
					qa2, inputs["ao"] & qa2, qb2, inputs["bo"] & qb2, qr2, inputs["ro"] & qr2,
					qa3, inputs["ao"] & qa3, qb3, inputs["bo"] & qb3, qr3, inputs["ro"] & qr3,
					qa4, inputs["ao"] & qa4, qb4, inputs["bo"] & qb4, qr4, inputs["ro"] & qr4,
					sum4 / 2,
				}
			}
		}(),
	}, {
		name: "Alu8",
		desc: "a1 a2 a3 a4 a5 a6 a7 a8 ai ao b1 b2 b3 b4 b5 b6 b7 b8 bi bo ri ro cin" +
			" => reg(a1,ai,ao) REG(a1,ai,ao) reg(b1,bi,bo) REG(b1,bi,bo)" +
			" reg(SUM(reg(a1,ai,ao),reg(b1,bi,bo),cin),ri,ro) REG(SUM(reg(a1,ai,ao),reg(b1,bi,bo),cin),ri,ro)" +
			" reg(a2,ai,ao) REG(a2,ai,ao) reg(b2,bi,bo) REG(b2,bi,bo)" +
			" reg(SUM(reg(a2,ai,ao),reg(b2,bi,bo),CARRY(reg(a1,ai,ao),reg(b1,bi,bo))),ri,ro)" +
			" REG(SUM(reg(a2,ai,ao),reg(b2,bi,bo),CARRY(reg(a1,ai,ao),reg(b1,bi,bo))),ri,ro)" +
			" reg(a3,ai,ao) REG(a3,ai,ao) reg(b3,bi,bo) REG(b3,bi,bo)" +
			" reg(SUM(reg(a3,ai,ao),reg(b3,bi,bo),CARRY(reg(a2,ai,ao),reg(b2,bi,bo))),ri,ro)" +
			" REG(SUM(reg(a3,ai,ao),reg(b3,bi,bo),CARRY(reg(a2,ai,ao),reg(b2,bi,bo))),ri,ro)" +
			" reg(a4,ai,ao) REG(a4,ai,ao) reg(b4,bi,bo) REG(b4,bi,bo)" +
			" reg(SUM(reg(a4,ai,ao),reg(b4,bi,bo),CARRY(reg(a3,ai,ao),reg(b3,bi,bo))),ri,ro)" +
			" REG(SUM(reg(a4,ai,ao),reg(b4,bi,bo),CARRY(reg(a3,ai,ao),reg(b3,bi,bo))),ri,ro)" +
			" reg(a5,ai,ao) REG(a5,ai,ao) reg(b5,bi,bo) REG(b5,bi,bo)" +
			" reg(SUM(reg(a5,ai,ao),reg(b5,bi,bo),CARRY(reg(a4,ai,ao),reg(b4,bi,bo))),ri,ro)" +
			" REG(SUM(reg(a5,ai,ao),reg(b5,bi,bo),CARRY(reg(a4,ai,ao),reg(b4,bi,bo))),ri,ro)" +
			" reg(a6,ai,ao) REG(a6,ai,ao) reg(b6,bi,bo) REG(b6,bi,bo)" +
			" reg(SUM(reg(a6,ai,ao),reg(b6,bi,bo),CARRY(reg(a5,ai,ao),reg(b5,bi,bo))),ri,ro)" +
			" REG(SUM(reg(a6,ai,ao),reg(b6,bi,bo),CARRY(reg(a5,ai,ao),reg(b5,bi,bo))),ri,ro)" +
			" reg(a7,ai,ao) REG(a7,ai,ao) reg(b7,bi,bo) REG(b7,bi,bo)" +
			" reg(SUM(reg(a7,ai,ao),reg(b7,bi,bo),CARRY(reg(a6,ai,ao),reg(b6,bi,bo))),ri,ro)" +
			" REG(SUM(reg(a7,ai,ao),reg(b7,bi,bo),CARRY(reg(a6,ai,ao),reg(b6,bi,bo))),ri,ro)" +
			" reg(a8,ai,ao) REG(a8,ai,ao) reg(b8,bi,bo) REG(b8,bi,bo)" +
			" reg(SUM(reg(a8,ai,ao),reg(b8,bi,bo),CARRY(reg(a7,ai,ao),reg(b7,bi,bo))),ri,ro)" +
			" REG(SUM(reg(a8,ai,ao),reg(b8,bi,bo),CARRY(reg(a7,ai,ao),reg(b7,bi,bo))),ri,ro)" +
			" CARRY(reg(a8,ai,ao),reg(b8,bi,bo))",
		want: []string{
			"11011010111000011000010=>1110111110110010111110111110110010111110110010111",
			"11010000000101101111010=>1000111011110000111011111011110000111011110011111",
			"00100010000101110110010=>1000111010110000111010111010110010111000110010111",
			"11010000010100110111111=>1100001111110000111100111111000011001100000011001",
			"10110100100010100010011=>1000000000111010111000110010001000000000000000000",
			"11100001010101011100011=>1100000000111110111100110010001100000000000000000",
			"01001110011000011100100=>1100100000001110001100000010001100000000100000000",
			"11000011000100001000001=>1000100000001010001000000010001000000000100000000",
			"00101101010010011010101=>1100000000101110001100000000101110000010000000100",
			"11100011000101110101001=>1000000000101011001000000000101011000011000000100",
		},
		isValidInt: func() func(inputs map[string]int) []int {
			qa1, qb1, qr1 := 1, 1, 1
			qa2, qb2, qr2 := 1, 1, 1
			qa3, qb3, qr3 := 1, 1, 1
			qa4, qb4, qr4 := 1, 1, 1
			qa5, qb5, qr5 := 1, 1, 1
			qa6, qb6, qr6 := 1, 1, 1
			qa7, qb7, qr7 := 1, 1, 1
			qa8, qb8, qr8 := 1, 1, 1
			return func(inputs map[string]int) []int {
				if inputs["ai"] == 1 {
					qa1, qa2, qa3, qa4 = inputs["a1"], inputs["a2"], inputs["a3"], inputs["a4"]
					qa5, qa6, qa7, qa8 = inputs["a5"], inputs["a6"], inputs["a7"], inputs["a8"]
				}
				if inputs["bi"] == 1 {
					qb1, qb2, qb3, qb4 = inputs["b1"], inputs["b2"], inputs["b3"], inputs["b4"]
					qb5, qb6, qb7, qb8 = inputs["b5"], inputs["b6"], inputs["b7"], inputs["b8"]
				}
				sum1 := qa1 + qb1 + inputs["cin"]
				sum2 := qa2 + qb2 + sum1/2
				sum3 := qa3 + qb3 + sum2/2
				sum4 := qa4 + qb4 + sum3/2
				sum5 := qa5 + qb5 + sum4/2
				sum6 := qa6 + qb6 + sum5/2
				sum7 := qa7 + qb7 + sum6/2
				sum8 := qa8 + qb8 + sum7/2
				if inputs["ri"] == 1 {
					qr1, qr2, qr3, qr4 = sum1%2, sum2%2, sum3%2, sum4%2
					qr5, qr6, qr7, qr8 = sum5%2, sum6%2, sum7%2, sum8%2
				}
				return []int{
					qa1, inputs["ao"] & qa1, qb1, inputs["bo"] & qb1, qr1, inputs["ro"] & qr1,
					qa2, inputs["ao"] & qa2, qb2, inputs["bo"] & qb2, qr2, inputs["ro"] & qr2,
					qa3, inputs["ao"] & qa3, qb3, inputs["bo"] & qb3, qr3, inputs["ro"] & qr3,
					qa4, inputs["ao"] & qa4, qb4, inputs["bo"] & qb4, qr4, inputs["ro"] & qr4,
					qa5, inputs["ao"] & qa5, qb5, inputs["bo"] & qb5, qr5, inputs["ro"] & qr5,
					qa6, inputs["ao"] & qa6, qb6, inputs["bo"] & qb6, qr6, inputs["ro"] & qr6,
					qa7, inputs["ao"] & qa7, qb7, inputs["bo"] & qb7, qr7, inputs["ro"] & qr7,
					qa8, inputs["ao"] & qa8, qb8, inputs["bo"] & qb8, qr8, inputs["ro"] & qr8,
					sum8 / 2,
				}
			}
		}(),
	}, {
		name: "Bus",
		desc: "bus a b r => BUS(bus) wa wb",
		want: []string{
			"0000=>000", "0001=>111", "0010=>111", "0011=>111",
			"0100=>111", "0101=>111", "0110=>111", "0111=>111",
			"1000=>111", "1001=>111", "1010=>111", "1011=>111",
			"1100=>111", "1101=>111", "1110=>111", "1111=>111",
		},
	}, {
		name: "Bus2",
		desc: "bus1 bus2 a1 a2 b1 b2 r1 r2 => BUS(bus1) BUS(bus2) wa1 wa2 wb1 wb2",
		want: []string{
			"11011010=>111111", "11100001=>111111", "10000101=>111111", "10100000=>101010",
			"00101101=>111111", "11101000=>111111", "10001000=>101010", "01011101=>111111",
			"10010110=>111111", "10000010=>101010",
		},
	}, {
		name: "Bus4",
		desc: "bus1 bus2 bus3 bus4 a1 a2 a3 a4 b1 b2 b3 b4 r1 r2 r3 r4" +
			" => BUS(bus1) BUS(bus2) BUS(bus3) BUS(bus4) wa1 wa2 wa3 wa4 wb1 wb2 wb3 wb4",
		want: []string{
			"1101101011100001=>111111111111", "1000010110100000=>111111111111",
			"0010110111101000=>111111111111", "1000100001011101=>110111011101",
			"1001011010000010=>111111111111", "1001101111111011=>111111111111",
			"0100100010100010=>111011101110", "0111110000101010=>111111111111",
			"1011100011010011=>111111111111", "1001100001110010=>111111111111",
		},
	}, {
		name: "Bus8",
		desc: "bus1 bus2 bus3 bus4 bus5 bus6 bus7 bus8 a1 a2 a3 a4 a5 a6 a7 a8 b1 b2 b3 b4 b5 b6 b7 b8 r1 r2 r3 r4 r5 r6 r7 r8" +
			" => BUS(bus1) BUS(bus2) BUS(bus3) BUS(bus4) BUS(bus5) BUS(bus6) BUS(bus7) BUS(bus8)" +
			" wa1 wa2 wa3 wa4 wa5 wa6 wa7 wa8 wb1 wb2 wb3 wb4 wb5 wb6 wb7 wb8",
		want: []string{
			"11011010111000011000010110100000=>111111111111111111111111",
			"00101101111010001000100001011101=>111111011111110111111101",
			"10010110100000101001101111111011=>111111111111111111111111",
			"01001000101000100111110000101010=>111111101111111011111110",
			"10111000110100111001100001110010=>111110111111101111111011",
			"01100001100010000100000100101101=>111011011110110111101101",
			"01001001101010111100011000101110=>111011111110111111101111",
			"10100110010110010110010101101000=>111111111111111111111111",
			"00101011100110001100001100111000=>111110111111101111111011",
			"11001100000110001010100010000111=>111111111111111111111111",
		},
	}, {
		name: "AluWithBus",
		desc: "bus ai ao bi bo ri ro cin" +
			" => BUS(bus) reg(ALU-bus-a,ai,ao) REG(ALU-bus-a,ai,ao) reg(ALU-bus-b,bi,bo) REG(ALU-bus-b,bi,bo)" +
			" reg(SUM(reg(ALU-bus-a,ai,ao),reg(ALU-bus-b,bi,bo),cin),ri,ro)" +
			" REG(SUM(reg(ALU-bus-a,ai,ao),reg(ALU-bus-b,bi,bo),cin),ri,ro)" +
			" CARRY(reg(ALU-bus-a,ai,ao),reg(ALU-bus-b,bi,bo))",
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
			" => BUS(bus1) reg(ALU-bus1-a,ai,ao) REG(ALU-bus1-a,ai,ao) reg(ALU-bus1-b,bi,bo) REG(ALU-bus1-b,bi,bo)" +
			" reg(SUM(reg(ALU-bus1-a,ai,ao),reg(ALU-bus1-b,bi,bo),cin),ri,ro)" +
			" REG(SUM(reg(ALU-bus1-a,ai,ao),reg(ALU-bus1-b,bi,bo),cin),ri,ro)" +
			" BUS(bus2) reg(ALU-bus2-a,ai,ao) REG(ALU-bus2-a,ai,ao) reg(ALU-bus2-b,bi,bo) REG(ALU-bus2-b,bi,bo)" +
			" reg(SUM(reg(ALU-bus2-a,ai,ao),reg(ALU-bus2-b,bi,bo),CARRY(reg(ALU-bus1-a,ai,ao),reg(ALU-bus1-b,bi,bo))),ri,ro)" +
			" REG(SUM(reg(ALU-bus2-a,ai,ao),reg(ALU-bus2-b,bi,bo),CARRY(reg(ALU-bus1-a,ai,ao),reg(ALU-bus1-b,bi,bo))),ri,ro)" +
			" CARRY(reg(ALU-bus2-a,ai,ao),reg(ALU-bus2-b,bi,bo))",
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
		name: "AluWithBus4",
		desc: "bus1 bus2 bus3 bus4 ai ao bi bo ri ro cin" +
			" => BUS(bus1) reg(ALU-bus1-a,ai,ao) REG(ALU-bus1-a,ai,ao) reg(ALU-bus1-b,bi,bo) REG(ALU-bus1-b,bi,bo)" +
			" reg(SUM(reg(ALU-bus1-a,ai,ao),reg(ALU-bus1-b,bi,bo),cin),ri,ro)" +
			" REG(SUM(reg(ALU-bus1-a,ai,ao),reg(ALU-bus1-b,bi,bo),cin),ri,ro)" +
			" BUS(bus2) reg(ALU-bus2-a,ai,ao) REG(ALU-bus2-a,ai,ao) reg(ALU-bus2-b,bi,bo) REG(ALU-bus2-b,bi,bo)" +
			" reg(SUM(reg(ALU-bus2-a,ai,ao),reg(ALU-bus2-b,bi,bo),CARRY(reg(ALU-bus1-a,ai,ao),reg(ALU-bus1-b,bi,bo))),ri,ro)" +
			" REG(SUM(reg(ALU-bus2-a,ai,ao),reg(ALU-bus2-b,bi,bo),CARRY(reg(ALU-bus1-a,ai,ao),reg(ALU-bus1-b,bi,bo))),ri,ro)" +
			" BUS(bus3) reg(ALU-bus3-a,ai,ao) REG(ALU-bus3-a,ai,ao) reg(ALU-bus3-b,bi,bo) REG(ALU-bus3-b,bi,bo)" +
			" reg(SUM(reg(ALU-bus3-a,ai,ao),reg(ALU-bus3-b,bi,bo),CARRY(reg(ALU-bus2-a,ai,ao),reg(ALU-bus2-b,bi,bo))),ri,ro)" +
			" REG(SUM(reg(ALU-bus3-a,ai,ao),reg(ALU-bus3-b,bi,bo),CARRY(reg(ALU-bus2-a,ai,ao),reg(ALU-bus2-b,bi,bo))),ri,ro)" +
			" BUS(bus4) reg(ALU-bus4-a,ai,ao) REG(ALU-bus4-a,ai,ao) reg(ALU-bus4-b,bi,bo) REG(ALU-bus4-b,bi,bo)" +
			" reg(SUM(reg(ALU-bus4-a,ai,ao),reg(ALU-bus4-b,bi,bo),CARRY(reg(ALU-bus3-a,ai,ao),reg(ALU-bus3-b,bi,bo))),ri,ro)" +
			" REG(SUM(reg(ALU-bus4-a,ai,ao),reg(ALU-bus4-b,bi,bo),CARRY(reg(ALU-bus3-a,ai,ao),reg(ALU-bus3-b,bi,bo))),ri,ro)" +
			" CARRY(reg(ALU-bus4-a,ai,ao),reg(ALU-bus4-b,bi,bo))",
		want: []string{
			"01101000000=>00010101101010110101000010101", "10001000100=>11010000001000000100000010001",
			"01011010000=>00000001101000000000011010001", "01000101000=>00000001111100000000011111001",
		},
		isValidInt: func() func(inputs map[string]int) []int {
			qa1, qb1, qr1 := 1, 1, 1
			qa2, qb2, qr2 := 1, 1, 1
			qa3, qb3, qr3 := 1, 1, 1
			qa4, qb4, qr4 := 1, 1, 1
			return func(inputs map[string]int) []int {
				ra1, rb1, rr1, bus1, sum1 := 0, 0, 0, 0, 0
				ra2, rb2, rr2, bus2, sum2 := 0, 0, 0, 0, 0
				ra3, rb3, rr3, bus3, sum3 := 0, 0, 0, 0, 0
				ra4, rb4, rr4, bus4, sum4 := 0, 0, 0, 0, 0
				for i := 0; i < 10; i++ {
					if inputs["ao"] == 1 {
						ra1, ra2, ra3, ra4 = qa1, qa2, qa3, qa4
					}
					if inputs["bo"] == 1 {
						rb1, rb2, rb3, rb4 = qb1, qb2, qb3, qb4
					}
					if inputs["ro"] == 1 {
						rr1, rr2, rr3, rr4 = qr1, qr2, qr3, qr4
					}
					bus1 = inputs["bus1"] | ra1 | rb1 | rr1
					bus2 = inputs["bus2"] | ra2 | rb2 | rr2
					bus3 = inputs["bus3"] | ra3 | rb3 | rr3
					bus4 = inputs["bus4"] | ra4 | rb4 | rr4
					if inputs["ai"] == 1 {
						qa1, qa2, qa3, qa4 = bus1, bus2, bus3, bus4
					}
					if inputs["bi"] == 1 {
						qb1, qb2, qb3, qb4 = bus1, bus2, bus3, bus4
					}
					sum1 = qa1 + qb1 + inputs["cin"]
					sum2 = qa2 + qb2 + sum1/2
					sum3 = qa3 + qb3 + sum2/2
					sum4 = qa4 + qb4 + sum3/2
					if inputs["ri"] == 1 {
						qr1, qr2, qr3, qr4 = sum1%2, sum2%2, sum3%2, sum4%2
					}
				}
				return []int{
					bus1, qa1, ra1, qb1, rb1, qr1, rr1,
					bus2, qa2, ra2, qb2, rb2, qr2, rr2,
					bus3, qa3, ra3, qb3, rb3, qr3, rr3,
					bus4, qa4, ra4, qb4, rb4, qr4, rr4,
					sum4 / 2}
			}
		}(),
	}, {
		name: "AluWithBus8",
		desc: "bus1 bus2 bus3 bus4 bus5 bus6 bus7 bus8 ai ao bi bo ri ro cin" +
			" => BUS(bus1) reg(ALU-bus1-a,ai,ao) REG(ALU-bus1-a,ai,ao) reg(ALU-bus1-b,bi,bo) REG(ALU-bus1-b,bi,bo)" +
			" reg(SUM(reg(ALU-bus1-a,ai,ao),reg(ALU-bus1-b,bi,bo),cin),ri,ro)" +
			" REG(SUM(reg(ALU-bus1-a,ai,ao),reg(ALU-bus1-b,bi,bo),cin),ri,ro)" +
			" BUS(bus2) reg(ALU-bus2-a,ai,ao) REG(ALU-bus2-a,ai,ao) reg(ALU-bus2-b,bi,bo) REG(ALU-bus2-b,bi,bo)" +
			" reg(SUM(reg(ALU-bus2-a,ai,ao),reg(ALU-bus2-b,bi,bo),CARRY(reg(ALU-bus1-a,ai,ao),reg(ALU-bus1-b,bi,bo))),ri,ro)" +
			" REG(SUM(reg(ALU-bus2-a,ai,ao),reg(ALU-bus2-b,bi,bo),CARRY(reg(ALU-bus1-a,ai,ao),reg(ALU-bus1-b,bi,bo))),ri,ro)" +
			" BUS(bus3) reg(ALU-bus3-a,ai,ao) REG(ALU-bus3-a,ai,ao) reg(ALU-bus3-b,bi,bo) REG(ALU-bus3-b,bi,bo)" +
			" reg(SUM(reg(ALU-bus3-a,ai,ao),reg(ALU-bus3-b,bi,bo),CARRY(reg(ALU-bus2-a,ai,ao),reg(ALU-bus2-b,bi,bo))),ri,ro)" +
			" REG(SUM(reg(ALU-bus3-a,ai,ao),reg(ALU-bus3-b,bi,bo),CARRY(reg(ALU-bus2-a,ai,ao),reg(ALU-bus2-b,bi,bo))),ri,ro)" +
			" BUS(bus4) reg(ALU-bus4-a,ai,ao) REG(ALU-bus4-a,ai,ao) reg(ALU-bus4-b,bi,bo) REG(ALU-bus4-b,bi,bo)" +
			" reg(SUM(reg(ALU-bus4-a,ai,ao),reg(ALU-bus4-b,bi,bo),CARRY(reg(ALU-bus3-a,ai,ao),reg(ALU-bus3-b,bi,bo))),ri,ro)" +
			" REG(SUM(reg(ALU-bus4-a,ai,ao),reg(ALU-bus4-b,bi,bo),CARRY(reg(ALU-bus3-a,ai,ao),reg(ALU-bus3-b,bi,bo))),ri,ro)" +
			" BUS(bus5) reg(ALU-bus5-a,ai,ao) REG(ALU-bus5-a,ai,ao) reg(ALU-bus5-b,bi,bo) REG(ALU-bus5-b,bi,bo)" +
			" reg(SUM(reg(ALU-bus5-a,ai,ao),reg(ALU-bus5-b,bi,bo),CARRY(reg(ALU-bus4-a,ai,ao),reg(ALU-bus4-b,bi,bo))),ri,ro)" +
			" REG(SUM(reg(ALU-bus5-a,ai,ao),reg(ALU-bus5-b,bi,bo),CARRY(reg(ALU-bus4-a,ai,ao),reg(ALU-bus4-b,bi,bo))),ri,ro)" +
			" BUS(bus6) reg(ALU-bus6-a,ai,ao) REG(ALU-bus6-a,ai,ao) reg(ALU-bus6-b,bi,bo) REG(ALU-bus6-b,bi,bo)" +
			" reg(SUM(reg(ALU-bus6-a,ai,ao),reg(ALU-bus6-b,bi,bo),CARRY(reg(ALU-bus5-a,ai,ao),reg(ALU-bus5-b,bi,bo))),ri,ro)" +
			" REG(SUM(reg(ALU-bus6-a,ai,ao),reg(ALU-bus6-b,bi,bo),CARRY(reg(ALU-bus5-a,ai,ao),reg(ALU-bus5-b,bi,bo))),ri,ro)" +
			" BUS(bus7) reg(ALU-bus7-a,ai,ao) REG(ALU-bus7-a,ai,ao) reg(ALU-bus7-b,bi,bo) REG(ALU-bus7-b,bi,bo)" +
			" reg(SUM(reg(ALU-bus7-a,ai,ao),reg(ALU-bus7-b,bi,bo),CARRY(reg(ALU-bus6-a,ai,ao),reg(ALU-bus6-b,bi,bo))),ri,ro)" +
			" REG(SUM(reg(ALU-bus7-a,ai,ao),reg(ALU-bus7-b,bi,bo),CARRY(reg(ALU-bus6-a,ai,ao),reg(ALU-bus6-b,bi,bo))),ri,ro)" +
			" BUS(bus8) reg(ALU-bus8-a,ai,ao) REG(ALU-bus8-a,ai,ao) reg(ALU-bus8-b,bi,bo) REG(ALU-bus8-b,bi,bo)" +
			" reg(SUM(reg(ALU-bus8-a,ai,ao),reg(ALU-bus8-b,bi,bo),CARRY(reg(ALU-bus7-a,ai,ao),reg(ALU-bus7-b,bi,bo))),ri,ro)" +
			" REG(SUM(reg(ALU-bus8-a,ai,ao),reg(ALU-bus8-b,bi,bo),CARRY(reg(ALU-bus7-a,ai,ao),reg(ALU-bus7-b,bi,bo))),ri,ro)" +
			" CARRY(reg(ALU-bus8-a,ai,ao),reg(ALU-bus8-b,bi,bo))",
		want: []string{
			"000100010000101=>010101001010100101010110101001010100101010010101011010101",
			"110110010110100=>111100011110101111010111101011110101111010111101011110101",
			"111011010010001=>110100011010101101010010001011010101101010010001011010101",
		},
		isValidInt: func() func(inputs map[string]int) []int {
			qa1, qb1, qr1 := 1, 1, 1
			qa2, qb2, qr2 := 1, 1, 1
			qa3, qb3, qr3 := 1, 1, 1
			qa4, qb4, qr4 := 1, 1, 1
			qa5, qb5, qr5 := 1, 1, 1
			qa6, qb6, qr6 := 1, 1, 1
			qa7, qb7, qr7 := 1, 1, 1
			qa8, qb8, qr8 := 1, 1, 1
			return func(inputs map[string]int) []int {
				ra1, rb1, rr1, bus1, sum1 := 0, 0, 0, 0, 0
				ra2, rb2, rr2, bus2, sum2 := 0, 0, 0, 0, 0
				ra3, rb3, rr3, bus3, sum3 := 0, 0, 0, 0, 0
				ra4, rb4, rr4, bus4, sum4 := 0, 0, 0, 0, 0
				ra5, rb5, rr5, bus5, sum5 := 0, 0, 0, 0, 0
				ra6, rb6, rr6, bus6, sum6 := 0, 0, 0, 0, 0
				ra7, rb7, rr7, bus7, sum7 := 0, 0, 0, 0, 0
				ra8, rb8, rr8, bus8, sum8 := 0, 0, 0, 0, 0
				for i := 0; i < 10; i++ {
					if inputs["ao"] == 1 {
						ra1, ra2, ra3, ra4 = qa1, qa2, qa3, qa4
						ra5, ra6, ra7, ra8 = qa5, qa6, qa7, qa8
					}
					if inputs["bo"] == 1 {
						rb1, rb2, rb3, rb4 = qb1, qb2, qb3, qb4
						rb5, rb6, rb7, rb8 = qb5, qb6, qb7, qb8
					}
					if inputs["ro"] == 1 {
						rr1, rr2, rr3, rr4 = qr1, qr2, qr3, qr4
						rr5, rr6, rr7, rr8 = qr5, qr6, qr7, qr8
					}
					bus1 = inputs["bus1"] | ra1 | rb1 | rr1
					bus2 = inputs["bus2"] | ra2 | rb2 | rr2
					bus3 = inputs["bus3"] | ra3 | rb3 | rr3
					bus4 = inputs["bus4"] | ra4 | rb4 | rr4
					bus5 = inputs["bus5"] | ra5 | rb5 | rr5
					bus6 = inputs["bus6"] | ra6 | rb6 | rr6
					bus7 = inputs["bus7"] | ra7 | rb7 | rr7
					bus8 = inputs["bus8"] | ra8 | rb8 | rr8
					if inputs["ai"] == 1 {
						qa1, qa2, qa3, qa4 = bus1, bus2, bus3, bus4
						qa5, qa6, qa7, qa8 = bus5, bus6, bus7, bus8
					}
					if inputs["bi"] == 1 {
						qb1, qb2, qb3, qb4 = bus1, bus2, bus3, bus4
						qb5, qb6, qb7, qb8 = bus5, bus6, bus7, bus8
					}
					sum1 = qa1 + qb1 + inputs["cin"]
					sum2 = qa2 + qb2 + sum1/2
					sum3 = qa3 + qb3 + sum2/2
					sum4 = qa4 + qb4 + sum3/2
					sum5 = qa5 + qb5 + sum4/2
					sum6 = qa6 + qb6 + sum5/2
					sum7 = qa7 + qb7 + sum6/2
					sum8 = qa8 + qb8 + sum7/2
					if inputs["ri"] == 1 {
						qr1, qr2, qr3, qr4 = sum1%2, sum2%2, sum3%2, sum4%2
						qr5, qr6, qr7, qr8 = sum5%2, sum2%2, sum7%2, sum8%2
					}
				}
				return []int{
					bus1, qa1, ra1, qb1, rb1, qr1, rr1,
					bus2, qa2, ra2, qb2, rb2, qr2, rr2,
					bus3, qa3, ra3, qb3, rb3, qr3, rr3,
					bus4, qa4, ra4, qb4, rb4, qr4, rr4,
					bus5, qa5, ra5, qb5, rb5, qr5, rr5,
					bus6, qa6, ra6, qb6, rb6, qr6, rr6,
					bus7, qa7, ra7, qb7, rb7, qr7, rr7,
					bus8, qa8, ra8, qb8, rb8, qr8, rr8,
					sum8 / 2}
			}
		}(),
	}, {
		name: "RAM",
		desc: "a d ei eo" +
			" => RAM RAM-s0 RAM-ei0 RAM-eo0 reg(d,RAM-ei0,RAM-eo0) REG(d,RAM-ei0,RAM-eo0)" +
			" RAM-s1 RAM-ei1 RAM-eo1 reg(d,RAM-ei1,RAM-eo1) REG(d,RAM-ei1,RAM-eo1)",
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
			" => RAM RAM-s0 RAM-ei0 RAM-eo0 reg(d,RAM-ei0,RAM-eo0) REG(d,RAM-ei0,RAM-eo0)" +
			" RAM-s1 RAM-ei1 RAM-eo1 reg(d,RAM-ei1,RAM-eo1) REG(d,RAM-ei1,RAM-eo1)" +
			" RAM-s2 RAM-ei2 RAM-eo2 reg(d,RAM-ei2,RAM-eo2) REG(d,RAM-ei2,RAM-eo2)" +
			" RAM-s3 RAM-ei3 RAM-eo3 reg(d,RAM-ei3,RAM-eo3) REG(d,RAM-ei3,RAM-eo3)",
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
			" reg(d0,RAM-ei0,RAM-eo0) REG(d0,RAM-ei0,RAM-eo0)" +
			" reg(d1,RAM-ei0,RAM-eo0) REG(d1,RAM-ei0,RAM-eo0)" +
			" RAM-s1 RAM-ei1 RAM-eo1" +
			" reg(d0,RAM-ei1,RAM-eo1) REG(d0,RAM-ei1,RAM-eo1)" +
			" reg(d1,RAM-ei1,RAM-eo1) REG(d1,RAM-ei1,RAM-eo1)",
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
			" reg(d0,RAM-ei0,RAM-eo0) REG(d0,RAM-ei0,RAM-eo0)" +
			" reg(d1,RAM-ei0,RAM-eo0) REG(d1,RAM-ei0,RAM-eo0)" +
			" RAM-s1 RAM-ei1 RAM-eo1" +
			" reg(d0,RAM-ei1,RAM-eo1) REG(d0,RAM-ei1,RAM-eo1)" +
			" reg(d1,RAM-ei1,RAM-eo1) REG(d1,RAM-ei1,RAM-eo1)" +
			" RAM-s2 RAM-ei2 RAM-eo2" +
			" reg(d0,RAM-ei2,RAM-eo2) REG(d0,RAM-ei2,RAM-eo2)" +
			" reg(d1,RAM-ei2,RAM-eo2) REG(d1,RAM-ei2,RAM-eo2)" +
			" RAM-s3 RAM-ei3 RAM-eo3" +
			" reg(d0,RAM-ei3,RAM-eo3) REG(d0,RAM-ei3,RAM-eo3)" +
			" reg(d1,RAM-ei3,RAM-eo3) REG(d1,RAM-ei3,RAM-eo3)",
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
				one = append(one, " => ")
				for i, output := range c.Outputs {
					one = append(one, sfmt.Sprintf("%s(%s)", output.Name, string(out[i+len(c.Inputs)+len("=>")])))
				}
				converted = append(converted, strings.Join(one, ""))
			}
		} else {
			converted = got
		}
		if diff := cmp.Diff(in.want, converted); diff != "" {
			t.Errorf("Simulate(%q) want %#v,\ngot %#v,\ndiff -want +got:\n%s", in.name, in.want, got, diff)
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
			" => BUS(bus) reg(ALU-bus-a,ai,ao) REG(ALU-bus-a,ai,ao) reg(ALU-bus-b,bi,bo) REG(ALU-bus-b,bi,bo)" +
			" reg(SUM(reg(ALU-bus-a,ai,ao),reg(ALU-bus-b,bi,bo),cin),ri,ro)" +
			" REG(SUM(reg(ALU-bus-a,ai,ao),reg(ALU-bus-b,bi,bo),cin),ri,ro)" +
			" CARRY(reg(ALU-bus-a,ai,ao),reg(ALU-bus-b,bi,bo))",
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
			" reg(d,RAM-ei0,RAM-eo0) REG(d,RAM-ei0,RAM-eo0)" +
			" RAM-s1 RAM-ei1 RAM-eo1" +
			" reg(d,RAM-ei1,RAM-eo1) REG(d,RAM-ei1,RAM-eo1)",
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
