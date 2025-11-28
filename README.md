# circuit-engine

Define and simulate circuits from transistors all the way to an 8-bit computer.

## Example Component

### Define

```go
func HalfSum(parent *group.Group, a, b *wire.Wire) []*wire.Wire {
	group := parent.Group(fmt.Sprintf("SUM(%s,%s)", a.Name, b.Name))
	res := Xor(group, a, b)
	res.Name = group.Name
	carry := And(group, a, b)
	carry.Name = fmt.Sprintf("CARRY(%s,%s)", a.Name, b.Name)
	return []*wire.Wire{res, carry}
}
```

### Create

```go
  "HalfSum": func(c *circuit.Circuit) []*wire.Wire {
    return HalfSum(c.Group(""), c.In("a"), c.In("b"))
  },
```

### Unit Test

```go
  name: "HalfSum",
  // a b => s carry
  want: []string{"00=>00", "01=>10", "10=>10", "11=>01"},
  isValidInt: func(inputs map[string]int) []int {
    sum := inputs["a"] + inputs["b"]
    return []int{sum % 2, sum / 2}
  },
```

### Draw

```console
$ go run main.go --example_name HalfSum --draw_graph --draw_single_graph | dot -Tsvg > doc/HalfSum.svg
$ google-chrome doc/HalfSum.svg
```

![HalfSum](doc/HalfSum.svg)

### Print

```console
$ go run main.go --example_name HalfSum
```

```console
Inputs:   a=0  b=1
Outputs:
  SUM(a,b)=1
  CARRY(a,b)=0
Components:

----------
|SUM(a,b)
|----------
||XOR(a,b)
||----------
|||OR(a,b)
|||----------
||||a=0    Vcc    OR(a,b)-wire1=0
||||b=1    Vcc    OR(a,b)-wire2=1
||||OR OR(a,b)-wire1=0    OR(a,b)-wire2=1    OR(a,b)=1
|||----------
|||NAND(a,b)
|||----------
||||a=0    Vcc    NAND(a,b)-wire=0    NAND(a,b)=1
||||b=1    NAND(a,b)-wire=0    Gnd
|||----------
|||AND(OR(a,b),NAND(a,b))
|||----------
||||OR(a,b)=1    Vcc    AND(OR(a,b),NAND(a,b))-wire=1
||||NAND(a,b)=1    AND(OR(a,b),NAND(a,b))-wire=1    SUM(a,b)=1
|||----------
||----------
||AND(a,b)
||----------
|||a=0    Vcc    AND(a,b)-wire=0
|||b=1    AND(a,b)-wire=0    CARRY(a,b)=0
||----------
|----------
----------
```

## Other Features

### Draw Multiple Graphs

```console
$ go run main.go --example_name HalfSum --draw_graph
$ for file in *.dot; do dot -Tsvg "${file}" > "${file}".svg; done
$ google-chrome *.svg
```

## Example Circuits

See [lib/lib.go](lib/lib.go).

### Unit Tests For Example Circuits

See [lib/lib_test.go](lib/lib_test.go).

## Development

### Run Unit Tests

```console
$ while inotifywait -r . 2>/dev/null; do clear; go test ./... | grep -v "no test files"; done
```

### Format Go Files

```console
$ find . -name '*.go' | xargs -n 1 go fmt
```
