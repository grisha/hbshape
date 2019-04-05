package hbshape

//
// #cgo CFLAGS: -I/usr/include/freetype2 -I/usr/include/harfbuzz
// #cgo LDFLAGS: -L/usr/lib64 -lfreetype -lharfbuzz
//
// #include <hb.h>
// #include <hb-ft.h>
//
// hb_glyph_position_t hbshape_glyph_pos_at(hb_glyph_position_t *pos, int i) {
//   return pos[i];
// }
//
import "C"
import (
	"fmt"
	"log"
	"sync"
	"unsafe"
)

var (
	ftLib   C.FT_Library
	mux     sync.Mutex
	initErr error
)

func init() {
	if rc, err := C.FT_Init_FreeType(&ftLib); rc != 0 {
		initErr = fmt.Errorf("FT_Init_FreeType(): %v (%d).", err, rc)
	}
}

func newFace(filename string) (C.FT_Face, error) {
	mux.Lock()
	defer mux.Unlock()

	var face C.FT_Face

	cs := C.CString(filename)
	defer C.free(unsafe.Pointer(cs))

	if rc, err := C.FT_New_Face(ftLib, cs, 0, &face); rc != 0 {
		return nil, fmt.Errorf("FT_New_Face(): %v (%d).", err, rc)
	}

	return face, nil
}

func newMemoryFace(fontBytes []byte) (C.FT_Face, error) {
	mux.Lock()
	defer mux.Unlock()

	var face C.FT_Face

	fb := C.CBytes(fontBytes)
	//TODO: This should be freed, but where? (not here)
	//defer C.free(fb)

	if rc, err := C.FT_New_Memory_Face(ftLib, (*_Ctype_uchar)(fb), C.long(len(fontBytes)), 0, &face); rc != 0 {
		return nil, fmt.Errorf("FT_New_Memory_Face(): %v (%d).", err, rc)
	}

	return face, nil
}

type shaper struct {
	face C.FT_Face
	hbf  *C.hb_font_t
	hbb  *C.hb_buffer_t
}

//func NewShaper(fontPath string, fontSize int) (*shaper, error) {
func NewShaper(fontBytes []byte, fontSize int) (*shaper, error) {

	var sh shaper

	face, err := newMemoryFace(fontBytes)
	//face, err := newFace(fontPath)
	if err != nil {
		return nil, err
	}

	log.Printf("ZZZ fontsize: %v", fontSize)

	C.FT_Set_Char_Size(face, C.long(fontSize*64), C.long(fontSize*64), 0, 0)

	sh.hbf = C.hb_ft_font_create(face, nil)

	// TODO: Errors, destroy?

	return &sh, nil
}

type GlyphPos struct {
	XAdvance, YAdvance, XOffset, YOffset float64
}

func (sh *shaper) ShapeText(text string) ([]*GlyphPos, error) {

	hbb := C.hb_buffer_create()
	C.hb_buffer_add_utf8(hbb, C.CString(text), -1, 0, -1)
	C.hb_buffer_guess_segment_properties(hbb)
	C.hb_shape(sh.hbf, hbb, nil, 0)

	l := int(C.hb_buffer_get_length(hbb))
	pos := C.hb_buffer_get_glyph_positions(hbb, nil)

	result := make([]*GlyphPos, l)

	for i := 0; i < l; i++ {
		p := C.hbshape_glyph_pos_at(pos, C.int(i))
		result[i] = &GlyphPos{
			float64(p.x_advance) / 64,
			float64(p.y_advance) / 64,
			float64(p.x_offset) / 64,
			float64(p.y_offset) / 64,
		}
	}

	return result, nil
}
