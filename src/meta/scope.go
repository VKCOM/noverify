package meta

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io"
	"strings"

	"github.com/VKCOM/noverify/src/ir"
	"github.com/VKCOM/noverify/src/types"
)

var debugScope = false

type VarFlags uint8

const (
	// varNoReplace - do not replace variable upon assignment (used for phpdoc @var declaration)
	varNoReplace VarFlags = 1 << iota
	VarAlwaysDefined
	VarImplicit
)

type ScopeVar struct {
	Type  types.Map
	Flags VarFlags
}

func (flags VarFlags) IsNoReplace() bool     { return flags&varNoReplace != 0 }
func (flags VarFlags) IsAlwaysDefined() bool { return flags&VarAlwaysDefined != 0 }
func (flags VarFlags) IsImplicit() bool      { return flags&VarImplicit != 0 }

func (flags *VarFlags) SetAlwaysDefined(v bool) {
	if v {
		*flags |= VarAlwaysDefined
	} else {
		*flags &^= VarAlwaysDefined
	}
}

// Scope contains variables with their types in the respective scope
type Scope struct {
	vars             map[string]*ScopeVar // variables declared in the scope
	inInstanceMethod bool
	inClosure        bool
}

// NewScope creates new empty scope
func NewScope() *Scope {
	return &Scope{vars: make(map[string]*ScopeVar)}
}

func (s *Scope) GobWrite(w io.Writer) error {
	encoder := gob.NewEncoder(w)
	err := encoder.Encode(s.vars)
	if err != nil {
		return err
	}
	err = encoder.Encode(s.inInstanceMethod)
	if err != nil {
		return err
	}
	err = encoder.Encode(s.inClosure)
	if err != nil {
		return err
	}
	return nil
}

