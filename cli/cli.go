package cli

import "fmt"

type Command struct {
	Name        string
	Func        CommandFunc
	Usage       string
	Summary     string
	Description string
}

type CommandFunc func(c Context) error

type Context struct {
}

type Settings struct {
	Name        string
	Summary     string
	Description string
}

var (
	commands = []Command{}
	settings = Settings{}
)

func Init(s Settings) {
	settings = s
}

func Register(c Command) {
	commands = append(commands, c)
}

func Run(args []string) {
	for _, c := range commands {
		if c.Name == args[1] {
			c.Func(Context{})
		}
	}
}

func (c *Context) Help() string {
	s := fmt.Sprintf("%s: %s\n\n", settings.Name, settings.Summary)
	s += fmt.Sprintf("usage:\n  %s <command> [args...]\n\n", settings.Name)
	s += fmt.Sprintf("subcommands:\n")

	t := Table{}

	for _, c := range commands {
		t.AddRow(fmt.Sprintf("  %s", c.Name), c.Summary)
	}

	s += t.Render()

	return s
}
