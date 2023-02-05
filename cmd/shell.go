package cmd

import (
	"bytes"
	"io"
	"log"
	"os"
	"os/exec"
)

const ShellToUse = "bash"

type OutMsg struct {
	cmd    string
	out    string
	errout string
}

func ExecCmd(command string) (string, string) {
	cmd := exec.Command(ShellToUse, "-c", command)

	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = io.MultiWriter(os.Stdout, &stdoutBuf)
	cmd.Stderr = io.MultiWriter(os.Stderr, &stderrBuf)

	err := cmd.Run()
	if err != nil {
		log.Fatalf("cmd.Run() failed with %s\n", err)
	}

	return stdoutBuf.String(), stderrBuf.String()
}
