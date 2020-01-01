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

	BLUE        = `%{$fg[blue]%}`
	GREEN       = `%{$fg[green]%}`
	MAGENTA     = `%{$fg[magenta]%}`
	RED         = `%{$fg[red]%}`
	YELLOW      = `%{$fg[yellow]%}`
	RESET_COLOR = `%{$reset_color%}`
)

func main() {
	exitCode := flag.Int("exit-code", 0, "exit code from the previous command (use $?)")
	flag.Parse()

	wd := getWD()
	arrow := getArrow(*exitCode)
	gitStatus := NewGitStatus()
	fmt.Printf(`echo "%s%s %s\n%s%s "`, BLUE, wd, gitStatus, arrow, RESET_COLOR)
}

func getArrow(code int) string {
	if code != 0 {
		return RED + PROMPT_ARROW
	}
	return YELLOW + PROMPT_ARROW
}

func getWD() string {
	pwd := strings.Replace(os.Getenv("PWD"), os.Getenv("HOME"), "~", 1)
	split := strings.Split(pwd, "/")

	// PWD is short enough, display as is
	if len(split) < 5 {
		return pwd
	}

	// PWD is too long, we'll obscure the upper levels
	pwd = path.Join(append(append(split[:1], WD_DOTS), split[len(split)-3:]...)...)

	return pwd
}

type GitStatus struct {
	branch string
	ahead  string
	behind string

	idxNew    int // "^A"
	idxMod    int // "^M"
	idxDel    int // "^D"
	uidxMod   int // "^.M"
	uidxDel   int // "^.D"
	untracked int // "^??"
}

func NewGitStatus() *GitStatus {
	status := &GitStatus{}

	// print the file status
	cmd := exec.Command("git", "status", "--branch", "--porcelain")
	bs, err := cmd.Output()
	if err != nil {
		return status
	}
	output := strings.Split(string(bs), "\n")

	status.ParseBranch(output[0])

	for _, l := range output[1:] {
		switch {
		case strings.HasPrefix(l, "A"):
			status.idxNew++
		case strings.HasPrefix(l, "M"):
			status.idxMod++
		case strings.HasPrefix(l, "D"):
			status.idxDel++
		case strings.HasPrefix(l, " M"):
			status.uidxMod++
		case strings.HasPrefix(l, " D"):
			status.uidxDel++
		case strings.HasPrefix(l, "??"):
			status.untracked++
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
			switch name {
			case "branch":
				s.branch = match[i]
			case "ahead":
				s.ahead = match[i]
			case "behind":
				s.behind = match[i]
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

	res := MAGENTA + GIT_BRANCH + " " + s.branch

	// add unindexed git files
	if s.untracked > 0 || s.uidxMod > 0 || s.uidxDel > 0 {
		res += RED
		if s.untracked > 0 {
			res += fmt.Sprintf(` +%d`, s.untracked)
		}
		if s.uidxMod > 0 {
			res += fmt.Sprintf(` ~%d`, s.uidxMod)
		}
		if s.uidxDel > 0 {
			res += fmt.Sprintf(` -%d`, s.uidxDel)
		}
	}

	// add indexed git files
	if s.idxNew > 0 || s.idxMod > 0 || s.idxDel > 0 {
		res += GREEN
		if s.idxNew > 0 {
			res += fmt.Sprintf(` +%d`, s.idxNew)
		}
		if s.idxMod > 0 {
			res += fmt.Sprintf(` ~%d`, s.idxMod)
		}
		if s.idxDel > 0 {
			res += fmt.Sprintf(` -%d`, s.idxDel)
		}
	}

	// add ahead/behind
	if s.ahead != "" || s.behind != "" {
		res += YELLOW
		if s.ahead != "" {
			res += " " + GIT_AHEAD + s.ahead
		}
		if s.behind != "" {
			res += " " + GIT_BEHIND + s.behind
		}
	}

	return res
}
