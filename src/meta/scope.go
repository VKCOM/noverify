package meta

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"strings"

	"github.com/z7zmey/php-parser/node"
	"github.com/z7zmey/php-parser/node/expr"
)

var debugScope = false

type scopeVar struct {
	typesMap      *TypesMap
	alwaysDefined bool
	noReplace     bool // do not replace variable upon assignment (used for phpdoc @var declaration)
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

func (s *Scope) Iterate(cb func(varName string, typ *TypesMap, alwaysDefined bool)) {
	for varName, v := range s.vars {
		cb(varName, v.typesMap, v.alwaysDefined)
	}
}

func (s *Scope) Len() int {
	return len(s.vars)
}

// AddVar adds variable with specified types to scope
func (s *Scope) AddVar(v *expr.Variable, typ *TypesMap, reason string, alwaysDefined bool) {
	name, ok := v.VarName.(*node.Identifier)
	if !ok {
		return
	}

	s.AddVarName(name.Value, typ, reason, alwaysDefined)
}

// ReplaceVar replaces variable with specified types to scope
func (s *Scope) ReplaceVar(v *expr.Variable, typ *TypesMap, reason string, alwaysDefined bool) {
	name, ok := v.VarName.(*node.Identifier)
	if !ok {
		return
	}

	s.ReplaceVarName(name.Value, typ, reason, alwaysDefined)
}

// DelVar deletes specified variable from scope
func (s *Scope) DelVar(v *expr.Variable, reason string) {
	name, ok := v.VarName.(*node.Identifier)
	if !ok {
		return
	}

	s.DelVarName(name.Value, reason)
}

// DelVarName deletes variable from the scope by it's name
func (s *Scope) DelVarName(name, reason string) {
	if debugScope {
		fmt.Println("unset $" + name + " - " + reason)
	}
	delete(s.vars, name)
}

// ReplaceVarName replaces variable with specified types to the scope
func (s *Scope) ReplaceVarName(name string, typ *TypesMap, reason string, alwaysDefined bool) {
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
func (s *Scope) addVarName(name string, typ *TypesMap, reason string, alwaysDefined, noReplace bool) {
	v, ok := s.vars[name]

	if !ok {
		s.vars[name] = &scopeVar{
			typesMap:      typ,
			alwaysDefined: alwaysDefined,
			noReplace:     noReplace,
		}
		return
	}

	if !v.alwaysDefined && alwaysDefined {
		v.alwaysDefined = true
	}

	if !v.noReplace && noReplace {
		v.noReplace = true
	}

	v.typesMap = v.typesMap.Append(typ)
	s.vars[name] = v
}

// AddVarName adds variable with specified types to the scope
func (s *Scope) AddVarName(name string, typ *TypesMap, reason string, alwaysDefined bool) {
	s.addVarName(name, typ, reason, alwaysDefined, false)
}

// AddVarFromPHPDoc adds variable with specified types to the scope
func (s *Scope) AddVarFromPHPDoc(name string, typ *TypesMap, reason string) {
	s.addVarName(name, typ, reason, true, true)
}

// HaveVar checks whether or not specified variable is present in the scope and that it is always defined
func (s *Scope) HaveVar(v *expr.Variable) bool {
	name, ok := v.VarName.(*node.Identifier)
	if !ok {
		return false
	}

	return s.HaveVarName(name.Value)
}

// MaybeHaveVar checks that variable is present in the scope (it may be not always defined)
func (s *Scope) MaybeHaveVar(v *expr.Variable) bool {
	name, ok := v.VarName.(*node.Identifier)
	if !ok {
		return false
	}

	return s.MaybeHaveVarName(name.Value)
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
func (s *Scope) GetVarNameType(name string) (m *TypesMap, ok bool) {
	res, ok := s.vars[name]
	if !ok {
		return &TypesMap{}, false
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
		m := make(map[string]struct{}, v.typesMap.Len())
		if v.typesMap != nil {
			for kk, vv := range v.typesMap.m {
				m[kk] = vv
			}
		}
		res.vars[k] = &scopeVar{
			typesMap:      &TypesMap{m: m},
			alwaysDefined: v.alwaysDefined,
		}
	}
	res.inInstanceMethod = s.inInstanceMethod
	res.inClosure = s.inClosure
	return res
}
