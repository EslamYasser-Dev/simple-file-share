package ports

type ConfigProvider interface {
	GetPort() string
	GetUsername() string
	GetPassword() string
	GetRootDir() string
}
