package utils

import (
	"fmt"
	"github.com/whoisfisher/mykubespray/pkg/entity"
	"github.com/whoisfisher/mykubespray/pkg/logger"
	"golang.org/x/crypto/ssh"
	"sync"
)

type SSHConnectionPool struct {
	Connections sync.Map
	mutex       sync.Mutex
}

type SSHSessionPool struct {
	Sessions sync.Map
	mutex    sync.Mutex
}

type SSHExecutorPool struct {
	Executors sync.Map
	mutex     sync.Mutex
}

func NewSSHConnectionPool() *SSHConnectionPool {
	return &SSHConnectionPool{}
}

func NewSSHSessionPool() *SSHSessionPool {
	return &SSHSessionPool{}
}

func NewSSHExecutorPool() *SSHExecutorPool {
	return &SSHExecutorPool{}
}

func (pool *SSHConnectionPool) GetSSHConnection(host entity.Host) (*SSHConnection, error) {
	//pool.mutex.Lock()
	//defer pool.mutex.Unlock()
	if conn, exists := pool.Connections.Load(host.Name); exists {
		return conn.(*SSHConnection), nil
	}
	conn, err := NewConnection(host)
	if err != nil {
		return nil, err
	}
	pool.Connections.Store(host.Name, conn)
	return conn, nil
}

func (pool *SSHSessionPool) GetSSHSession(host entity.Host) (*ssh.Session, error) {
	//pool.mutex.Lock()
	//defer pool.mutex.Unlock()
	if session, exists := pool.Sessions.Load(host.Name); exists {
		return session.(*ssh.Session), nil
	}
	conn, err := NewConnection(host)
	if err != nil {
		return nil, err
	}
	session, err := conn.Client.NewSession()
	if err != nil {
		return nil, err
	}
	pool.Sessions.Store(host.Name, session)
	return session, nil
}

func (pool *SSHSessionPool) Close() {
	pool.Sessions.Range(func(key, value interface{}) bool {
		client := value.(*ssh.Client)
		client.Close()
		return true
	})
}

func (pool *SSHExecutorPool) GetSSHExecutor(host entity.Host) (*SSHExecutor, error) {
	//pool.mutex.Lock()
	//defer pool.mutex.Unlock()
	if executor, exists := pool.Executors.Load(host.Name); exists {
		return executor.(*SSHExecutor), nil
	}
	conn, err := NewConnection(host)
	if err != nil {
		return nil, err
	}
	executor := &SSHExecutor{
		Connection: *conn,
		Host:       host,
	}
	pool.Executors.Store(host.Name, executor)
	return executor, nil
}

func (pool *SSHExecutorPool) Close() {
	pool.Executors.Range(func(key, value interface{}) bool {
		client := value.(*SSHExecutor)
		client.Connection.Client.Close()
		return true
	})
}

func (pool *SSHExecutorPool) ExecuteShortCommand(command string, host entity.Host) (string, error) {
	pool.mutex.Lock()
	executor, err := pool.GetSSHExecutor(host)
	pool.mutex.Unlock()
	if err != nil {
		logger.GetLogger().Errorf("Failed to get SSH executor: %s", err.Error())
		return "", err
	}
	return executor.ExecuteShortCommand(command)
}

func (pool *SSHExecutorPool) ExecuteShortCMD(command string, host entity.Host) ([]byte, error) {
	pool.mutex.Lock()
	executor, err := pool.GetSSHExecutor(host)
	pool.mutex.Unlock()
	if err != nil {
		logger.GetLogger().Errorf("Failed to get SSH executor: %s", err.Error())
		return nil, err
	}
	return executor.ExecuteShortCMD(command)
}

func (pool *SSHExecutorPool) ExecuteCommand(command string, host entity.Host, logChan chan LogEntry) error {
	pool.mutex.Lock()
	executor, err := pool.GetSSHExecutor(host)
	pool.mutex.Unlock()
	if err != nil {
		logger.GetLogger().Errorf("Failed to get SSH executor: %s", err.Error())
		return err
	}
	return executor.ExecuteCommand(command, logChan)
}

func (pool *SSHExecutorPool) ExecuteCommandWithoutReturn(command string, host entity.Host) error {
	pool.mutex.Lock()
	executor, err := pool.GetSSHExecutor(host)
	pool.mutex.Unlock()
	if err != nil {
		logger.GetLogger().Errorf("Failed to get SSH executor: %s", err.Error())
		return err
	}
	return executor.ExecuteCommandWithoutReturn(command)
}

