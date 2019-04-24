package commander

import (
	"fmt"
	"io"
	"os"
	"strings"

	prompt "github.com/c-bata/go-prompt"
)

// Command represents an interactive command
type Command struct {
	// name of the command (REQUIRED)
	Cmd string
	// description of the command (OPTIONAL)
	Description string
	// callback which gets executed (OPTIONAL)
	RunCallback func(Commander, Command, []string) error
	// callback to provide command argument suggestions (OPTIONAL)
	ArgumentComplete func(...string) []prompt.Suggest
	// commander to support nesting interfaces (OPTIONAL)
	SubCommander Commander
	// list of allowed first arguments (OPTIONAL)
	Options    []string
	optionsMap map[string]struct{}
	// validator function to validate arguments (OPTIONAL)
	Validate func(args []string) bool
}

// Commander implements the cli parsing and handling
type Commander interface {
	// Processes commands in a loop until ctrl+d/ctrl+c is pressed. Optionally takes a list of values for the prefix string
	Run(prefixArgs ...string) error
	// add a new command to the current commander
	AddCommand(c Command) error
	// current prefix args (not threadsafe!)
	PrefixArgs() []string
	// set a custom writer for status messages (uses os.Stdout by default)
	SetWriter(io.Writer)
}

// New creates a new commander with the given prefix format template and prompt options
func New(prefixTemplate string, options ...prompt.Option) Commander {
	c := &commander{
		suggestions:    []prompt.Suggest{},
		commands:       commandMap{},
		prefixTemplate: prefixTemplate,
		prefixArgs:     []string{},
		writer:         os.Stdout,
	}
	options = append(options, prompt.OptionPrefix(fmt.Sprintf(prefixTemplate)), prompt.OptionLivePrefix(c.livePrefix))
	p := prompt.New(c.executor, c.completer, options...)
	c.prompt = p
	return c
}

type commandMap map[string]Command

type commander struct {
	prompt         *prompt.Prompt
	commands       commandMap
	suggestions    []prompt.Suggest
	prefixTemplate string
	prefixArgs     []string
	writer         io.Writer
}

func (c *commander) PrefixArgs() []string {
	return c.prefixArgs
}

func (c *commander) SetWriter(w io.Writer) {
	c.writer = w
}

func (c *commander) Run(prefixArgs ...string) error {
	c.prefixArgs = prefixArgs
	c.prompt.Run()
	return nil
}

func (c *commander) AddCommand(cmd Command) error {
	if cmd.Cmd == "" {
		return fmt.Errorf("Command name is missing")
	}
	if _, found := c.commands[cmd.Cmd]; found {
		return fmt.Errorf("Command %s already registered", cmd.Cmd)
	}
	if cmd.SubCommander != nil && cmd.RunCallback != nil {
		return fmt.Errorf("Using a sub commander and a run callback is not supported")
	}
	if cmd.Options != nil {
		cmd.optionsMap = make(map[string]struct{}, len(cmd.Options))
		for _, o := range cmd.Options {
			cmd.optionsMap[o] = struct{}{}
		}
	}
	c.commands[cmd.Cmd] = cmd
	c.suggestions = append(c.suggestions, prompt.Suggest{
		Text:        cmd.Cmd,
		Description: cmd.Description,
	})
	return nil
}

func (c *commander) livePrefix() (string, bool) {
	if len(c.prefixArgs) > 0 {
		prefixArgs := make([]interface{}, len(c.prefixArgs))
		for i, a := range c.prefixArgs {
			prefixArgs[i] = a
		}
		return fmt.Sprintf(c.prefixTemplate, prefixArgs...), true
	}
	return "", false
}

func (c *commander) executor(t string) {
	args := splitArguments(t)
	if len(args) == 0 {
		return
	}
	if cmd, ok := c.commands[args[0]]; ok {
		if !validArguments(cmd, args[1:]) {
			fmt.Fprintf(c.writer, "Invalid arguments %s\n", args[1:])
			return
		}
		if cmd.Options != nil {
			valid := len(args) > 1
			if valid {
				_, valid = cmd.optionsMap[args[1]]
			}
		}

		if cmd.RunCallback != nil {
			if err := cmd.RunCallback(c, cmd, args[1:]); err != nil {
				fmt.Fprintf(c.writer, "ERROR: %s\n", err)
			}
		} else if cmd.SubCommander != nil {
			cmd.SubCommander.Run(args[1:]...)
		} else {
			fmt.Fprintf(c.writer, "Command %s incomplete\n", args[0])
		}
	} else {
		fmt.Fprintf(c.writer, "Command %s not found\n", args[0])
	}
}

func (c *commander) completer(d prompt.Document) []prompt.Suggest {
	t := d.TextBeforeCursor()
	idx := strings.IndexByte(t, ' ')
	if idx != -1 {
		t = t[:idx]
	}
	suggestions := c.suggestions
	if cmd, ok := c.commands[t]; ok {
		if cmd.ArgumentComplete != nil {
			args := []string{}
			if idx != -1 {
				args = splitArguments(d.TextBeforeCursor()[idx:])
			}
			suggestions = cmd.ArgumentComplete(args...)
		} else {
			suggestions = []prompt.Suggest{}
		}
	}
	return prompt.FilterHasPrefix(suggestions, d.GetWordBeforeCursor(), true)
}

func splitArguments(line string) []string {
	tmp := strings.Split(line, " ")
	result := []string{}
	for _, arg := range tmp {
		if arg != "" {
			result = append(result, arg)
		}
	}
	return result
}

// SimpleSuggestions converts strings into prompt.Suggest structs
func SimpleSuggestions(text ...string) []prompt.Suggest {
	s := make([]prompt.Suggest, len(text))
	for i, t := range text {
		s[i].Text = t
	}
	return s
}

func validArguments(cmd Command, args []string) bool {
	if cmd.Validate != nil {
		return cmd.Validate(args)
	} else if cmd.Options != nil {
		valid := len(args) > 0
		if valid {
			_, valid = cmd.optionsMap[args[0]]
		}
		return valid
	}
	return true
}
