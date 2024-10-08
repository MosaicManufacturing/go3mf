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

func TestValidate(t *testing.T) {
	tests := []struct {
		name  string
		model *go3mf.Model
		want  []string
	}{
		{"child", &go3mf.Model{Childs: map[string]*go3mf.ChildModel{
			"/other.model": {Resources: go3mf.Resources{Assets: []go3mf.Asset{
				&ColorGroup{ID: 1},
			}}},
			"/that.model": {Resources: go3mf.Resources{Assets: []go3mf.Asset{
				&MultiProperties{ID: 2},
			}}},
		}}, []string{
			fmt.Sprintf("/other.model@Resources@ColorGroup#0: %v", errors.ErrEmptyResourceProps),
			fmt.Sprintf("/that.model@Resources@MultiProperties#0: %v", &errors.MissingFieldError{Name: attrPIDs}),
			fmt.Sprintf("/that.model@Resources@MultiProperties#0: %v", ErrMultiBlend),
			fmt.Sprintf("/that.model@Resources@MultiProperties#0: %v", errors.ErrEmptyResourceProps),
		}},
		{"multi", &go3mf.Model{
			Resources: go3mf.Resources{Assets: []go3mf.Asset{
				&MultiProperties{ID: 4},
				&MultiProperties{ID: 5, Multis: []Multi{{PIndices: []uint32{}}}, PIDs: []uint32{4, 100}},
				&go3mf.BaseMaterials{ID: 1, Materials: []go3mf.Base{
					{Name: "a", Color: color.RGBA{R: 1}},
					{Name: "b", Color: color.RGBA{G: 1}},
				}},
				&ColorGroup{ID: 6, Colors: []color.RGBA{{R: 1}, {R: 2, G: 3, B: 4, A: 5}}},
				&CompositeMaterials{ID: 3, MaterialID: 1, Indices: []uint32{0, 1}, Composites: []Composite{{Values: []float32{1, 2}}}},
				&MultiProperties{ID: 2, Multis: []Multi{{PIndices: []uint32{1, 0}}}, PIDs: []uint32{1, 6}},
				&MultiProperties{ID: 7, Multis: []Multi{{PIndices: []uint32{1, 3}}}, PIDs: []uint32{1, 6}},
				&MultiProperties{ID: 8, Multis: []Multi{{PIndices: []uint32{}}}, PIDs: []uint32{6, 1, 6}},
				&MultiProperties{ID: 9, Multis: []Multi{{PIndices: []uint32{}}}, PIDs: []uint32{1, 3}},
			}},
		}, []string{
			fmt.Sprintf("Resources@MultiProperties#0: %v", &errors.MissingFieldError{Name: attrPIDs}),
			fmt.Sprintf("Resources@MultiProperties#0: %v", ErrMultiBlend),
			fmt.Sprintf("Resources@MultiProperties#0: %v", errors.ErrEmptyResourceProps),
			fmt.Sprintf("Resources@MultiProperties#1: %v", ErrMultiRefMulti),
			fmt.Sprintf("Resources@MultiProperties#1: %v", errors.ErrMissingResource),
			fmt.Sprintf("Resources@MultiProperties#6@Multi#0: %v", errors.ErrIndexOutOfBounds),
			fmt.Sprintf("Resources@MultiProperties#7: %v", ErrMaterialMulti),
			fmt.Sprintf("Resources@MultiProperties#7: %v", ErrMultiColors),
			fmt.Sprintf("Resources@MultiProperties#8: %v", ErrMaterialMulti),
		}},
		{"missingTextPart", &go3mf.Model{
			Resources: go3mf.Resources{Assets: []go3mf.Asset{
				&Texture2D{ID: 1},
				&Texture2D{ID: 2, ContentType: TextureTypePNG, Path: "/a.png"},
			}},
		}, []string{
			fmt.Sprintf("Resources@Texture2D#0: %v", &errors.MissingFieldError{Name: attrPath}),
			fmt.Sprintf("Resources@Texture2D#0: %v", &errors.MissingFieldError{Name: attrContentType}),
			fmt.Sprintf("Resources@Texture2D#1: %v", ErrMissingTexturePart),
		}},
		{"textureGroup", &go3mf.Model{
			Attachments: []go3mf.Attachment{{Path: "/a.png"}},
			Resources: go3mf.Resources{Assets: []go3mf.Asset{
				&Texture2D{ID: 1, ContentType: TextureTypePNG, Path: "/a.png"},
				&Texture2DGroup{ID: 2},
				&Texture2DGroup{ID: 3, TextureID: 1, Coords: []TextureCoord{{}}},
				&Texture2DGroup{ID: 4, TextureID: 2, Coords: []TextureCoord{{}}},
				&Texture2DGroup{ID: 5, TextureID: 100, Coords: []TextureCoord{{}}},
			}},
		}, []string{
			fmt.Sprintf("Resources@Texture2DGroup#1: %v", &errors.MissingFieldError{Name: attrTexID}),
			fmt.Sprintf("Resources@Texture2DGroup#1: %v", errors.ErrEmptyResourceProps),
			fmt.Sprintf("Resources@Texture2DGroup#3: %v", ErrTextureReference),
			fmt.Sprintf("Resources@Texture2DGroup#4: %v", ErrTextureReference),
		}},
		{"colorGroup", &go3mf.Model{
			Resources: go3mf.Resources{Assets: []go3mf.Asset{
				&ColorGroup{ID: 1},
				&ColorGroup{ID: 2, Colors: []color.RGBA{{R: 1}, {R: 2, G: 3, B: 4, A: 5}}},
				&ColorGroup{ID: 3, Colors: []color.RGBA{{R: 1}, {}}},
			}},
		}, []string{
			fmt.Sprintf("Resources@ColorGroup#0: %v", errors.ErrEmptyResourceProps),
			fmt.Sprintf("Resources@ColorGroup#2@RGBA#1: %v", &errors.MissingFieldError{Name: attrColor}),
		}},
		{"composite", &go3mf.Model{
			Resources: go3mf.Resources{Assets: []go3mf.Asset{
				&go3mf.BaseMaterials{ID: 1, Materials: []go3mf.Base{
					{Name: "a", Color: color.RGBA{R: 1}},
					{Name: "b", Color: color.RGBA{G: 1}},
				}},
				&CompositeMaterials{ID: 2},
				&CompositeMaterials{ID: 3, MaterialID: 1, Indices: []uint32{0, 1}, Composites: []Composite{{Values: []float32{1, 2}}}},
				&CompositeMaterials{ID: 4, MaterialID: 1, Indices: []uint32{100, 100}, Composites: []Composite{{Values: []float32{1, 2}}}},
				&CompositeMaterials{ID: 5, MaterialID: 2, Indices: []uint32{0, 1}, Composites: []Composite{{Values: []float32{1, 2}}}},
				&CompositeMaterials{ID: 6, MaterialID: 100, Indices: []uint32{0, 1}, Composites: []Composite{{Values: []float32{1, 2}}}},
			}}}, []string{
			fmt.Sprintf("Resources@CompositeMaterials#1: %v", &errors.MissingFieldError{Name: attrMatID}),
			fmt.Sprintf("Resources@CompositeMaterials#1: %v", &errors.MissingFieldError{Name: attrMatIndices}),
			fmt.Sprintf("Resources@CompositeMaterials#1: %v", errors.ErrEmptyResourceProps),
			fmt.Sprintf("Resources@CompositeMaterials#3: %v", errors.ErrIndexOutOfBounds),
			fmt.Sprintf("Resources@CompositeMaterials#4: %v", ErrCompositeBase),
			fmt.Sprintf("Resources@CompositeMaterials#5: %v", errors.ErrMissingResource),
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.model.Extensions = []go3mf.Extension{DefaultExtension}
			err := tt.model.Validate()
			if err == nil {
				t.Fatal("error expected")
			}
			var errs []string
			for _, err := range err.(*errors.List).Errors {
				errs = append(errs, err.Error())
			}
			if diff := deep.Equal(errs, tt.want); diff != nil {
				t.Errorf("Validate() = %v", diff)
			}
		})
	}
}
