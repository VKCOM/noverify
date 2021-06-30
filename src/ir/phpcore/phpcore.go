package phpcore

import (
	"github.com/VKCOM/noverify/src/ir"
)

func ResolveAlias(function ir.Node) ir.Node {
	nm, ok := function.(*ir.Name)
	if !ok {
		return function
	}
	alias, ok := funcAliases[nm.Value]
	if ok {
		return alias
	}
	return function
}

var funcAliases = map[string]*ir.Name{
	// See https://www.php.net/manual/ru/aliases.php

	`doubleval`: {Value: `floatval`},

	`ini_alter`:    {Value: `ini_set`},
	`is_integer`:   {Value: `is_int`},
	`is_long`:      {Value: `is_int`},
	`is_real`:      {Value: `is_float`},
	`is_double`:    {Value: `is_float`},
	`is_writeable`: {Value: `is_writable`},

	`join`:       {Value: `implode`},
	`chop`:       {Value: `rtrim`},
	`strchr`:     {Value: `strstr`},
	`pos`:        {Value: `current`},
	`key_exists`: {Value: `array_key_exists`},
	`sizeof`:     {Value: `count`},

	`close`:                {Value: `closedir`},
	`fputs`:                {Value: `fwrite`},
	`magic_quotes_runtime`: {Value: `set_magic_quotes_runtime`},
	`show_source`:          {Value: `highlight_file`},
}
