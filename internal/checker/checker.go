package checker

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"
)

func ml(command string) string {
	cmd := exec.Command("sh", "-c", command)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}

func printResult(buffer *bytes.Buffer, check string, result bool) {
	if result {
		buffer.WriteString(fmt.Sprintf("[√] %s\n", check))
	} else {
		buffer.WriteString(fmt.Sprintf("[x] %s\n", check))
	}
}

func checkAlerts(buffer *bytes.Buffer) {
	_, err := exec.Command("sh", "-c", "alias").Output()
	printResult(buffer, "Alerts后门", err == nil)
}

func checkSSHKey(buffer *bytes.Buffer) {
	currentUser, err := user.Current()
	if err != nil {
		buffer.WriteString(fmt.Sprintf("Error getting current user: %v\n", err))
		return
	}

	var filePath string
	if currentUser.Username == "root" {
		filePath = "/root/.ssh/authorized_keys"
	} else {
		filePath = filepath.Join("/home", currentUser.Username, ".ssh/authorized_keys")
	}

	_, err = os.Stat(filePath)
	printResult(buffer, "SSH公私密钥后门", err == nil)

	sshdConfigPath := "/etc/ssh/sshd_config"
	if _, err := os.Stat(sshdConfigPath); err == nil {
		if isWritable(sshdConfigPath) {
			buffer.WriteString("   可以修改sshd_config配置文件\n")
		} else {
			buffer.WriteString("   没有权限修改sshd_config文件\n")
		}
	} else {
		buffer.WriteString("   sshd_config配置文件文件不存在\n")
	}
}

func isWritable(path string) bool {
	file, err := os.OpenFile(path, os.O_WRONLY, 0666)
	if err != nil {
		if os.IsPermission(err) {
			return false
		}
	}
	file.Close()
	return true
}

func checkAddUser(buffer *bytes.Buffer) {
	buffer.WriteString("----------->Backdoor Check<-------------\n")
	buffer.WriteString("OpenSSH后门[只测试过 Ubuntu 14 版本成功]\n")
	currentUser, err := user.Current()
	if err != nil {
		buffer.WriteString(fmt.Sprintf("Error getting current user: %v\n", err))
		return
	}

	printResult(buffer, "SSH后门用户", currentUser.Gid == "0")
}

func checkCrontab(buffer *bytes.Buffer) {
	currentUser, err := user.Current()
	if err != nil {
		buffer.WriteString(fmt.Sprintf("Error getting current user: %v\n", err))
		return
	}

	userCronPath := filepath.Join("/var/spool/cron", currentUser.Username)
	cronFiles := []string{"/etc/crontab", userCronPath, "/var/spool/cron/crontabs"}

	buffer.WriteString("计划任务后门\n")
	for _, cronFile := range cronFiles {
		if isWritable(cronFile) {
			buffer.WriteString(fmt.Sprintf("  [√] %s\n", cronFile))
		} else {
			buffer.WriteString(fmt.Sprintf("  [x] %s\n", cronFile))
		}
	}
}

func checkStrace(buffer *bytes.Buffer) {
	output := ml("strace -V")
	printResult(buffer, "Strace后门", strings.Contains(output, "strace -- version"))
}

func checkSSHSoftLink(buffer *bytes.Buffer) {
	sshConfig := ml("cat /etc/ssh/sshd_config|grep UsePAM")
	currentUser := ml("whoami")

	hasSSHSoftLink := strings.Contains(sshConfig, "UsePAM yes") && strings.Contains(currentUser, "root")
	printResult(buffer, "SSH软链接后门", hasSSHSoftLink)
	if !hasSSHSoftLink {
		buffer.WriteString("  [如果是root权限，可以直接SSH软链接模块运行开启]\n")
	}
}

func checkRootkit(buffer *bytes.Buffer) {
	kernelVersion := ml("uname -r")
	osInfo := ml("cat /etc/os-release")

	minKernelVersion := map[string]string{
		"Centos 6.10":        "2.6.32-754.6.3.el6.x86_64",
		"Centos 7":           "3.10.0-862.3.2.el7.x86_64",
		"Centos 8":           "4.18.0-147.5.1.el8_1.x86_64",
		"Ubuntu 18.04.1 LTS": "4.15.0-38-generic",
	}

	maxKernelVersion := map[string]string{
		"Centos 6.10":        "2.6.32",
		"Centos 7":           "3.10.0",
		"Centos 8":           "4.18.0",
		"Ubuntu 18.04.1 LTS": "4.15.0",
	}

	hasRootkit := false
	for osName, minVersion := range minKernelVersion {
		if strings.Contains(osInfo, osName) {
			maxVersion := maxKernelVersion[osName]
			if kernelVersion >= minVersion && kernelVersion <= maxVersion {
				hasRootkit = true
				break
			}
		}
	}

	printResult(buffer, "Rootkit后门", hasRootkit)
	if hasRootkit {
		buffer.WriteString("  https://github.com/f0rb1dd3n/Reptile/\n")
	}
}

