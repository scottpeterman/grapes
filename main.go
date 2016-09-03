package main

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"sync"
)

type (
	grape struct {
		input  input
		ssh    grapeSSH
		config config
	}
	serverOutput struct {
		Server server             `yaml:"server"`
		Output []*grapeCommandStd `yaml:"stds"`
	}
)

const (
	version = "0.1.4"
	welcome = `
//      ____ __________ _____  ___  _____
//     / __  / ___/ __  / __ \/ _ \/ ___/
//    / /_/ / /  / /_/ / /_/ /  __(__  )
//    \__, /_/   \__,_/ .___/\___/____/
//   /____/          /_/ v %s // Yaron Sumel [yaronsu@gmail.com]
//
`
)

var wg sync.WaitGroup

func main() {
	fmt.Printf(welcome, version)
	newGrape().runApp()
}

func newGrape() *grape {
	grape := grape{}
	//parse flags and validate it
	grape.input.parse()
	grape.input.validate()

	grape.ssh.setKey(grape.input.keyPath)
	grape.config.set(grape.input.configPath)

	return &grape
}

func (app *grape) runApp() {

	servers := app.config.getServersFromConfig(app.input.serverGroup)
	app.input.verifyAction(servers)

	for _, server := range servers {
		if app.input.asyncFlag {
			wg.Add(1)
			go app.asyncRunCommand(server, &wg)
		} else {
			app.runCommandsOnServer(server)
		}
	}
	if app.input.asyncFlag {
		wg.Wait()
	}
}

func (app *grape) asyncRunCommand(server server, wg *sync.WaitGroup) {
	app.runCommandsOnServer(server)
	wg.Done()
}

func (app *grape) runCommandsOnServer(server server) {

	commands := app.config.getCommandsFromConfig(app.input.commandName)
	client := app.ssh.newClient(server)

	so := serverOutput{
		Server: server,
	}

	for _, command := range commands {
		so.Output = append(so.Output, client.exec(command))
	}

	//done with all commands for this server
	so.print()
}

func (so *serverOutput) print() {
	out, err := yaml.Marshal(so)
	if err != nil {
		fatal("something went wrong with the output")
	}
	fmt.Println(string(out))
}
