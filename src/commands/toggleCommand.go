package commands

import (
    "os/exec"
)

type ToggleCommand struct {
    commandOn string
    commandOff string
    argsOn []string
    argsOff []string
}

const ToggleCommandTypeString = "toggle"

func (cmd ToggleCommand) Execute(parameter interface{}) string {
    strprm := parameter.(string)
    command := ""
    args := make([]string, 0)
    if strprm == "on" {
        command = cmd.commandOn
        args = cmd.argsOn
    } else if strprm == "off" {
        command = cmd.commandOff
        args = cmd.argsOff
    } else {
        return "Neither on nor off was specified."
    }
    
    retval, err := exec.Command(command, args...).CombinedOutput()
    
    if err != nil {
		return err.Error()
	}
    
    return "Success:\n\n" + string(retval)
}

func CreateToggleCommand(data JsonObject) *ToggleCommand {
    cmd := ToggleCommand{}
    cmd.commandOn = data["command_on"].(string)
    a := data["args_on"].([]interface{})
    cmd.argsOn = make([]string, 0)
    for _, arg := range a {
        cmd.argsOn = append(cmd.argsOn, arg.(string))
    }
    cmd.commandOff = data["command_off"].(string)
    a = data["args_off"].([]interface{})
    cmd.argsOff = make([]string, 0)
    for _, arg := range a {
        cmd.argsOff = append(cmd.argsOff, arg.(string))
    }
    return &cmd
}