package cmd

import (
	"bytes"
	"embed"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/VKCOM/noverify/src/ir"
	"github.com/VKCOM/noverify/src/rules"
)

//go:embed embeddedrules
var embeddedRulesData embed.FS

func AddEmbeddedRules(rset *rules.Set, filter func(r rules.Rule) bool) ([]*rules.Set, error) {
	embeddedRuleSets, err := ParseEmbeddedRules()
	if err != nil {
		return nil, err
	}

	for _, embeddedRuleSet := range embeddedRuleSets {
		appendRuleSet(rset, embeddedRuleSet, filter)
	}

	return embeddedRuleSets, nil
}

func ParseEmbeddedRules() ([]*rules.Set, error) {
	var ruleSets []*rules.Set
	p := rules.NewParser()

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

func ParseExternalRules(externalRules string) ([]*rules.Set, error) {
	if externalRules == "" {
		return nil, nil
	}

	var ruleSets []*rules.Set

	for _, filename := range strings.Split(externalRules, ",") {
		stat, err := os.Stat(filename)
		if err != nil {
			return nil, err
		}

		if stat.IsDir() {
			dir := filename
			files, err := ioutil.ReadDir(dir)
			if err != nil {
				return nil, err
			}

			for _, file := range files {
				if file.IsDir() {
					continue
				}

				ruleSets, err = readAndParseRuleFile(filepath.Join(dir, file.Name()), ruleSets)
				if err != nil {
					return nil, err
				}
			}
		} else {
			ruleSets, err = readAndParseRuleFile(filename, ruleSets)
			if err != nil {
				return nil, err
			}
		}
	}

	return ruleSets, nil
}

func readAndParseRuleFile(filename string, ruleSets []*rules.Set) ([]*rules.Set, error) {
	p := rules.NewParser()

	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	ruleSet, err := p.Parse(filename, bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	ruleSets = append(ruleSets, ruleSet)

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
