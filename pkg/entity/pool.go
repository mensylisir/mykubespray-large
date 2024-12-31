package entity

type AddHostsParallel struct {
	Hosts  []Host
	Record Record
}

type AddDNSParallel struct {
	Hosts []Host
	DNS   string
}

type CopyFileParallel struct {
	Hosts    []Host
	SrcFile  string
	DestFile string
}

type CommandParallel struct {
	Hosts   []Host
	Command string
}
