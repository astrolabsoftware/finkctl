package cmd

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"text/template"
)

const ShellToUse = "bash"

func ExecCmd(command string) (string, string) {

	var stdoutBuf, stderrBuf bytes.Buffer
	if !dryRun {
		slog.Info("Run", "command", command)
		cmd := exec.Command(ShellToUse, "-c", command)

		cmd.Stdout = io.MultiWriter(os.Stdout, &stdoutBuf)
		cmd.Stderr = io.MultiWriter(os.Stderr, &stderrBuf)

		err := cmd.Run()
		if err != nil {
			slog.Error("cmd.Run() failed", "error", err)
			syscall.Exit(1)
		}
		slog.Info("stdout", "buffer", stdoutBuf)
		slog.Info("stderr", "buffer", stderrBuf)

	} else {
		slog.Info("Dry run")
		fmt.Println(command)
	}
	return stdoutBuf.String(), stderrBuf.String()
}

func format(s string, v interface{}) string {
	t, b := new(template.Template), new(strings.Builder)
	err := template.Must(t.Parse(s)).Execute(b, v)
	if err != nil {
		slog.Error("Error while formatting string", "string", s, "error", err)
		syscall.Exit(1)
	}
	return b.String()
}
