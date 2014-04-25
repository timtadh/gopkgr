package goenv

import (
	"os"
	"fmt"
	"path"
	"strings"
)

var Shell string

func init() {
	shell := os.Getenv("SHELL")
	if shell == "" {
		// assume bash
		Shell = "bash"
	}
	Shell = path.Base(shell)
}

type Context struct {
	Exports map[string]string
	GoPaths []string
}

func NewContext() *Context {
	return &Context{
		Exports: make(map[string]string),
		GoPaths: make([]string, 0),
	}
}

func (self *Context) WriteAll() error {
	err := self.deactivate()
	if err != nil {
		return err
	}
	fmt.Println("echo exporting")
	err = self.export_all()
	if err != nil {
		return err
	}
	return nil
}

func (self *Context) AddGoPath(path string) {
	self.GoPaths = append(self.GoPaths, path)
}

func (self *Context) Export(name, value string) {
	self.Exports[name] = value
}

func (self *Context) deactivate() error {
	switch Shell {
	case "bash":
		return self.deactivate_bash()
	default:
		return fmt.Errorf("Shell, %s, is not yet supported", Shell)
	}
	return nil
}

func (self *Context) deactivate_bash() error {
	reset := func(name string) {
		value := os.Getenv(name)
		fmt.Printf("unset %s; ", name)
		if value != "" {
			fmt.Printf("export %s=%s; ", name, value)
		}
	}
	fmt.Print("deactivate () { ")
	fmt.Print("echo deactivating; ")
	for name := range self.Exports {
		reset(name)
	}
	reset("GOPATH")
	fmt.Print("unset -f deactivate; ")
	fmt.Println(" }")
	return nil
}

func (self *Context) export_all() error {
	for name, value := range self.Exports {
		if err := self.export(name, value); err != nil {
			return err
		}
	}
	GOPATH := strings.Join(self.GoPaths, ":")
	if err := self.export("GOPATH", GOPATH); err != nil {
		return err
	}

	return nil
}

func (self *Context) export(name, value string) error {
	switch Shell {
	case "bash":
		return self.export_bash(name, value)
	default:
		return fmt.Errorf("Shell, %s, is not yet supported", Shell)
	}
	return nil
}

func (self *Context) export_bash(name, value string) error {
	fmt.Printf("export %s=%s\n", name, value)
	return nil
}


