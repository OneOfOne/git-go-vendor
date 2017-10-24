package main

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"github.com/mattn/go-colorable"

	cli "gopkg.in/urfave/cli.v2"
)

var (
	stdout         = colorable.NewColorableStdout()
	stderr         = colorable.NewColorableStderr()
	boldAt         = color.New(color.Bold, color.FgBlue).Sprint("@")
	boldBlueStar   = color.New(color.Bold, color.FgBlue).Sprint("*")
	boldYellowStar = color.New(color.Bold, color.FgYellow).Sprint("*")
	boldRedStar    = color.New(color.Bold, color.FgRed).Sprint("*")
	bold           = color.New(color.Bold).Sprint

	verbose bool
	dryRun  bool
	gitPath string

	app = cli.App{
		Name:    "git-go-vendor",
		Version: "v0.1",
		Usage:   "Simple vendoring using git submodules.",
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
						Usage:   "branch to checkout the repo as, defaults to master",
					},
				},
				Usage:  "adds or replaces a vendor package {git-repo}[@branch|tag|hash] [alias]",
				Action: addSubModule,
			},
			{
				Name:    "update",
				Aliases: []string{"up"},
				Usage:   "updates a vendored package or all of them if non is specified",
				Action:  upSubModule,
			},
			{
				Name:    "remove",
				Aliases: []string{"rm"},
				Usage:   "removes the vendor package",
				Action:  rmSubModule,
			},
		},
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:        "verbose",
				Aliases:     []string{"v"},
				Usage:       "verbose output",
				Destination: &verbose,
			},
			&cli.BoolFlag{
				Name:        "dry-run",
				Aliases:     []string{"n"},
				Usage:       "dry-run, don't actually execute any git commands",
				Destination: &dryRun,
			},
			&cli.StringFlag{
				Name:        "git-path",
				Aliases:     []string{"git"},
				Value:       "git",
				Usage:       "path to the git executable",
				Destination: &gitPath,
			},
		},
		Action: func(ctx *cli.Context) error {
			return ctx.App.Command("list").Run(ctx)
		},
	}
)

func main() {
	log.SetFlags(log.Lshortfile)
	cli.VersionFlag = &cli.BoolFlag{
		Name:    "version",
		Aliases: []string{"V"},
		Usage:   "print the version",
	}
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
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
		return cli.Exit("add requires a package path", 1)
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

	if _, err = runCmd(gitPath, "submodule", "add", "--force", path, alias); err != nil {
		return cli.Exit(err, 2)
	}

	// TODO:
	// if _, err = runCmd(gitPath, "config", "-f", ".gitmodules", "submodule"+path+".shallow", "true"); err != nil {
	// 	return cli.Exit(err, 2)
	// }

	if branch == commit {
		_, err = runCmd(gitPath, "-C", alias, "checkout", "-t", "-B", branch, commit)
	} else {
		_, err = runCmd(gitPath, "-C", alias, "checkout", commit)
	}
	if err != nil {
		return cli.Exit(err, 2)
	}

	// 	fmt.Printf("%s Successfully vendored %s as %s %s %s.\n", boldBlueStar, path, alias, boldAt, commit)
	if s := repoString(alias, ""); s != "" {
		printf("%s %s", bold("Added"), s)
	}
	return
}

func upSubModule(c *cli.Context) error {
	sms := c.Args().Slice()
	if len(sms) == 0 {
		sms = allSubModules()
	}
	for _, sm := range sms {
		if !strings.HasPrefix(sm, "vendor/") {
			sm = "vendor/" + sm
		}
		if _, err := runCmd(gitPath, "-C", sm, "pull", "--prune"); err != nil {
			if strings.Contains(err.Error(), "not currently on a branch") {
				errPrintf("%s %s, not on a branch", bold("Skipping"), sm[7:])
				continue
			} else {
				return cli.Exit(sm+" git pull failed: "+err.Error(), 2)
			}
		}
		if s := repoString(sm, ""); s != "" {
			printf("%s %s", bold("Updated"), s)
		}
	}
	return nil
}

func rmSubModule(c *cli.Context) error {
	alias := c.Args().Get(0)
	if alias == "" {
		return cli.Exit("delete requires a package path", 1)

	}

	if !strings.HasPrefix(alias, "vendor/") {
		alias = "vendor/" + alias
	}

	if _, err := runCmd(gitPath, "submodule", "deinit", "--force", alias); err != nil {
		return cli.Exit(err, 2)
	}

	if _, err := runCmd(gitPath, "rm", "--force", alias); err != nil {
		return cli.Exit(err, 2)
	}

	if dryRun {
		return nil
	}
	// debug
	if err := os.RemoveAll(filepath.Join(".git/modules/", alias)); err != nil {
		return cli.Exit(err, 3)
	} else {
		verbosePrintf("%s .git/modules/%s", bold("Deleting"), alias)
	}

	if st, err := os.Stat(".gitmodules"); err == nil && st.Size() == 0 {
		if err := os.Remove(".gitmodules"); err != nil {
			return cli.Exit(err, 3)
		}
		verbosePrintf("%s .gitmodules because it was empty", bold("Deleting"))
	}

	printf("%s %s", bold("Removed"), alias)
	return nil
}

func listSubModules(c *cli.Context) error {
	for _, sm := range allSubModules() {
		if s := repoString(sm, ""); s != "" {
			printf("%s", s)
		}
	}
	return nil
}

func allSubModules() (sms []string) {
	out, err := runCmd(gitPath, "submodule", "status", "--recursive", "vendor/")
	if err != nil {
		return nil
	}
	for _, l := range out {
		p := strings.Split(l, " ")
		sms = append(sms, p[1])
	}
	return
}

func submoduleURL(path string) string {
	if out, _ := runCmd(gitPath, "config", "submodule."+path+".url"); len(out) == 1 {
		return out[0]
	}
	return ""
}

func repoString(path, hash string) string {
	addr := submoduleURL(path)
	if addr == "" && !dryRun {
		return ""
	}

	if idx := strings.Index(addr, "://"); idx > -1 {
		addr = addr[idx+3:]
	}

	if hash == "" {
		out, _ := runCmd(gitPath, "-C", path, "describe", "--always", "--all", "--abbrev=10")
		if len(out) == 0 {
			return ""
		}
		hash = strings.TrimPrefix(out[0], "heads/")
	}

	if addr == path[7:] {
		return fmt.Sprintf("%s %s %s", path, boldAt, hash)
	}

	return fmt.Sprintf("%s %s %s → %s", addr, boldAt, hash, path)

}

func runCmd(name string, args ...string) ([]string, error) {
	cmd := exec.Command(name, args...)
	cmd.Env = append(cmd.Env, "LANG=C") // try to run the english version

	verbosePrintf("%s %s", bold("Executing"), strings.Join(cmd.Args, " "))
	if dryRun {
		return nil, nil
	}
	out, err := cmd.CombinedOutput()
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

func errPrintf(f string, args ...interface{}) {
	fmt.Fprintf(stderr, boldRedStar+" "+f+"\n", args...)
}

func printf(f string, args ...interface{}) {
	fmt.Fprintf(stdout, boldBlueStar+" "+f+"\n", args...)
}

func verbosePrintf(f string, args ...interface{}) {
	if !verbose && !dryRun {
		return
	}
	fmt.Fprintf(stderr, boldYellowStar+" "+f+"\n", args...)
}
