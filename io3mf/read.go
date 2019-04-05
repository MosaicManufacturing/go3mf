package io3mf

import (
	"bytes"
	"encoding/xml"
	"errors"
	"image/color"
	"io"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/go-gl/mathgl/mgl32"
	go3mf "github.com/qmuntal/go3mf"
)

type relationship interface {
	Type() string
	TargetURI() string
}

type packageFile interface {
	Name() string
	FindFileFromRel(string) (packageFile, bool)
	Relationships() []relationship
	Open() (io.ReadCloser, error)
}

type packageReader interface {
	FindFileFromRel(string) (packageFile, bool)
	FindFileFromName(string) (packageFile, bool)
}

// ReadCloser wrapps a Reader than can be closed.
type ReadCloser struct {
	f *os.File
	*Reader
}

// OpenReader will open the 3MF file specified by name and return a ReadCloser.
func OpenReader(name string) (*ReadCloser, error) {
	f, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	fi, err := f.Stat()
	if err != nil {
		f.Close()
		return nil, err
	}
	r, err := NewReader(f, fi.Size())
	return &ReadCloser{f: f, Reader: r}, err
}

// Close closes the 3MF file, rendering it unusable for I/O.
func (r *ReadCloser) Close() error {
	return r.f.Close()
}

type nodeDecoder interface {
	Open() error
	Attributes([]xml.Attr) error
	Text([]byte) error
	Child(xml.Name) nodeDecoder
	Close() error
	SetModelFile(f *modelFile)
	ModelFile() *modelFile
}

type emptyDecoder struct {
	file *modelFile
}

func (d *emptyDecoder) Open() error                 { return nil }
func (d *emptyDecoder) Attributes([]xml.Attr) error { return nil }
func (d *emptyDecoder) Text([]byte) error           { return nil }
func (d *emptyDecoder) Child(xml.Name) nodeDecoder  { return nil }
func (d *emptyDecoder) Close() error                { return nil }
func (d *emptyDecoder) ModelFile() *modelFile       { return d.file }
func (d *emptyDecoder) SetModelFile(f *modelFile)   { d.file = f }

type topLevelDecoder struct {
	emptyDecoder
	isRoot bool
}

func (d *topLevelDecoder) Child(name xml.Name) (child nodeDecoder) {
	modelName := xml.Name{Space: nsCoreSpec, Local: attrModel}
	if name == modelName {
		child = new(modelDecoder)
	}
	return
}

// modelFile cannot be reused between goroutines.
type modelFile struct {
	r            *Reader
	path         string
	isRoot       bool
	warnings     []error
	resourcesMap map[uint64]go3mf.Identifier
	resources    []go3mf.Identifier
	namespaces   map[string]string
}

func (d *modelFile) AddWarning(err error) {
	d.warnings = append(d.warnings, err)
}

func (d *modelFile) AddResource(r go3mf.Identifier) {
	_, id := r.Identify()
	d.resourcesMap[id] = r
	d.resources = append(d.resources, r)
}

func (d *modelFile) FindResource(path string, id uint64) (r go3mf.Identifier, ok bool) {
	if path == "" {
		path = d.r.Model.Path
	}
	if path == d.path {
		r, ok = d.resourcesMap[id]
	} else {
		r, ok = d.r.Model.FindResource(path, id)
	}
	return
}

func (d *modelFile) NamespaceRegistered(ns string) bool {
	for _, space := range d.namespaces {
		if ns == space {
			return true
		}
	}
	return false
}

func (d *modelFile) Path() string {
	return d.path
}

func (d *modelFile) IsRoot() bool {
	return d.isRoot
}

