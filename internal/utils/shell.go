package utils

import (
	"bytes"
	"os/exec"
	"strings"
)

// ExecuteCommand 执行shell命令并返回输出
func ExecuteCommand(command string) (string, error) {
	cmd := exec.Command("sh", "-c", command)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	return strings.TrimSpace(out.String()), err
}

// CheckRoot 检查是否具有root权限
func CheckRoot() bool {
	output, err := ExecuteCommand("id -u")
	if err != nil {
		return false
	}
	return output == "0"
}
