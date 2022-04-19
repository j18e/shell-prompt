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
	ColorTpl  = "\033[0;%dm"
	ResetCode = "\033[0m"

	BLUE    Color = 34
	GREEN   Color = 32
	MAGENTA Color = 35
	RED     Color = 31
	YELLOW  Color = 33
)

type Color int

func (c Color) String() string {
	if *ZSH {
		// this tells zsh that the color code is not part of the prompt length
		return fmt.Sprintf("%%{%s%%}", fmt.Sprintf(ColorTpl, c))
	}
	return fmt.Sprintf(ColorTpl, c)
}

var (
	SHELL = ""
	ZSH   = flag.Bool("zsh", false, "if we're using zsh (affects tab completion)")
)

func main() {
	exitCode := flag.Int("exit-code", 0, "exit code from the previous command (use $?)")
	flag.Parse()

	fmt.Printf("%s %s\n%s%s ", getWD(), gitStatus(), getArrow(*exitCode), reset())
}

func reset() string {
	if *ZSH {
		return fmt.Sprintf("%%{%s%%}", ResetCode)
	}
	return ResetCode
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
	branch   string
	ahead    string
	behind   string
	noBranch bool

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

func handleBranchless() string {
	const rebasePattern = `rebase in progress`
	bs, err := exec.Command("git", "describe", "--tags").Output()
	if err == nil && len(bs) > 0 {
		return strings.TrimSpace(string(bs))
	}
	bs, err = exec.Command("git", "status", "nosuchpath").Output()
	if err == nil && strings.HasPrefix(string(bs), rebasePattern) {
		return "REBASING"
	}
	bs, err = exec.Command("git", "rev-parse", "--short", "HEAD").Output()
	if err == nil {
		return strings.TrimSpace(string(bs))
	}
	return "DETACHED"
}

func gitStatus() *GitStatus {
	// print the file status
	bs, err := exec.Command("git", "status", "--branch", "--porcelain").Output()
	if err != nil {
		return &GitStatus{}
	}
	output := strings.Split(string(bs), "\n")

	status := &GitStatus{}
	status.ParseBranch(output[0])

	for _, l := range output[1:] {
		switch {
		case strings.HasPrefix(l, "??"):
			status.Unindexed.Added++
		case strings.HasPrefix(l, "A "):
			status.Indexed.Added++
		case strings.HasPrefix(l, "M "):
			status.Indexed.Modified++
		case strings.HasPrefix(l, "D "):
			status.Indexed.Deleted++
		case strings.HasPrefix(l, "R "):
			status.Indexed.Modified++
		case strings.HasPrefix(l, " M"):
			status.Unindexed.Modified++
		case strings.HasPrefix(l, " D"):
			status.Unindexed.Deleted++
		case strings.HasPrefix(l, " T"):
			status.Unindexed.Modified++
		case strings.HasPrefix(l, "T "):
			status.Indexed.Modified++
		case strings.HasPrefix(l, "MM"):
			status.Indexed.Modified++
			status.Unindexed.Modified++
		case strings.HasPrefix(l, "UU"):
			status.Unindexed.Added++
		}
	}
	return status
}

func (s *GitStatus) ParseBranch(line string) {
	const (
		noBranch = "## HEAD (no branch)"
	)
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

	if line == noBranch {
		s.branch = handleBranchless()
		return
	}
	s.branch = "UNKNOWN"
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
