package compgen

import (
	"flag"
	"fmt"
	"os/exec"
	"strings"
)

/*
this file contains a few useful Compgens
*/

//ValueGen returns a Compgen that filters out 'values'
func ValueGen(values []string) Compgen {
	return func(prefix string) (predict []string) {
		//log.Printf("value gen %v", prefix)
		f := make([]string, 0, len(values))
		for _, v := range values {
			if strings.HasPrefix(v, prefix) {
				f = append(f, v)
			}
		}
		return f
	}
}

//FlagValueGen returns a Compgen that return the default value for the given key in the flagset
//
//This is pretty useless 'as is' but it's the default compgen associated with each key
func FlagValueGen(fs *flag.FlagSet, key string) Compgen {
	return func(prefix string) (predict []string) {
		// build the result
		fs.VisitAll(func(f *flag.Flag) {
			if f.Name == key {
				if strings.HasPrefix(f.DefValue, prefix) {
					predict = []string{f.DefValue}
				}
				return
			}
		})
		return predict
	}
}

// FlagNameGen returns a Compgen that generate a list of unused flag names
//
// If the flag set has been parsed and if some values have been set, this comgen return only the not set ones.
func FlagNameGen(fs *flag.FlagSet) Compgen {
	return func(prefix string) (predict []string) {

		// we need to extract the name part of the prefix (to use in compare)
		name := strings.TrimLeft(prefix, "-")
		// and get the dash part (to reuse it)
		dash := strings.TrimSuffix(prefix, name)
		if dash == "" { // empty prefix lead to empty dash, this is unfortunate
			dash = "-"
		}
		predict = make([]string, 0, 10)

		//buid a set of already set flags
		actual := make(map[string]interface{})
		fs.Visit(func(f *flag.Flag) {
			actual[f.Name] = nil
		})
		// build the result
		fs.VisitAll(func(f *flag.Flag) {
			if _, exists := actual[f.Name]; !exists && strings.HasPrefix(f.Name, name) {
				predict = append(predict, dash+f.Name)
			}
		})

		return predict
	}
}

//CompgenCmd execute the bash builtin 'compgen' command to return values
//
//The action may be one of the following to generate a list of possible completions:
//
//    alias     Alias names.
//    arrayvar  Array variable names.
//    binding   Readline key binding names.
//    builtin   Names of shell builtin commands. May also be specified as -b.
//    command   Command names. May also be specified as -c.
//    directory Directory names. May also be specified as -d.
//    disabled  Names of disabled shell builtins.
//    enabled   Names of enabled shell builtins.
//    export    Names of exported shell variables. May also be specified as -e.
//    file      File names. May also be specified as -f.
//    function  Names of shell functions.
//    group     Group names. May also be specified as -g.
//    helptopic Help topics as accepted by the help builtin.
//    hostname  Hostnames, as taken from the file specified by the HOSTFILE shell variable.
//    job       Job names, if job control is active. May also be specified as -j.
//    keyword   Shell reserved words. May also be specified as -k.
//    running   Names of running jobs, if job control is active.
//    service   Service names. May also be specified as -s.
//    setopt    Valid arguments for the -o option to the set builtin.
//    shopt     Shell option names as accepted by the shopt builtin.
//    signal    Signal names.
//    stopped   Names of stopped jobs, if job control is active.
//    user      User names. May also be specified as -u.
//    variable  Names of all shell variables. May also be specified as -v.
//
func CompgenCmd(action string) Compgen {

	return func(prefix string) (predict []string) {
		var cmd *exec.Cmd
		cmd = exec.Command("bash", "-i", "-c", fmt.Sprintf(`compgen -A %s %s`, action, prefix))

		out, err := cmd.CombinedOutput()
		//log.Printf("%s %v", string(out), err)
		if err != nil { //err are ignored
			return
		}
		return strings.Split(string(out), "\n")
	}
}
