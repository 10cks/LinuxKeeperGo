package modules

type Module struct {
	ID               int
	Name             string
	Description      string
	RequiredPrivs    string
	SupportedSystems []string
	RiskLevel        string
	Execute          func() error
}

var AvailableModules = map[int]Module{
	1: {
		ID:               1,
		Name:             "SSH Backdoor",
		Description:      "Create SSH backdoor with custom port",
		RequiredPrivs:    "root",
		SupportedSystems: []string{"Ubuntu", "CentOS"},
		RiskLevel:        "High",
	},
	2: {
		ID:               2,
		Name:             "Crontab Backdoor",
		Description:      "Create persistent crontab backdoor",
		RequiredPrivs:    "root",
		SupportedSystems: []string{"Ubuntu", "CentOS"},
		RiskLevel:        "Medium",
	},
}
