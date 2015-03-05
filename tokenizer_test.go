package compgen

import (
	"strings"
	"testing"
)

var bench = map[string][]Arg{
	//random example found on the web
	// or to help debugging a particular case
	`ab cd `:             []Arg{NewArg("ab", 0, 2), NewArg("cd", 3, 2)},
	"a\u2345b cd":        []Arg{NewArg("a\u2345b", 0, 3), NewArg("cd", 4, 2)},
	`ab cd`:              []Arg{NewArg("ab", 0, 2), NewArg("cd", 3, 2)},
	`'ab' 'cd'`:          []Arg{NewArg("ab", 0, 4), NewArg("cd", 5, 4)},
	`ab   cd`:            []Arg{NewArg("ab", 0, 2), NewArg("cd", 5, 2)},
	`sh -c 'echo 123'`:   []Arg{NewArg("sh", 0, 2), NewArg("-c", 3, 2), NewArg("echo 123", 6, 10)},
	`find -name '*.go'`:  []Arg{NewArg("find", 0, 4), NewArg("-name", 5, 5), NewArg("*.go", 11, 6)},
	`ls /go/pkg/l*/sy*/`: []Arg{NewArg("ls", 0, 2), NewArg("/go/pkg/l*/sy*/", 3, 15)},
	`echo $CWD`:          []Arg{NewArg("echo", 0, 4), NewArg("$CWD", 5, 4)},
	//`echo $(echo abc)`:   []Arg{NewArg("echo", 0), NewArg("$(echo abc)", 5)},
	`echo  xxx # yyy`:  []Arg{NewArg("echo", 0, 4), NewArg("xxx", 6, 3), NewArg("#", 10, 1), NewArg("yyy", 12, 3)},
	`find ... \;`:      []Arg{NewArg("find", 0, 4), NewArg("...", 5, 3), NewArg(";", 9, 2)},
	`echo "\\\"1 2 3"`: []Arg{NewArg("echo", 0, 4), NewArg(`\"1 2 3`, 5, 11)},

	//from http://www.gnu.org/software/bash/manual/bashref.html#Escape-Character
	//3.1.2.1 Escape Character
	//
	// A non-quoted backslash ‘\’ is the Bash escape character.
	//It preserves the literal value of the next character that follows, with
	//the exception of newline.
	//If a \newline pair appears, and the backslash itself is not quoted,
	//the \newline is treated as a line continuation (that is, it is removed from
	//the input stream and effectively ignored).
	//
	// testing: any char, space, and quotes
	`ab\cd abcd`: []Arg{NewArg("abcd", 0, 5), NewArg("abcd", 6, 4)},
	`a\ b cd`:    []Arg{NewArg("a b", 0, 4), NewArg("cd", 5, 2)},
	`a\"b cd`:    []Arg{NewArg(`a"b`, 0, 4), NewArg("cd", 5, 2)},
	`a\'b cd`:    []Arg{NewArg(`a'b`, 0, 4), NewArg("cd", 5, 2)},
	"a\\`b cd":   []Arg{NewArg("a`b", 0, 4), NewArg("cd", 5, 2)},

	//3.1.2.2 Single Quotes
	//
	//Enclosing characters in single quotes (‘'’) preserves the literal value of
	// each character within the quotes. A single quote may not occur between
	//single quotes, even when preceded by a backslash.
	//testing other special ones
	`ab 'abcd'`:  []Arg{NewArg("ab", 0, 2), NewArg("abcd", 3, 6)},
	`ab 'ab"cd'`: []Arg{NewArg("ab", 0, 2), NewArg(`ab"cd`, 3, 7)},
	`ab 'ab cd'`: []Arg{NewArg("ab", 0, 2), NewArg(`ab cd`, 3, 7)},
	"ab 'ab`cd'": []Arg{NewArg("ab", 0, 2), NewArg("ab`cd", 3, 7)},

	/*
		3.1.2.3 Double Quotes

		Enclosing characters in double quotes (‘"’) preserves the literal value of all characters
		 within the quotes, with the exception of

		 ‘$’, ‘`’, ‘\’, and, when history expansion is enabled, ‘!’.

		 The characters ‘$’ and ‘`’ retain their special meaning within double quotes (see Shell Expansions).
		 The backslash retains its special meaning only when followed by one of the following characters:
		 ‘$’, ‘`’, ‘"’, ‘\’, or newline.
		 Within double quotes, backslashes that are followed by one of these characters are removed.
		 Backslashes preceding characters without a special meaning are left unmodified.
		 A double quote may be quoted within double quotes by preceding it with a backslash.
		 If enabled, history expansion will be performed unless an ‘!’ appearing in double quotes
		 is escaped using a backslash. The backslash preceding the ‘!’ is not removed.

		The special parameters ‘*’ and ‘@’ have special meaning when in double quotes (see Shell Parameter Expansion).


	*/
	`ab "ab 'cd`:      []Arg{NewArg("ab", 0, 2), NewArg("ab 'cd", 3, 7)},    //all char values is preserved
	`ab "ab \$\"\\cd`: []Arg{NewArg("ab", 0, 2), NewArg(`ab $"\cd`, 3, 12)}, // in double quote escaping
	"ab \"ab \\`cd":   []Arg{NewArg("ab", 0, 2), NewArg("ab `cd", 3, 8)},    // in double quote escaping backtick(for readability)

	/*
		`cmd` $'string' ${var} and $(cmd) requires more than a lexer (it's a grammar that we need)


	*/

}

func TestSimpleLexer(t *testing.T) {
	chk(t, `'ab' 'cd'`)
}

func chk(t *testing.T, k string) {
	CheckLexer(t, k, bench[k])
}
func TestLexer(t *testing.T) {
	for k, v := range bench {
		CheckLexer(t, k, v)
	}
}

func CheckLexer(t *testing.T, k string, v []Arg) {
	args, err := Tokenize(strings.NewReader(k))
	if err != nil {
		panic(err)
	}
	t.Logf("Testing %q", k)
	la, lv := len(args), len(v)
	if la != lv {
		t.Errorf("invalid NArgs %v vs %v", la, lv)
	}

	for i := 0; i < Min(la, lv); i++ {
		a, av := args[i], v[i]
		if a.Offset != av.Offset {
			t.Errorf("%q invalid %q Arg[%v].Offset %v vs %v", k, a.Val, i, a.Offset, av.Offset)
		}
		if a.Length != av.Length {
			t.Errorf("%q invalid %q Arg[%v].Length %v vs %v", k, a.Val, i, a.Length, av.Length)
		}
		if a.Val != av.Val {
			t.Errorf("%q invalid %q Arg[%v].Val %v vs %v", k, a.Val, i, a.Val, av.Val)
		}
	}
}

func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
