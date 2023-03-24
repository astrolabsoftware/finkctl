package cmd

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
	"text/template"
)

const ShellToUse = "bash"

type OutMsg struct {
	cmd    string
	out    string
	errout string
}

func ExecCmd(command string) (string, string) {

	var stdoutBuf, stderrBuf bytes.Buffer
	if !dryRun {
		log.Printf("Launch command: %v", command)
		cmd := exec.Command(ShellToUse, "-c", command)

		cmd.Stdout = io.MultiWriter(os.Stdout, &stdoutBuf)
		cmd.Stderr = io.MultiWriter(os.Stderr, &stderrBuf)

		err := cmd.Run()
		if err != nil {
			log.Fatalf("cmd.Run() failed with %s\n", err)
		}
		log.Printf("%v", stdoutBuf)
		log.Printf("%v", stderrBuf)

	} else {
		log.Printf("Dry run")
		fmt.Println(command)
	}
	return stdoutBuf.String(), stderrBuf.String()
}

func format(s string, v interface{}) string {
	t, b := new(template.Template), new(strings.Builder)
	err := template.Must(t.Parse(s)).Execute(b, v)
	if err != nil {
		log.Fatalf("Error while formatting string %s: %v", s, err)
	}
	return b.String()
}
