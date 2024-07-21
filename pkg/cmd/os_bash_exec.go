package cmd

import (
	"bytes"
	"os/exec"
)

type OsBashExec struct{}

func NewOsBashExec() OsBashExec {
	return OsBashExec{}
}

func (osBashExec OsBashExec) Exec(command string) (stdout, stderr string, err error) {
	cmd := exec.Command("/bin/bash", "-c", command)

	var stdoutBuffer, stderrBuffer bytes.Buffer
	cmd.Stdout = &stdoutBuffer
	cmd.Stderr = &stderrBuffer

	if err := cmd.Run(); err != nil {
		return "", "", err
	}

	return stdoutBuffer.String(), stderrBuffer.String(), nil
}
