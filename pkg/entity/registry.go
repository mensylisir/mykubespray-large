package entity

type Registry struct {
	Name               string
	Url                string
	User               string
	Password           string
	Type               string
	KeyPath            string
	CertPath           string
	SkipTLS            bool
	PlainHttp          bool
	InsecureRegistries []string
	NodeName           string
}
