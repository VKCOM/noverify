package meta

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"strings"

	"github.com/VKCOM/noverify/src/php/parser/node"
)

var debugScope = false

type VarFlags uint8

const (
	// varNoReplace - do not replace variable upon assignment (used for phpdoc @var declaration)
	varNoReplace VarFlags = 1 << iota
	VarAlwaysDefined
)

type scopeVar struct {
	typesMap TypesMap
	flags    VarFlags
}

func (flags VarFlags) IsNoReplace() bool     { return flags&varNoReplace != 0 }
func (flags VarFlags) IsAlwaysDefined() bool { return flags&VarAlwaysDefined != 0 }

func (flags *VarFlags) SetAlwaysDefined(v bool) {
	if v {
		*flags |= VarAlwaysDefined
	} else {
		*flags &^= VarAlwaysDefined
	}
}

// Scope contains variables with their types in the respective scope
type Scope struct {
	vars             map[string]*scopeVar // variables declared in the scope
	inInstanceMethod bool
	inClosure        bool

	CallerFunction     FuncInfo    // function that is called with this closure
	CallerFunctionArgs []node.Node // and the arguments with which it is called
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
	err = encoder.Encode(s.flags)
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
	return decoder.Decode(&s.flags)
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

func (s *Scope) Iterate(cb func(varName string, typ TypesMap, flags VarFlags)) {
	for varName, v := range s.vars {
		cb(varName, v.typesMap, v.flags)
	}
}

func (s *Scope) Len() int {
	return len(s.vars)
}

// AddVar adds variable with specified types to scope
func (s *Scope) AddVar(v node.Node, typ TypesMap, reason string, flags VarFlags) {
	name, ok := scopeVarName(v)
	if !ok {
		return
	}
	s.AddVarName(name, typ, reason, flags)
}

// ReplaceVar replaces variable with specified types to scope
func (s *Scope) ReplaceVar(v node.Node, typ TypesMap, reason string, flags VarFlags) {
	name, ok := scopeVarName(v)
	if !ok {
		return
	}

	s.ReplaceVarName(name, typ, reason, flags)
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
func (s *Scope) ReplaceVarName(name string, typ TypesMap, reason string, flags VarFlags) {
	oldVar, ok := s.vars[name]
	if ok && oldVar.flags.IsNoReplace() {
		oldVar.typesMap = oldVar.typesMap.Append(typ)
		return
	}

	s.vars[name] = &scopeVar{
		typesMap: typ,
		flags:    flags,
	}
}

// AddVarName adds variable with specified types to the scope
func (s *Scope) addVarName(name string, typ TypesMap, reason string, flags VarFlags) {
	v, ok := s.vars[name]

	if !ok {
		s.vars[name] = &scopeVar{
			typesMap: typ,
			flags:    flags,
		}
		return
	}

	if !v.flags.IsAlwaysDefined() && flags.IsAlwaysDefined() {
		v.flags |= VarAlwaysDefined
	}

	if !v.flags.IsNoReplace() && flags.IsNoReplace() {
		v.flags |= varNoReplace
	}

	v.typesMap = v.typesMap.Append(typ)
	s.vars[name] = v
}

// AddVarName adds variable with specified types to the scope
func (s *Scope) AddVarName(name string, typ TypesMap, reason string, flags VarFlags) {
	s.addVarName(name, typ, reason, flags)
}

// AddVarFromPHPDoc adds variable with specified types to the scope
func (s *Scope) AddVarFromPHPDoc(name string, typ TypesMap, reason string) {
	s.addVarName(name, typ, reason, varNoReplace|VarAlwaysDefined)
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
	return v.flags.IsAlwaysDefined()
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
		res = append(res, fmt.Sprintf("%s: alwaysDefined=%v, typ=%s", name, v.flags.IsAlwaysDefined(), v.typesMap))
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
			typesMap: v.typesMap.Clone(),
			flags:    v.flags,
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
