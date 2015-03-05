package compgen

import (
	"bytes"
	"errors"
	"fmt"
	"unicode"

	"io"
)

var (
	ErrShellParameterExpansion = errors.New("Cannot Lex with Shell Parameter Expansion")
)

//Arg is the result of parsing a full command line
type Arg struct {
	Val    string
	Offset int // position in the original
	Length int // length occupied in the original
}

func NewArg(val string, offset, length int) Arg {
	return Arg{Val: val, Offset: offset, Length: length}
}

func printf(format string, a ...interface{}) {
	//fmt.Printf(format, a...)
}

func Tokenize(reader io.RuneScanner) (args []Arg, err error) {

	args = make([]Arg, 0, 10)
	state := newstate() // create all the "state" into a struct (zero is the initial)

	for {
		//read one and deal with errors
		r, _, err := reader.ReadRune()
		printf("rune %4s ", fmt.Sprintf("%q", string(r)))
		if err != nil && err != io.EOF {
			return args, err
		}

		// ok I've got this new char, deal with all the cases
		// depends on the rune, and the internal state
		switch {
		// cases have to be sorted by priority
		// case represent the top state selector

		//TOP MODE: ERROR
		case err == io.EOF: // end of an arg. need to cut the arg
			printf("%12s : ", "Is EOF")
			//fire the tokens
			if state.initpos > 0 {

				a := state.Pull()
				printf("Pull %v\n", a)
				args = append(args, a)
			}
			return args, nil

		//TOP MODE: SINGLE QUOTE
		case state.InDollar:
			switch {
			case r == '(' || r == '{':
				//that's bad, this is command substitution, I just can't really deal with that
				return args, ErrShellParameterExpansion
			default:
				//any other case are fine
				state.InDollar = false // moving out of this state
				state.Move(-1)         //move back (caveat: we know that the position will always be moved )
				reader.UnreadRune()    // this character does not belong to us rewing
			}

		//TOP MODE: SINGLE QUOTE
		case state.InSingleQuote:
			switch {

			case r == '\'': //end of single quote start
				printf("%15s : %s\n", "SingleQuote End", "Cons")
				state.InSingleQuote = false

			default:
				printf("%15s : %s\n", "SingleQuote Cont", "Push")
				state.Push(r)
			}

		//TOP MODE: DOUBLE QUOTE
		case state.InDoubleQuote:
			switch {

			case state.InDoubleQuoteEscape:
				state.InDoubleQuoteEscape = false
				switch {
				// I've found a \ previsouly
				case r == '$' || r == '`' || r == '"' || r == '\\':
					//this is a valid escape
					printf("%15s : %s\n", "DoubleQuoteEsc OK", "Push")
					state.Push(r)
				default:
					// all other chara are pushed as along witht he backslash
					printf("%15s : %s\n", "DoubleQuoteEsc KO", "Push")
					state.Push('\\')
					state.Push(r)
				}

			case r == '\\': //entering double quote escape mode
				printf("%15s : %s\n", "DoubleQuoteEsc Start", "Cons")
				state.InDoubleQuoteEscape = true //simply consume it

			case r == '$': // Dollar mode ? I push it anyway but remember it
				printf("%15s : %s\n", "DoubleQuote $", "Push")
				state.InDollar = true
				state.Push(r)
			case r == '`':
				//that's bad, this is command substitution, I just can't really deal with that
				return args, ErrShellParameterExpansion

			case r == '"': //exit mode
				printf("%15s : %s\n", "DoubleQuote End", "Cons")
				state.InDoubleQuote = false //simply consume it

			//todo submode $ and ` (which is the same I guess)

			default:
				printf("%15s : %s\n", "DoubleQuote Cont", "Push")
				state.Push(r)
			}

		//TOP MODE: ESCAPE CHAR
		case state.InEscape:
			state.InEscape = false
			state.InSeparator = false
			printf("%15s : %s\n", "Esc", "Push")
			//psuh the rune but I need to push it like if it has started one char before (the \)
			state.Push(r)

		//TOP MODE: SEPARATOR
		case state.InSeparator:
			switch {
			// Separator
			case unicode.IsSpace(r):
				printf("%15s : %s\n", "Sep Cont", "Consume")
				//just consume it

			case !unicode.IsSpace(r): // end of separator
				printf("%15s : ", "Sep End")
				state.InSeparator = false
				state.initpos = state.pos
				state.Move(-1)      //move back (caveat: we know that the position will always be moved )
				reader.UnreadRune() // this character does not belong to us rewing

			}

		//TOP MODE DEFAULT (means in word)
		default:
			switch {

			// entering modes

			case r == '`':
				//that's bad, this is command substitution, I just can't really deal with that
				return args, ErrShellParameterExpansion

			case r == '\'':
				printf("%15s : %s\n", "SingleQuote Start", "Cons")
				state.InSingleQuote = true

			case r == '"':
				printf("%15s : %s\n", "DoubleQuote Start", "Cons")
				state.InDoubleQuote = true

			case r == '\\':
				printf("%15s : %s\n", "Escape Start", "Cons")
				state.InEscape = true //simply consume it

			case r == '$': // Dollar mode ? I push it anyway but remember it
				printf("%15s : %s\n", "DoubleQuote $", "Push")
				state.InDollar = true
				state.Push(r)

			case unicode.IsSpace(r):
				printf("%15s : ", "Sep Start")
				//fire the tokens
				a := state.Pull()
				printf("Pull %v\n", a)
				args = append(args, a)
				state.InSeparator = true

			default:
				printf("%15s : %s\n", "Default", "Push")
				state.Push(r)
			}
		}
		//alway move the position
		state.Move(1)

	}
}

type lexstate struct {
	buf *bytes.Buffer
	//initpos = pos when an arg has started
	//pos the current pos
	initpos, pos                 int
	InSingleQuote, InDoubleQuote bool
	InSeparator                  bool
	InDoubleQuoteEscape          bool
	InEscape, InDollar           bool
}

func newstate() lexstate {
	return lexstate{
		buf: new(bytes.Buffer),
	}
}

//Pull get an arg out of the lexstate and reset the state
func (l *lexstate) Move(size int) {
	l.pos += size
}

func (l *lexstate) Push(r rune) {
	if l.initpos < 0 {
		l.initpos = l.pos
	}
	l.buf.WriteRune(r)
}
func (l *lexstate) Pull() Arg {
	length := l.pos - l.initpos
	a := NewArg(l.buf.String(), l.initpos, length)
	l.buf.Reset()
	l.initpos = -1
	return a
}
