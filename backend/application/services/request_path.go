package services

// requestPath maps an omitted (empty) client-supplied path to the virtual root.
// The path value object stays strict about empty strings, so this normalization
// lives once in the application layer instead of being duplicated in every
// primary adapter.
func requestPath(path string) string {
	if path == "" {
		return "/"
	}
	return path
}