type MachineResult struct {
	Machine string
	Success bool
	Error   string
}

type CopyResult struct {
	OverallSuccess bool
	Results        []MachineResult
}

func (pool *SSHExecutorPool) ExecuteCommandParallel(command string, hosts []entity.Host) *CopyResult {
	var wg sync.WaitGroup
	results := make(chan MachineResult, len(hosts))
	for _, host := range hosts {
		wg.Add(1)
		go func(host entity.Host) {
			defer wg.Done()
			pool.mutex.Lock()
			executor, err := pool.GetSSHExecutor(host)
			pool.mutex.Unlock()
			if err != nil {
				logger.GetLogger().Errorf("Failed to get SSH executor: %s", err.Error())
				results <- MachineResult{Machine: host.Address, Success: false, Error: fmt.Sprintf("Failed to connect to %s: %s", host.Address, err.Error())}
				return
			}
			err = executor.ExecuteCMDWithoutReturn(command, func(msg string) {
				results <- MachineResult{Machine: host.Address, Success: true, Error: ""}
			})
			if err != nil {
				results <- MachineResult{Machine: host.Address, Success: false, Error: fmt.Sprintf("Failed to execute command on %s: %s", host.Address, err.Error())}
			}
		}(host)
	}

	go func() {
		wg.Wait()
		close(results)
	}()
	var successCount, failureCount int
	var copyResult CopyResult
	var machineResults []MachineResult
	for result := range results {
		if result.Success {
			logger.GetLogger().Infof("Successfully to execute command on %s\n", result.Machine)
			successCount++
		} else {
			logger.GetLogger().Errorf("Failedto execute command on %s: %s\n", result.Machine, result.Error)
			failureCount++
		}
		machineResults = append(machineResults, result)
	}
	copyResult.Results = machineResults
	if failureCount > 0 {
		copyResult.OverallSuccess = false
	} else {
		copyResult.OverallSuccess = true
	}
	return &copyResult
}

func (pool *SSHExecutorPool) ExecuteCommandParallelWithoutPool(command string, hosts []entity.Host) *CopyResult {
	var wg sync.WaitGroup
	results := make(chan MachineResult, len(hosts))
	for _, host := range hosts {
		wg.Add(1)
		go func(host entity.Host) {
			defer wg.Done()
			conn, err := NewConnection(host)
			if err != nil {
				return
			}
			executor := &SSHExecutor{
				Connection: *conn,
			}
			if err != nil {
				logger.GetLogger().Errorf("Failed to get SSH executor: %s", err.Error())
				results <- MachineResult{Machine: host.Address, Success: false, Error: fmt.Sprintf("Failed to connect to %s: %s", host.Address, err.Error())}
				return
			}
			err = executor.ExecuteCMDWithoutReturn(command, func(msg string) {
				results <- MachineResult{Machine: host.Address, Success: true, Error: ""}
			})
			if err != nil {
				results <- MachineResult{Machine: host.Address, Success: false, Error: fmt.Sprintf("Failed to execute command on %s: %s", host.Address, err.Error())}
			}
		}(host)
	}

	go func() {
		wg.Wait()
		close(results)
	}()
	var successCount, failureCount int
	var copyResult CopyResult
	var machineResults []MachineResult
	for result := range results {
		if result.Success {
			logger.GetLogger().Infof("Successfully to execute command on %s\n", result.Machine)
			successCount++
		} else {
			logger.GetLogger().Errorf("Failedto execute command on %s: %s\n", result.Machine, result.Error)
			failureCount++
		}
		machineResults = append(machineResults, result)
	}
	copyResult.Results = machineResults
	if failureCount > 0 {
		copyResult.OverallSuccess = false
	} else {
		copyResult.OverallSuccess = true
	}
	return &copyResult
}

