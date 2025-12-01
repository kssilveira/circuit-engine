// Package sfmt contains string formatting functions.
package sfmt

import "fmt"

// Sprintf formats according to a format specifier and returns the resulting string.
func Sprintf(format string, a ...any) string {
	return fmt.Sprintf(format, a...)
}
