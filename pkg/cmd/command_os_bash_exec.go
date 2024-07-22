package cmd

import (
	"bytes"
	"fmt"
	"os/exec"

	"github.com/rogeriofbrito/sinth-chaos-poc/pkg/log"
)

type OsBashExecCommand struct{}

func NewOsBashExecCommand() OsBashExecCommand {
	log.Info("Creating new OsBashExecCommand cmd")
	return OsBashExecCommand{}
}

func (o OsBashExecCommand) Exec(command string) (string, string, error) { // TODO: change to pointer
	log.Infof("Executing command with bash: %s", command)
	cmd := exec.Command("/bin/bash", "-c", command)

	var stdoutBuffer, stderrBuffer bytes.Buffer
	cmd.Stdout = &stdoutBuffer
	cmd.Stderr = &stderrBuffer

	if err := cmd.Run(); err != nil {
		stderr := stderrBuffer.String() // TODO: when command fails, there is stderr?
		return "", "", fmt.Errorf("OsBashExecCommand.Exec - error on executing command: %w, stderr: %s", err, stderr)
	}

	stdout := stdoutBuffer.String()
	stderr := stderrBuffer.String()

	log.Debugf("Stdout: %s", stdout)
	log.Debugf("Stderr: %s", stderr)

	return stdout, stderr, nil
}