func (pool *SSHExecutorPool) CopyFileParallel(srcFile, destFile string, hosts []entity.Host) *CopyResult {
	var wg sync.WaitGroup
	results := make(chan MachineResult, len(hosts))
	for _, host := range hosts {
		wg.Add(1)
		go func(host entity.Host) {
			defer wg.Done()
			pool.mutex.Lock()
			executor, err := pool.GetSSHExecutor(host)
			pool.mutex.Unlock()
			if err != nil {
				logger.GetLogger().Errorf("Failed to get SSH executor: %s", err.Error())
				results <- MachineResult{Machine: host.Address, Success: false, Error: fmt.Sprintf("Failed to connect to %s: %s", host.Address, err.Error())}
				return
			}
			err = executor.CopyFile(srcFile, destFile, func(msg string) {
				results <- MachineResult{Machine: host.Address, Success: true, Error: ""}
			})
			if err != nil {
				results <- MachineResult{Machine: host.Address, Success: false, Error: fmt.Sprintf("Failed to copy file to %s: %s", host.Address, err.Error())}
			}
		}(host)
	}

	go func() {
		wg.Wait()
		close(results)
	}()
	var successCount, failureCount int
	var copyResult CopyResult
	var machineResults []MachineResult
	for result := range results {
		if result.Success {
			logger.GetLogger().Infof("Successfully copied file to %s\n", result.Machine)
			successCount++
		} else {
			logger.GetLogger().Errorf("Failed to copy file to %s: %s\n", result.Machine, result.Error)
			failureCount++
		}
		machineResults = append(machineResults, result)
	}
	copyResult.Results = machineResults
	if failureCount > 0 {
		copyResult.OverallSuccess = false
	} else {
		copyResult.OverallSuccess = true
	}
	return &copyResult
}

func (pool *SSHExecutorPool) CopyFileParallelWithoutPool(srcFile, destFile string, hosts []entity.Host) *CopyResult {
	var wg sync.WaitGroup
	results := make(chan MachineResult, len(hosts))
	for _, host := range hosts {
		wg.Add(1)
		go func(host entity.Host) {
			defer wg.Done()
			conn, err := NewConnection(host)
			if err != nil {
				return
			}
			executor := &SSHExecutor{
				Connection: *conn,
			}
			if err != nil {
				logger.GetLogger().Errorf("Failed to get SSH executor: %s", err.Error())
				results <- MachineResult{Machine: host.Address, Success: false, Error: fmt.Sprintf("Failed to connect to %s: %s", host.Address, err.Error())}
				return
			}
			err = executor.CopyFile(srcFile, destFile, func(msg string) {
				results <- MachineResult{Machine: host.Address, Success: true, Error: ""}
			})
			if err != nil {
				results <- MachineResult{Machine: host.Address, Success: false, Error: fmt.Sprintf("Failed to copy file to %s: %s", host.Address, err.Error())}
			}
		}(host)
	}

	go func() {
		wg.Wait()
		close(results)
	}()
	var successCount, failureCount int
	var copyResult CopyResult
	var machineResults []MachineResult
	for result := range results {
		if result.Success {
			logger.GetLogger().Infof("Successfully copied file to %s\n", result.Machine)
			successCount++
		} else {
			logger.GetLogger().Errorf("Failed to copy file to %s: %s\n", result.Machine, result.Error)
			failureCount++
		}
		machineResults = append(machineResults, result)
	}
	copyResult.Results = machineResults
	if failureCount > 0 {
		copyResult.OverallSuccess = false
	} else {
		copyResult.OverallSuccess = true
	}
	return &copyResult
}

func (pool *SSHExecutorPool) CopyMultiFileParallel(files []entity.FileSrcDest, hosts []entity.Host) *CopyResult {
	var wg sync.WaitGroup
	results := make(chan MachineResult, len(hosts))
	for _, host := range hosts {
		wg.Add(1)
		go func(host entity.Host) {
			defer wg.Done()
			pool.mutex.Lock()
			executor, err := pool.GetSSHExecutor(host)
			pool.mutex.Unlock()
			if err != nil {
				logger.GetLogger().Errorf("Failed to get SSH executor: %s", err.Error())
				results <- MachineResult{Machine: host.Address, Success: false, Error: fmt.Sprintf("Failed to connect to %s: %s", host.Address, err.Error())}
				return
			}
			_ = executor.CopyMultiFile(files, func(msg string) {
				results <- MachineResult{Machine: host.Address, Success: true, Error: ""}
			})
			if err != nil {
				results <- MachineResult{Machine: host.Address, Success: false, Error: fmt.Sprintf("Failed to copy file to %s: %s", host.Address, err.Error())}
			}
		}(host)
	}

	go func() {
		wg.Wait()
		close(results)
	}()
	var successCount, failureCount int
	var copyResult CopyResult
	var machineResults []MachineResult
	for result := range results {
		if result.Success {
			logger.GetLogger().Infof("Successfully copied file to %s\n", result.Machine)
			successCount++
		} else {
			logger.GetLogger().Errorf("Failed to copy file to %s: %s\n", result.Machine, result.Error)
			failureCount++
		}
		machineResults = append(machineResults, result)
	}
	copyResult.Results = machineResults
	if failureCount > 0 {
		copyResult.OverallSuccess = false
	} else {
		copyResult.OverallSuccess = true
	}
	return &copyResult
}

