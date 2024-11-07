package stages

import "os/exec"

// ResetCmd returns the command to reset the nodeadm installation.
func ResetCmd() *exec.Cmd {
	return exec.Command("bash", resetScript, runtimeRoot)
}
