package ports

type AuthProvider interface {
	Authenticate(username, password string) bool
}
