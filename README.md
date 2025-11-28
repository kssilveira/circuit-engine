# circuit-engine

Define and simulate circuits from transistors all the way to an 8-bit computer.

## Features

### Simulate

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

### Draw

Single graph:

```console
$ go run main.go --example_name HalfSum --draw_graph --draw_single_graph \
| dot -Tsvg > doc/HalfSum.svg
$ google-chrome doc/HalfSum.svg
```

![HalfSum](doc/HalfSum.svg)

Multiple graphs:

```console
$ go run main.go --example_name HalfSum --draw_graph
$ for file in *.dot; do dot -Tsvg "${file}" > "${file}".svg; done
$ google-chrome *.svg
```

## Development

Run unit tests:

```console
$ while inotifywait -r . 2>/dev/null; \
do clear; go test ./... | grep -v "no test files"; done
```

Format files:

```console
$ find . -name '*.go' | xargs -n 1 go fmt
```
