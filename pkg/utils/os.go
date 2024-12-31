package utils

import (
	"fmt"
	"github.com/whoisfisher/mykubespray/pkg/entity"
	"github.com/whoisfisher/mykubespray/pkg/logger"
	"strings"
)

type OSConf struct {
	Arch           string
	Version        string
	Name           string
	CPU            string
	CPUCores       string
	MemorySize     string
	DiskSize       string
	NetCardList    []string
	SpecifyNetCard string
}

type OSClient struct {
	OSConf        OSConf
	SSExecutor    SSHExecutor
	LocalExecutor LocalExecutor
}

func NewClient(host entity.Host) *OSClient {
	osConf := OSConf{}
	localExecutor := NewLocalExecutor()
	sshExecutor := NewExecutor(host)
	osclient := &OSClient{
		OSConf:        osConf,
		SSExecutor:    *sshExecutor,
		LocalExecutor: *localExecutor,
	}
	osclient.GetOSConf()
	osclient.GetCPU()
	osclient.GetCPUCores()
	osclient.GetMemorySize()
	osclient.GetDiskSize()
	osclient.GetNetCardList()
	return osclient
}

func NewOSClient(osConf OSConf, sshExecutor SSHExecutor, localExecutor LocalExecutor) *OSClient {
	osclient := &OSClient{
		OSConf:        osConf,
		SSExecutor:    sshExecutor,
		LocalExecutor: localExecutor,
	}
	osclient.GetOSConf()
	osclient.GetCPU()
	osclient.GetCPUCores()
	osclient.GetMemorySize()
	osclient.GetDiskSize()
	osclient.GetNetCardList()
	return osclient
}

func (client *OSClient) GetOSConf() bool {
	command := fmt.Sprintf("cat /etc/os-release")
	output, err := client.SSExecutor.ExecuteShortCommand(command)
	if err != nil {
		return false
	}
	client.OSConf.Name = "Unknown"
	client.OSConf.Version = "Unknown"
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "ID=") {
			client.OSConf.Name = strings.TrimPrefix(line, "ID=")
		} else if strings.HasPrefix(line, "VERSION_ID=") {
			client.OSConf.Version = strings.TrimPrefix(line, "VERSION_ID=")
		}
	}
	res, err := client.SSExecutor.ExecuteShortCommand("arch")
	if err != nil {
		logger.GetLogger().Errorf("Failed to get os arch: %s", err.Error())
		client.OSConf.Arch = "Unknown"
		return false
	}
	arch := strings.TrimSpace(string(res))
	client.OSConf.Arch = arch
	return true
}

func (client *OSClient) GetDistribution() (string, error) {
	command := fmt.Sprintf("cat /etc/os-release")
	output, err := client.SSExecutor.ExecuteShortCommand(command)
	if err != nil {
		logger.GetLogger().Errorf("Failed to get distribution: %s", err.Error())
		return "", err
	}
	res := parseOSRelease(output)
	return res, nil
}

func (client *OSClient) DaemonReload() error {
	command := fmt.Sprintf("systemctl daemon-reload")
	if client.WhoAmI() != "root" {
		command = SudoPrefixWithPassword(command, client.SSExecutor.Host.Password)
	}
	_, err := client.SSExecutor.ExecuteShortCommand(command)
	if err != nil {
		logger.GetLogger().Errorf("Failed to reload daemon: %s", err.Error())
		return err
	}
	return nil
}

func (client *OSClient) RestartService(service string) error {
	command := fmt.Sprintf("systemctl restart %s", service)
	if client.WhoAmI() != "root" {
		command = SudoPrefixWithPassword(command, client.SSExecutor.Host.Password)
	}
	_, err := client.SSExecutor.ExecuteShortCommand(command)
	if err != nil {
		logger.GetLogger().Errorf("Failed to restart %s: %s", service, err.Error())
		return err
	}
	return nil
}

