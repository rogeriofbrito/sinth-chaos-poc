package cmd

import (
	"bytes"
	"os/exec"

	"github.com/rogeriofbrito/sinth-chaos-poc/pkg/log"
)

type OsBashExec struct{}

func NewOsBashExec() OsBashExec {
	log.Info("Creating new command impl (OsBashExec)")
	return OsBashExec{}
}

func (osBashExec OsBashExec) Exec(command string) (string, string, error) {
	log.Infof("Executing command %s", command)
	cmd := exec.Command("/bin/bash", "-c", command)

	var stdoutBuffer, stderrBuffer bytes.Buffer
	cmd.Stdout = &stdoutBuffer
	cmd.Stderr = &stderrBuffer

	if err := cmd.Run(); err != nil {
		stderr := stderrBuffer.String()
		log.Errorf("Error on executing command: %s, stderr: %s", err, stderr)
		return "", "", err
	}

	stdout := stdoutBuffer.String()
	stderr := stderrBuffer.String()

	log.Infof("Stdout: %s", stdout)
	log.Infof("Stderr: %s", stderr)

	return stdout, stderr, nil
}
