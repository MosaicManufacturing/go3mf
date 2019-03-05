package io3mf

import (
	"bytes"
	"encoding/xml"
	"errors"
	"image/color"
	"io"
	"strconv"
	"strings"

	"github.com/go-gl/mathgl/mgl32"
	mdl "github.com/qmuntal/go3mf/internal/model"
	"github.com/qmuntal/go3mf/internal/progress"
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

// Reader implements a 3mf file reader.
type Reader struct {
	Model               *mdl.Model
	Warnings            []error
	AttachmentRelations []string
	progress            progress.Monitor
	r                   packageReader
	namespaces          []string
}

// NewReader returns a new Reader reading a 3mf file from r.
func NewReader(r io.ReaderAt, size int64) (*Reader, error) {
	opcr, err := newOPCReader(r, size)
	if err != nil {
		return nil, err
	}
	return &Reader{
		r: opcr,
	}, nil
}

func (r *Reader) namespaceRegistered(ns string) bool {
	for _, space := range r.namespaces {
		if ns == space {
			return true
		}
	}
	return false
}

// SetProgressCallback specifies the callback to be executed on every step of the progress.
func (r *Reader) SetProgressCallback(callback progress.ProgressCallback, userData interface{}) {
	r.progress.SetProgressCallback(callback, userData)
}

// decode reads the 3mf file and unmarshall its content into the model.
func (r *Reader) decode() error {
	r.progress.ResetLevels()
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
	if !r.progress.Progress(r.nonRootProgress(), progress.StageReadRootModel) {
		return ErrUserAborted
	}
	rootFile, ok := r.r.FindFileFromRel(relTypeModel3D)
	if !ok {
		return errors.New("go3mf: package does not have root model")
	}
	f, err := rootFile.Open()
	if err != nil {
		return err
	}
	defer f.Close()
	x := xml.NewDecoder(f)
	var t xml.Token
mainLoop:
	for {
		t, err = x.Token()
		if err != nil {
			break
		}
		switch tp := t.(type) {
		case xml.StartElement:
			if tp.Name.Local == attrModel {
				md := modelDecoder{x: x, r: r, model: r.Model}
				err = md.Decode(tp)
				break mainLoop
			}
		}
	}
	if err != io.EOF {
		return err
	}
	return nil
}

func (r *Reader) processNonRootModels() error {
	if !r.progress.Progress(0.1, progress.StageReadNonRootModels) {
		return ErrUserAborted
	}
	r.progress.PushLevel(0.1, r.nonRootProgress())
	// read production attachments
	r.progress.PopLevel()
	return nil
}

func (r *Reader) processOPC() error {
	if !r.progress.Progress(0.05, progress.StageExtractOPCPackage) {
		return ErrUserAborted
	}
	rootFile, ok := r.r.FindFileFromRel(relTypeModel3D)
	if !ok {
		return errors.New("go3mf: package does not have root model")
	}

	r.Model.RootPath = rootFile.Name()
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
	for _, rel := range rootFile.Relationships() {
		if rel.Type() != relTypeModel3D {
			continue
		}

		if file, ok := rootFile.FindFileFromRel(rel.TargetURI()); ok {
			r.Model.ProductionAttachments = r.addAttachment(r.Model.ProductionAttachments, file, rel.Type())
		}
	}
}

func (r *Reader) addAttachment(attachments []*mdl.Attachment, file packageFile, relType string) []*mdl.Attachment {
	buff, err := copyFile(file)
	if err == nil {
		return append(attachments, &mdl.Attachment{
			RelationshipType: relType,
			Path:             file.Name(),
			Stream:           buff,
		})
	}
	return attachments
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