func (client *OSClient) StartService(service string) error {
	command := fmt.Sprintf("systemctl start %s", service)
	if client.WhoAmI() != "root" {
		command = SudoPrefixWithPassword(command, client.SSExecutor.Host.Password)
	}
	_, err := client.SSExecutor.ExecuteShortCommand(command)
	if err != nil {
		logger.GetLogger().Errorf("Failed to start %s: %s", service, err.Error())
		return err
	}
	return nil
}

func (client *OSClient) StopService(service string) error {
	command := fmt.Sprintf("systemctl stop %s", service)
	if client.WhoAmI() != "root" {
		command = SudoPrefixWithPassword(command, client.SSExecutor.Host.Password)
	}
	_, err := client.SSExecutor.ExecuteShortCommand(command)
	if err != nil {
		logger.GetLogger().Errorf("Failed to stop %s: %s", service, err.Error())
		return err
	}
	return nil
}

func (client *OSClient) DisableService(service string) error {
	command := fmt.Sprintf("systemctl disable %s", service)
	if client.WhoAmI() != "root" {
		command = SudoPrefixWithPassword(command, client.SSExecutor.Host.Password)
	}
	_, err := client.SSExecutor.ExecuteShortCommand(command)
	if err != nil {
		logger.GetLogger().Errorf("Failed to disable %s: %s", service, err.Error())
		return err
	}
	return nil
}

func (client *OSClient) EnableService(service string) error {
	command := fmt.Sprintf("systemctl enable %s", service)
	if client.WhoAmI() != "root" {
		command = SudoPrefixWithPassword(command, client.SSExecutor.Host.Password)
	}
	_, err := client.SSExecutor.ExecuteShortCommand(command)
	if err != nil {
		logger.GetLogger().Errorf("Failed to enable %s: %s", service, err.Error())
		return err
	}
	return nil
}

func (client *OSClient) MaskService(service string) error {
	command := fmt.Sprintf("systemctl mask %s", service)
	if client.WhoAmI() != "root" {
		command = SudoPrefixWithPassword(command, client.SSExecutor.Host.Password)
	}
	_, err := client.SSExecutor.ExecuteShortCommand(command)
	if err != nil {
		logger.GetLogger().Errorf("Failed to mask %s: %s", service, err.Error())
		return err
	}
	return nil
}

func (client *OSClient) UNMaskService(service string) error {
	command := fmt.Sprintf("systemctl unmask %s", service)
	if client.WhoAmI() != "root" {
		command = SudoPrefixWithPassword(command, client.SSExecutor.Host.Password)
	}
	_, err := client.SSExecutor.ExecuteShortCommand(command)
	if err != nil {
		logger.GetLogger().Errorf("Failed to unmask %s: %s", service, err.Error())
		return err
	}
	return nil
}

func (client *OSClient) StatusService(service string) bool {
	command := fmt.Sprintf("systemctl status %s | grep -iE active", service)
	if client.WhoAmI() != "root" {
		command = SudoPrefixWithPassword(command, client.SSExecutor.Host.Password)
	}
	res, err := client.SSExecutor.ExecuteShortCommand(command)
	if err != nil {
		logger.GetLogger().Errorf("Failed to view %s status: %s", service, err.Error())
		return false
	}
	if strings.Contains(res, "inactive") {
		return false
	}
	return true
}

func (client *OSClient) GetCPUCores() bool {
	command := "grep -c ^processor /proc/cpuinfo"
	if client.WhoAmI() != "root" {
		command = SudoPrefixWithPassword(command, client.SSExecutor.Host.Password)
	}
	res, err := client.SSExecutor.ExecuteShortCommand(command)
	if err != nil {
		logger.GetLogger().Errorf("Failed to get cpu cores: %s", err.Error())
		client.OSConf.CPUCores = "Unknown"
		return false
	}
	cores := strings.TrimSpace(res)
	client.OSConf.CPUCores = cores
	return true
}

func (client *OSClient) GetCPU() bool {
	command := "grep -iE \"^model\\s+name\\s+:\" /proc/cpuinfo | awk -F':' '{print $NF}' | sort -u"
	if client.WhoAmI() != "root" {
		command = SudoPrefixWithPassword(command, client.SSExecutor.Host.Password)
	}
	res, err := client.SSExecutor.ExecuteShortCommand(command)
	if err != nil {
		logger.GetLogger().Errorf("Failed to get cpu info: %s", err.Error())
		client.OSConf.CPU = "Unknown"
		return false
	}
	cpu := strings.TrimSpace(res)
	client.OSConf.CPU = cpu
	return true
}

