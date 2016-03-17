package commands

import (
    "os/exec"
)

type BasicCommand struct {
    command string
    args []string
}

const BasicCommandTypeString = "basic"

func (cmd BasicCommand) Execute(parameter interface{}) bool {
    if err := exec.Command(cmd.command, cmd.args...).Run(); err != nil {
		return false
	}
    
    return true
}

func CreateBasicCommand(data JsonObject) *BasicCommand {
    cmd := BasicCommand{}
    cmd.command = data["command"].(string)
    a := data["args"].([]interface{})
    cmd.args = make([]string, 0)
    for _, arg := range a {
        cmd.args = append(cmd.args, arg.(string))
    }
    return &cmd
}