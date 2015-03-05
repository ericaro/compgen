package compgen

import (
	"strings"
	"testing"
)

func TestParseArgs(t *testing.T) {
	testArgs(t, "tester tototata", 11, []string{"tester", "toto"}, true)
	testArgs(t, "tester toto tata", 12, []string{"tester", "toto"}, false)

}
func testArgs(t *testing.T, line string, pos int, xargs []string, xinwords bool) {
	args, inwords, err := parseArgs(line, pos)
	if err != nil {
		panic(err)
	}

	if inwords != xinwords || !EqStrings(args, xargs) {
		t.Errorf("Invalid Args parsing (%v,%v) vs (%v,%v)", args, inwords, xargs, xinwords)
	}
}

func EqStrings(v, x []string) bool {
	if len(v) != len(x) {
		return false
	}
	for i := range x {
		if v[i] != x[i] {
			return false
		}
	}
	return true
}

func TestArgs(t *testing.T) {
	// for a single line (example below) we are going to test every position for to detect inwords
	m := "ab cde f gh "
	args, _ := Tokenize(strings.NewReader(m))

	for i, a := range args {
		t.Logf("%v %v %v-> %v", i, a.Val, a.Offset, a.Length)
	}

	//for each position of cursor the expected current and last
	x_current := []int{
		-1, //|ab_cde_f_gh_
		0,  //a|b_cde_f_gh_
		0,  //b|_cde_f_gh_
		-1, //_|cde_f_gh_
		1,  //c|de_f_gh_
		1,  //d|e_f_gh_
		1,  //e|_f_gh_
		-1, //_|f_gh_
		2,  //f|_gh_
		-1, //_|gh_
		3,  //g|h_
		3,  //h|_
		-1, //_|

	}

	for i := 0; i < len(m); i++ { //there is not difference between runes and byte in this example
		current := position(args, i)
		if current != x_current[i] {
			t.Errorf("Invalid current '%s|%s' %v vs %v", m[0:i], m[i:], current, x_current[i])
		}
	}
}
