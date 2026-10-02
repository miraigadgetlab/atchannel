package service

import "strings"

// likePattern turns user input into a literal ILIKE pattern: the wildcard
// characters are escaped first (backslash is LIKE's default escape in
// Postgres) so searching for "100%" or "snake_case" matches those exact
// characters instead of acting as a pattern.
func likePattern(q string) string {
	q = strings.ReplaceAll(q, `\`, `\\`)
	q = strings.ReplaceAll(q, `%`, `\%`)
	q = strings.ReplaceAll(q, `_`, `\_`)

	return "%" + q + "%"
}
