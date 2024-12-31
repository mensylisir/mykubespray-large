package utils

import (
	"bufio"
	"fmt"
	"github.com/whoisfisher/mykubespray/pkg/entity"
	"github.com/whoisfisher/mykubespray/pkg/logger"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// SSHExecutor implements Executor for SSH connections.
type SSHExecutor struct {
	Connection SSHConnection
	Host       entity.Host
}

func NewExecutor(host entity.Host) *SSHExecutor {
	connection, err := NewConnection(host)
	if err != nil {
		return nil
	}
	return &SSHExecutor{Connection: *connection, Host: host}
}

// NewSSHExecutor creates a new instance of SSHExecutor.
func NewSSHExecutor(connection SSHConnection) *SSHExecutor {
	return &SSHExecutor{Connection: connection}
}

// ExecuteCommand executes a command over SSH.

func (executor *SSHExecutor) ExecuteShortCommand(command string) (string, error) {
	session, err := executor.Connection.Client.NewSession()
	if err != nil {
		logger.GetLogger().Errorf("Failed to create SSH session: %s", err.Error())
		return "", err
	}
	defer session.Close()
	res, err := session.CombinedOutput(command)
	if err != nil {
		logger.GetLogger().Errorf("Failed to execute command: %s, %s", err.Error(), res)
		return "", err
	}
	return string(res), nil
}

func (executor *SSHExecutor) ExecuteShortCMD(command string) ([]byte, error) {
	session, err := executor.Connection.Client.NewSession()
	if err != nil {
		logger.GetLogger().Errorf("Failed to create SSH session: %s", err.Error())
		return nil, err
	}
	defer session.Close()
	res, err := session.CombinedOutput(command)
	if err != nil {
		logger.GetLogger().Errorf("Failed to execute command: %s", err.Error())
		return nil, err
	}
	return res, nil
}

func (executor *SSHExecutor) ExecuteCommandWithoutReturn(command string) error {
	session, err := executor.Connection.Client.NewSession()
	if err != nil {
		logger.GetLogger().Errorf("Failed to create SSH session: %s", err.Error())
		return err
	}
	defer session.Close()
	err = session.Run(command)
	if err != nil {
		logger.GetLogger().Errorf("Failed to execute command: %s", err.Error())
		return err
	}
	return nil
}

func (executor *SSHExecutor) ExecuteCMDWithoutReturn(command string, outputHandler func(string)) error {
	session, err := executor.Connection.Client.NewSession()
	if err != nil {
		logger.GetLogger().Errorf("Failed to create SSH session: %s", err.Error())
		return err
	}
	defer session.Close()
	err = session.Run(command)
	if err != nil {
		logger.GetLogger().Errorf("Failed to execute command: %s", err.Error())
		return err
	}
	outputHandler(fmt.Sprintf("Successfully to execute command:%s", command))
	return nil
}

func (executor *SSHExecutor) WhoAmI() string {
	command := fmt.Sprintf("whoami")
	user, err := executor.ExecuteShortCommand(command)
	if err != nil {
		logger.GetLogger().Warnf("Read username failed: %v", err.Error())
		return ""
	}
	return strings.TrimSpace(user)
}

func (executor *SSHExecutor) ExecuteCommand(command string, logChan chan LogEntry) error {
	session, err := executor.Connection.Client.NewSession()
	if err != nil {
		logger.GetLogger().Errorf("Failed to create SSH session: %s", err.Error())
		return err
	}
	defer session.Close()
	//session.RequestPty("xterm", 80, 40, ssh.TerminalModes{})

	stdin, err := session.StdinPipe()
	if err != nil {
		logger.GetLogger().Errorf("Unable to setup stdin for session: %v", err)
		return err
	}

	stdoutPipe, err := session.StdoutPipe()
	if err != nil {
		logger.GetLogger().Errorf("Unable to create stdout pipe: %v", err.Error())
		return err
	}
	stderrPipe, err := session.StderrPipe()
	if err != nil {
		logger.GetLogger().Errorf("Failed to create stderr pipe: %s", err.Error())
		return err
	}

	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.GetLogger().Errorf("Recovered from panic in stderr pipe: %v", r)
			}
		}()
		scanner := bufio.NewScanner(stdoutPipe)
		for scanner.Scan() {
			go fmt.Fprintln(stdin, "yes\n")
			text := scanner.Text()
			if strings.Contains(text, "[yes/no]") {
				continue
			} else {
				logChan <- LogEntry{Message: text, IsError: false}
			}
		}
	}()

	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.GetLogger().Errorf("Recovered from panic in stderr pipe: %v", r)
			}
		}()
		scanner := bufio.NewScanner(stderrPipe)
		for scanner.Scan() {
			go fmt.Fprintln(stdin, "yes\n")
			text := scanner.Text()
			if strings.Contains(text, "[yes/no]") {
				continue
			} else {
				logChan <- LogEntry{Message: text, IsError: false}
			}
		}
	}()

	err = session.Start(command)
	if err != nil {
		logChan <- LogEntry{Message: "Pipeline Done", IsError: true}
		logger.GetLogger().Errorf("Failed to run SSH command: %s", err.Error())
		return err
	}

	err = session.Wait()
	if err != nil {
		logger.GetLogger().Errorf("SSH command execution failed: %s", err.Error())
		logChan <- LogEntry{Message: "Pipeline Done", IsError: true}
		return err
	}
	logChan <- LogEntry{Message: "Pipeline Done", IsError: false}
	return nil
}

