# Shell Prompt
A lightning fast terminal prompt including detailed Git status info

## Installation
`go get github.com/j18e/shell-prompt`

### Bash
Add the following to your `.bashrc`:
```
function _update_ps1() {
    PS1="$(shell-prompt -exit-code $?)"
}

PROMPT_COMMAND="_update_ps1; $PROMPT_COMMAND"
```

### ZSH
Add the following to your `.zshrc` (you should be using oh-my-zsh):
```
which shell-prompt >> /dev/null || go install github.com/j18e/shell-prompt
PROMPT='$(shell-prompt -exit-code $? -zsh)'
```