func (pool *SSHExecutorPool) CopyFile(srcFile, destFile string, host entity.Host) error {
	pool.mutex.Lock()
	defer pool.mutex.Unlock()
	executor, err := pool.GetSSHExecutor(host)
	if err != nil {
		logger.GetLogger().Errorf("Failed to get SSH executor: %s", err.Error())
		return err
	}
	outputHandler := func(string) { logger.GetLogger().Infof("Copy file") }
	return executor.CopyFile(srcFile, destFile, outputHandler)
}

func (pool *SSHExecutorPool) CopyMultiFile(files []entity.FileSrcDest, host entity.Host) (*CopyResult, error) {
	pool.mutex.Lock()
	defer pool.mutex.Unlock()
	executor, err := pool.GetSSHExecutor(host)
	if err != nil {
		logger.GetLogger().Errorf("Failed to get SSH executor: %s", err.Error())
		return nil, err
	}
	outputHandler := func(string) { logger.GetLogger().Infof("Copy file") }
	return executor.CopyMultiFile(files, outputHandler), nil
}

func (pool *SSHExecutorPool) AddHostsParallel(record entity.Record, hosts []entity.Host) *CopyResult {
	var wg sync.WaitGroup
	results := make(chan MachineResult, len(hosts))
	for _, host := range hosts {
		wg.Add(1)
		go func(host entity.Host) {
			defer wg.Done()
			pool.mutex.Lock()
			executor, err := pool.GetSSHExecutor(host)
			pool.mutex.Unlock()
			if err != nil {
				logger.GetLogger().Errorf("Failed to get SSH executor: %s", err.Error())
				results <- MachineResult{Machine: host.Address, Success: false, Error: fmt.Sprintf("Failed to connect to %s: %s", host.Address, err.Error())}
				return
			}
			err = executor.AddHosts(record, func(msg string) {
				results <- MachineResult{Machine: host.Address, Success: true, Error: ""}
			})
			if err != nil {
				results <- MachineResult{Machine: host.Address, Success: false, Error: fmt.Sprintf("Failed to add hosts to %s: %s", host.Address, err.Error())}
			}
		}(host)
	}

	go func() {
		wg.Wait()
		close(results)
	}()
	var successCount, failureCount int
	var copyResult CopyResult
	var machineResults []MachineResult
	for result := range results {
		if result.Success {
			logger.GetLogger().Infof("Successfully add hosts to %s\n", result.Machine)
			successCount++
		} else {
			logger.GetLogger().Errorf("Failed to add hosts %s: %s\n", result.Machine, result.Error)
			failureCount++
		}
		machineResults = append(machineResults, result)
	}
	copyResult.Results = machineResults
	if failureCount > 0 {
		copyResult.OverallSuccess = false
	} else {
		copyResult.OverallSuccess = true
	}
	return &copyResult
}

func (pool *SSHExecutorPool) AddHostsParallelWithoutPool(record entity.Record, hosts []entity.Host) *CopyResult {
	var wg sync.WaitGroup
	results := make(chan MachineResult, len(hosts))
	for _, host := range hosts {
		wg.Add(1)
		go func(host entity.Host) {
			defer wg.Done()
			conn, err := NewConnection(host)
			if err != nil {
				return
			}
			executor := &SSHExecutor{
				Connection: *conn,
			}
			if err != nil {
				logger.GetLogger().Errorf("Failed to get SSH executor: %s", err.Error())
				results <- MachineResult{Machine: host.Address, Success: false, Error: fmt.Sprintf("Failed to connect to %s: %s", host.Address, err.Error())}
				return
			}
			err = executor.AddHosts(record, func(msg string) {
				results <- MachineResult{Machine: host.Address, Success: true, Error: ""}
			})
			if err != nil {
				results <- MachineResult{Machine: host.Address, Success: false, Error: fmt.Sprintf("Failed to add hosts to %s: %s", host.Address, err.Error())}
			}
		}(host)
	}

	go func() {
		wg.Wait()
		close(results)
	}()
	var successCount, failureCount int
	var copyResult CopyResult
	var machineResults []MachineResult
	for result := range results {
		if result.Success {
			logger.GetLogger().Infof("Successfully add hosts to %s\n", result.Machine)
			successCount++
		} else {
			logger.GetLogger().Errorf("Failed to add hosts %s: %s\n", result.Machine, result.Error)
			failureCount++
		}
		machineResults = append(machineResults, result)
	}
	copyResult.Results = machineResults
	if failureCount > 0 {
		copyResult.OverallSuccess = false
	} else {
		copyResult.OverallSuccess = true
	}
	return &copyResult
}

