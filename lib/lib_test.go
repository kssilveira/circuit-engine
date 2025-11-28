package lib

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/kssilveira/circuit-engine/circuit"
	"github.com/kssilveira/circuit-engine/config"
)

func TestOutputsCombinational(t *testing.T) {
	inputs := []struct {
		name        string
		want        []string
		isValidInt  func(inputs map[string]int) []int
		isValidBool func(inputs map[string]bool) []bool
	}{{
		name: "TransistorEmitter",
		// base collector => emitter
		want: []string{"00=>0", "01=>0", "10=>0", "11=>1"},
		isValidBool: func(inputs map[string]bool) []bool {
			return []bool{inputs["base"] && inputs["collector"]}
		},
	}, {
		name: "TransistorGnd",
		// base collector => collectorOut
		want: []string{"00=>0", "01=>1", "10=>0", "11=>0"},
		isValidBool: func(inputs map[string]bool) []bool {
			return []bool{!inputs["base"] && inputs["collector"]}
		},
	}, {
		name: "Transistor",
		// base collector => emitter collectorOut
		want: []string{"00=>00", "01=>01", "10=>00", "11=>11"},
		isValidBool: func(inputs map[string]bool) []bool {
			return []bool{inputs["base"] && inputs["collector"], inputs["collector"]}
		},
	}, {
		name: "Not",
		want: []string{"0=>1", "1=>0"},
		isValidBool: func(inputs map[string]bool) []bool {
			return []bool{!inputs["a"]}
		},
	}, {
		name: "And",
		want: []string{"00=>0", "01=>0", "10=>0", "11=>1"},
		isValidBool: func(inputs map[string]bool) []bool {
			return []bool{inputs["a"] && inputs["b"]}
		},
	}, {
		name: "Or",
		want: []string{"00=>0", "01=>1", "10=>1", "11=>1"},
		isValidBool: func(inputs map[string]bool) []bool {
			return []bool{inputs["a"] || inputs["b"]}
		},
	}, {
		name: "OrRes",
		want: []string{"0=>0", "1=>1"},
		isValidBool: func(inputs map[string]bool) []bool {
			return []bool{inputs["a"]}
		},
	}, {
		name: "Nand",
		want: []string{"00=>1", "01=>1", "10=>1", "11=>0"},
		isValidBool: func(inputs map[string]bool) []bool {
			return []bool{!(inputs["a"] && inputs["b"])}
		},
	}, {
		name: "Nand(Nand)",
		want: []string{"000=>1", "001=>1", "010=>1", "011=>1", "100=>0", "101=>0", "110=>0", "111=>1"},
		isValidBool: func(inputs map[string]bool) []bool {
			return []bool{!(inputs["a"] && !(inputs["b"] && inputs["c"]))}
		},
	}, {
		name: "Xor",
		want: []string{"00=>0", "01=>1", "10=>1", "11=>0"},
		isValidBool: func(inputs map[string]bool) []bool {
			return []bool{inputs["a"] != inputs["b"]}
		},
	}, {
		name: "Nor",
		want: []string{"00=>1", "01=>0", "10=>0", "11=>0"},
		isValidBool: func(inputs map[string]bool) []bool {
			return []bool{!(inputs["a"] || inputs["b"])}
		},
	}, {
		name: "HalfSum",
		// a b => s carry
		want: []string{"00=>00", "01=>10", "10=>10", "11=>01"},
		isValidInt: func(inputs map[string]int) []int {
			sum := inputs["a"] + inputs["b"]
			return []int{sum % 2, sum / 2}
		},
	}, {
		name: "Sum",
		// a b cin => s cout
		want: []string{"000=>00", "001=>10", "010=>10", "011=>01", "100=>10", "101=>01", "110=>01", "111=>11"},
		isValidInt: func(inputs map[string]int) []int {
			sum := inputs["a"] + inputs["b"] + inputs["c"]
			return []int{sum % 2, sum / 2}
		},
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
		isValidInt: func(inputs map[string]int) []int {
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
		isValidInt: func(inputs map[string]int) []int {
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
		isValidInt: func(inputs map[string]int) []int {
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
		// s r e => q !q
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
		// d e => q !q
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
		// d ei eo => q r
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
		// d1 d2 ei eo => q1 r1 q2 r2
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
		// d1 d2 d3 d4 ei eo => q1 r1 q2 r2 q3 r3 q4 r4
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
		// d1 d2 d3 d4 d5 d6 d7 d8 ei eo => q1 r1 q2 r2 q3 r3 q4 r4 q5 r5 q6 r6 q7 r7 q8 r8
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
		// a ai ao b bi bo ri ro cin => qa ra qb rb qr rr cout
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
		// a1 a2 ai ao b1 b2 bi bo ri ro carry => qa1 ra1 qb1 rb1 qr1 rr1 qa2 ra2 qb2 rb2 qr2 rr2
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
		// a1 a2 a3 a4 ai ao b1 b2 b3 b4 bi bo ri ro carry
		// => qa1 ra1 qb1 rb1 qr1 rr1 qa2 ra2 qb2 rb2 qr2 rr2
		//    qa3 ra3 qb3 rb3 qr3 rr3 qa4 ra4 qb4 rb4 qr4 rr4
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
		// a1 a2 a3 a4 a5 a6 a7 a8 ai ao b1 b2 b3 b4 b5 b6 b7 b8 bi bo ri ro carry
		// => qa1 ra1 qb1 rb1 qr1 rr1 qa2 ra2 qb2 rb2 qr2 rr2
		//    qa3 ra3 qb3 rb3 qr3 rr3 qa4 ra4 qb4 rb4 qr4 rr4
		//    qa5 ra5 qb5 rb5 qr5 rr5 qa6 ra6 qb6 rb6 qr6 rr6
		//    qa7 ra7 qb7 rb7 qr7 rr7 qa8 ra8 qb8 rb8 qr8 rr8
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
		// bus a b r => rbus wa wb
		want: []string{
			"0000=>000", "0001=>111", "0010=>111", "0011=>111",
			"0100=>111", "0101=>111", "0110=>111", "0111=>111",
			"1000=>111", "1001=>111", "1010=>111", "1011=>111",
			"1100=>111", "1101=>111", "1110=>111", "1111=>111",
		},
	}, {
		name: "AluWithBus",
		// bus ai ao bi bo ri ro cin => rbus qa ra qb rb qr rr cout
		want: []string{
			"11011010=>11011111", "11100001=>11110101", "10000101=>11010101", "10100000=>11110101",
			"00101101=>11111101", "11101000=>11111101", "10001000=>11011101", "01011101=>11011101",
			"10000010=>11010111",
		},
		isValidInt: func() func(inputs map[string]int) []int {
			qa, qb, qr := 1, 1, 1
			return func(inputs map[string]int) []int {
				ra, rb, rr, bus, sum := 0, 0, 0, 0, 0
				for i := 0; i < 10; i++ {
					ra = inputs["ao"] & qa
					rb = inputs["bo"] & qb
					rr = inputs["ro"] & qr
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
	}}
	for _, in := range inputs {
		c := circuit.NewCircuit(config.Config{IsUnitTest: true})
		c.Outs(Example(c, in.name))
		got := c.Simulate()
		if diff := cmp.Diff(in.want, got); diff != "" {
			t.Errorf("Simulate(%q) want %#v,\ngot %#v,\ndiff -want +got:\n%s", in.name, in.want, got, diff)
		}
		if in.isValidInt != nil {
			for _, out := range got {
				inputs := map[string]int{}
				for i, input := range c.Inputs {
					inputs[input.Name] = int(out[i] - '0')
				}
				wants := in.isValidInt(inputs)
				for i, want := range wants {
					got := int(out[len(c.Inputs)+len("=>")+i] - '0')
					if want != got {
						t.Errorf("Simulate(%q) out %s output %s want %d got %#v", in.name, out, c.Outputs[i].Name, want, got)
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
	}, {
		name: "AluWithBus",
		// bus ai ao bi bo ri ro cin => rbus qa ra qb rb qr rr cout
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
