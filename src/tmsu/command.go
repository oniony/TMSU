package main

type Command interface {
    Name() string
    Description() string
    Exec(args []string) error
}
