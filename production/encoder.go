// © Copyright 2021 HP Development Company, L.P.
// SPDX-License Identifier: BSD-2-Clause

package production

import (
	"encoding/xml"

	"github.com/MosaicManufacturing/go3mf/spec"
)

// Marshal3MFAttr encodes the resource attributes.
func (u *BuildAttr) Marshal3MFAttr(_ spec.Encoder) ([]xml.Attr, error) {
	return []xml.Attr{
		{Name: xml.Name{Space: Namespace, Local: attrProdUUID}, Value: u.UUID},
	}, nil
}

// Marshal3MFAttr encodes the resource attributes.
func (u *ObjectAttr) Marshal3MFAttr(_ spec.Encoder) ([]xml.Attr, error) {
	return []xml.Attr{
		{Name: xml.Name{Space: Namespace, Local: attrProdUUID}, Value: u.UUID},
	}, nil
}

// Marshal3MFAttr encodes the resource attributes.
func (p *ItemAttr) Marshal3MFAttr(_ spec.Encoder) ([]xml.Attr, error) {
	if p.Path == "" {
		return []xml.Attr{
			{Name: xml.Name{Space: Namespace, Local: attrProdUUID}, Value: p.UUID},
		}, nil
	}
	return []xml.Attr{
		{Name: xml.Name{Space: Namespace, Local: attrPath}, Value: p.Path},
		{Name: xml.Name{Space: Namespace, Local: attrProdUUID}, Value: p.UUID},
	}, nil
}

// Marshal3MFAttr encodes the resource attributes.
func (p *ComponentAttr) Marshal3MFAttr(_ spec.Encoder) ([]xml.Attr, error) {
	if p.Path == "" {
		return []xml.Attr{
			{Name: xml.Name{Space: Namespace, Local: attrProdUUID}, Value: p.UUID},
		}, nil
	}
	return []xml.Attr{
		{Name: xml.Name{Space: Namespace, Local: attrPath}, Value: p.Path},
		{Name: xml.Name{Space: Namespace, Local: attrProdUUID}, Value: p.UUID},
	}, nil
}
