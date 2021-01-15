package cmd

import (
	"fmt"
)

type SubCommand struct {
	Name        string
	Main        func(*MainConfig) (int, error)
	Description string
	Examples    []SubCommandExample
}

type SubCommandExample struct {
	Description string
	Line        string
}

func (s *SubCommand) String() string {
	var res string
	res += fmt.Sprintf("\n\tnoverify %s\n", s.Name)
	res += fmt.Sprintf("\tDescription: %s.\n", s.Description)
	for _, ex := range s.Examples {
		res += fmt.Sprintf("\t%s:\n", ex.Description)
		res += fmt.Sprintf("\t\t$ noverify %s %s\n", s.Name, ex.Line)
	}
	return res
}
