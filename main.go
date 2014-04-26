package main

import (
	"fmt"
	"log"
	"path"
	"path/filepath"
	"os"
	"os/exec"
	"io/ioutil"
	"strings"
)

import (
	"github.com/timtadh/getopt"
	"github.com/timtadh/gopkgr/tar"
	"github.com/timtadh/gopkgr/goenv"
)

var ErrorCodes map[string]int = map[string]int{
	"usage":1,
	"version":2,
	"opts":3,
	"badint":5,
}

var UsageMessage string = "gopkgr -h"
var ExtendedMessage string = `
Before getting started run

    $ eval $(gopkgr --goenv-function)

This installs the shell function, goenv, you need to utilize the virtualenv
like functionality.

Commands
    goenv <project>                     start a goenv (don't run this directly
                                        instead use the goenv command)
    install <gopath> <tarball>          install a package to the gopath
    remove <gopath> <tarball>           remove a package from the gopath
    mkpkg -o name.tar.gz <gopath>       packages the src directory at this go path

gopkgr Options
    -h, --help                          print this message
    --goenv-function                    print the goenv function
`

var GoEnvUsageMessage string = "goenv -h"
var GoEnvMessage string = `
goenv allows you to manage your shell environment. It sets three environment
variables and a deactivate function.

Environment variables manipulated:
    $GOPATH - adds your project dir and the virtual env
    $PATH - adds bin dirs for your project and virtual env
    $GOENV - sets so future commands can find the root

The shell function added:
    deactivate - resets the variables and removes this function

goenv Commands
    activate <project>                  activate the virtual env
    deactivate                          deactivate the virtual env
    install <tarball>                   install a package to this env
    remove <tarball>                    remove a package from this env
    getpkg -o name.tar.gz <url>         gets the go gettable repository and
                                        packages it as tar.gz. Note: this
                                        is scoped because it uses the goenv
                                        you are currently using to limit which
                                        dependencies to download. You can use
                                        this property to make minimal tarballs

<project> is a path to your project directory.
<tarball> is a path to a source tarball.
<url> is an import url ie. github.com/timtadh/gopkgr
`

func Usage(code int) {
    fmt.Fprintln(os.Stderr, UsageMessage)
    if code == 0 {
        fmt.Fprintln(os.Stderr, ExtendedMessage)
        code = ErrorCodes["usage"]
    } else {
        fmt.Fprintln(os.Stderr, "Try -h or --help for help")
    }
    os.Exit(code)
}

func GoEnvUsage(code int) {
    fmt.Fprintln(os.Stderr, GoEnvUsageMessage)
    if code == 0 {
        fmt.Fprintln(os.Stderr, GoEnvMessage)
        code = ErrorCodes["usage"]
    } else {
        fmt.Fprintln(os.Stderr, "Try -h or --help for help")
    }
    os.Exit(code)
}

func pack_unpack_test() {
	cwd, _ := os.Getwd()
	prefix := path.Join(cwd, "test")
	if err := tar.Archive(prefix, "src", "ex.tar.gz"); err != nil {
		log.Fatalln(err)
	}
	if err := tar.Unpack(path.Join(cwd, "extract"), "ex.tar.gz"); err != nil {
		log.Fatalln(err)
	}
	fmt.Println("success!")
}

func Pkger() error {
	bin, err := exec.LookPath("gopkgr")
	if err != nil {
		return err
	}
	switch goenv.Shell {
	case "bash":
		fmt.Printf("goenv () { IFS=$'\\n'; for x in $(%s goenv $@); do eval $x; done ; unset IFS; }\n", bin)
	default:
		return fmt.Errorf("Shell, %s, is not yet supported", goenv.Shell)
	}
	return nil
}

func activate(project string) {
	if os.Getenv("GOENV") != "" {
		fmt.Fprintln(os.Stderr, "already in project, run `goenv deactivate` first")
		GoEnvUsage(ErrorCodes["opts"])
	}
	project, err := filepath.Abs(project)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		GoEnvUsage(ErrorCodes["opts"])
	}
	if err := os.MkdirAll(project, 0775); err != nil {
		fmt.Fprintln(os.Stderr, err)
		GoEnvUsage(ErrorCodes["opts"])
	}

	venv := path.Join(project, "venv")
	if err := os.MkdirAll(venv, 0775); err != nil {
		fmt.Fprintln(os.Stderr, err)
		Usage(ErrorCodes["opts"])
	}

	PATH := path.Join(venv, "bin") + ":" + os.Getenv("PATH")
	PATH = path.Join(project, "bin") + ":" + PATH

	c := goenv.NewContext()
	c.AddGoPath(project)
	c.AddGoPath(venv)
	c.Export("PATH", PATH)
	c.Export("GOENV", venv)
	c.WriteAll()
}

