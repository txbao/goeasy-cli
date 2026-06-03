package utils

import (
	"strings"
	"unicode"
)

// ToPascal converts user, user_order -> User, UserOrder.
func ToPascal(name string) string {
	parts := splitName(name)
	for i, p := range parts {
		if p == "" {
			continue
		}
		runes := []rune(strings.ToLower(p))
		runes[0] = unicode.ToUpper(runes[0])
		parts[i] = string(runes)
	}
	return strings.Join(parts, "")
}

// ToSnake converts UserOrder -> user_order.
func ToSnake(name string) string {
	var b strings.Builder
	parts := splitName(name)
	for i, p := range parts {
		if i > 0 {
			b.WriteByte('_')
		}
		b.WriteString(strings.ToLower(p))
	}
	return b.String()
}

func splitName(name string) []string {
	name = strings.TrimSpace(name)
	name = strings.ReplaceAll(name, "-", "_")
	if strings.Contains(name, "_") {
		return strings.Split(name, "_")
	}
	var parts []string
	var cur strings.Builder
	for _, r := range name {
		if unicode.IsUpper(r) && cur.Len() > 0 {
			parts = append(parts, cur.String())
			cur.Reset()
		}
		cur.WriteRune(r)
	}
	if cur.Len() > 0 {
		parts = append(parts, cur.String())
	}
	if len(parts) == 0 {
		return []string{name}
	}
	return parts
}
