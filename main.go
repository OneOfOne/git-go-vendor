package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
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
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "branch",
						Aliases: []string{"b"},
						Value:   "master",
						Usage:   "branch to checkout the repo as, defaults to master.",
					},
				},
				Usage:  "adds or replaces a vendor package {git-repo}[@branch|tag|hash] [alias]",
				Action: addSubModule,
			},
			{
				// TODO
				Name:    "update",
				Aliases: []string{"up"},
				Usage:   "updates a vendored package or all of them if non is specified.",
				Action:  upSubModule,
			},
			{
				Name:    "remove",
				Aliases: []string{"rm"},
				Usage:   "removes the vendor package",
				Action:  rmSubModule,
			},
		},
		Action: func(c *cli.Context) (err error) {
			return listSubModules(c)
		},
	}
)

func main() {
	log.SetFlags(log.Lshortfile)
	app.Run(os.Args)
}

func addSubModule(c *cli.Context) (err error) {
	var (
		args   = c.Args()
		path   = args.Get(0)
		alias  = args.Get(1)
		branch = c.String("branch")
		commit = branch
	)

	if path == "" {
		return cli.Exit("add requires a package path.", 1)
	}

	if idx := strings.LastIndex(path, "@"); idx > -1 {
		commit = path[idx+1:]
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

	if strings.HasSuffix(alias, ".git") {
		alias = alias[:len(alias)-3]
	}

	if _, err = runCmd("git", "submodule", "add", "--force", path, alias); err != nil {
		return cli.Exit(err, 2)
	}

	if branch == commit {
		_, err = runCmd("git", "-C", alias, "checkout", "-t", "-B", branch, commit)
	} else {
		_, err = runCmd("git", "-C", alias, "checkout", commit)
	}
	if err != nil {
		return cli.Exit(err, 2)
	}

	// 	fmt.Printf("%s Successfully vendored %s as %s %s %s.\n", boldBlueStar, path, alias, boldAt, commit)
	printRepo(alias, "")
	return
}

func upSubModule(c *cli.Context) error {
	log.Println("x")
	return io.EOF
}

func rmSubModule(c *cli.Context) error {
	alias := c.Args().Get(0)
	if alias == "" {
		return cli.Exit("delete requires a package path.", 1)

	}

	if !strings.HasPrefix(alias, "vendor/") {
		alias = "vendor/" + alias
	}

	if _, err := runCmd("git", "submodule", "deinit", "--force", alias); err != nil {
		log.Printf("%s", err)
		return err
	}

	if _, err := runCmd("git", "rm", "--force", alias); err != nil {
		return cli.Exit(err, 2)
	}

	fmt.Printf("%s Successfully removed %s.\n", boldBlueStar, alias)
	return nil
}

func listSubModules(c *cli.Context) error {
	out, err := runCmd("git", "submodule", "status", "--recursive", "vendor/")
	if err != nil {
		return cli.Exit(err, 2)
	}
	for _, l := range out {
		p := strings.Split(l, " ")
		printRepo(p[1], p[0])
	}
	return nil
}

func submoduleURL(path string) string {
	if out, _ := runCmd("git", "config", "submodule."+path+".url"); len(out) == 1 {
		return out[0]
	}
	return ""
}

func printRepo(path, hash string) {
	addr := submoduleURL(path)
	if addr == "" {
		return
	}

	if idx := strings.Index(addr, "://"); idx > -1 {
		addr = addr[idx+3:]
	}

	if hash == "" {
		out, _ := runCmd("git", "-C", path, "describe", "--always", "--abbrev=8")
		if len(out) == 0 {
			return
		}
		hash = strings.TrimPrefix(out[0], "heads/")
	}

	if addr == path[7:] {
		fmt.Println(boldBlueStar, path, boldAt, hash)
	} else {
		fmt.Println(boldBlueStar, addr, boldAt, hash, "→", path)
	}
}

func runCmd(name string, args ...string) ([]string, error) {
	out, err := exec.Command(name, args...).CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("%s", out)
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
