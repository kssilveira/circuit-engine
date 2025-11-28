# circuit-engine

Define and simulate circuits from transistors all the way to an 8-bit computer.

## Development

Run unit test:

```sh
$ while inotifywait -r .; do clear; go test ./... | grep -v "no test files"; done
```

Format files:

```sh
$ find . -name '*.go' | xargs -n 1 go fmt
```
