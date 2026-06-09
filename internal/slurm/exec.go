package slurm

import "os/exec"

func Run(command string, args []string) ([]byte, error) {
	return exec.Command(command, args...).Output()
}
