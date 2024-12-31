package entity

type DiskConf struct {
	Host   Host
	Device string
	LVS
}

type LVS struct {
	LVName string
	VGName string
	Size   string
}

type RecordConf struct {
	Host   Host
	Record Record
}

type CertConf struct {
	Host     Host
	CertPath string
	DestPath string
}

type Record struct {
	IP     string
	Domain string
}

type FileSrcDest struct {
	SrcFile  string
	DestFile string
}
