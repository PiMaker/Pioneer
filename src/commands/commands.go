package commands

import (
)

var CommandsAvailable []DisplayCommand

type JsonObject map[string]interface{}

type DisplayCommand struct {
    ExecutableCommand Command
    ID int
    Name, Description, Type string
}

type Command interface {
    Execute(parameter interface{}) bool
}

func ParseCommands(config JsonObject) {
    CommandsAvailable = make([]DisplayCommand, 0)
    id := 0
    for _, cmd := range config["commands"].([]interface{}) {
        c := cmd.(map[string]interface {})
        command := DisplayCommand {
            Name: c["name"].(string),
            Description: c["description"].(string),
            Type: c["type"].(string) }
        switch command.Type {
        case BasicCommandTypeString:
            command.ExecutableCommand = CreateBasicCommand(c["data"].(map[string]interface {}))
        }
        command.ID = id
        id++
        CommandsAvailable = append(CommandsAvailable, command)
    }
}