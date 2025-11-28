package component

type Component interface {
	Update()
	String(int) string
	Graph(int) string
}