func (executor *SSHExecutor) ExecuteCommandNew(command string, logChan chan LogEntry) error {
	session, err := executor.Connection.Client.NewSession()
	if err != nil {
		logger.GetLogger().Errorf("Failed to create SSH session: %s", err.Error())
		logChan <- LogEntry{Message: "Pipeline Failed", IsError: true}
		return err
	}
	defer session.Close()
	//session.RequestPty("xterm", 80, 40, ssh.TerminalModes{})

	stdin, err := session.StdinPipe()
	if err != nil {
		logger.GetLogger().Errorf("Unable to setup stdin for session: %v", err)
		logChan <- LogEntry{Message: "Pipeline Failed", IsError: true}
		return err
	}

	stdoutPipe, err := session.StdoutPipe()
	if err != nil {
		logger.GetLogger().Errorf("Unable to create stdout pipe: %v", err.Error())
		logChan <- LogEntry{Message: "Pipeline Failed", IsError: true}
		return err
	}
	stderrPipe, err := session.StderrPipe()
	if err != nil {
		logger.GetLogger().Errorf("Failed to create stderr pipe: %s", err.Error())
		logChan <- LogEntry{Message: "Pipeline Failed", IsError: true}
		return err
	}

	var wg sync.WaitGroup

	doneStdout := make(chan struct{})
	doneStderr := make(chan struct{})

	wg.Add(1)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.GetLogger().Errorf("Recovered from panic in stdout pipe: %v", r)
			}
			wg.Done()
		}()
		defer close(doneStdout)
		scanner := bufio.NewScanner(stdoutPipe)
		for scanner.Scan() {
			go fmt.Fprintln(stdin, "yes\n")
			text := scanner.Text()
			if strings.Contains(text, "[yes/no]") {
				continue
			} else {
				select {
				case logChan <- LogEntry{Message: text, IsError: false}:
				case <-doneStdout:
					return
				}
			}
		}
		if err := scanner.Err(); err != nil {
			logger.GetLogger().Errorf("Error reading stdout pipe: %v", err)
		}
	}()

	wg.Add(1)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.GetLogger().Errorf("Recovered from panic in stderr pipe: %v", r)
			}
			wg.Done()
		}()
		defer close(doneStderr)
		scanner := bufio.NewScanner(stderrPipe)
		for scanner.Scan() {
			go fmt.Fprintln(stdin, "yes\n")
			text := scanner.Text()
			if strings.Contains(text, "[yes/no]") {
				continue
			} else {
				select {
				case logChan <- LogEntry{Message: text, IsError: true}:
				case <-doneStderr:
					return
				}
			}
		}
		if err := scanner.Err(); err != nil {
			logger.GetLogger().Errorf("Error reading stderr pipe: %v", err)
		}
	}()

	err = session.Start(command)
	if err != nil {
		logger.GetLogger().Errorf("Failed to run SSH command: %s", err.Error())
		logChan <- LogEntry{Message: "Pipeline Failed", IsError: true}
		return err
	}

	err = session.Wait()
	if err != nil {
		logger.GetLogger().Errorf("SSH command execution failed: %s", err.Error())
		logChan <- LogEntry{Message: "Pipeline Failed", IsError: true}
		return err
	}
	logChan <- LogEntry{Message: "Pipeline Success", IsError: false}
	close(logChan)
	return nil
}

