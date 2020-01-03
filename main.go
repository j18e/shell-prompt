package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path"
	"regexp"
	"strings"
)

const (
	GIT_BRANCH   = ``
	GIT_AHEAD    = `↑`
	GIT_BEHIND   = `↓`
	WD_DOTS      = `…`
	PROMPT_ARROW = `❯`

	// colors codes
	ResetCode = "\033[0m"

	RESET   Color = 0
	BLUE    Color = 34
	GREEN   Color = 32
	MAGENTA Color = 35
	RED     Color = 31
	YELLOW  Color = 33
)

type Color int

func (c Color) String() string {
	res := ""
	if c == RESET {
		res = ResetCode
	} else {
		res = fmt.Sprintf("\033[0;%dm", c)
	}

	if *ZSH {
		// this tells zsh that the color code is not part of the prompt length
		return fmt.Sprintf("%%{%s%%}", res)
	}
	return res
}

var (
	SHELL = ""
	ZSH   = flag.Bool("zsh", false, "if we're using zsh (affects tab completion)")
)

func main() {
	exitCode := flag.Int("exit-code", 0, "exit code from the previous command (use $?)")
	flag.Parse()

	fmt.Printf("%s %s\n%s%s ", getWD(), gitStatus(), getArrow(*exitCode), RESET)
}

func getArrow(code int) string {
	if code != 0 {
		return fmt.Sprintf("%s%s", RED, PROMPT_ARROW)
	}
	return fmt.Sprintf("%s%s", YELLOW, PROMPT_ARROW)
}

func getWD() string {
	pwd := strings.Replace(os.Getenv("PWD"), os.Getenv("HOME"), "~", 1)
	split := strings.Split(pwd, "/")

	// PWD is short enough, display as is
	if len(split) < 5 {
		return fmt.Sprintf("%s%s", BLUE, pwd)
	}

	// PWD is too long, we'll obscure the upper levels
	pwd = path.Join(append(append(split[:1], WD_DOTS), split[len(split)-3:]...)...)

	return fmt.Sprintf("%s%s", BLUE, pwd)
}

type GitStatus struct {
	branch string
	ahead  string
	behind string

	Unindexed Changes
	Indexed   Changes
}

type Changes struct {
	Color    string
	Added    int
	Modified int
	Deleted  int
}

func (c Changes) String() string {
	res := ""
	if c.Added > 0 {
		res += fmt.Sprintf(" +%d", c.Added)
	}
	if c.Modified > 0 {
		res += fmt.Sprintf(" ~%d", c.Modified)
	}
	if c.Deleted > 0 {
		res += fmt.Sprintf(" -%d", c.Deleted)
	}
	return res
}

func gitStatus() *GitStatus {
	// print the file status
	cmd := exec.Command("git", "status", "--branch", "--porcelain")
	bs, err := cmd.Output()
	if err != nil {
		return &GitStatus{}
	}
	output := strings.Split(string(bs), "\n")

	status := &GitStatus{}
	status.ParseBranch(output[0])

	for _, l := range output[1:] {
		switch {
		case strings.HasPrefix(l, "A"):
			status.Indexed.Added++
		case strings.HasPrefix(l, "M"):
			status.Indexed.Modified++
		case strings.HasPrefix(l, "D"):
			status.Indexed.Deleted++
		case strings.HasPrefix(l, " M"):
			status.Unindexed.Modified++
		case strings.HasPrefix(l, " D"):
			status.Unindexed.Deleted++
		case strings.HasPrefix(l, "??"):
			status.Unindexed.Added++
		}
	}
	return status
}

func (s *GitStatus) ParseBranch(line string) {
	var (
		reFull      = regexp.MustCompile(`^## (?P<branch>\S+)\.{3}\S+( \[(?:ahead (?P<ahead>\d+)(?:, )?)?(?:behind (?P<behind>\d+)?)?\])?$`)
		reNoRemote  = regexp.MustCompile(`^## (\S+)$`)
		reNoCommits = regexp.MustCompile(`^## No commits yet on (\S+)$`)
	)

	match := reFull.FindStringSubmatch(line)
	if len(match) > 1 {
		for i, name := range reFull.SubexpNames() {
			if match[i] == "" {
				continue
			}
			switch name {
			case "branch":
				s.branch = match[i]
			case "ahead":
				s.ahead = " " + GIT_AHEAD + match[i]
			case "behind":
				s.behind = " " + GIT_BEHIND + match[i]
			}
		}
		return
	}

	match = reNoRemote.FindStringSubmatch(line)
	if len(match) > 1 {
		s.branch = match[1]
		return
	}

	match = reNoCommits.FindStringSubmatch(line)
	if len(match) > 1 {
		s.branch = match[1]
		return
	}
}

func (s *GitStatus) String() string {
	if s.branch == "" {
		return ""
	}

	res := fmt.Sprintf("%s%s %s", MAGENTA, GIT_BRANCH, s.branch) // magenta

	if uidx := s.Unindexed.String(); uidx != "" {
		res += fmt.Sprintf("%s%s", RED, uidx)
	}
	if idx := s.Indexed.String(); idx != "" {
		res += fmt.Sprintf("%s%s", GREEN, idx)
	}

	if s.ahead == "" && s.behind == "" {
		return res
	}
	res += fmt.Sprintf("%s%s%s", YELLOW, s.ahead, s.behind)
	return res
}