func (pool *SSHExecutorPool) AddMultiHostsParallel(records []entity.Record, hosts []entity.Host) *CopyResult {
	var wg sync.WaitGroup
	results := make(chan MachineResult, len(hosts))
	for _, host := range hosts {
		wg.Add(1)
		go func(host entity.Host) {
			defer wg.Done()
			pool.mutex.Lock()
			executor, err := pool.GetSSHExecutor(host)
			pool.mutex.Unlock()
			if err != nil {
				logger.GetLogger().Errorf("Failed to get SSH executor: %s", err.Error())
				results <- MachineResult{Machine: host.Address, Success: false, Error: fmt.Sprintf("Failed to connect to %s: %s", host.Address, err.Error())}
				return
			}
			err = executor.AddMultiHosts(records, func(msg string) {
				results <- MachineResult{Machine: host.Address, Success: true, Error: ""}
			})
			if err != nil {
				results <- MachineResult{Machine: host.Address, Success: false, Error: fmt.Sprintf("Failed to add hosts to %s: %s", host.Address, err.Error())}
			}
		}(host)
	}

	go func() {
		wg.Wait()
		close(results)
	}()
	var successCount, failureCount int
	var copyResult CopyResult
	var machineResults []MachineResult
	for result := range results {
		if result.Success {
			logger.GetLogger().Infof("Successfully add hosts to %s\n", result.Machine)
			successCount++
		} else {
			logger.GetLogger().Errorf("Failed to add hosts %s: %s\n", result.Machine, result.Error)
			failureCount++
		}
		machineResults = append(machineResults, result)
	}
	copyResult.Results = machineResults
	if failureCount > 0 {
		copyResult.OverallSuccess = false
	} else {
		copyResult.OverallSuccess = true
	}
	return &copyResult
}

func (pool *SSHExecutorPool) AddHosts(record entity.Record, host entity.Host) error {
	pool.mutex.Lock()
	defer pool.mutex.Unlock()
	executor, err := pool.GetSSHExecutor(host)
	if err != nil {
		logger.GetLogger().Errorf("Failed to get SSH executor: %s", err.Error())
		return err
	}
	outputHandler := func(string) { logger.GetLogger().Infof("Add hosts") }
	return executor.AddHosts(record, outputHandler)
}

func (pool *SSHExecutorPool) AddMultiHosts(records []entity.Record, host entity.Host) error {
	pool.mutex.Lock()
	defer pool.mutex.Unlock()
	executor, err := pool.GetSSHExecutor(host)
	if err != nil {
		logger.GetLogger().Errorf("Failed to get SSH executor: %s", err.Error())
		return err
	}
	outputHandler := func(string) { logger.GetLogger().Infof("Add hosts") }
	return executor.AddMultiHosts(records, outputHandler)
}

func (pool *SSHExecutorPool) AddDNSParallel(dns string, hosts []entity.Host) *CopyResult {
	var wg sync.WaitGroup
	results := make(chan MachineResult, len(hosts))
	for _, host := range hosts {
		wg.Add(1)
		go func(host entity.Host) {
			defer wg.Done()
			pool.mutex.Lock()
			executor, err := pool.GetSSHExecutor(host)
			pool.mutex.Unlock()
			if err != nil {
				logger.GetLogger().Errorf("Failed to get SSH executor: %s", err.Error())
				results <- MachineResult{Machine: host.Address, Success: false, Error: fmt.Sprintf("Failed to connect to %s: %s", host.Address, err.Error())}
				return
			}
			err = executor.UpdateResolvFile(dns)
			if err != nil {
				results <- MachineResult{Machine: host.Address, Success: false, Error: fmt.Sprintf("Failed to add dns to %s: %s", host.Address, err.Error())}
			}
		}(host)
	}

	go func() {
		wg.Wait()
		close(results)
	}()
	var successCount, failureCount int
	var copyResult CopyResult
	var machineResults []MachineResult
	for result := range results {
		if result.Success {
			logger.GetLogger().Infof("Successfully add dns to %s\n", result.Machine)
			successCount++
		} else {
			logger.GetLogger().Errorf("Failed to add dns %s: %s\n", result.Machine, result.Error)
			failureCount++
		}
		machineResults = append(machineResults, result)
	}
	copyResult.Results = machineResults
	if failureCount > 0 {
		copyResult.OverallSuccess = false
	} else {
		copyResult.OverallSuccess = true
	}
	return &copyResult
}
