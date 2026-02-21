package installer

// Phase identifies each installation phase.
type Phase int

const (
	PhasePrepare Phase = iota
	PhasePartition
	PhaseLUKS
	PhaseBtrfs
	PhaseBaseInstall
	PhaseSystemConfig
	PhaseSwap
	PhaseBootloader
	PhaseServices
	PhaseSSHD
	PhaseDocker
	PhaseDesktop
	PhaseSoftware
	PhaseDotfiles
	phaseCount
)

func (p Phase) String() string {
	switch p {
	case PhasePrepare:
		return "Preparing system"
	case PhasePartition:
		return "Partitioning disk"
	case PhaseLUKS:
		return "Setting up LUKS encryption"
	case PhaseBtrfs:
		return "Configuring btrfs subvolumes"
	case PhaseBaseInstall:
		return "Installing base system"
	case PhaseSystemConfig:
		return "Configuring system"
	case PhaseSwap:
		return "Setting up ZRAM swap"
	case PhaseBootloader:
		return "Installing bootloader"
	case PhaseServices:
		return "Enabling services"
	case PhaseSSHD:
		return "Configuring SSH server"
	case PhaseDocker:
		return "Installing Docker"
	case PhaseSoftware:
		return "Installing software"
	case PhaseDesktop:
		return "Installing desktop environment"
	case PhaseDotfiles:
		return "Installing dotfiles"
	default:
		return "Unknown phase"
	}
}

// PhaseUpdate is sent through the progress channel to report status to the TUI.
type PhaseUpdate struct {
	Phase       Phase
	Description string
	Percent     float64
	LogLine     string
	Done        bool
	Err         error
}
