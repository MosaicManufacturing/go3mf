// © Copyright 2021 HP Development Company, L.P.
// SPDX-License Identifier: BSD-2-Clause

package materials

import (
	"fmt"
	"image/color"
	"testing"

	"github.com/MosaicManufacturing/go3mf"
	"github.com/MosaicManufacturing/go3mf/errors"
	"github.com/go-test/deep"
)

func TestDecode(t *testing.T) {
	baseTexture := &Texture2D{ID: 6, Path: "/3D/Texture/msLogo.png", ContentType: TextureTypePNG, TileStyleU: TileWrap, TileStyleV: TileMirror, Filter: TextureFilterAuto}
	colorGroup := &ColorGroup{ID: 1, Colors: []color.RGBA{{R: 255, G: 255, B: 255, A: 255}, {R: 0, G: 0, B: 0, A: 255}, {R: 26, G: 181, B: 103, A: 255}, {R: 223, G: 4, B: 90, A: 255}}}
	texGroup := &Texture2DGroup{ID: 2, TextureID: 6, Coords: []TextureCoord{{0.3, 0.5}, {0.3, 0.8}, {0.5, 0.8}, {0.5, 0.5}}}
	compositeGroup := &CompositeMaterials{ID: 4, MaterialID: 5, Indices: []uint32{1, 2}, Composites: []Composite{{Values: []float32{0.5, 0.5}}, {Values: []float32{0.2, 0.8}}}}
	multiGroup := &MultiProperties{ID: 9, BlendMethods: []BlendMethod{BlendMultiply}, PIDs: []uint32{5, 2}, Multis: []Multi{{PIndices: []uint32{0, 0}}, {PIndices: []uint32{1, 0}}, {PIndices: []uint32{2, 3}}}}
	want := &go3mf.Model{
		Path:       "/3D/3dmodel.model",
		Extensions: []go3mf.Extension{DefaultExtension},
	}
	want.Resources.Assets = append(want.Resources.Assets, baseTexture, colorGroup, texGroup, compositeGroup, multiGroup)
	got := new(go3mf.Model)
	got.Path = "/3D/3dmodel.model"
	rootFile := `
	<model xmlns="http://schemas.microsoft.com/3dmanufacturing/core/2015/02" xmlns:m="http://schemas.microsoft.com/3dmanufacturing/material/2015/02">
		<resources>
			<m:texture2d id="6" path="/3D/Texture/msLogo.png" contenttype="image/png" tilestyleu="wrap" tilestylev="mirror" filter="auto" />
			<m:colorgroup id="1">
				<m:color color="#FFFFFF" /> <m:color color="#000000" /> <m:color color="#1AB567" /> <m:color color="#DF045A" />
			</m:colorgroup>
			<m:texture2dgroup id="2" texid="6">
				<m:tex2coord u="0.3" v="0.5" /> <m:tex2coord u="0.3" v="0.8" />	<m:tex2coord u="0.5" v="0.8" />	<m:tex2coord u="0.5" v="0.5" />
			</m:texture2dgroup>
			<m:compositematerials id="4" matid="5" matindices="1 2">
				<m:composite values="0.5 0.5"/>
				<m:composite values="0.2 0.8"/>
			</m:compositematerials>
			<m:multiproperties id="9" pids="5 2" blendmethods="multiply">
				<m:multi pindices="0 0" />
				<m:multi pindices="1 0" />
				<m:multi pindices="2 3" />
			</m:multiproperties>
		</resources>
		<build>
		</build>
	</model>`
	t.Run("base", func(t *testing.T) {
		if err := go3mf.UnmarshalModel([]byte(rootFile), got); err != nil {
			t.Errorf("DecodeRawModel() unexpected error = %v", err)
			return
		}
		if diff := deep.Equal(got, want); diff != nil {
			t.Errorf("DecodeRawModell() = %v", diff)
			return
		}
	})
}

