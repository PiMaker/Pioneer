package commands

import (
    "os/exec"
)

type BasicCommand struct {
    command string
    args []string
}

const BasicCommandTypeString = "basic"

func (cmd BasicCommand) Execute(parameter interface{}) string {
    retval, err := exec.Command(cmd.command, cmd.args...).CombinedOutput()
    
    if err != nil {
		return err.Error()
	}
    
    return "Success:\n\n" + string(retval)
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