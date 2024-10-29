package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/10cks/LinuxKeeperGo/internal/checker"
	"github.com/10cks/LinuxKeeperGo/internal/core/generator"
	"github.com/10cks/LinuxKeeperGo/internal/modules"
	"github.com/10cks/LinuxKeeperGo/internal/utils"
)

func main() {
	var (
		moduleFlag = flag.Int("m", 0, "Select module number")
		checkFlag  = flag.Int("c", 0, "Check system information (1: basic, 2: detailed)")
		listFlag   = flag.Bool("l", false, "List all available modules")
		helpFlag   = flag.Bool("h", false, "Show help information")
	)
	flag.Parse()

	utils.ShowBanner()

	if *helpFlag {
		showHelp()
		os.Exit(0)
	}

	if *listFlag {
		listModules()
		os.Exit(0)
	}

	// 运行系统检查
	if result := checker.Start(); result != nil {
		fmt.Println("\n[*] System Check Results:")
		fmt.Println(string(result))
	}

	// 处理模块选择
	switch {
	case *checkFlag > 0:
		showModuleInfo(*checkFlag)
	case *moduleFlag > 0:
		if err := executeModule(*moduleFlag); err != nil {
			log.Fatalf("[x] Module execution failed: %v", err)
		}
	default:
		showHelp()
	}
}

func showHelp() {
	fmt.Println("\nUsage:")
	fmt.Println("  -m <number>  Select and execute a module")
	fmt.Println("  -c <1|2>    Show module information")
	fmt.Println("  -l          List all available modules")
	fmt.Println("  -h          Show this help message")
}

func listModules() {
	fmt.Println("\nAvailable Modules:")
	for id, module := range modules.AvailableModules {
		fmt.Printf("[%d] %s - %s\n", id, module.Name, module.Description)
	}
}

func showModuleInfo(level int) {
	fmt.Println("\nModule Information:")
	for id, module := range modules.AvailableModules {
		if level == 1 {
			fmt.Printf("[%d] %s - %s\n", id, module.Name, module.Description)
		} else {
			fmt.Printf("\n[Module %d]\n", id)
			fmt.Printf("Name: %s\n", module.Name)
			fmt.Printf("Description: %s\n", module.Description)
			fmt.Printf("Required Privileges: %s\n", module.RequiredPrivs)
			fmt.Printf("Supported Systems: %v\n", module.SupportedSystems)
			fmt.Printf("Risk Level: %s\n", module.RiskLevel)
		}
	}
}

func executeModule(moduleID int) error {
	if !utils.CheckRoot() {
		return fmt.Errorf("root privileges required")
	}

	gen := generator.NewGenerator()
	if err := gen.Generate(moduleID); err != nil {
		return fmt.Errorf("failed to generate payload: %v", err)
	}

	fmt.Printf("\n[+] Payload generated successfully!")
	fmt.Printf("\n[+] Check the 'payloads' directory for your files\n")
	return nil
}
