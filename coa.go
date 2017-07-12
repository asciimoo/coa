package main

import (
	"fmt"
	"os"
	"runtime"

	"github.com/asciimoo/coa/config"
	"github.com/asciimoo/coa/notification"
	"github.com/asciimoo/coa/server"

	"github.com/mitchellh/go-homedir"
	"github.com/jawher/mow.cli"
)

var configFolder = ""

func init() {
	switch runtime.GOOS {
	case "linux":
		// Use the XDG_CONFIG_HOME variable if it is set, otherwise
		// $HOME/.config/wuzz/config.toml
		xdgConfigHome := os.Getenv("XDG_CONFIG_HOME")
		if xdgConfigHome != "" {
			configFolder = xdgConfigHome
		} else {
			configFolder, _ = homedir.Expand("~/.config/coa/")
		}

	default:
		// On other platforms we just use $HOME/.wuzz
		configFolder, _ = homedir.Expand("~/.coa/")
	}
}

func main() {
	var c *config.Config
	app := cli.App("coa", "Coding assistant")
	cf := app.StringOpt("c config", configFolder, "Folder of Coa configuration files")

	app.Before = func() {
		var err error
		c, err = config.Load(*cf)
		if err != nil {
			fmt.Println("Configuration error!", err.Error())
			os.Exit(1)
		}
		if c.Notifiers == nil {
			fmt.Println("Configuration error! Please specify at least one notification backend in your Coa settings")
			os.Exit(4)
		}
		if err := notification.Initialize(c.Notifiers); err != nil {
			fmt.Println("Configuration error! Cannot initialize notification backend:", err.Error())
			os.Exit(5)
		}
	}

	app.Command("add", "Add new project", func(cmd *cli.Cmd) {
		projectFile := cmd.StringArg("FILE", "", "Project settings file")

		cmd.Action = func() {
			if *projectFile == "" {
				fmt.Println("Error! Missing project file")
				os.Exit(2)
			}
			err := server.Call(c.ServerAddress + "/api/add", map[string]string{"path": *projectFile})
			if err != nil {
				fmt.Println("Error!", err.Error())
				os.Exit(3)
			}

			fmt.Println("Project added")
		}

	})

	app.Command("listen", "Start Coa server", func(cmd *cli.Cmd) {
		cmd.Action = func() {
			server.Listen(c)
		}
	})

	app.Command("reload", "Reload Coa server", func(cmd *cli.Cmd) {
		cmd.Action = func() {
			err := server.Call(c.ServerAddress + "/api/reload", nil)
			if err != nil {
				fmt.Println("Server reload error!", err.Error())
				os.Exit(6)
			}

			fmt.Println("Server reloaded")
		}
	})

	app.Action = func() {
		app.PrintHelp()
	}
	app.Run(os.Args)
}