func (client *OSClient) GetAvailableCPU() string {
	command := "top -bn1 | grep 'Cpu(s)' | awk '{print $8\"%\"}'"
	if client.WhoAmI() != "root" {
		command = SudoPrefixWithPassword(command, client.SSExecutor.Host.Password)
	}
	res, err := client.SSExecutor.ExecuteShortCommand(command)
	if err != nil {
		logger.GetLogger().Errorf("Failed to get cpu info: %s", err.Error())
		return ""
	}
	availableCPU := strings.TrimSpace(res)
	return availableCPU
}

func (client *OSClient) GetMemorySize() bool {
	command := "free -m | grep Mem | awk '{print $2}'"
	if client.WhoAmI() != "root" {
		command = SudoPrefixWithPassword(command, client.SSExecutor.Host.Password)
	}
	res, err := client.SSExecutor.ExecuteShortCommand(command)
	if err != nil {
		logger.GetLogger().Errorf("Failed to get memory info: %s", err.Error())
		client.OSConf.MemorySize = "Unknown"
		return false
	}
	memCapacity := strings.TrimSpace(res) + "MB"
	client.OSConf.MemorySize = memCapacity
	return true
}

func (client *OSClient) GetAvailableMemory() string {
	command := "free -m | grep Mem | awk '{print $4}'"
	if client.WhoAmI() != "root" {
		command = SudoPrefixWithPassword(command, client.SSExecutor.Host.Password)
	}
	res, err := client.SSExecutor.ExecuteShortCommand(command)
	if err != nil {
		logger.GetLogger().Errorf("Failed to get memory info: %s", err.Error())
		return ""
	}
	availableMemCapacity := strings.TrimSpace(res) + "MB"
	return availableMemCapacity
}

func (client *OSClient) GetDiskSize() bool {
	command := "df -h / | tail -n 1 | awk '{print $2}'"
	if client.WhoAmI() != "root" {
		command = SudoPrefixWithPassword(command, client.SSExecutor.Host.Password)
	}
	res, err := client.SSExecutor.ExecuteShortCommand(command)
	if err != nil {
		logger.GetLogger().Errorf("Failed to get disk size: %s", err.Error())
		client.OSConf.DiskSize = "Unknown"
		return false
	}
	diskCapacity := strings.TrimSpace(res)
	client.OSConf.DiskSize = diskCapacity
	return true
}

func (client *OSClient) GetNetCardList() bool {
	command := "ip addr show | grep -o '^[0-9]\\+: [a-zA-Z0-9]*' | awk '{print $2}'"
	if client.WhoAmI() != "root" {
		command = SudoPrefixWithPassword(command, client.SSExecutor.Host.Password)
	}
	res, err := client.SSExecutor.ExecuteShortCommand(command)
	if err != nil {
		logger.GetLogger().Errorf("Failed to get netcard list: %s", err.Error())
		client.OSConf.NetCardList = []string{"Unknown"}
		return false
	}
	client.OSConf.NetCardList = strings.Split(res, " ")
	return true
}

func (client *OSClient) GetSpecifyNetCard(ipaddr string) string {
	command := fmt.Sprintf("ip addr | grep -B 2 '%s' | head -n 1 | awk -F':' '{print $2}'", ipaddr)
	if client.WhoAmI() != "root" {
		command = SudoPrefixWithPassword(command, client.SSExecutor.Host.Password)
	}
	res, err := client.SSExecutor.ExecuteShortCommand(command)
	if err != nil {
		logger.GetLogger().Errorf("Failed to get netcard info for %s: %s", ipaddr, err.Error())
		client.OSConf.SpecifyNetCard = res
	}
	client.OSConf.SpecifyNetCard = strings.TrimSpace(res)
	return client.OSConf.SpecifyNetCard
}

