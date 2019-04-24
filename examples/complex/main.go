package main

import (
	"fmt"
	"strings"

	commander "github.com/bluec0re/go-commander"
	prompt "github.com/c-bata/go-prompt"
)

func addBuild(c commander.Commander, cb func(string, string) error) {
	c.AddCommand(commander.Command{
		Cmd: "build",
		ArgumentComplete: func(args ...string) []prompt.Suggest {
			if len(args) == 0 {
				return commander.SimpleSuggestions("windows", "linux")
			}
			switch args[0] {
			case "linux":
				return commander.SimpleSuggestions("x86", "x64", "arm")
			case "windows":
				return commander.SimpleSuggestions("x86", "x64")
			default:
				return []prompt.Suggest{}
			}
		},
		RunCallback: func(cr commander.Commander, c commander.Command, args []string) error {
			env := "default"
			if len(cr.PrefixArgs()) > 0 {
				env = cr.PrefixArgs()[0]
			}
			return cb(env, strings.Join(args, "-"))
		},
		Validate: func(args []string) bool {
			if len(args) < 2 {
				return false
			}
			switch args[0] {
			case "linux":
				switch args[1] {
				case "x86", "x64", "arm":
					return true
				default:
					return false
				}
			case "windows":
				switch args[1] {
				case "x86", "x64":
					return true
				default:
					return false
				}
			default:
				return false
			}
		},
	})
}

func build(env string, target string) error {
	fmt.Printf("Building for %s in %s\n", target, env)
	if env == "prod" {
		return fmt.Errorf("Compile error: %s", "out of memory")
	}
	return nil
}

func main() {
	subcommander := commander.New("%s>")
	addBuild(subcommander, build)
	c := commander.New(">>> ")
	c.AddCommand(commander.Command{
		Cmd: "env",
		ArgumentComplete: func(...string) []prompt.Suggest {
			return []prompt.Suggest{
				prompt.Suggest{
					Text: "prod",
				},
				prompt.Suggest{
					Text: "dev",
				},
			}
		},
		SubCommander: subcommander,
		Options:      []string{"prod", "dev"},
	})
	addBuild(c, build)
	c.Run()
}
