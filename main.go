package compgen

import (
	"os"
	"strconv"
	"strings"
)

const (
	COMP_LINE  = "COMP_LINE"
	COMP_POINT = "COMP_POINT"
)

//IsCompletionMode returns true if the the execution has been made in a bash_completion environnement.
//
//More specifically:
//
// Returns true iif  "COMP_LINE" and "COMP_POINT" are set to a non empty value
func IsCompletionMode() bool {
	return os.Getenv(COMP_LINE) != "" && os.Getenv(COMP_POINT) != ""
}

//CompletionLine return the completion line
func CompletionLine() string { return os.Getenv(COMP_LINE) }

//CompletionPoint return the completion point variable
func CompletionPoint() int {
	pos, err := strconv.Atoi(os.Getenv(COMP_POINT))
	if err != nil {
		return -1
	}
	return pos
}

// Args read the completion line ($COMP_LINE) and completion point ($COMP_POINT) from env
// and returns
//
// args is the convertion-truncated comp_line as a []string
//
// inword is true if the tab is pressed within a word "toto<TAB>" or to<TAB>to but false when "toto <TAB>"
// the pseudo args are all args from the begining of the line up to the position. values after are NOT returned
//
// err is not nil if the comp_line cannot be tokenized
func Args() (args []string, inword bool, err error) {

	//read the <TAB> position
	pos, err := strconv.Atoi(os.Getenv(COMP_POINT))
	if err != nil {
		return
	}
	// args is splitted so parseArgs can be tested
	return parseArgs(os.Getenv(COMP_LINE), pos)
}

// parseArgs split the comp_line based on pos
//
// comp_line is the bash command line
//
// args is convertion-truncated comp_line as a []string
//
// inword is true if the tab is pressed within a word "toto<TAB>" or to<TAB>to but false when "toto <TAB>"
// the pseudo args are all args from the begining of the line up to the position. values after are NOT returned
//
// err is not nil if the comp_line cannot be tokenized
func parseArgs(comp_line string, pos int) (args []string, inword bool, err error) {

	// parse the command line upto the position
	r := strings.NewReader(comp_line[0:pos])

	aargs, err := Tokenize(r)
	if err != nil {
		return
	}
	current := position(aargs, pos)
	inword = current > 0

	args = make([]string, len(aargs))
	for i, a := range aargs {
		args[i] = a.Val
	}
	return
}

//Prefix compute the completion prefix and position
func Prefix(args []string, inword bool) (pos int, prefix string) {
	pos = len(args)
	if inword {
		pos--
		// prefix is always last
		if pos >= 0 {
			prefix = args[pos]
		}
	}
	return
}

//read the current position from args
func position(args []Arg, pos int) (current int) {
	current = -1 // if the pos is not inside any word current remains -1
	//last := -1   // last will increase unti
	for i, a := range args {
		if a.Offset+a.Length >= pos && a.Offset < pos {
			current = i
		}
	}
	return
}
