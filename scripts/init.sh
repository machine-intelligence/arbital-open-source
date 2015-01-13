# Shared bash setup for scripts.

# Fail if any command fails (returns != 0).
set -e
set -o pipefail

# cfg returns a value loaded from config.yaml.
function cfg() {
	if [ "$#" -ne 1 ]; then
		echo "Usage: $0 [value]" >&2
		return 1
	fi

	if ! /usr/bin/env shyaml 2> /dev/null; then
			echo "No shyaml installed. Try 'sudo pip install shyaml'." >&2
			return 3
	fi
	VAL=$(cat config.yaml | /usr/bin/env shyaml get-value $1)
	if [ -z "$VAL" ]; then
		echo "No value $1 in config.yaml. Buggy script? Or config.yaml has changed in the repo and you need to decrypt_config.sh again?" >&2
		return 4
	fi
	echo ${VAL}
	return 0
}

function check_config() {
		if [ ! -e "config.yaml" ]; then
				echo "No config.yaml present in root of repo. Did you decrypt it with decrypt_config.sh (if you have keys locally) or decrypt_config_remote.sh (if not)?" >&2
				exit 1
		fi
}

function check_git_hooks() {
   if [ ! -e ".git/hooks/pre-commit" ]; then
			 echo "Missing git precommit hooks file (.git/hooks/pre-commit). Please run 'symlink_git_hooks.sh'." >&2
			 exit 2
	 elif [ ! -e ".git/hooks/pre-push" ]; then
			 echo "Missing git prepush hooks file (.git/hooks/pre-push). Please run 'symlink_git_hooks.sh'." >&2
			 exit 3
	 fi
}

check_config
check_git_hooks