func TestDecode_warns(t *testing.T) {
	want := []string{
		fmt.Sprintf("Resources@Texture2D#1: %v", errors.NewParseAttrError("id", true)),
		fmt.Sprintf("Resources@ColorGroup#2@RGBA#0: %v", errors.NewParseAttrError("color", true)),
		fmt.Sprintf("Resources@Texture2DGroup#3: %v", errors.NewParseAttrError("texid", true)),
		fmt.Sprintf("Resources@Texture2DGroup#3@TextureCoord#0: %v", errors.NewParseAttrError("u", true)),
		fmt.Sprintf("Resources@Texture2DGroup#3@TextureCoord#1: %v", errors.NewParseAttrError("v", true)),
		fmt.Sprintf("Resources@CompositeMaterials#4: %v", errors.NewParseAttrError("matid", true)),
		fmt.Sprintf("Resources@CompositeMaterials#4@Composite#1: %v", errors.NewParseAttrError("values", true)),
		fmt.Sprintf("Resources@MultiProperties#5: %v", errors.NewParseAttrError("pids", true)),
	}
	got := new(go3mf.Model)
	got.Path = "/3D/3dmodel.model"
	rootFile := `
	<model xmlns="http://schemas.microsoft.com/3dmanufacturing/core/2015/02" xmlns:m="http://schemas.microsoft.com/3dmanufacturing/material/2015/02">
		<resources>
			<m:texture2d id="6" qm:mq="other" path="/3D/Texture/msLogo.png" contenttype="image/png" tilestyleu="wrap" tilestylev="mirror" filter="auto" />
			<m:texture2d id="b" contenttype="image/png" tilestyleu="wrap" tilestylev="mirror" filter="auto" />
			<m:colorgroup id="1">
				<m:color color="#FFFFF" /> <m:color color="#000000" /> <m:color color="#1AB567" /> <m:color color="#DF045A" />
			</m:colorgroup>
			<m:texture2dgroup qm:mq="other" id="2" texid="a">
				<m:tex2coord qm:mq="other" u="b" v="0.5" /> <m:tex2coord u="0.3" v="c" />	<m:tex2coord u="0.5" v="0.8" />	<m:tex2coord u="0.5" v="0.5" />
			</m:texture2dgroup>
			<m:compositematerials id="4" matid="a" qm:mq="other">
				<m:composite/>
				<m:composite values="a 0.8"/>
			</m:compositematerials>
			<m:multiproperties id="9" qm:mq="other" pids="a 2">
				<m:multi />
			</m:multiproperties>
			<m:multiproperties id="19" />
			<object id="8" name="Box 1" pid="5" pindex="0" type="model">
				<mesh>
					<vertices>
						<vertex x="0" y="0" z="0" />
						<vertex x="100.00000" y="0" z="0" />
						<vertex x="100.00000" y="100.00000" z="0" />
						<vertex x="0" y="100.00000" z="0" />
						<vertex x="0" y="0" z="100.00000" />
						<vertex x="100.00000" y="0" z="100.00000" />
						<vertex x="100.00000" y="100.00000" z="100.00000" />
						<vertex x="0" y="100.00000" z="100.00000" />
					</vertices>
					<triangles>
						<triangle v1="2" v2="3" v3="1" />
						<triangle v1="3" v2="2" v3="1" />
						<triangle v1="3" v2="2" v3="1" />
						<triangle v1="1" v2="0" v3="3" />
						<triangle v1="4" v2="5" v3="6" p1="1" />
						<triangle v1="6" v2="7" v3="4" pid="5" p1="1" />
						<triangle v1="0" v2="1" v3="5" pid="2" p1="0" p2="1" p3="2"/>
						<triangle v1="5" v2="4" v3="0" pid="2" p1="3" p2="0" p3="2"/>
						<triangle v1="1" v2="2" v3="6" pid="1" p1="0" p2="1" p3="2"/>
						<triangle v1="6" v2="5" v3="1" pid="1" p1="2" p2="1" p3="3"/>
						<triangle v1="2" v2="3" v3="7" />
						<triangle v1="7" v2="6" v3="2" />
						<triangle v1="3" v2="0" v3="4" />
						<triangle v1="4" v2="7" v3="3" />
					</triangles>
				</mesh>
			</object>
		</resources>
		<build>
		</build>
	</model>`
	t.Run("base", func(t *testing.T) {
		err := go3mf.UnmarshalModel([]byte(rootFile), got)
		if err == nil {
			t.Fatal("error expected")
		}
		var errs []string
		for _, err := range err.(*errors.List).Errors {
			errs = append(errs, err.Error())
		}
		if diff := deep.Equal(errs, want); diff != nil {
			t.Errorf("UnmarshalModel_warn() = %v", diff)
			return
		}
	})
}
