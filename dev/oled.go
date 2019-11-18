package dev

import (
	"image"
	"image/draw"
	"io/ioutil"

	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
	"github.com/mdp/monochromeoled"
	"golang.org/x/exp/io/i2c"
)

const (
	fontFile = "casio-fx-9860gii.ttf"
)

// OLED ...
type OLED struct {
	oled   *monochromeoled.OLED
	width  int
	height int
	font   *truetype.Font
}

// NewOLED ...
func NewOLED(width, heigth int) (*OLED, error) {
	oled, err := monochromeoled.Open(&i2c.Devfs{Dev: "/dev/i2c-1"}, 0x3c, width, heigth)
	if err != nil {
		return nil, err
	}
	a, err := ioutil.ReadFile(fontFile)
	if err != nil {
		return nil, err
	}
	font, err := truetype.Parse(a)
	if err != nil {
		return nil, err
	}
	return &OLED{
		oled:   oled,
		width:  width,
		height: heigth,
		font:   font,
	}, nil
}

// Display ...
func (o *OLED) Display(text string, fontSize float64, x, y int) error {
	image, err := o.drawText(text, fontSize, x, y)
	if err != nil {
		return err
	}
	if err := o.oled.SetImage(0, 0, image); err != nil {
		return err
	}
	if err := o.oled.Draw(); err != nil {
		return err
	}
	return nil
}

// Clear ...
func (o *OLED) Clear() error {
	if err := o.oled.Clear(); err != nil {
		return err
	}
	return nil
}

// Close ...
func (o *OLED) Close() {
	o.oled.Clear()
	o.oled.Close()
}

func (o *OLED) drawText(text string, size float64, x, y int) (image.Image, error) {
	dst := image.NewRGBA(image.Rect(0, 0, o.width, o.height))
	draw.Draw(dst, dst.Bounds(), image.Transparent, image.ZP, draw.Src)

	c := freetype.NewContext()
	c.SetDst(dst)
	c.SetClip(dst.Bounds())
	c.SetSrc(image.White)
	c.SetFont(o.font)
	c.SetFontSize(size)

	if _, err := c.DrawString(text, freetype.Pt(x, y)); err != nil {
		return nil, err
	}

	return dst, nil
}