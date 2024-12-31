package utils

// Executor interface defines methods for executing commands and copying files.
type Executor interface {
	ExecuteCommand(command string, logChan chan LogEntry) error
	CopyFile(srcFile, destFile string, outputHandler func(string)) error
	CopyRemoteToRemote(srcHost, srcFile, destHost, destFile string, outputHandler func(string)) error
}

// Connection interface defines methods for establishing a connection.
type Connection interface {
	Connect(config SSHConfig) error
}
