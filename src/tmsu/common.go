package main

import (
    "fmt"
    "os"
)

func die(format string, a ...interface{}) {
    fmt.Fprintf(os.Stderr, format + "\n", a...)
    os.Exit(1)
}
