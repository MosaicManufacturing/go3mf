// © Copyright 2021 HP Development Company, L.P.
// SPDX-License Identifier: BSD-2-Clause

package materials

import (
	"image/color"
	"testing"

	"github.com/MosaicManufacturing/go3mf"
	"github.com/go-test/deep"
)

func TestMarshalModel(t *testing.T) {
	baseTexture := &Texture2D{ID: 6, Path: "/3D/Texture/msLogo.png", ContentType: TextureTypePNG, TileStyleU: TileWrap, TileStyleV: TileMirror, Filter: TextureFilterAuto}
	colorGroup := &ColorGroup{ID: 1, Colors: []color.RGBA{{R: 255, G: 255, B: 255, A: 255}, {R: 0, G: 0, B: 0, A: 255}, {R: 26, G: 181, B: 103, A: 255}, {R: 223, G: 4, B: 90, A: 255}}}
	texGroup := &Texture2DGroup{ID: 2, TextureID: 6, Coords: []TextureCoord{{0.3, 0.5}, {0.3, 0.8}, {0.5, 0.8}, {0.5, 0.5}}}
	compositeGroup := &CompositeMaterials{ID: 4, MaterialID: 5, Indices: []uint32{1, 2}, Composites: []Composite{{Values: []float32{0.5, 0.5}}, {Values: []float32{0.2, 0.8}}}}
	multiGroup := &MultiProperties{ID: 9, BlendMethods: []BlendMethod{BlendMultiply}, PIDs: []uint32{5, 2}, Multis: []Multi{{PIndices: []uint32{0, 0}}, {PIndices: []uint32{1, 0}}, {PIndices: []uint32{2, 3}}}}
	m := &go3mf.Model{Path: "/3D/3dmodel.model"}
	m.Resources.Assets = append(m.Resources.Assets, baseTexture, colorGroup, texGroup, compositeGroup, multiGroup)
	m.Extensions = []go3mf.Extension{DefaultExtension}
	t.Run("base", func(t *testing.T) {
		b, err := go3mf.MarshalModel(m)
		if err != nil {
			t.Errorf("materials.MarshalModel() error = %v", err)
			return
		}
		newModel := new(go3mf.Model)
		newModel.Path = m.Path
		if err := go3mf.UnmarshalModel(b, newModel); err != nil {
			t.Errorf("materials.MarshalModel() error decoding = %v, s = %s", err, string(b))
			return
		}
		if diff := deep.Equal(m, newModel); diff != nil {
			t.Errorf("materials.MarshalModel() = %v, s = %s", diff, string(b))
		}
	})
}
