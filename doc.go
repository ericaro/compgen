// compgen package contains helper functions to write bash completion to command.
//like
//
//      $ git comm<TAB>
//      $ git commit
//
//
// Even though it can be any command, it makes more sense if both program share a bit of code. It is even possible to integrate both programs into a single one: the completion program and the execution program are just the same.
//
// If you use any CLI framework there are chances that you can "introspect" the CLI struct (options, sub commands etc.) and use it to build the completion program.
//
// To register a program that is self completing you just need to:
//
//    complete -C cmd cmd
//
// Or copy the above statement in a file into  `/etc/bash_completion.d/`
//
//
// Usage
//
// Create a Terminator object, it's the basic object to runn completion on command.
//
// Configure the terminator by associating Compgen to flags or positional arguments.
//
// Terminate the command line, by printing to stdout the list of propositions.
//
//
//
//
//
//
package compgen
