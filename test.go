package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// ExternalPlugin represents an external program with name and path
type ExternalPlugin struct {
	Name       string
	Version    string
	Path       string
	DirContext string
}

// PluginRequest contains all information kubebuilder received from the CLI
// and plugins executed before it.
type PluginRequest struct {
	Command  string            `json:"command"`
	Args     []string          `json:"args"`
	Universe map[string]string `json:"universe"`
}

// PluginResponse is returned to kubebuilder by the plugin and contains all files
// written by the plugin following a certain command.
type PluginResponse struct {
	Command  string            `json:"command"`
	Universe map[string]string `json:"universe"`
	Error    bool              `json:"error,omitempty"`
	ErrorMsg string            `json:"error_msg,omitempty"`
}

// runExternalProgram invokes an external program with the provided program path and
// sends/receives messages from stdin/stdout/stderr
func (p *ExternalPlugin) runExternalProgram(req PluginRequest) (res PluginResponse, err error) {
	b, err := json.Marshal(req)
	if err != nil {
		return res, err
	}

	// Invoke the command specified by p in the context dir.
	cmd := exec.Command(p.Path)
	cmd.Dir = p.DirContext
	cmd.Stdin = bytes.NewBuffer(b)
	cmd.Stderr = os.Stderr

	out, err := cmd.Output()
	if err != nil {
		fmt.Fprint(os.Stdout, string(out))
		return res, err
	}

	if json.Unmarshal(out, &res); err != nil {
		return res, err
	}

	return res, nil
}

func main() {
	wd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	// Create the test plugin context.
	p := &ExternalPlugin{
		Name:       "pythonplugin.my.domain",
		Version:    "v1-alpha",
		DirContext: "testproject",
		Path:       filepath.Join(wd, "pythonplugin.sh"),
	}

	var args []string
	if len(os.Args) == 1 {
		// Some test args.
		args = []string{"init", "--domain", "example.com"}
	} else {
		args = make([]string, len(os.Args)-1)
		copy(args, os.Args[1:])
	}

	fmt.Fprintln(os.Stdout, "Running:", p.Path, strings.Join(args, " "))

	// Plugin request setup.
	// Much of this will actually be done by the cli library when implemented in kubebuilder.
	req := PluginRequest{}
	maybeCmdArgs := args[:2]
	if args[0] == "init" {
		if _, err := os.Stat(p.DirContext); err == nil {
			log.Fatalf("project directory %q must not exist", p.DirContext)
		}
		if err := os.MkdirAll(p.DirContext, 0755); err != nil {
			log.Fatalln("failed to create project dir:", err)
		}
		req.Command = args[0]
		req.Args = args[1:]
	} else {
		if info, err := os.Stat(p.DirContext); err != nil || !info.IsDir() {
			log.Fatalf("project directory %q has error: %v", p.DirContext, err)
		}
		if args[0] == "create" {
			req.Command = strings.Join(args[:2], " ")
			req.Args = args[2:]
		}
	}
	if req.Command == "" {
		log.Fatalf("unknown command: %q", maybeCmdArgs[0])
	}
	req.Universe = map[string]string{}

	// Run the plugin.
	res, err := p.runExternalProgram(req)
	if err != nil {
		log.Fatalln("unable to run external program:", err)
	}

	// Error if the plugin failed.
	if res.Error {
		log.Fatal(res.ErrorMsg)
	}

	output, err := json.MarshalIndent(res, "", "  ")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Fprint(os.Stdout, "Universe:\n", string(output))

	// Write all files in the returned universe.
	for subpath, data := range res.Universe {
		if err := os.MkdirAll(filepath.Dir(subpath), 0755); err != nil {
			log.Fatal(err)
		}
		p := filepath.Join(p.DirContext, subpath)
		if err := os.WriteFile(p, []byte(data), 0644); err != nil {
			log.Fatalln("failed to write project file:", err)
		}
	}
}
