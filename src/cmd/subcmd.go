package cmd

import (
	"fmt"

	"github.com/i582/cfmt"
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
	res += cfmt.Sprintf("  {{%s}}::green                %s\n", s.Name, s.Description)
	res += fmt.Sprintln("    Recipes:")
	for _, ex := range s.Examples {
		res += cfmt.Sprintf("      {{$}}::gray noverify %s %s      %s\n", s.Name, ex.Line, ex.Description)
	}
	res += "\n"
	return res
}
