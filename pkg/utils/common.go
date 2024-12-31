package utils

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/whoisfisher/mykubespray/pkg/entity"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
	"io"
	"log"
	"os"
	"regexp"
	"strings"
)

//var kubekeyLogo = `
//
// _   __      _          _   __
//| | / /     | |        | | / /
//| |/ / _   _| |__   ___| |/ /  ___ _   _
//|    \| | | | '_ \ / _ \    \ / _ \ | | |
//| |\  \ |_| | |_) |  __/ |\  \  __/ |_| |
//\_| \_/\__,_|_.__/ \___\_| \_/\___|\__, |
//                                    __/ |
//                                   |___/
//
//`
//
//var taichuLogo = `
//  _______     _         _____      _____     _    _     _    _
// |_______|   / \       |_   _|    / ____|   | |  | |   | |  | |
//    | |     / _ \        | |     | |        | |  | |   | |  | |
//    | |    / ___ \       | |     | |        | ---- |   | |  | |
//    | |   /_/   \_\     _| |_    | |____    | |  | |   | |__| |
//    |_|  / /     \ \   |_____|    \_____|   |_|  |_|    \____/
//`

type LogEntry struct {
	Message string // 日志消息
	IsError bool   // 是否为错误日志
}

func DecodeBytes(data []byte, decoder *encoding.Decoder) (string, error) {
	reader := transform.NewReader(bytes.NewReader(data), decoder)
	decodedData, err := io.ReadAll(reader)
	if err != nil {
		return "", err
	}
	return string(decodedData), nil
}

func DecodeGBK(data []byte) (string, error) {
	return DecodeBytes(data, simplifiedchinese.GBK.NewDecoder())
}

func GetDistribution(executor *SSHExecutor) (string, error) {
	output, err := executor.ExecuteShortCommand("cat /etc/os-release")
	if err != nil {
		log.Printf("Get distribution failed: %s", err.Error())
		return "", err
	}
	res := parseOSRelease(output)
	return res, nil
}

func parseOSRelease(output string) string {
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "ID=") {
			return strings.TrimPrefix(line, "ID=")
		}
	}
	return "Unknown"
}

func contains(slice []interface{}, item interface{}) bool {
	for _, elem := range slice {
		if elem == item {
			return true
		}
	}
	return false
}

func toInterfaceSlice(slice []string) []interface{} {
	var result []interface{}
	for _, s := range slice {
		result = append(result, s)
	}
	return result
}

func AddHosts(record entity.Record, host entity.Host) error {
	filepath := "/etc/hosts"
	file, err := os.Open(filepath)
	if err != nil {
		return fmt.Errorf("error opening file: %w", err)
	}
	defer file.Close()
	var lines []string
	var found bool
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, record.Domain) {
			line = record.IP + " " + record.Domain
			found = true
		}
		lines = append(lines, line)
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading file: %w", err)
	}
	if !found {
		lines = append(lines, fmt.Sprintf("%s %s", record.IP, record.Domain))
	}
	file, err = os.Create(filepath)
	if err != nil {
		return fmt.Errorf("error creating file: %w", err)
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	for _, line := range lines {
		if _, err := writer.WriteString(line + "\n"); err != nil {
			return fmt.Errorf("error writing to file: %w", err)
		}
	}

	if err := writer.Flush(); err != nil {
		return fmt.Errorf("error flushing writer: %w", err)
	}

	return nil
}

func ChangeLogo(logChan chan LogEntry) {
	// 定义原始的 kubekeyLogo 正则表达式，用于匹配日志输出中的 logo
	kubekeyLogo := regexp.MustCompile(`(?s)_   __.*?\|___/\|_| \\_\|\|___/`)

	// 替换后的 taichuLogo
	taichuLogo := `
  _______     _         _____      _____     _    _     _    _  
 |_______|   / \       |_   _|    / ____|   | |  | |   | |  | | 
    | |     / _ \        | |     | |        | |  | |   | |  | | 
    | |    / ___ \       | |     | |        | ---- |   | |  | | 
    | |   /_/   \_\     _| |_    | |____    | |  | |   | |__| | 
    |_|  / /     \ \   |_____|    \_____|   |_|  |_|    \____/  
	`

	// 使用字符串构建器收集日志内容
	var logBuffer strings.Builder
	logProcessing := false

	// 处理日志流
	for logEntry := range logChan {
		message := logEntry.Message

		// 如果开始捕捉 logo，多行拼接到 logBuffer
		if kubekeyLogo.MatchString(message) || logProcessing {
			logBuffer.WriteString(message + "\n")
			logProcessing = true

			// 如果匹配到 logo 结束，进行替换
			if strings.Contains(message, "|___/") {
				completeLogo := logBuffer.String()
				updatedLogo := kubekeyLogo.ReplaceAllString(completeLogo, taichuLogo)
				fmt.Println(updatedLogo) // 输出替换后的 taichuLogo
				logProcessing = false
				logBuffer.Reset()
			}
			continue
		}

		// 输出未替换的普通日志行
		fmt.Println(message)
	}
}
