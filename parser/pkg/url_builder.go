package pkg

import "strings"

func UrlBuilder(url, id string) string {
	var builder strings.Builder

	builder.Grow(2) // Оптимизация производительности (так как у нас 2 строки)
	builder.WriteString(url)
	builder.WriteString("/")
	builder.WriteString(id)

	return builder.String()
}
