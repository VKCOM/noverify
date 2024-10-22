package pathChecker

import (
	"path/filepath"
	"strings"
	"testing"
)

// Узел дерева правил
type RuleNode struct {
	Children map[string]*RuleNode // Дочерние узлы (поддиректории и файлы)
	Enable   map[string]bool      // Включенные правила на этом уровне
	Disable  map[string]bool      // Отключенные правила на этом уровне
}

func NewRuleNode() *RuleNode {
	return &RuleNode{
		Children: make(map[string]*RuleNode),
		Enable:   make(map[string]bool),
		Disable:  make(map[string]bool),
	}
}

type PathRuleSet struct {
	Enable  map[string]bool // Правила, которые включены для данного пути
	Disable map[string]bool // Правила, которые отключены для данного пути
}

func BuildRuleTree(pathRules map[string]*PathRuleSet) *RuleNode {
	root := NewRuleNode()

	for path, ruleSet := range pathRules {
		normalizedPath := filepath.ToSlash(filepath.Clean(path))
		parts := strings.Split(normalizedPath, "/")
		currentNode := root

		for _, part := range parts {
			if part == "" {
				continue
			}
			if _, exists := currentNode.Children[part]; !exists {
				currentNode.Children[part] = NewRuleNode()
			}
			currentNode = currentNode.Children[part]
		}

		// Добавляем правила в конечный узел
		for rule := range ruleSet.Enable {
			currentNode.Enable[rule] = true
		}
		for rule := range ruleSet.Disable {
			currentNode.Disable[rule] = true
		}
	}

	return root
}

func IsRuleEnabled(root *RuleNode, filePath string, checkRule string) bool {
	normalizedPath := filepath.ToSlash(filepath.Clean(filePath))
	parts := strings.Split(normalizedPath, "/")
	currentNode := root
	ruleState := "unknown"

	for _, part := range parts {
		if part == "" {
			continue
		}
		if node, exists := currentNode.Children[part]; exists {
			// Проверяем отключенные правила
			if node.Disable[checkRule] {
				ruleState = "disabled"
			}
			// Проверяем включенные правила
			if node.Enable[checkRule] {
				ruleState = "enabled"
			}
			currentNode = node
		} else {
			break
		}
	}

	// По умолчанию считаем, что правило включено, если не указано обратное
	if ruleState == "disabled" {
		return false
	}
	return true
}

func TestIsRuleEnabled(t *testing.T) {
	pathRules := map[string]*PathRuleSet{
		"A": {
			Enable:  map[string]bool{"rule1": true},
			Disable: map[string]bool{},
		},
		"A/B": {
			Enable:  map[string]bool{},
			Disable: map[string]bool{"rule1": true},
		},
		"A/B/C": {
			Enable:  map[string]bool{"rule2": true, "rule4": true},
			Disable: map[string]bool{},
		},
		"A/B/C/file.php": {
			Enable:  map[string]bool{"rule3": true},
			Disable: map[string]bool{},
		},
		"A/B/C/another.php": {
			Enable:  map[string]bool{"rule3": true},
			Disable: map[string]bool{"rule4": true},
		},
	}

	// Строим дерево правил
	ruleTree := BuildRuleTree(pathRules)

	tests := []struct {
		name      string
		filePath  string
		checkRule string
		want      bool
	}{
		{
			name:      "Rule1 отключено в A/B",
			filePath:  "A/B/D",
			checkRule: "rule1",
			want:      false,
		},
		{
			name:      "Rule2 включено по умолчанию",
			filePath:  "A/B/D",
			checkRule: "rule2",
			want:      true,
		},
		{
			name:      "Rule1 включено в A/X/Y",
			filePath:  "A/X/Y",
			checkRule: "rule1",
			want:      true,
		},
		{
			name:      "Rule2 включено в A/B/C",
			filePath:  "A/B/C/other.php",
			checkRule: "rule2",
			want:      true,
		},
		{
			name:      "Rule3 включено в A/B/C/file.php",
			filePath:  "A/B/C/file.php",
			checkRule: "rule3",
			want:      true,
		},
		{
			name:      "Rule3 неявно включено в A/B/C/other.php",
			filePath:  "A/B/C/other.php",
			checkRule: "rule3",
			want:      true,
		},
		{
			name:      "Rule4 отключено в A/B/C/another.php, несмотря на то что включено в A/B/C",
			filePath:  "A/B/C/another.php",
			checkRule: "rule4",
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsRuleEnabled(ruleTree, tt.filePath, tt.checkRule)
			if got != tt.want {
				t.Errorf("IsRuleEnabled(%q, %q) = %v; ожидается %v", tt.filePath, tt.checkRule, got, tt.want)
			}
		})
	}
}
