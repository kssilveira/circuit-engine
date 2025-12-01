// Package component encapsulates the component interface.
package component

import (
	"github.com/kssilveira/circuit-engine/config"
)

// Component contains the component interface.
type Component interface {
	Update()
	String(depth int, cfg config.Config) string
	Graph(depth int, cfg config.Config) string
}
