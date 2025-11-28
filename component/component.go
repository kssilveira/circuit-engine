package component

type Component interface {
	Update()
	String(depth int) string
	Graph(depth int) string
}
