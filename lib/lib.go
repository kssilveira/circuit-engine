package lib

import (
	"github.com/kssilveira/circuit-engine/circuit"
	"github.com/kssilveira/circuit-engine/group"
	"github.com/kssilveira/circuit-engine/wire"
)

func Transistor(parent *group.Group, base, collector *wire.Wire) []*wire.Wire {
	group := parent.Group("transistor")
	emitter := &wire.Wire{Name: "emitter"}
	collectorOut := &wire.Wire{Name: "collector_out"}
	group.Transistor(base, collector, emitter, collectorOut)
	return []*wire.Wire{emitter, collectorOut}
}

func TransistorGnd(parent *group.Group, base, collector *wire.Wire) []*wire.Wire {
	group := parent.Group("transistorGnd")
	collectorOut := &wire.Wire{Name: "collector_out"}
	group.Transistor(base, collector, group.Gnd, collectorOut)
	return []*wire.Wire{collectorOut}
}

func Example(c *circuit.Circuit, name string) []*wire.Wire {
	res, ok := examples[name]
	if !ok {
		return nil
	}
	return res(c)
}

func ExampleNames() []string {
	var res []string
	for name := range examples {
		res = append(res, name)
	}
	return res
}

var (
	examples = map[string]func(*circuit.Circuit) []*wire.Wire{
		"Transistor": func(c *circuit.Circuit) []*wire.Wire {
			return Transistor(c.Group(""), c.In("base"), c.In("collector"))
		},
		"TransistorGnd": func(c *circuit.Circuit) []*wire.Wire {
			return TransistorGnd(c.Group(""), c.In("base"), c.In("collector"))
		},
	}
)
