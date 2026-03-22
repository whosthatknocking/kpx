package cmd

import "io"

const bashCompletionCompat = `
# Minimal compatibility shim for shells where bash-completion is not loaded.
if ! declare -F _get_comp_words_by_ref >/dev/null 2>&1; then
_get_comp_words_by_ref()
{
    local exclude=""
    local curvar="" prevvar="" wordsvar="" cwordvar=""

    while [[ $# -gt 0 ]]; do
        case "$1" in
        -n)
            shift
            exclude="$1"
            ;;
        *)
            if [[ -z $curvar ]]; then
                curvar="$1"
            elif [[ -z $prevvar ]]; then
                prevvar="$1"
            elif [[ -z $wordsvar ]]; then
                wordsvar="$1"
            else
                cwordvar="$1"
            fi
            ;;
        esac
        shift
    done

    local words=("${COMP_WORDS[@]}")
    local cword=$COMP_CWORD
    local cur="${words[cword]}"
    local prev=""
    if (( cword > 0 )); then
        prev="${words[cword-1]}"
    fi

    if [[ -n $exclude ]]; then
        cur="${cur//[$exclude]/}"
        prev="${prev//[$exclude]/}"
    fi

    printf -v "$curvar" '%s' "$cur"
    printf -v "$prevvar" '%s' "$prev"
    eval "$wordsvar=(\"\${words[@]}\")"
    printf -v "$cwordvar" '%s' "$cword"
}
fi
`

// GenerateBashCompletion writes the shipped Bash completion script.
func GenerateBashCompletion(w io.Writer) error {
	if _, err := io.WriteString(w, bashCompletionCompat); err != nil {
		return err
	}
	return rootCmd.GenBashCompletionV2(w, true)
}
