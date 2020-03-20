package meta

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"strings"

	"github.com/VKCOM/noverify/src/php/parser/node"
)

var debugScope = false

type scopeVar struct {
	typesMap      TypesMap
	alwaysDefined bool
	noReplace     bool // do not replace variable upon assignment (used for phpdoc @var declaration)

	// whether a variable is implicitly defined (by isset or instanceof).
	// Implicit vars are not tracked and they are not required to be "used".
	// They're also more scoped than normal variables and rarely outlive
	// the block context in which they were introduced.
	implicit bool
}

// Scope contains variables with their types in the respective scope
type Scope struct {
	vars             map[string]*scopeVar // variables declared in the scope
	inInstanceMethod bool
	inClosure        bool
}

// NewScope creates new empty scope
func NewScope() *Scope {
	return &Scope{vars: make(map[string]*scopeVar)}
}

// GobEncode is a custom gob marshaller
func (s *Scope) GobEncode() ([]byte, error) {
	w := new(bytes.Buffer)
	encoder := gob.NewEncoder(w)
	err := encoder.Encode(s.vars)
	if err != nil {
		return nil, err
	}
	err = encoder.Encode(s.inInstanceMethod)
	if err != nil {
		return nil, err
	}
	err = encoder.Encode(s.inClosure)
	if err != nil {
		return nil, err
	}
	return w.Bytes(), nil
}

// GobDecode is custom gob unmarshaller
func (s *Scope) GobDecode(buf []byte) error {
	r := bytes.NewBuffer(buf)
	decoder := gob.NewDecoder(r)
	err := decoder.Decode(&s.vars)
	if err != nil {
		return err
	}
	err = decoder.Decode(&s.inInstanceMethod)
	if err != nil {
		return err
	}
	return decoder.Decode(&s.inClosure)
}

// GobEncode is a custom gob marshaller
func (s *scopeVar) GobEncode() ([]byte, error) {
	w := new(bytes.Buffer)
	encoder := gob.NewEncoder(w)
	err := encoder.Encode(s.typesMap)
	if err != nil {
		return nil, err
	}
	err = encoder.Encode(s.alwaysDefined)
	if err != nil {
		return nil, err
	}
	return w.Bytes(), nil
}

// GobDecode is custom gob unmarshaller
func (s *scopeVar) GobDecode(buf []byte) error {
	r := bytes.NewBuffer(buf)
	decoder := gob.NewDecoder(r)
	err := decoder.Decode(&s.typesMap)
	if err != nil {
		return err
	}
	return decoder.Decode(&s.alwaysDefined)
}

// IsInInstanceMethod returns whether or not this scope exists in instance method (and thus closures must capture $this)
func (s *Scope) IsInInstanceMethod() bool {
	return s.inInstanceMethod
}

// IsInClosure returns whether or not this scope is inside a closure and thus $this can be late-bound.
func (s *Scope) IsInClosure() bool {
	return s.inClosure
}

// SetInInstanceMethod updates "inInstanceMethod" flag that indicated whether or not scope is located inside instance method
// and that "$this" needs to be captured
func (s *Scope) SetInInstanceMethod(v bool) {
	s.inInstanceMethod = v
}

// SetInClosure updates "inClosure" flag that indicates whether or not we are inside a closure
// and thus late $this binding is possible.
func (s *Scope) SetInClosure(v bool) {
	s.inClosure = v
}

func (s *Scope) Iterate(cb func(varName string, typ TypesMap, alwaysDefined bool)) {
	for varName, v := range s.vars {
		cb(varName, v.typesMap, v.alwaysDefined)
	}
}

func (s *Scope) IterateExplicit(cb func(varName string, typ TypesMap, alwaysDefined bool)) {
	for varName, v := range s.vars {
		if !v.implicit {
			cb(varName, v.typesMap, v.alwaysDefined)
		}
	}
}

func (s *Scope) Len() int {
	return len(s.vars)
}

// AddVar adds variable with specified types to scope
func (s *Scope) AddVar(v node.Node, typ TypesMap, reason string, alwaysDefined bool) {
	name, ok := scopeVarName(v)
	if !ok {
		return
	}
	s.AddVarName(name, typ, reason, alwaysDefined)
}

// AddImplicitVar adds implicit variable with specified types to scope
func (s *Scope) AddImplicitVar(varNode node.Node, typ TypesMap, reason string, alwaysDefined bool) {
	name, ok := scopeVarName(varNode)
	if !ok {
		return
	}
	v := s.addVarName(name, typ, reason, alwaysDefined)
	v.implicit = true
}

