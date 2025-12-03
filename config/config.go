// Package config encapsulates configuration.
package config

// Config contains configuration.
type Config struct {
	MaxPrintDepth   int
	DrawGraph       bool
	DrawSingleGraph bool
	DrawNodes       bool
	DrawShapePoint  bool
	DrawEdges       bool
	IsUnitTest      bool
	SimulateInputs  []string
}
