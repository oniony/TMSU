package main

import (
          "errors"
          "exec"
          "os"
       )

type UnmountCommand struct {}

func (this UnmountCommand) Name() string {
    return "unmount"
}

func (this UnmountCommand) Description() string {
    return "unmounts the virtual file-system"
}

func (this UnmountCommand) Exec(args []string) error {
    if len(args) < 1 { errors.New("Path to unmount not speciified.") }

    path := args[0]

    _, error := os.Stat(path)
    if error != nil { return error }

    fusermountPath, error := exec.LookPath("fusermount")
    if error != nil { return error }

    process, error := os.StartProcess(fusermountPath, []string{fusermountPath, "-u", path}, &os.ProcAttr{})
    if error != nil { return error }

    message, error := os.Wait(process.Pid, 0)
    if error != nil { return error }
    if message.ExitStatus() != 0 { return errors.New("Could not unmount.") }

    return nil
}
