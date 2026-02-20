package tui

import tea "github.com/charmbracelet/bubbletea"

// Step identifies each screen in the wizard.
type Step int

const (
	StepWelcome Step = iota
	StepDevice
	StepPartSize
	StepEncrypt
	StepPassphrase
	StepHostname
	StepTimezone
	StepUsername
	StepUserPassword
	StepRootPassword
	StepZRAMSize
	StepDesktop
	StepShell
	StepSSHD
	StepSSHPubKey
	StepConfirm
	StepInstall
	stepCount
)

// StepModel is the interface each wizard step must implement.
type StepModel interface {
	tea.Model
	// Title returns the step's heading.
	Title() string
}
