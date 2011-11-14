package main

type Command interface {
	Name() string
	Summary() string
	Help() string
	Exec(args []string) error
}
