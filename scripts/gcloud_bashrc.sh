# A simple .bashrc for GCE hosts.

# If not running interactively, don't do anything
[ -z "$PS1" ] && return

# Append to the history file, don't overwrite it.
shopt -s histappend

# set a fancy prompt (non-color, unless we know we "want" color)
case "$TERM" in
    xterm-color) color_prompt=yes;;
esac

force_color_prompt=yes
if [ -n "$force_color_prompt" ]; then
    if [ -x /usr/bin/tput ] && tput setaf 1 >&/dev/null; then
	# We have color support; assume it's compliant with Ecma-48
	# (ISO/IEC-6429). (Lack of such support is extremely rare, and such
	# a case would tend to support setf rather than setaf.)
	color_prompt=yes
    else
	color_prompt=
    fi
fi

# Show git branch in shell info
git_branch() {
  git branch 2> /dev/null | sed -e '/^[^*]/d' -e 's/* \(.*\)/(\1)/'
}

if [ "$color_prompt" = yes ]; then
    PS1="${debian_chroot:+($debian_chroot)}\[\033[01;11m\]\$(git_branch) \[\033[01;32m\]\u@\h\[\033[00m\]:\[\033[01;34m\]\w\[\033[00m\]\$ "
else
    PS1='${debian_chroot:+($debian_chroot)}\u@\h:\w\$ '
fi
unset color_prompt force_color_prompt

STARTCOLOR='\e[1;40m';
ENDCOLOR="\e[0m"

# enable color support of ls.
if [ -x /usr/bin/dircolors ]; then
    eval "`dircolors -b`"
    alias ls='ls --color=auto'
fi

# enable programmable completion features.
if [ -f /etc/bash_completion ]; then
    . /etc/bash_completion
fi

alias e="emacsclient -nw $1"
alias pp="git pull && git push"
alias tl="tmux list-sessions"
alias tc="tmux new -s $1"
alias ta="tmux attach -d -t $1"
alias become_daemon="su - xelaiedaemon"

export EDITOR=emacsclient
export GOPATH=${HOME}
export PROJ_DIR=${HOME}/src/xelaie
export PATH=${HOME}/go_appengine:${HOME}/bin:${HOME}/go/bin:${PROJ_DIR}/scripts:.:${PATH}
export PYTHONPATH=${PROJ_DIR}/src/py:${PYTHONPATH}

# The next line updates PATH for the Google Cloud SDK.
source "${HOME}/google-cloud-sdk/path.bash.inc"

# The next line enables bash completion for gcloud.
source "${HOME}/google-cloud-sdk/completion.bash.inc"

export CLOUDSDK_PYTHON=python2
