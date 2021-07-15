package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
)

// Program represents an external program with name and path
type Program struct {
	Name              string
	Path              string
	SupportedLanguage string
}

// runExternalProgram invokes an external program with the provided program path and
// sends/receives messages from stdin/stdout/stderr
func (p *Program) runExternalProgram() (string, error) {
	b, err := json.Marshal(&Program{Name: p.Name, Path: p.Path, SupportedLanguage: p.SupportedLanguage})
	if err != nil {
		fmt.Printf("JSON marshal error: %s", err)
		return "", err
	}

	exampleJSON := string(b)

	// invoking a python program using exec.Command
	cmd := exec.Command(p.SupportedLanguage, p.Path, "test-flag", exampleJSON)

	var rawBytes bytes.Buffer
	rawBytes.Write([]byte("Writing to external program: hello" + "\n"))

	// cmd.Stdout = os.Stdout
	cmd.Stdin = &rawBytes
	cmd.Stderr = os.Stderr

	out, err := cmd.Output()
	if err != nil {
		fmt.Println(err.Error())
		return "", err
	}

	return string(out), nil
}

func main() {
	p := &Program{Name: "myExternalProgram",
		Path:              "/tmp/test.py",
		SupportedLanguage: "python"}

	output, err := p.runExternalProgram()
	if err != nil {
		fmt.Printf("unable to run external program: %v", err)
		return
	}

	// todo(rashmigottipati): add logic to unmarshal into a struct that this program expects

	fmt.Printf("Output: %s \n", string(output))
}