func (executor *SSHExecutor) CopyMultiFile(files []entity.FileSrcDest, outputHandler func(string)) *CopyResult {
	var wg sync.WaitGroup
	results := make(chan MachineResult, len(files))
	for _, file := range files {
		wg.Add(1)
		go func(file entity.FileSrcDest) {
			defer wg.Done()
			byteData, err := os.ReadFile(file.SrcFile)
			if err != nil {
				logger.GetLogger().Errorf("Failed to open source file: %s", err.Error())
				results <- MachineResult{Machine: "", Success: false, Error: fmt.Sprintf("Failed to open source file: %s", err.Error())}
				return
			}

			fileName := filepath.Base(file.DestFile)
			tempFile := fmt.Sprintf("/tmp/%s", fileName)

			touchCommand := fmt.Sprintf("echo '%s' > %s", string(byteData), tempFile)
			err = executor.ExecuteCommandWithoutReturn(touchCommand)
			if err != nil {
				logger.GetLogger().Errorf("Failed to copy file to destination: %s", err.Error())
				return
			}

			command := fmt.Sprintf("cp -f %s %s", tempFile, file.DestFile)
			if executor.WhoAmI() != "root" {
				command = SudoPrefixWithPassword(command, executor.Host.Password)
			}
			err = executor.ExecuteCommandWithoutReturn(command)
			if err != nil {
				logger.GetLogger().Errorf("Failed to copy file to destination: %s", err.Error())
				results <- MachineResult{Machine: "", Success: false, Error: fmt.Sprintf("Failed to copy file to destination: %s", err.Error())}
				return
			}

			rmCommand := fmt.Sprintf("rm -f %s", tempFile)
			err = executor.ExecuteCommandWithoutReturn(rmCommand)
			if err != nil {
				logger.GetLogger().Errorf("Failed to delete temp file: %s", err.Error())
				return
			}

			results <- MachineResult{Machine: "", Success: true, Error: ""}
			outputHandler(fmt.Sprintf("Copied file %s to %s", file.SrcFile, file.DestFile))
			return
		}(file)
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

// CopyFile copies a file over SSH using SCP.
func (executor *SSHExecutor) CopyFile(srcFile, destFile string, outputHandler func(string)) error {
	byteData, err := os.ReadFile(srcFile)
	if err != nil {
		logger.GetLogger().Errorf("Failed to open source file: %s", err.Error())
		return err
	}

	fileName := filepath.Base(destFile)
	tempFile := fmt.Sprintf("/tmp/%s", fileName)

	touchCommand := fmt.Sprintf("echo '%s' > %s", string(byteData), tempFile)
	err = executor.ExecuteCommandWithoutReturn(touchCommand)
	if err != nil {
		logger.GetLogger().Errorf("Failed to copy file to destination: %s", err.Error())
		return err
	}

	command := fmt.Sprintf("cp -f %s %s", tempFile, destFile)
	if executor.WhoAmI() != "root" {
		command = SudoPrefixWithPassword(command, executor.Host.Password)
	}
	err = executor.ExecuteCommandWithoutReturn(command)
	if err != nil {
		logger.GetLogger().Errorf("Failed to copy file to destination: %s", err.Error())
		return err
	}

	rmCommand := fmt.Sprintf("rm -f %s", tempFile)
	err = executor.ExecuteCommandWithoutReturn(rmCommand)
	if err != nil {
		logger.GetLogger().Errorf("Failed to delete temp file: %s", err.Error())
		return err
	}

	outputHandler(fmt.Sprintf("Copied file %s to %s", srcFile, destFile))
	return nil
}

func (executor *SSHExecutor) MkDirALL(path string, outputHandler func(string)) error {
	path = filepath.ToSlash(path)
	command := fmt.Sprintf("mkdir -p %s", path)
	if executor.WhoAmI() != "root" {
		command = SudoPrefixWithPassword(command, executor.Host.Password)
	}
	err := executor.ExecuteCommandWithoutReturn(command)
	if err != nil {
		errMsg := fmt.Sprintf("Failed to create directory '%s' on remote host: %s", path, err)
		log.Println("%s: %s", errMsg, err.Error())
		return err
	}
	_ = fmt.Sprintf("Directory '%s' created successfully on remote host\n", path)
	outputHandler(fmt.Sprintf("Mkdir Directory: %s", path))
	return nil
}

func (executor *SSHExecutor) AddHosts(record entity.Record, outputHandler func(string)) error {
	getHostContentCMD := "cat /etc/hosts"
	hostContent, err := executor.ExecuteShortCommand(getHostContentCMD)
	if err != nil {
		errMsg := fmt.Errorf("failed to read /etc/hosts: %w", err)
		log.Println("%s: %s", errMsg, err.Error())
		return err
	}
	if strings.TrimSpace(hostContent) != "" {
		cmdUpdate := fmt.Sprintf(`
		#!/bin/bash
        # Remove all lines containing the hostname
        sudo sed -i "/^.* %s$/d" /etc/hosts
        # Add new entry
        echo "%s %s" | sudo tee -a /etc/hosts > /dev/null
    `, record.Domain, record.IP, record.Domain)
		cmdUpdate = fmt.Sprintf("bash -c '%s'", cmdUpdate)
		if executor.WhoAmI() != "root" {
			cmdUpdate = SudoPrefixWithPassword(cmdUpdate, executor.Host.Password)
		}
		_, err = executor.ExecuteShortCommand(cmdUpdate)
		if err != nil {
			logger.GetLogger().Errorf("failed to update /etc/hosts: %v", err)
			return fmt.Errorf("failed to update /etc/hosts: %v", err)
		}
		logger.GetLogger().Infof("Updated %s to IP %s\n", record.Domain, record.IP)
		fmt.Printf("Updated %s to IP %s\n", record.Domain, record.IP)
	} else {
		cmdAdd := fmt.Sprintf(`bash -c 'echo "%s %s" >> /etc/hosts'`, record.IP, record.Domain)
		if executor.WhoAmI() != "root" {
			cmdAdd = SudoPrefixWithPassword(cmdAdd, executor.Host.Password)
		}
		_, err = executor.ExecuteShortCommand(cmdAdd)
		if err != nil {
			logger.GetLogger().Errorf("failed to add to /etc/hosts: %v", err)
			return fmt.Errorf("failed to add to /etc/hosts: %v", err)
		}
		logger.GetLogger().Infof("Added %s with IP %s\n", record.Domain, record.IP)
		fmt.Printf("Added %s with IP %s\n", record.Domain, record.IP)
	}
	outputHandler(fmt.Sprintf("Add Hosts"))
	return nil
}

func (executor *SSHExecutor) AddMultiHosts(records []entity.Record, outputHandler func(string)) error {
	getHostContentCMD := "cat /etc/hosts"
	hostContent, err := executor.ExecuteShortCommand(getHostContentCMD)
	if err != nil {
		errMsg := fmt.Errorf("failed to read /etc/hosts: %w", err)
		log.Println("%s: %s", errMsg, err.Error())
		return err
	}
	lines := strings.Split(hostContent, "\n")
	var updateContent strings.Builder
	domainMap := make(map[string]string)

	for _, line := range lines {
		parts := strings.Fields(line)
		if len(parts) >= 2 {
			ip, domain := parts[0], parts[1]
			domainMap[domain] = ip
		}
	}

	for _, record := range records {
		domainMap[record.Domain] = record.IP
	}

	for key, value := range domainMap {
		updateContent.WriteString(fmt.Sprintf("%s %s\n", value, key))
	}

	tmpFile := "/tmp/hosts"
	err = os.WriteFile(tmpFile, []byte(updateContent.String()), 0644)
	if err != nil {
		logger.GetLogger().Error("Failed to write "+
			""+
			" temporary file: %s", err)
		return fmt.Errorf("Failed to write to temporary file: %s", err)
	}
	cmd := fmt.Sprintf("cp %s /etc/hosts", tmpFile)
	if executor.WhoAmI() != "root" {
		cmd = SudoPrefixWithPassword(cmd, executor.Host.Password)
	}
	_, err = executor.ExecuteShortCommand(cmd)
	if err != nil {
		logger.GetLogger().Errorf("failed to add to /etc/hosts: %v", err)
		return fmt.Errorf("failed to add to /etc/hosts: %v", err)
	}
	outputHandler(fmt.Sprintf("Add Hosts"))
	return nil
}

func (executor *SSHExecutor) UpdateHostsFile(ip string, domain string) error {
	getHostContentCMD := "cat /etc/hosts"
	hostContent, err := executor.ExecuteShortCommand(getHostContentCMD)
	if err != nil {
		logger.GetLogger().Errorf("读取 /etc/hosts 出错: %v\", err")
		return fmt.Errorf("读取 /etc/hosts 出错: %v", err)
	}
	lines := strings.Split(hostContent, "\n")
	domainExists := false
	var updatedLines []string

	for _, line := range lines {
		if strings.Contains(line, domain) {
			domainExists = true
			parts := strings.Fields(line)
			if len(parts) > 0 {
				updatedLines = append(updatedLines, fmt.Sprintf("%s %s", ip, domain))
			}
		} else {
			updatedLines = append(updatedLines, line)
		}
	}

	if !domainExists {
		updatedLines = append(updatedLines, fmt.Sprintf("%s %s", ip, domain))
	}

	// 写入更新后的内容
	command := fmt.Sprintf("bash -c \"echo -n '%s' | sudo tee /etc/hosts\"", strings.Join(updatedLines, "\n"))
	if executor.WhoAmI() != "root" {
		command = SudoPrefixWithPassword(command, executor.Host.Password)
	}
	_, err = executor.ExecuteShortCommand(command)
	if err != nil {
		logger.GetLogger().Errorf("写入 /etc/hosts 出错: %v", err)
		return fmt.Errorf("写入 /etc/hosts 出错: %v", err)
	}
	return nil
}

func (executor *SSHExecutor) UpdateResolvFile(ip string) error {
	getDNSContentCMD := "cat /etc/resolv.conf"
	dnsContent, err := executor.ExecuteShortCommand(getDNSContentCMD)
	if err != nil {
		logger.GetLogger().Errorf("读取 /etc/resolv.conf 出错: %v", err)
		return fmt.Errorf("读取 /etc/resolv.conf 出错: %v", err)
	}

	lines := strings.Split(dnsContent, "\n")
	ipExists := false

	for _, line := range lines {
		if strings.Contains(line, ip) {
			ipExists = true
			break
		}
	}

	if !ipExists {
		command := fmt.Sprintf("bash -c \" echo -n 'nameserver %s\n' | sudo tee -a /etc/resolv.conf\"", ip)
		if executor.WhoAmI() != "root" {
			command = SudoPrefixWithPassword(command, executor.Host.Password)
		}
		_, err = executor.ExecuteShortCommand(command)
		if err != nil {
			logger.GetLogger().Errorf("追加到 /etc/resolv.conf 出错: %v", err)
			return fmt.Errorf("追加到 /etc/resolv.conf 出错: %v", err)
		}
	} else {
		logger.GetLogger().Infof("IP 已存在，跳过追加")
		fmt.Println("IP 已存在，跳过追加")
	}
	return nil
}
