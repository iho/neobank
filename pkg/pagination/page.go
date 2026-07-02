package pagination

// Trim returns up to limit items and the next cursor when more rows exist.
func Trim[T any](items []T, limit int, cursorFn func(T) Cursor) ([]T, string) {
	if limit <= 0 {
		limit = 20
	}
	if len(items) <= limit {
		return items, ""
	}
	last := items[limit-1]
	return items[:limit], Encode(cursorFn(last))
}