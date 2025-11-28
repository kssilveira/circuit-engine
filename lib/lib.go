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

func Example(c *circuit.Circuit, name string) []*wire.Wire {
	switch name {
	case "transistor":
		return Transistor(c.Group(""), c.In("base"), c.In("collector"))
	}
	return nil
}
