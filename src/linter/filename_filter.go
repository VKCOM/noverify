package linter

import (
	"regexp"

	"github.com/monochromegane/go-gitignore"
)

type FilenameFilter struct {
	exclude *regexp.Regexp

	gitignoreEnabled bool
	matchers         []gitignore.IgnoreMatcher
	initialMatchers  map[string]struct{}
}

func NewFilenameFilter(exclude *regexp.Regexp) *FilenameFilter {
	return &FilenameFilter{
		exclude:         exclude,
		initialMatchers: make(map[string]struct{}),
	}
}

func (filter *FilenameFilter) EnableGitignore() { filter.gitignoreEnabled = true }

func (filter *FilenameFilter) GitignoreIsEnabled() bool { return filter.gitignoreEnabled }

func (filter *FilenameFilter) InitialGitignorePush(path string, matcher gitignore.IgnoreMatcher) {
	if !filter.gitignoreEnabled {
		panic("add when gitignore is disabled")
	}
	DebugMessage("gitignore: add %s/.gitignore", path)
	filter.initialMatchers[path] = struct{}{}
	filter.matchers = append(filter.matchers, matcher)
}

func (filter *FilenameFilter) GitignorePush(path string, matcher gitignore.IgnoreMatcher) {
	if !filter.gitignoreEnabled {
		panic("pop when gitignore is disabled")
	}
	if _, ok := filter.initialMatchers[path]; ok {
		DebugMessage("gitignore: don't push %s/.gitignore", path)
		return
	}
	DebugMessage("gitignore: push %s/.gitignore", path)
	filter.matchers = append(filter.matchers, matcher)
}

func (filter *FilenameFilter) GitignorePop(path string) {
	if !filter.gitignoreEnabled {
		panic("pop when gitignore is disabled")
	}
	if _, ok := filter.initialMatchers[path]; ok {
		DebugMessage("gitignore: don't pop %s/.gitignore", path)
		return
	}
	DebugMessage("gitignore: pop %s/.gitignore", path)
	filter.matchers = filter.matchers[:len(filter.matchers)-1]
}

func (filter *FilenameFilter) IgnoreFile(path string) bool {
	return filter.ignore(path, false)
}

func (filter *FilenameFilter) IgnoreDir(path string) bool {
	return filter.ignore(path, true)
}

func (filter *FilenameFilter) ignore(path string, isDir bool) bool {
	if filter.exclude != nil && filter.exclude.MatchString(path) {
		return true
	}
	for _, matcher := range filter.matchers {
		if matcher.Match(path, isDir) {
			return true
		}
	}
	return false
}
