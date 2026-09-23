package ports

import "errors"

// ErrUnauthorized is returned when credentials are missing or invalid in a
// primary adapter. It is intentionally distinct from domain login errors so
// middleware can map it to 401 without leaking credential details.
var ErrUnauthorized = errors.New("unauthorized")