func (client *OSClient) IsProcessExist(processName string) bool {
	command := fmt.Sprintf("pgrep %s", processName)
	if client.WhoAmI() != "root" {
		command = SudoPrefixWithPassword(command, client.SSExecutor.Host.Password)
	}
	_, err := client.SSExecutor.ExecuteShortCommand(command)
	if err != nil {
		logger.GetLogger().Warnf("The process %s is non-exist: %s", processName, err.Error())
		return false
	}
	return true
}

func (client *OSClient) WhoAmI() string {
	command := fmt.Sprintf("whoami")
	user, err := client.SSExecutor.ExecuteShortCommand(command)
	if err != nil {
		logger.GetLogger().Warnf("Read username failed: %v", err.Error())
		return ""
	}
	return strings.TrimSpace(user)
}

func (client *OSClient) Chmod(file string, mode string) error {
	cmd := fmt.Sprintf("chmod %s %s", mode, file)
	if client.WhoAmI() != "root" {
		cmd = SudoPrefixWithPassword(cmd, client.SSExecutor.Host.Password)
	}
	_, err := client.SSExecutor.ExecuteShortCommand(cmd)
	if err != nil {
		logger.GetLogger().Errorf("Chmod %s failed: %v", file, err)
		return err
	}
	return nil
}

func (client *OSClient) ReadFile(file string) (string, error) {
	cmd := fmt.Sprintf("cat %s", file)
	data, err := client.SSExecutor.ExecuteShortCommand(cmd)
	if err != nil {
		logger.GetLogger().Errorf("Read %s failed: %v", file, err)
		return "", err
	}
	return strings.TrimSpace(data), nil
}

func (client *OSClient) ReadBytes(file string) ([]byte, error) {
	cmd := fmt.Sprintf("cat %s", file)
	data, err := client.SSExecutor.ExecuteShortCMD(cmd)
	if err != nil {
		logger.GetLogger().Errorf("Read %s failed: %v", file, err)
		return nil, err
	}
	return data, nil
}

func (client *OSClient) WriteFile(content, file string) error {
	cmd := fmt.Sprintf("bash -c \"echo '%s' > %s\"", content, file)
	if client.WhoAmI() != "root" {
		cmd = SudoPrefixWithPassword(cmd, client.SSExecutor.Host.Password)
	}
	err := client.SSExecutor.ExecuteCommandWithoutReturn(cmd)
	if err != nil {
		logger.GetLogger().Errorf("Write %s failed: %v", file, err)
		return err
	}
	return nil
}

func (client *OSClient) QueryVGName() (*entity.LVS, error) {
	lvs := &entity.LVS{}
	cmd := "lvs"
	if client.WhoAmI() != "root" {
		cmd = SudoPrefixWithPassword(cmd, client.SSExecutor.Host.Password)
	}
	data, err := client.SSExecutor.ExecuteShortCommand(cmd)
	if err != nil {
		logger.GetLogger().Errorf("Query VGName failed: %v", err)
		return nil, err
	}
	lines := strings.Split(data, "\n")
	for _, line := range lines {
		item := strings.TrimSpace(line)
		if strings.HasPrefix(item, "root") {
			items := strings.Split(item, " ")
			lvs.LVName = strings.TrimSpace(items[0])
			lvs.VGName = strings.TrimSpace(items[1])
		}
	}
	return lvs, nil
}

func (client *OSClient) CreatePV(diskConf entity.DiskConf) error {
	cmd := fmt.Sprintf("pvcreate %s", diskConf.Device)
	if client.WhoAmI() != "root" {
		cmd = SudoPrefixWithPassword(cmd, client.SSExecutor.Host.Password)
	}
	data, err := client.SSExecutor.ExecuteShortCommand(cmd)
	if err != nil {
		logger.GetLogger().Errorf("pvcreate failed: %v,%s", err, data)
		return err
	}
	fmt.Println(data)
	return nil
}