// ReplaceVar replaces variable with specified types to scope
func (s *Scope) ReplaceVar(v node.Node, typ TypesMap, reason string, alwaysDefined bool) {
	name, ok := scopeVarName(v)
	if !ok {
		return
	}

	s.ReplaceVarName(name, typ, reason, alwaysDefined)
}

// DelVar deletes specified variable from scope
func (s *Scope) DelVar(v node.Node, reason string) {
	name, ok := scopeVarName(v)
	if !ok {
		return
	}

	s.DelVarName(name, reason)
}

// DelVarName deletes variable from the scope by it's name
func (s *Scope) DelVarName(name, reason string) {
	if debugScope {
		fmt.Println("unset $" + name + " - " + reason)
	}
	delete(s.vars, name)
}

// ReplaceVarName replaces variable with specified types to the scope
func (s *Scope) ReplaceVarName(name string, typ TypesMap, reason string, alwaysDefined bool) {
	oldVar, ok := s.vars[name]
	if ok && oldVar.noReplace {
		oldVar.typesMap = oldVar.typesMap.Append(typ)
		return
	}

	s.vars[name] = &scopeVar{
		typesMap:      typ,
		alwaysDefined: alwaysDefined,
	}
}

// AddVarName adds variable with specified types to the scope
func (s *Scope) addVarName(name string, typ TypesMap, reason string, alwaysDefined bool) *scopeVar {
	v, ok := s.vars[name]

	if !ok {
		v := &scopeVar{
			typesMap:      typ,
			alwaysDefined: alwaysDefined,
		}
		s.vars[name] = v
		return v
	}

	if !v.alwaysDefined && alwaysDefined {
		v.alwaysDefined = true
	}

	v.typesMap = v.typesMap.Append(typ)
	s.vars[name] = v
	return v
}

// AddVarName adds variable with specified types to the scope
func (s *Scope) AddVarName(name string, typ TypesMap, reason string, alwaysDefined bool) {
	s.addVarName(name, typ, reason, alwaysDefined)
}

// AddVarFromPHPDoc adds variable with specified types to the scope
func (s *Scope) AddVarFromPHPDoc(name string, typ TypesMap, reason string) {
	v := s.addVarName(name, typ, reason, true)
	v.noReplace = true
}

// HaveVar checks whether or not specified variable is present in the scope and that it is always defined
func (s *Scope) HaveVar(v node.Node) bool {
	name, ok := scopeVarName(v)
	if !ok {
		return false
	}

	return s.HaveVarName(name)
}

// MaybeHaveVar checks that variable is present in the scope (it may be not always defined)
func (s *Scope) MaybeHaveVar(v node.Node) bool {
	name, ok := scopeVarName(v)
	if !ok {
		return false
	}

	return s.MaybeHaveVarName(name)
}

// HaveVarName checks whether or not specified variable is present in the scope and that it is always defined
func (s *Scope) HaveVarName(name string) bool {
	v, ok := s.vars[name]
	if !ok {
		return false
	}
	return v.alwaysDefined
}

// GetVarNameType returns type map for variable if it exists
func (s *Scope) GetVarNameType(name string) (m TypesMap, ok bool) {
	res, ok := s.vars[name]
	if !ok {
		return TypesMap{}, false
	}
	return res.typesMap, ok
}

// MaybeHaveVarName checks that variable is present in the scope (it may be not always defined)
func (s *Scope) MaybeHaveVarName(name string) bool {
	_, ok := s.vars[name]
	return ok
}

// String returns vars contents (for debug purposes)
func (s *Scope) String() string {
	var res []string

	for name, v := range s.vars {
		res = append(res, fmt.Sprintf("%s: alwaysDefined=%v, typ=%s", name, v.alwaysDefined, v.typesMap))
	}

	return strings.Join(res, "\n")
}

// Clone creates a full scope copy (used in branches)
func (s *Scope) Clone() *Scope {
	if s == nil {
		return NewScope()
	}

	res := &Scope{vars: make(map[string]*scopeVar, len(s.vars))}
	for k, v := range s.vars {
		res.vars[k] = &scopeVar{
			typesMap:      v.typesMap.clone(),
			alwaysDefined: v.alwaysDefined,
		}
	}
	res.inInstanceMethod = s.inInstanceMethod
	res.inClosure = s.inClosure
	return res
}

func scopeVarName(v node.Node) (string, bool) {
	switch v := v.(type) {
	case *node.SimpleVar:
		return v.Name, true
	case *node.Var:
		vv, ok := v.Expr.(*node.SimpleVar)
		if !ok {
			return "", false // Don't go further than 1 level
		}
		return "$" + vv.Name, true
	default:
		return "", false
	}
}
