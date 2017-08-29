package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/fatih/color"

	cli "gopkg.in/urfave/cli.v2"
)

var (
	boldAt       = color.New(color.Bold, color.FgBlue).Sprint("@")
	boldBlueStar = color.New(color.Bold, color.FgBlue).Sprint("*")

	app = cli.App{
		Name:    "git-go-dep",
		Version: "v0.1",
		Commands: []*cli.Command{
			{
				Name:    "list",
				Aliases: []string{"ls"},
				Usage:   "list all current directly vendored packages",
				Action:  listSubModules,
			},
			{
				Name:    "add",
				Aliases: []string{"a"},
				Usage:   "add a new vendor {git-repo}[@branch|tag|hash] [alias]",
				Action:  addSubModule,
			},
		},
	}
)

func main() {
	app.Run(os.Args)
}

func addSubModule(c *cli.Context) error {
	var (
		args        = c.Args()
		path, alias = args.Get(0), args.Get(1)
		spec        string
	)

	if idx := strings.LastIndex(path, "@"); idx > -1 {
		spec = path[idx+1:]
		path = path[:idx]
	}

	if idx := strings.Index(path, ":"); idx == -1 {
		path = "https://" + path
	}

	if alias == "" {
		alias = path[strings.Index(path, "://")+3:]
	}

	if !strings.HasPrefix(alias, "vendor/") {
		alias = "vendor/" + alias
	}
	out, err := runCmd("git", "submodule", "status", "--recursive", "vendor/")
	fmt.Printf("%q %q %q\n", path, alias, spec)
	return nil
}

func listSubModules(c *cli.Context) error {
	out, err := runCmd("git", "submodule", "status", "--recursive", "vendor/")
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v: %s", err, out)
		return err
	}
	for _, l := range out {
		p := strings.Split(l, " ")
		printRepo(p[1], p[0])
	}
	return nil
}

func printRepo(path, hash string) {
	out, _ := runCmd("git", "config", "submodule."+path+".url")
	if len(out) == 0 {
		return
	}

	addr := out[0]
	if idx := strings.Index(addr, "://"); idx > -1 {
		addr = addr[idx+3:]
	}

	if addr == path[7:] {
		fmt.Println(boldBlueStar, path, boldAt, hash[:8])
	} else {
		fmt.Println(boldBlueStar, addr, boldAt, hash[:8], "→", path)
	}
}

func runCmd(name string, args ...string) ([]string, error) {
	out, err := exec.Command(name, args...).CombinedOutput()
	if err != nil {
		return []string{string(out)}, err
	}

	var (
		sc    = bufio.NewScanner(bytes.NewReader(out))
		lines []string
	)

	for sc.Scan() {
		lines = append(lines, strings.TrimSpace(sc.Text()))
	}

	return lines, nil
}