func (client *OSClient) ExtendVG(diskConf entity.DiskConf) error {
	cmd := fmt.Sprintf("vgextend %s %s", diskConf.VGName, diskConf.Device)
	if client.WhoAmI() != "root" {
		cmd = SudoPrefixWithPassword(cmd, client.SSExecutor.Host.Password)
	}
	data, err := client.SSExecutor.ExecuteShortCommand(cmd)
	if err != nil {
		logger.GetLogger().Errorf("vgextend failed: %v,%s", err, data)
		return err
	}
	return nil
}

func (client *OSClient) ExtendLVPercent100(diskConf entity.DiskConf) error {
	cmd := fmt.Sprintf("lvextend -l +100%%FREE /dev/mapper/%s-%s", diskConf.VGName, diskConf.LVName)
	if client.WhoAmI() != "root" {
		cmd = SudoPrefixWithPassword(cmd, client.SSExecutor.Host.Password)
	}
	data, err := client.SSExecutor.ExecuteShortCommand(cmd)
	if err != nil {
		logger.GetLogger().Errorf("lvextend failed: %v,%s", err, data)
		return err
	}
	return nil
}

func (client *OSClient) ExtendLV(diskConf entity.DiskConf) error {
	cmd := fmt.Sprintf("lvextend -L +%s /dev/mapper/%s-%s", diskConf.Size, diskConf.VGName, diskConf.LVName)
	if client.WhoAmI() != "root" {
		cmd = SudoPrefixWithPassword(cmd, client.SSExecutor.Host.Password)
	}
	data, err := client.SSExecutor.ExecuteShortCommand(cmd)
	if err != nil {
		logger.GetLogger().Errorf("lvextend failed: %v,%s", err, data)
		return err
	}
	return nil
}

func (client *OSClient) XGrowFS(diskConf entity.DiskConf) error {
	cmd := fmt.Sprintf("xfs_growfs /dev/mapper/%s-%s", diskConf.VGName, diskConf.LVName)
	if client.WhoAmI() != "root" {
		cmd = SudoPrefixWithPassword(cmd, client.SSExecutor.Host.Password)
	}
	data, err := client.SSExecutor.ExecuteShortCommand(cmd)
	if err != nil {
		logger.GetLogger().Errorf("xfs_growfs failed: %v,%s", err, data)
		return err
	}
	return nil
}

func (client *OSClient) Resize2FS(diskConf entity.DiskConf) error {
	cmd := fmt.Sprintf("resize2fs /dev/mapper/%s-%s", diskConf.VGName, diskConf.LVName)
	if client.WhoAmI() != "root" {
		cmd = SudoPrefixWithPassword(cmd, client.SSExecutor.Host.Password)
	}
	data, err := client.SSExecutor.ExecuteShortCommand(cmd)
	if err != nil {
		logger.GetLogger().Errorf("xfs_growfs failed: %v,%s", err, data)
		return err
	}
	return nil
}

func (client *OSClient) CopyFile(srcFile, destFile string) error {
	outputHandler := func(string) { logger.GetLogger().Infof("Copy file") }
	return client.SSExecutor.CopyFile(srcFile, destFile, outputHandler)
}

func (client *OSClient) CopyMultiFile(files []entity.FileSrcDest) *CopyResult {
	outputHandler := func(string) { logger.GetLogger().Infof("Copy file") }
	return client.SSExecutor.CopyMultiFile(files, outputHandler)
}

func (client *OSClient) AddHost(record entity.Record) error {
	outputHandler := func(string) { logger.GetLogger().Infof("Add Hosts") }
	return client.SSExecutor.AddHosts(record, outputHandler)
}

func (client *OSClient) AddMultiHost(records []entity.Record) error {
	outputHandler := func(string) { logger.GetLogger().Infof("Add Hosts") }
	return client.SSExecutor.AddMultiHosts(records, outputHandler)
}

func SudoPrefixWithEOF(cmd string) string {
	return fmt.Sprintf("sudo -S -E /bin/bash <<EOF\n%s\nEOF", cmd)
}

func SudoPrefix(cmd string) string {
	return fmt.Sprintf("sudo %s", cmd)
}

func SudoPrefixWithPassword(cmd, sudoPassword string) string {
	return fmt.Sprintf("echo %s | sudo -S %s", sudoPassword, SudoPrefix(cmd))
}
