package compgen

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

//Comgen is a function to generate a single kind of values
type Compgen func(prefix string) []string

//Argsgen is the interface any object need to implement to to be able to deal with varargs.
type Argsgen interface {
	Compgen(args []string, inword bool) (comp []string, err error)
}

// Terminator is the basic object to deal with completion.
//
// Basically, you create one `NewTerminator` with a flag.FlagSet, you Configure it and run Terminate
//
// Terminate() will generate a list of suggestions for the current bash completion line
// and write them to the stdout.
//
// Terminate() can be called anytime, if will do nothing if not in completion mode.
//
// Caveat: bash_completion mode require a clean stdout (and stderr) so be careful to not output anything before calling Terminate.
//
//
// Terminator Configuration
//
// For each flag value (like `cmd -name toto<TAB>` ) Terminator need to generate a list of suggestion for this flag.
// By default, it will generate the default value as the only suggestion, you can override it by mapping a Compgen to the flag
//
//    Flag(name, compgen)
//
// When Terminator look for suggestion for an 'arg' ( like `cmd toto<TAB>`) it will try first
// the Argsgen if not nil, then a positional Compgen
//
//    Arg(position, compgen)
//    Argsgen(argsgen)
//
// Argsgen receive the full FlagSet.Args() (i.e. removed from the flags arguments)
//
// Position is zero-indexed argument position:
//
//    `cmd toto<TAB> titi`     0
//    `cmd toto <TAB> titi`    1
//
// If there is a Compgen mapped to the actual completion position, then it is used.
//
//
// Recursion
//
// In addition to this usage, Terminator also implements the Argsgen interface.
// It is then possible to use it recursively.
//
// One common usecase is to use it to implement subcommands.
// Each subcommand has it's own "terminator" configured, and you just need to implement
// a Argsgen to dispatch to the right subcommander.
type Terminator struct {
	fs        *flag.FlagSet
	keyvalgen map[string]Compgen // ability to set a Comgen for each key val
	arggen    map[int]Compgen    // positional Compgen
	argsgen   Argsgen            // the compgen for varargs
}

//NewTerminator creates a new Terminator
func NewTerminator(fs *flag.FlagSet) (t *Terminator) {
	t = new(Terminator)
	t.fs = fs
	return t
}

//Flag maps a Compgen to a given flag by name
func (t *Terminator) Flag(name string, gen Compgen) {
	if t.keyvalgen == nil {
		t.keyvalgen = make(map[string]Compgen)
	}
	t.keyvalgen[name] = gen
}

//Arg maps a Compgen to a positional argument
func (t *Terminator) Arg(pos int, gen Compgen) {
	if t.arggen == nil {
		t.arggen = make(map[int]Compgen)
	}
	t.arggen[pos] = gen
}

// Argsgen set the interface to be used to deal with varargs
func (t *Terminator) Argsgen(a Argsgen) {
	t.argsgen = a
}

//Terminate the current command line.
// if the executable is in completion mode, this methods tries to complete
// the current command line and *exit*
//
// If it is not in completion mode, then this methods simply returns
func (t *Terminator) Terminate() {

	if !IsCompletionMode() {
		return
	}
	args, inword, err := Args()
	if err != nil {
		os.Exit(-1)
	}
	pred, err := t.Compgen(args, inword)
	if err != nil {
		os.Exit(-1)
	}
	fmt.Println(strings.Join(pred, "\n"))
	os.Exit(0)
}

//Compgen is the method required by the Argsgen interface
func (t *Terminator) Compgen(args []string, inword bool) (comp []string, err error) {

	//log.Printf("terminator %v inword:%t", args, inword)
	// quick exit on non completion mode
	if !IsCompletionMode() {
		return
	}

	// hack the flagset to discard output and use a permissive parsing
	t.fs.Init("terminators", flag.ContinueOnError)
	t.fs.SetOutput(ioutil.Discard)

	// compute a few reused vars
	la := len(args)
	last := ""
	if la-1 < len(args) && la-1 >= 0 {
		last = args[la-1]
	}

	prefix := ""
	if inword {
		prefix = last
	}

	// find out the completion case we are in
	c := findCase(t.fs, args, inword)
	//log.Printf("completing %v", c)
	switch c {

	case CompErr:
		err = errors.New("Invalid Flags")
		return

	case CompFlagKey:
		return FlagNameGen(t.fs)(prefix), nil

	case CompFlagVal:
		// find out the key
		key := last
		if inword {
			//key is the one before if available
			key = ""
			if la-2 >= 0 {
				key = args[la-2]
			}
		}

		// clean up the leading dashes
		key = strings.TrimLeft(key, "-")

		// ok the key is ready

		//get the key compgen
		if gen, exists := t.keyvalgen[key]; exists {
			return gen(prefix), nil
		} else { // uses the default based one
			return FlagValueGen(t.fs, key)(prefix), nil
		}
		return

	case CompArgs:
		// there is no way to find out any compgen by default, I really need to rely on the one passed.
		if t.argsgen != nil {
			return t.argsgen.Compgen(t.fs.Args(), inword)
		}

		if len(t.arggen) > 0 { // there are some positional arguments
			// which is the current position?
			position := t.fs.NArg()
			if inword {
				position--
			}
			if gen, exists := t.arggen[position]; exists {
				return gen(prefix), nil
			}
		}
		return

	default:
		err = errors.New("Unknown case")
		return

	}
	return
}

const (
	CompErr = iota
	CompFlagKey
	CompFlagVal
	CompArgs
)

type CompCase int // one of the above
func (c CompCase) String() string {
	switch c {
	case CompErr:
		return "CompErr"
	case CompFlagKey:
		return "CompFlagKey"
	case CompFlagVal:
		return "CompFlagVal"
	case CompArgs:
		return "CompArgs"
	default:
		return "<unknown>"
	}
}

func findCase(fs *flag.FlagSet, args []string, inword bool) CompCase {
	//log.Printf("case for %v %v", args, inword)
	la := len(args)
	if la == 0 {
		return CompArgs
	}
	endIsKey := strings.HasPrefix(args[la-1], "-")

	err := fs.Parse(args[1:])

	if err != nil {
		//log.Printf("flag parse err %v", err)
		if endIsKey {
			if inword {
				return CompFlagKey
			} else {

				// it depends on the key in fact. if the key is double ( bool ) or double (string)
				//TODO handle the case of "bool": I need to lookup the flagset for args[la-1] and find out if it's a single or double flag
				// and also check that it's not a single flag syntax (kinda -name=toto ) in which cas the output should be comperr
				return CompFlagVal
			}
		}

		return CompErr
	}

	remaining := fs.Args()
	//log.Printf("rem=%v", remaining)
	//log.Printf("err=%v parsed=%v -> %v", err, fs.Parsed(), remaining)
	lr := len(remaining)

	switch {
	case lr == 0: // there is no args, all have been parsed by flag
		if inword {
			// there is no args but i was completing a word, this word is either a valid key or a flag
			if endIsKey {
				return CompFlagKey
			}
			return CompFlagVal
		}

		// not in word so this a first arg
		return CompArgs

	case lr == 1:
		if endIsKey {
			return CompFlagKey
		}
		return CompArgs

	case lr > 0:
		// there are not just only flags
		return CompArgs
	}

	return CompErr // unexpected outcome
}
