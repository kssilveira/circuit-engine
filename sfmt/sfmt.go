package sfmt

import "fmt"

func Sprintf(format string, a ...any) string {
	return fmt.Sprintf(format, a...)
}