func sshSoftLinkCrontab(buffer *bytes.Buffer) {
	currentUser, err := user.Current()
	if err != nil {
		buffer.WriteString(fmt.Sprintf("Error getting current user: %v\n", err))
		return
	}

	userCronPath := filepath.Join("/var/spool/cron", currentUser.Username)
	cronFiles := []string{"/etc/crontab", userCronPath, "/var/spool/cron/crontabs"}

	buffer.WriteString("计划任务&SSH软链接后门\n")
	for _, cronFile := range cronFiles {
		if isWritable(cronFile) {
			buffer.WriteString(fmt.Sprintf("  [√] %s\n", cronFile))
		} else {
			buffer.WriteString(fmt.Sprintf("  [x] %s\n", cronFile))
		}
	}
}

func sshCrontabSSHKey(buffer *bytes.Buffer) {
	currentUser, err := user.Current()
	if err != nil {
		buffer.WriteString(fmt.Sprintf("Error getting current user: %v\n", err))
		return
	}

	userCronPath := filepath.Join("/var/spool/cron", currentUser.Username)
	cronFiles := []string{"/etc/crontab", userCronPath, "/var/spool/cron/crontabs", "/etc/ssh/sshd_config"}

	buffer.WriteString("计划任务&SSH Key后门\n")
	for _, cronFile := range cronFiles {
		if isWritable(cronFile) {
			buffer.WriteString(fmt.Sprintf("  [√] %s\n", cronFile))
		} else {
			buffer.WriteString(fmt.Sprintf("  [x] %s\n", cronFile))
		}
	}
}

func dockerK8s(buffer *bytes.Buffer) {
	cgroupContent, err := os.ReadFile("/proc/1/cgroup")
	if err != nil {
		buffer.WriteString(fmt.Sprintf("Error reading cgroup file: %v\n", err))
		return
	}

	if strings.Contains(string(cgroupContent), "kubepods") {
		buffer.WriteString("----------------->container<---------------\n")
		buffer.WriteString("{kubepods k8s}\n")
		dockerK8sEsc(buffer)
	} else if strings.Contains(string(cgroupContent), "docker") {
		buffer.WriteString("----------------->container<---------------\n")
		buffer.WriteString("{docker}\n")
		dockerK8sEsc(buffer)
	}
}

func dockerK8sEsc(buffer *bytes.Buffer) {
	capEff := ml("cat /proc/self/status | grep CapEff")
	printResult(buffer, "Docker特权逃逸", strings.Contains(capEff, "0000001fffffffff"))

	_, err := os.Stat("/var/run/docker.sock")
	printResult(buffer, "Docker Socket逃逸", err == nil)

	corePatterPath := ml("find / -name core_pattern")
	hasProcfsEscape := strings.Contains(corePatterPath, "/host/proc/sys/kernel/core_pattern") &&
		strings.Contains(corePatterPath, "/proc/sys/kernel/core_pattern")
	printResult(buffer, "Docker procfs逃逸", hasProcfsEscape)
}

func checkUser(buffer *bytes.Buffer) {
	buffer.WriteString("---------------->Privilege<-----------------\n")
	cmd := exec.Command("id")
	output, err := cmd.Output()
	if err != nil {
		buffer.WriteString(fmt.Sprintf("Error getting user information: %v\n", err))
		return
	}
	buffer.WriteString(fmt.Sprintf("%s\n", strings.TrimSpace(string(output))))
}

func checkPath(buffer *bytes.Buffer) {
	buffer.WriteString("------------------->Env<--------------------\n")
	languages := map[string]string{
		"Python2":    "python2",
		"Python3":    "python3",
		"Java":       "java",
		"Docker":     "docker",
		"PHP":        "php",
		"Kubernetes": "kubectl",
		"Rust":       "rustc",
		"C++":        "g++",
	}

	for langName, langCmd := range languages {
		_, err := exec.LookPath(langCmd)
		printResult(buffer, langName, err == nil)
	}
}

func Start() []byte {
	var buffer bytes.Buffer

	checkUser(&buffer)
	checkAddUser(&buffer)
	checkAlerts(&buffer)
	checkCrontab(&buffer)
	checkSSHSoftLink(&buffer)
	checkSSHKey(&buffer)
	checkStrace(&buffer)
	checkRootkit(&buffer)
	sshSoftLinkCrontab(&buffer)
	sshCrontabSSHKey(&buffer)
	dockerK8s(&buffer)
	checkPath(&buffer)

	return buffer.Bytes()
}

func main() {
	result := Start()
	fmt.Println(string(result))
}
