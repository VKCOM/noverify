package meta

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"strings"

	"github.com/VKCOM/noverify/src/ir"
)

var debugScope = false

type VarFlags uint8

const (
	// varNoReplace - do not replace variable upon assignment (used for phpdoc @var declaration)
	varNoReplace VarFlags = 1 << iota
	VarAlwaysDefined
)

type ScopeVar struct {
	TypesMap TypesMap
	Flags    VarFlags
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
	Vars             map[string]*ScopeVar // variables declared in the scope
	InInstanceMethod bool
	InClosure        bool
}

// NewScope creates new empty scope
func NewScope() *Scope {
	return &Scope{Vars: make(map[string]*ScopeVar)}
}

// GobEncode is a custom gob marshaller
func (s *Scope) GobEncode() ([]byte, error) {
	w := new(bytes.Buffer)
	encoder := gob.NewEncoder(w)
	err := encoder.Encode(s.Vars)
	if err != nil {
		return nil, err
	}
	err = encoder.Encode(s.InInstanceMethod)
	if err != nil {
		return nil, err
	}
	err = encoder.Encode(s.InClosure)
	if err != nil {
		return nil, err
	}
	return w.Bytes(), nil
}

// GobDecode is custom gob unmarshaller
func (s *Scope) GobDecode(buf []byte) error {
	r := bytes.NewBuffer(buf)
	decoder := gob.NewDecoder(r)
	err := decoder.Decode(&s.Vars)
	if err != nil {
		return err
	}
	err = decoder.Decode(&s.InInstanceMethod)
	if err != nil {
		return err
	}
	return decoder.Decode(&s.InClosure)
}

// GoString is used when %#v print format is requested.
func (s *ScopeVar) GoString() string {
	return fmt.Sprintf("&meta.ScopeVar{TypesMap: %#v, Flags: %#v}", s.TypesMap, s.Flags)
}

// GobEncode is a custom gob marshaller
func (s *ScopeVar) GobEncode() ([]byte, error) {
	w := new(bytes.Buffer)
	encoder := gob.NewEncoder(w)
	err := encoder.Encode(s.TypesMap)
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
	err := decoder.Decode(&s.TypesMap)
	if err != nil {
		return err
	}
	return decoder.Decode(&s.Flags)
}

// IsInInstanceMethod returns whether or not this scope exists in instance method (and thus closures must capture $this)
func (s *Scope) IsInInstanceMethod() bool {
	return s.InInstanceMethod
}

// IsInClosure returns whether or not this scope is inside a closure and thus $this can be late-bound.
func (s *Scope) IsInClosure() bool {
	return s.InClosure
}

// SetInInstanceMethod updates "inInstanceMethod" flag that indicated whether or not scope is located inside instance method
// and that "$this" needs to be captured
func (s *Scope) SetInInstanceMethod(v bool) {
	s.InInstanceMethod = v
}

// SetInClosure updates "inClosure" flag that indicates whether or not we are inside a closure
// and thus late $this binding is possible.
func (s *Scope) SetInClosure(v bool) {
	s.InClosure = v
}

func (s *Scope) Iterate(cb func(varName string, typ TypesMap, flags VarFlags)) {
	for varName, v := range s.Vars {
		cb(varName, v.TypesMap, v.Flags)
	}
}

func (s *Scope) Len() int {
	return len(s.Vars)
}

// AddVar adds variable with specified types to scope
func (s *Scope) AddVar(v ir.Node, typ TypesMap, reason string, flags VarFlags) {
	name, ok := ScopeVarName(v)
	if !ok {
		return
	}
	s.AddVarName(name, typ, reason, flags)
}

// ReplaceVar replaces variable with specified types to scope
func (s *Scope) ReplaceVar(v ir.Node, typ TypesMap, reason string, flags VarFlags) {
	name, ok := ScopeVarName(v)
	if !ok {
		return
	}

	s.ReplaceVarName(name, typ, reason, flags)
}

// DelVar deletes specified variable from scope
func (s *Scope) DelVar(v ir.Node, reason string) {
	name, ok := ScopeVarName(v)
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
	delete(s.Vars, name)
}

// ReplaceVarName replaces variable with specified types to the scope
func (s *Scope) ReplaceVarName(name string, typ TypesMap, reason string, flags VarFlags) {
	oldVar, ok := s.Vars[name]
	if ok && oldVar.Flags.IsNoReplace() {
		oldVar.TypesMap = oldVar.TypesMap.Append(typ)
		return
	}

	s.Vars[name] = &ScopeVar{
		TypesMap: typ,
		Flags:    flags,
	}
}

// AddVarName adds variable with specified types to the scope
func (s *Scope) addVarName(name string, typ TypesMap, reason string, flags VarFlags) {
	v, ok := s.Vars[name]

	if !ok {
		s.Vars[name] = &ScopeVar{
			TypesMap: typ,
			Flags:    flags,
		}
		return
	}

	if !v.Flags.IsAlwaysDefined() && flags.IsAlwaysDefined() {
		v.Flags |= VarAlwaysDefined
	}

	if !v.Flags.IsNoReplace() && flags.IsNoReplace() {
		v.Flags |= varNoReplace
	}

	v.TypesMap = v.TypesMap.Append(typ)
	s.Vars[name] = v
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
func (s *Scope) HaveVar(v ir.Node) bool {
	name, ok := ScopeVarName(v)
	if !ok {
		return false
	}

	return s.HaveVarName(name)
}

// MaybeHaveVar checks that variable is present in the scope (it may be not always defined)
func (s *Scope) MaybeHaveVar(v ir.Node) bool {
	name, ok := ScopeVarName(v)
	if !ok {
		return false
	}

	return s.MaybeHaveVarName(name)
}

// HaveVarName checks whether or not specified variable is present in the scope and that it is always defined
func (s *Scope) HaveVarName(name string) bool {
	v, ok := s.Vars[name]
	if !ok {
		return false
	}
	return v.Flags.IsAlwaysDefined()
}

// GetVarNameType returns type map for variable if it exists
func (s *Scope) GetVarNameType(name string) (m TypesMap, ok bool) {
	res, ok := s.Vars[name]
	if !ok {
		return TypesMap{}, false
	}
	return res.TypesMap, ok
}

// MaybeHaveVarName checks that variable is present in the scope (it may be not always defined)
func (s *Scope) MaybeHaveVarName(name string) bool {
	_, ok := s.Vars[name]
	return ok
}

// String returns vars contents (for debug purposes)
func (s *Scope) String() string {
	var res []string

	for name, v := range s.Vars {
		res = append(res, fmt.Sprintf("%s: alwaysDefined=%v, typ=%s", name, v.Flags.IsAlwaysDefined(), v.TypesMap))
	}

	return strings.Join(res, "\n")
}

// Clone creates a full scope copy (used in branches)
func (s *Scope) Clone() *Scope {
	if s == nil {
		return NewScope()
	}

	res := &Scope{Vars: make(map[string]*ScopeVar, len(s.Vars))}
	for k, v := range s.Vars {
		res.Vars[k] = &ScopeVar{
			TypesMap: v.TypesMap.Clone(),
			Flags:    v.Flags,
		}
	}
	res.InInstanceMethod = s.InInstanceMethod
	res.InClosure = s.InClosure
	return res
}

func ScopeVarName(v ir.Node) (string, bool) {
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
