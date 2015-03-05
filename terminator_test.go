package compgen

import (
	"flag"
	"io/ioutil"

	"testing"
)

func TestFindCase(t *testing.T) {

	CheckCase(t, []string{"cmd", "-"}, true, CompFlagKey)
	CheckCase(t, []string{"cmd", "-name"}, false, CompFlagVal)
	CheckCase(t, []string{"cmd", "-name", "toto"}, true, CompFlagVal)
	CheckCase(t, []string{"cmd", "-yes", "toto"}, true, CompArgs)
	CheckCase(t, []string{"cmd", "-yes", "toto", "tata"}, true, CompArgs)
	CheckCase(t, []string{"cmd", "-no", "toto", "tata"}, true, CompErr)

}

func CheckCase(t *testing.T, args []string, inw bool, x CompCase) {
	// always chec with the same flagset
	fs := flag.NewFlagSet("t", flag.ContinueOnError)
	fs.String("name", "name", "to set a name")
	fs.Bool("yes", false, "to say yes")
	fs.SetOutput(ioutil.Discard)

	c := findCase(fs, args, inw)
	if c != x {
		t.Errorf("Invalid case %v %v: (%v vs %v)", args, inw, x, c)
	}

}