// GobEncode is a custom gob marshaller
func (s *Scope) GobEncode() ([]byte, error) {
	w := new(bytes.Buffer)
	if err := s.GobWrite(w); err != nil {
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
func (s *ScopeVar) GobEncode() ([]byte, error) {
	w := new(bytes.Buffer)
	encoder := gob.NewEncoder(w)
	err := encoder.Encode(s.Type)
	if err != nil {
		return nil, err
	}
	err = encoder.Encode(s.Flags)
	if err != nil {
		return nil, err
	}
	return w.Bytes(), nil
}

// GobDecode is custom gob unmarshaller
func (s *ScopeVar) GobDecode(buf []byte) error {
	r := bytes.NewBuffer(buf)
	decoder := gob.NewDecoder(r)
	err := decoder.Decode(&s.Type)
	if err != nil {
		return err
	}
	return decoder.Decode(&s.Flags)
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

func (s *Scope) Iterate(cb func(varName string, typ types.Map, flags VarFlags)) {
	for varName, v := range s.vars {
		cb(varName, v.Type, v.Flags)
	}
}

func (s *Scope) Len() int {
	return len(s.vars)
}

// AddVar adds variable with specified types to scope
func (s *Scope) AddVar(v ir.Node, typ types.Map, reason string, flags VarFlags) {
	name, ok := scopeVarName(v)
	if !ok {
		return
	}
	s.AddVarName(name, typ, reason, flags)
}

// AddImplicitVar adds implicit variable with specified types to scope
func (s *Scope) AddImplicitVar(varNode ir.Node, typ types.Map, reason string, flags VarFlags) {
	name, ok := scopeVarName(varNode)
	if !ok {
		return
	}
	s.addVarName(name, typ, reason, flags|VarImplicit)
}

// ReplaceVar replaces variable with specified types to scope
func (s *Scope) ReplaceVar(v ir.Node, typ types.Map, reason string, flags VarFlags) {
	name, ok := scopeVarName(v)
	if !ok {
		return
	}

	s.ReplaceVarName(name, typ, reason, flags)
}

// DelVar deletes specified variable from scope
func (s *Scope) DelVar(v ir.Node, reason string) {
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
func (s *Scope) ReplaceVarName(name string, typ types.Map, reason string, flags VarFlags) {
	oldVar, ok := s.vars[name]
	if ok && oldVar.Flags.IsNoReplace() {
		oldVar.Type = oldVar.Type.Append(typ)
		return
	}

	s.vars[name] = &ScopeVar{
		Type:  typ,
		Flags: flags,
	}
}

// AddVarName adds variable with specified types to the scope
func (s *Scope) addVarName(name string, typ types.Map, reason string, flags VarFlags) {
	v, ok := s.vars[name]

	if !ok {
		s.vars[name] = &ScopeVar{
			Type:  typ,
			Flags: flags,
		}
		return
	}

	if !v.Flags.IsAlwaysDefined() && flags.IsAlwaysDefined() {
		v.Flags |= VarAlwaysDefined
	}

	if !v.Flags.IsNoReplace() && flags.IsNoReplace() {
		v.Flags |= varNoReplace
	}

	v.Type = v.Type.Append(typ)
	s.vars[name] = v
}

// AddVarName adds variable with specified types to the scope
func (s *Scope) AddVarName(name string, typ types.Map, reason string, flags VarFlags) {
	s.addVarName(name, typ, reason, flags)
}

// AddVarFromPHPDoc adds variable with specified types to the scope
func (s *Scope) AddVarFromPHPDoc(name string, typ types.Map, reason string) {
	s.addVarName(name, typ, reason, varNoReplace|VarAlwaysDefined)
}

// HaveVar checks whether or not specified variable is present in the scope
// and that it is always defined
func (s *Scope) HaveVar(v ir.Node) bool {
	name, ok := scopeVarName(v)
	if !ok {
		return false
	}

	return s.HaveVarName(name)
}

// HaveImplicitVar checks whether or not specified implicit variable is present
// in the scope and that it is always defined
func (s *Scope) HaveImplicitVar(v ir.Node) bool {
	name, ok := scopeVarName(v)
	if !ok {
		return false
	}

	return s.HaveImplicitVarName(name)
}

// MaybeHaveVar checks that variable is present in the scope (it may be not always defined)
func (s *Scope) MaybeHaveVar(v ir.Node) bool {
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
	return v.Flags.IsAlwaysDefined()
}

// HaveImplicitVarName checks whether or not specified implicit variable is present in the scope and that it is always defined
func (s *Scope) HaveImplicitVarName(name string) bool {
	v, ok := s.vars[name]
	if !ok {
		return false
	}
	return v.Flags.IsImplicit()
}

// GetVarName returns variable if it exists
func (s *Scope) GetVarName(name string) (m *ScopeVar, ok bool) {
	res, ok := s.vars[name]
	return res, ok
}

// GetVar returns variable if it exists
func (s *Scope) GetVar(v ir.Node) (m *ScopeVar, ok bool) {
	name, ok := scopeVarName(v)
	if !ok {
		return nil, false
	}
	return s.GetVarName(name)
}

// GetVarType returns type map for variable if it exists
func (s *Scope) GetVarType(v ir.Node) (m types.Map, ok bool) {
	name, ok := scopeVarName(v)
	if !ok {
		return types.Map{}, false
	}
	return s.GetVarNameType(name)
}

// GetVarNameType returns type map for variable if it exists
func (s *Scope) GetVarNameType(name string) (m types.Map, ok bool) {
	res, ok := s.vars[name]
	if !ok {
		return types.Map{}, false
	}
	return res.Type, ok
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
		res = append(res, fmt.Sprintf("%s: alwaysDefined=%v, typ=%s", name, v.Flags.IsAlwaysDefined(), v.Type))
	}

	return strings.Join(res, "\n")
}

// Clone creates a full scope copy (used in branches)
func (s *Scope) Clone() *Scope {
	if s == nil {
		return NewScope()
	}

	res := &Scope{vars: make(map[string]*ScopeVar, len(s.vars))}
	for k, v := range s.vars {
		res.vars[k] = &ScopeVar{
			Type:  v.Type.Clone(),
			Flags: v.Flags,
		}
	}
	res.inInstanceMethod = s.inInstanceMethod
	res.inClosure = s.inClosure
	return res
}

func scopeVarName(v ir.Node) (string, bool) {
	switch v := v.(type) {
	case *ir.SimpleVar:
		return v.Name, true
	case *ir.Var:
		vv, ok := v.Expr.(*ir.SimpleVar)
		if !ok {
			return "", false // Don't go further than 1 level
		}
		return "$" + vv.Name, true
	default:
		return "", false
	}
}