func deactivate() {
	if os.Getenv("GOENV") == "" {
		fmt.Fprintln(os.Stderr, "not in a virtual environment")
		GoEnvUsage(ErrorCodes["opts"])
	}
	fmt.Println("deactivate")
}

func goenv_cmd(argv []string) {
	args, optargs, err := getopt.GetOpt(
		argv,
		"h",
		[]string{
			"help",
		},
	)

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		GoEnvUsage(ErrorCodes["opts"])
	}

	for _, oa := range optargs {
		switch oa.Opt() {
		case "-h", "--help": GoEnvUsage(0)
		}
	}

	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "Expected a sub-command")
		GoEnvUsage(ErrorCodes["opts"])
	}

	goenv := os.Getenv("GOENV")
	cmd := args[0]
	args = args[1:]
	switch cmd {
	case "getpkg":
		if goenv == "" {
			fmt.Fprintln(os.Stderr, "getpkg requires you to be in a virtual env")
			GoEnvUsage(ErrorCodes["opts"])
		}
		getpkg(goenv, args)
	case "activate":
		if len(args) == 0 {
			fmt.Fprintln(os.Stderr, "activate require a project dir")
			GoEnvUsage(ErrorCodes["opts"])
		}
		activate(args[0])
	case "deactivate":
		deactivate()
	case "install":
		if goenv == "" {
			fmt.Fprintln(os.Stderr, "install requires you to be in a virtual env")
			GoEnvUsage(ErrorCodes["opts"])
		}
		if len(args) == 0 {
			fmt.Fprintln(os.Stderr, "install requires a tarball")
			GoEnvUsage(ErrorCodes["opts"])
		}
		if err := install(goenv, args[0]); err != nil {
			fmt.Fprintln(os.Stderr, err)
			GoEnvUsage(ErrorCodes["opts"])
		}
	case "remove":
		if goenv == "" {
			fmt.Fprintln(os.Stderr, "remove requires you to be in a virtual env")
			GoEnvUsage(ErrorCodes["opts"])
		}
		if len(args) == 0 {
			fmt.Fprintln(os.Stderr, "remove requires a tarball")
			GoEnvUsage(ErrorCodes["opts"])
		}
		if err := remove(goenv, args[0]); err != nil {
			fmt.Fprintln(os.Stderr, err)
			GoEnvUsage(ErrorCodes["opts"])
		}
	default:
		fmt.Fprintf(os.Stderr, "command, %s, not found\n", cmd)
		GoEnvUsage(ErrorCodes["opts"])
	}
}

func mkpkg(argv []string) {
	args, optargs, err := getopt.GetOpt(
		argv,
		"ho:",
		[]string{
			"help", "output=",
		},
	)

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		Usage(ErrorCodes["opts"])
	}

	var output string = ""
	for _, oa := range optargs {
		switch oa.Opt() {
		case "-h", "--help": Usage(0)
		case "-o", "--output": output = oa.Arg()
		}
	}

	if output == "" {
		fmt.Fprintln(os.Stderr, "You must supply an output location")
		Usage(ErrorCodes["opts"])
	}

	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "You must supply a source gopath")
		Usage(ErrorCodes["opts"])
	}
	gopath := args[0]

	if err := tar.Archive(gopath, "src", output); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}

func getpkg(goenv string, argv []string) {
	args, optargs, err := getopt.GetOpt(
		argv,
		"ho:r:",
		[]string{
			"help", "output=", "revision=",
		},
	)

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		GoEnvUsage(ErrorCodes["opts"])
	}

	var output string = ""
	// var revision string = ""
	for _, oa := range optargs {
		switch oa.Opt() {
		case "-h", "--help": GoEnvUsage(0)
		case "-o", "--output": output = oa.Arg()
		// case "-r", "--revision": revision = oa.Arg()
		}
	}

	if output == "" {
		fmt.Fprintln(os.Stderr, "You must supply an output location")
		GoEnvUsage(ErrorCodes["opts"])
	}

	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "You must supply a source repo")
		GoEnvUsage(ErrorCodes["opts"])
	}
	url := args[0]

	gopath, err := ioutil.TempDir("", "gopkgr-gopath-")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(12)
	}
	defer os.RemoveAll(gopath)
	if err := go_get(goenv, gopath, url); err != nil {
		fmt.Fprintln(os.Stderr, err)
		GoEnvUsage(ErrorCodes["opts"])
	}
	if err := go_install(goenv, gopath, url); err != nil {
		fmt.Fprintln(os.Stderr, err)
		GoEnvUsage(ErrorCodes["opts"])
	}

	if err := tar.Archive(gopath, "src", output); err != nil {
		fmt.Fprintln(os.Stderr, err)
		GoEnvUsage(ErrorCodes["opts"])
	}
}

