package cmd

import (
	"bytes"
	"embed"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/VKCOM/noverify/src/ir"
	"github.com/VKCOM/noverify/src/rules"
)

//go:embed embeddedrules
var embeddedRulesData embed.FS

func AddEmbeddedRules(rset *rules.Set, p *rules.Parser, filter func(r rules.Rule) bool) ([]*rules.Set, error) {
	embeddedRuleSets, err := parseEmbeddedRules(p)
	if err != nil {
		return nil, err
	}

	for _, embeddedRuleSet := range embeddedRuleSets {
		appendRuleSet(rset, embeddedRuleSet, filter)
	}

	return embeddedRuleSets, nil
}

func parseRules(externalRules string) ([]*rules.Set, error) {
	p := rules.NewParser()

	ruleSets, err := parseEmbeddedRules(p)
	if err != nil {
		return nil, fmt.Errorf("embedded rules: %v", err)
	}

	if externalRules != "" {
		for _, filename := range strings.Split(externalRules, ",") {
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

func parseEmbeddedRules(p *rules.Parser) ([]*rules.Set, error) {
	var ruleSets []*rules.Set

	entries, err := embeddedRulesData.ReadDir("embeddedrules")
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		filename := filepath.ToSlash(filepath.Join("embeddedrules", entry.Name()))
		data, err := embeddedRulesData.ReadFile(filename)
		if err != nil {
			return nil, err
		}

		rset, err := p.Parse(entry.Name(), bytes.NewReader(data))
		if err != nil {
			return nil, err
		}
		rset.Builtin = true
		ruleSets = append(ruleSets, rset)
	}

	return ruleSets, nil
}

func appendRuleSet(dstSet *rules.Set, srcSet *rules.Set, filter func(r rules.Rule) bool) {
	appendRules := func(dst, src *rules.ScopedSet) {
		for kind, ruleByKind := range &src.RulesByKind {
			for _, rule := range ruleByKind {
				if !filter(rule) {
					continue
				}

				dst.Add(ir.NodeKind(kind), rule)
			}
		}
	}
	appendRules(dstSet.Any, srcSet.Any)
	appendRules(dstSet.Root, srcSet.Root)
	appendRules(dstSet.Local, srcSet.Local)
}
