package go3mf

import (
	"encoding/xml"
	"reflect"
)

var marshalerAttrType = reflect.TypeOf((*MarshalerAttr)(nil)).Elem()
var marshalerType = reflect.TypeOf((*Marshaler)(nil)).Elem()

type Spec interface {
	Space() string
	Local() string
	Required() bool
	SetRequired(bool)
	SetLocal(string)
}

type UnknownSpec struct {
	SpaceName  string
	LocalName  string
	IsRequired bool
}

func (u *UnknownSpec) Space() string      { return u.SpaceName }
func (u *UnknownSpec) Local() string      { return u.LocalName }
func (u *UnknownSpec) Required() bool     { return u.IsRequired }
func (u *UnknownSpec) SetLocal(l string)  { u.LocalName = l }
func (u *UnknownSpec) SetRequired(r bool) { u.IsRequired = r }

type ExtensionDecoder interface {
	Spec
	NewNodeDecoder(interface{}, string) NodeDecoder
	DecodeAttribute(*Scanner, interface{}, xml.Attr)
}

type PropertyGroup interface {
	Len() int
}

// AnyAttr is an extension point containing <anyAttribute> information.
// The key should be the extension namespace.
type AnyAttr []MarshalerAttr

// Get will panic if target is not a non-nil pointer to either a type that implements
// MarshallerAttr, or to any interface type.
func (e AnyAttr) Get(target interface{}) bool {
	if e == nil || len(e) == 0 {
		return false
	}
	if target == nil {
		panic("go3mf: target cannot be nil")
	}

	val := reflect.ValueOf(target)
	typ := val.Type()
	if typ.Kind() != reflect.Ptr || val.IsNil() {
		panic("go3mf: target must be a non-nil pointer")
	}
	if el := typ.Elem(); el.Kind() != reflect.Interface && !el.Implements(marshalerAttrType) {
		panic("go3mf: *target must be interface or implement MarshalerAttr")
	}
	targetType := typ.Elem()
	for _, v := range e {
		if v != nil && reflect.TypeOf(v).AssignableTo(targetType) {
			val.Elem().Set(reflect.ValueOf(v))
			return true
		}
	}
	return false
}

func (e AnyAttr) encode(x *XMLEncoder, start *xml.StartElement) {
	for _, ext := range e {
		if att, err := ext.Marshal3MFAttr(x); err == nil {
			start.Attr = append(start.Attr, att...)
		}
	}
}

// Extension is an extension point containing <any> information.
// The key should be the extension namespace.
type Extension []Marshaler

// Get finds the first Marshaller that matches target, and if so, sets
// target to that extension value and returns true.

// A Marshaller matches target if the marshaller's concrete value is assignable to the value
// pointed to by target.

// Get will panic if target is not a non-nil pointer to either a type that implements
// Marshaller, or to any interface type.
func (e Extension) Get(target interface{}) bool {
	if e == nil || len(e) == 0 {
		return false
	}
	if target == nil {
		panic("go3mf: target cannot be nil")
	}

	val := reflect.ValueOf(target)
	typ := val.Type()
	if typ.Kind() != reflect.Ptr || val.IsNil() {
		panic("go3mf: target must be a non-nil pointer")
	}
	if el := typ.Elem(); el.Kind() != reflect.Interface && !el.Implements(marshalerType) {
		panic("go3mf: *target must be interface or implement Marshaler")
	}
	targetType := typ.Elem()
	for _, v := range e {
		if v != nil && reflect.TypeOf(v).AssignableTo(targetType) {
			val.Elem().Set(reflect.ValueOf(v))
			return true
		}
	}
	return false
}

func (e Extension) encode(x *XMLEncoder) error {
	for _, ext := range e {
		if err := ext.Marshal3MF(x); err == nil {
			return err
		}
	}
	return nil
}
