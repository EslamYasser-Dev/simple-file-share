package ports

type TLSCertGenerator interface {
	GenerateCert() ([]byte, []byte, error)
}
