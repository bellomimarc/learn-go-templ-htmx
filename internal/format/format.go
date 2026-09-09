package format

import "fmt"

func FormatCents(value int64) string {
	return fmt.Sprintf("%d.%02d", value/100, value%100)
}
