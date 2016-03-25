package commands

import (
    "time"
)

var CommandsAvailable map[int]DisplayCommand

type JsonObject map[string]interface{}

type DisplayCommand struct {
    ExecutableCommand Command
    ID int
    Name, Description, Type string
    
    IsBasic bool
    IsToggle bool
    
    AllowedUsers []string
}

type Command interface {
    Execute(parameter interface{}) string
}

func ParseCommands(config JsonObject) {
    CommandsAvailable = make(map[int]DisplayCommand)
    id := 0
    for _, cmd := range config["commands"].([]interface{}) {
        c := cmd.(map[string]interface {})
        
        command := DisplayCommand {
            Name: c["name"].(string),
            Description: c["description"].(string),
            Type: c["type"].(string) }
            
        data := c["data"].(map[string]interface {})
        switch command.Type {
        case BasicCommandTypeString:
            command.ExecutableCommand = CreateBasicCommand(data)
            command.IsBasic = true
        case ToggleCommandTypeString:
            command.ExecutableCommand = CreateToggleCommand(data)
            command.IsToggle = true
        }
        
        command.AllowedUsers = make([]string, 0)
        for _, username := range c["users"].([]interface{}) {
            command.AllowedUsers = append(command.AllowedUsers, username.(string))
        }
        
        if periodic, ok := c["periodic_exec"]; ok {
            RegisterCommandForPeriodicExecution(command, periodic.(int))
        }
        
        command.ID = id
        id++
        
        CommandsAvailable[command.ID] = command
    }
}

func RegisterCommandForPeriodicExecution(cmd DisplayCommand, secs int) {
    ticker := time.NewTicker(time.Duration(secs) * time.Second)
    go func() {
    for {
       <-ticker.C
       cmd.ExecutableCommand.Execute("on")
    }
 }()
}