func (d *modelFile) Decode(x *xml.Decoder) (err error) {
	d.namespaces = make(map[string]string)
	d.resourcesMap = make(map[uint64]go3mf.Identifier)

	state := make([]nodeDecoder, 0, 10)
	names := make([]xml.Name, 0, 10)
	var (
		currentDecoder nodeDecoder
		tmpDecoder     nodeDecoder
		currentName    xml.Name
		t              xml.Token
	)
	currentDecoder = &topLevelDecoder{isRoot: d.isRoot}
	for {
		t, err = x.Token()
		if err != nil {
			break
		}
		switch tp := t.(type) {
		case xml.StartElement:
			tmpDecoder = currentDecoder.Child(tp.Name)
			if tmpDecoder != nil {
				tmpDecoder.SetModelFile(d)
				state = append(state, currentDecoder)
				names = append(names, currentName)
				currentName = tp.Name
				currentDecoder = tmpDecoder
				err = currentDecoder.Open()
				if err == nil {
					err = currentDecoder.Attributes(tp.Attr)
				}
			} else {
				err = x.Skip()
			}
		case xml.CharData:
			err = currentDecoder.Text(tp)
		case xml.EndElement:
			if currentName == tp.Name {
				err = currentDecoder.Close()
				if err == nil {
					currentDecoder, state = state[len(state)-1], state[:len(state)-1]
					currentName, names = names[len(names)-1], names[:len(names)-1]
				}
			}
		}
		if err != nil {
			break
		}
	}
	return
}

// Reader implements a 3mf file reader.
type Reader struct {
	Model               *go3mf.Model
	Warnings            []error
	AttachmentRelations []string
	r                   packageReader
	productionModels    map[string]packageFile
}

// NewReader returns a new Reader reading a 3mf file from r.
func NewReader(r io.ReaderAt, size int64) (*Reader, error) {
	opcr, err := newOPCReader(r, size)
	if err != nil {
		return nil, err
	}
	return &Reader{
		r:     opcr,
		Model: new(go3mf.Model),
	}, nil
}

func (r *Reader) addResource(res go3mf.Identifier) {
	r.Model.Resources = append(r.Model.Resources, res)
}

// Decode reads the 3mf file and unmarshall its content into the model.
func (r *Reader) Decode() error {
	if err := r.processOPC(); err != nil {
		return err
	}
	if err := r.processNonRootModels(); err != nil {
		return err
	}
	if err := r.processRootModel(); err != nil {
		return err
	}
	return nil
}

func (r *Reader) processRootModel() error {
	rootFile, ok := r.r.FindFileFromRel(relTypeModel3D)
	if !ok {
		return errors.New("go3mf: package does not have root model")
	}
	f, err := rootFile.Open()
	if err != nil {
		return err
	}
	defer f.Close()
	d := modelFile{r: r, path: rootFile.Name(), isRoot: true}
	err = d.Decode(xml.NewDecoder(f))
	if err != io.EOF {
		return err
	}
	r.addModelFile(&d)
	return nil
}

func (r *Reader) addModelFile(f *modelFile) {
	for _, res := range f.resources {
		r.addResource(res)
	}
	for _, res := range f.warnings {
		r.Warnings = append(r.Warnings, res)
	}
}

func (r *Reader) processNonRootModels() error {
	var mu sync.Mutex
	prodAttCount := len(r.Model.ProductionAttachments)
	for i := prodAttCount - 1; i >= 0; i-- {
		f, err := r.readProductionAttachmentModel(i)
		if err != nil {
			return err
		}
		mu.Lock()
		r.addModelFile(f)
		mu.Unlock()
	}
	return nil
}

func (r *Reader) processOPC() error {
	rootFile, ok := r.r.FindFileFromRel(relTypeModel3D)
	if !ok {
		return errors.New("go3mf: package does not have root model")
	}

	r.Model.Path = rootFile.Name()
	r.extractTexturesAttachments(rootFile)
	r.extractCustomAttachments(rootFile)
	r.extractModelAttachments(rootFile)
	for _, a := range r.Model.ProductionAttachments {
		file, _ := r.r.FindFileFromName(a.Path)
		r.extractCustomAttachments(file)
		r.extractTexturesAttachments(file)
	}
	thumbFile, ok := rootFile.FindFileFromRel(relTypeThumbnail)
	if ok {
		if buff, err := copyFile(thumbFile); err == nil {
			r.Model.SetThumbnail(buff)
		}
	}

	return nil
}

func (r *Reader) nonRootProgress() float64 {
	if len(r.Model.ProductionAttachments) == 0 {
		return 0.1
	}
	return 0.6
}