func git_clone(url string) (tree string, err error) {
	return "", fmt.Errorf("unimplemented")
}

func git_checkout(url, tree string) (err error) {
	return fmt.Errorf("unimplemented")
}

func locate(url, tree string) (gopath string, err error) {
	return "", fmt.Errorf("unimplemented")
}

func go_get(goenv, gopath, spec string) error {
	gobin, err := exec.LookPath("go")
	if err != nil {
		return err
	}
	env := os.Environ()
	for i, kv := range env {
		if strings.HasPrefix(kv, "GOPATH") {
			env[i] = fmt.Sprintf("GOPATH=%s", gopath + ":" + goenv)
		}
	}
	go_get := &exec.Cmd{
		Path: gobin,
		Args: []string{gobin, "get", spec},
		Env: env,
		Stdin: os.Stdin,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}
	if err := go_get.Run(); err != nil {
		return err
	}
	return nil
}

func go_install(goenv, gopath, spec string) error {
	gobin, err := exec.LookPath("go")
	if err != nil {
		return err
	}
	env := os.Environ()
	for i, kv := range env {
		if strings.HasPrefix(kv, "GOPATH") {
			env[i] = fmt.Sprintf("GOPATH=%s", gopath + ":" + goenv)
		}
	}
	go_install := &exec.Cmd{
		Path: gobin,
		Args: []string{gobin, "install", spec},
		Env: env,
		Stdin: os.Stdin,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}
	if err := go_install.Run(); err != nil {
		return err
	}
	return nil
}

func install(gopath, tarball string) error {
	gopath, err := filepath.Abs(gopath)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(gopath, 0775); err != nil {
		return err
	}
	if !tar.Exists(tarball) {
		return fmt.Errorf("supplied tarball, %s, does not exist\n", tarball)
	}
	if err := tar.Unpack(gopath, tarball); err != nil {
		return err
	}
	return go_install("", gopath, "...")
}

func install_cmd(argv []string) {
	args, optargs, err := getopt.GetOpt(
		argv,
		"h",
		[]string{
			"help", 
		},
	)

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		Usage(ErrorCodes["opts"])
	}

	for _, oa := range optargs {
		switch oa.Opt() {
		case "-h", "--help": Usage(0)
		}
	}

	if len(args) < 2 {
		fmt.Fprintln(os.Stderr, "You must supply a target gopath and tarball")
		Usage(ErrorCodes["opts"])
	}
	gopath := args[0]
	tarball := args[1]

	if err := install(gopath, tarball); err != nil {
		fmt.Fprintln(os.Stderr, err)
		Usage(ErrorCodes["opts"])
	}
}

func remove(gopath, tarball string) error {
	gopath, err := filepath.Abs(gopath)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(gopath, 0775); err != nil {
		return err
	}
	if !tar.Exists(tarball) {
		return fmt.Errorf("supplied tarball, %s, does not exist\n", tarball)
	}
	if err := tar.Remove(gopath, tarball); err != nil {
		return err
	}
	os.RemoveAll(path.Join(gopath, "bin"))
	os.RemoveAll(path.Join(gopath, "pkg"))
	return go_install("", gopath, "...")
}

func remove_cmd(argv []string) {
	args, optargs, err := getopt.GetOpt(
		argv,
		"h",
		[]string{
			"help", 
		},
	)

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		Usage(ErrorCodes["opts"])
	}

	for _, oa := range optargs {
		switch oa.Opt() {
		case "-h", "--help": Usage(0)
		}
	}

	if len(args) < 2 {
		fmt.Fprintln(os.Stderr, "You must supply a target gopath and tarball")
		Usage(ErrorCodes["opts"])
	}
	gopath := args[0]
	tarball := args[1]

	if err := remove(gopath, tarball); err != nil {
		fmt.Fprintln(os.Stderr, err)
		Usage(ErrorCodes["opts"])
	}
}

func main() {

	args, optargs, err := getopt.GetOpt(
		os.Args[1:],
		"h",
		[]string{
			"help",
			"goenv-function",
		},
	)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		Usage(ErrorCodes["opts"])
	}

	for _, oa := range optargs {
		switch oa.Opt() {
		case "-h", "--help": Usage(0)
		case "--goenv-function":
			err := Pkger()
			if err != nil {
				panic(err)
			}
			return
		}
	}

	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "Expected a sub-command")
		Usage(ErrorCodes["opts"])
	}

	cmd := args[0]
	switch cmd {
	case "goenv": goenv_cmd(args[1:])
	case "mkpkg": mkpkg(args[1:])
	case "install": install_cmd(args[1:])
	case "remove": remove_cmd(args[1:])
	default:
		fmt.Fprintf(os.Stderr, "command, %s, not found\n", cmd)
		Usage(ErrorCodes["opts"])
	}
}

