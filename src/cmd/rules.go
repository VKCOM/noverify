package cmd

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/VKCOM/noverify/src/cmd/embeddedrules"
	"github.com/VKCOM/noverify/src/linter"
	"github.com/VKCOM/noverify/src/rules"
)

func InitEmbeddedRules(config *linter.Config, p *rules.Parser, filter func(r rules.Rule) bool) ([]*rules.Set, error) {
	ruleSets, err := parseEmbeddedRules(p)
	if err != nil {
		return nil, err
	}

	for _, rset := range ruleSets {
		appendRuleSet(config, rset, filter)
	}
	return ruleSets, nil
}

func parseRules() ([]*rules.Set, error) {
	p := rules.NewParser()

	ruleSets, err := parseEmbeddedRules(p)
	if err != nil {
		return nil, fmt.Errorf("embedded rules: %v", err)
	}

	rulesFlag, ok := findRulesFlag()
	if ok && rulesFlag != "" {
		for _, filename := range strings.Split(rulesFlag, ",") {
			data, err := ioutil.ReadFile(filename)
			if err != nil {
				return nil, err
			}
			rset, err := p.Parse(filename, bytes.NewReader(data))
			if err != nil {
				return nil, err
			}
			ruleSets = append(ruleSets, rset)
		}
	}

	return ruleSets, nil
}

func findRulesFlag() (string, bool) {
	// Prefix can be "-" or "--".
	// Value can be "=" or " " separated.
	// If value is " " separated, then it's located in the next argument.
	for i, arg := range os.Args {
		arg = strings.TrimLeft(arg, "-")
		switch {
		case strings.HasPrefix(arg, "rules="):
			parts := strings.Split(arg, "=")
			return parts[1], true
		case arg == "rules":
			if i+1 < len(os.Args) {
				return os.Args[i+1], true
			}
		}
	}
	return "", false
}

func parseEmbeddedRules(p *rules.Parser) ([]*rules.Set, error) {
	var ruleSets []*rules.Set
	for _, filename := range embeddedrules.AssetNames() {
		data, err := embeddedrules.Asset(filename)
		if err != nil {
			return nil, err
		}
		rset, err := p.Parse(filename, bytes.NewReader(data))
		if err != nil {
			return nil, err
		}
		rset.Builtin = true
		ruleSets = append(ruleSets, rset)
	}
	return ruleSets, nil
}

func appendRuleSet(config *linter.Config, rset *rules.Set, filter func(r rules.Rule) bool) {
	appendRules := func(dst, src *rules.ScopedSet) {
		for i, list := range &src.RulesByKind {
			for _, r := range list {
				if !filter(r) {
					continue
				}
				dst.RulesByKind[i] = append(dst.RulesByKind[i], r)
			}
		}
	}
	appendRules(config.Rules.Any, rset.Any)
	appendRules(config.Rules.Root, rset.Root)
	appendRules(config.Rules.Local, rset.Local)
}