func (r *Reader) extractTexturesAttachments(rootFile packageFile) {
	for _, rel := range rootFile.Relationships() {
		if rel.Type() != relTypeTexture3D && rel.Type() != relTypeThumbnail {
			continue
		}

		if file, ok := rootFile.FindFileFromRel(rel.TargetURI()); ok {
			r.Model.Attachments = r.addAttachment(r.Model.Attachments, file, rel.Type())
		}
	}
}

func (r *Reader) extractCustomAttachments(rootFile packageFile) {
	for _, rel := range r.AttachmentRelations {
		if file, ok := rootFile.FindFileFromRel(rel); ok {
			r.Model.Attachments = r.addAttachment(r.Model.Attachments, file, rel)
		}
	}
}

func (r *Reader) extractModelAttachments(rootFile packageFile) {
	r.productionModels = make(map[string]packageFile)
	for _, rel := range rootFile.Relationships() {
		if rel.Type() != relTypeModel3D {
			continue
		}

		if file, ok := rootFile.FindFileFromRel(rel.TargetURI()); ok {
			r.Model.ProductionAttachments = append(r.Model.ProductionAttachments, &go3mf.ProductionAttachment{
				RelationshipType: rel.Type(),
				Path:             file.Name(),
			})
			r.productionModels[file.Name()] = file
		}
	}
}

func (r *Reader) addAttachment(attachments []*go3mf.Attachment, file packageFile, relType string) []*go3mf.Attachment {
	buff, err := copyFile(file)
	if err == nil {
		return append(attachments, &go3mf.Attachment{
			RelationshipType: relType,
			Path:             file.Name(),
			Stream:           buff,
		})
	}
	return attachments
}

func (r *Reader) readProductionAttachmentModel(i int) (*modelFile, error) {
	attachment := r.Model.ProductionAttachments[i]
	file, err := r.productionModels[attachment.Path].Open()
	if err != nil {
		return nil, err
	}
	defer file.Close()
	d := modelFile{r: r, path: attachment.Path, isRoot: false}
	err = d.Decode(xml.NewDecoder(file))
	if err != io.EOF {
		return nil, err
	}
	return &d, nil
}

func copyFile(file packageFile) (io.Reader, error) {
	stream, err := file.Open()
	if err != nil {
		return nil, err
	}
	buff := new(bytes.Buffer)
	_, err = io.Copy(buff, stream)
	stream.Close()
	return buff, err
}

func strToSRGB(s string) (c color.RGBA, err error) {
	var errInvalidFormat = errors.New("gltf: invalid color format")

	if len(s) == 0 || s[0] != '#' {
		return c, errInvalidFormat
	}

	hexToByte := func(b byte) byte {
		switch {
		case b >= '0' && b <= '9':
			return b - '0'
		case b >= 'a' && b <= 'f':
			return b - 'a'
		case b >= 'A' && b <= 'F':
			return b - 'A'
		}
		err = errInvalidFormat
		return 0
	}

	switch len(s) {
	case 9:
		c.R = hexToByte(s[1])<<4 + hexToByte(s[2])
		c.G = hexToByte(s[3])<<4 + hexToByte(s[4])
		c.B = hexToByte(s[5])<<4 + hexToByte(s[6])
		c.A = hexToByte(s[7])<<4 + hexToByte(s[8])
	case 7:
		c.R = hexToByte(s[1])<<4 + hexToByte(s[2])
		c.G = hexToByte(s[3])<<4 + hexToByte(s[4])
		c.B = hexToByte(s[5])<<4 + hexToByte(s[6])
		c.A = 0xff
	default:
		err = errInvalidFormat
	}
	return
}

func strToMatrix(s string) (mgl32.Mat4, error) {
	var matrix mgl32.Mat4
	values := strings.Fields(s)
	if len(values) != 12 {
		return matrix, errors.New("go3mf: matrix string does not have 12 values")
	}
	var t [12]float32
	for i := 0; i < 12; i++ {
		val, err := strconv.ParseFloat(values[i], 32)
		if err != nil {
			return matrix, errors.New("go3mf: matrix string contain characters other than numbers")
		}
		t[i] = float32(val)
	}
	return mgl32.Mat4{t[0], t[3], t[6], t[9],
		t[1], t[4], t[7], t[10],
		t[2], t[5], t[8], t[11],
		0.0, 0.0, 0.0, 1.0}, nil
}
