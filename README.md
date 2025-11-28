# circuit-engine

Define and simulate circuits from transistors all the way to an 8-bit computer.

## Development

Run unit tests:

```sh
$ while inotifywait -r .; do clear; go test ./... | grep -v "no test files"; done
```

Run main:

```sh
$ go run main.go --example_name TransistorEmitter
```

Format files:

```sh
$ find . -name '*.go' | xargs -n 1 go fmt
```
