package entity

type KeepalivedConf struct {
	State    string
	IntFace  string
	Priority int
	AuthType string
	AuthPass string
	SrcIP    string
	Peers    []string
	StrPeers string
	VIP      string
	Host     Host
}
