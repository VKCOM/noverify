package linter

var superGlobals = map[string]struct{}{
	"GLOBALS":  {},
	"_SERVER":  {},
	"_GET":     {},
	"_POST":    {},
	"_REQUEST": {},
	"_COOKIE":  {},
	"_FILES":   {},
	"_SESSION": {},
	"_ENV":     {},
}

func isSuperGlobal(name string) bool {
	_, ok := superGlobals[name]
	return ok
}
