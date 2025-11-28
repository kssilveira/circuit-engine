package component

import (
	"github.com/kssilveira/circuit-engine/config"
)

type Component interface {
	Update()
	String(depth int) string
	Graph(depth int, cfg config.Config) string
}
