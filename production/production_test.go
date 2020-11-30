package production

import (
	"testing"

	"github.com/qmuntal/go3mf"
	"github.com/qmuntal/go3mf/spec"
	"github.com/qmuntal/go3mf/spec/xml"
)

var _ xml.Decoder = new(Spec)
var _ xml.PostProcessorDecoder = new(Spec)
var _ xml.PreProcessEncoder = new(Spec)
var _ go3mf.Spec = new(Spec)
var _ spec.ModelValidator = new(Spec)
var _ spec.ObjectValidator = new(Spec)
var _ xml.MarshalerAttr = new(BuildAttr)
var _ xml.MarshalerAttr = new(ItemAttr)
var _ spec.ObjectPather = new(ItemAttr)
var _ xml.MarshalerAttr = new(ComponentAttr)
var _ spec.ObjectPather = new(ComponentAttr)
var _ xml.MarshalerAttr = new(ObjectAttr)

func TestComponentAttr_ObjectPath(t *testing.T) {
	tests := []struct {
		name string
		p    *ComponentAttr
		want string
	}{
		{"empty", new(ComponentAttr), ""},
		{"path", &ComponentAttr{Path: "/a.model"}, "/a.model"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.p.ObjectPath(); got != tt.want {
				t.Errorf("ComponentAttr.ObjectPath() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestItemAttr_ObjectPath(t *testing.T) {
	tests := []struct {
		name string
		p    *ItemAttr
		want string
	}{
		{"empty", new(ItemAttr), ""},
		{"path", &ItemAttr{Path: "/a.model"}, "/a.model"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.p.ObjectPath(); got != tt.want {
				t.Errorf("ItemAttr.ObjectPath() = %v, want %v", got, tt.want)
			}
		})
	}
}
