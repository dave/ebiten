// Copyright 2015 Hajime Hoshi
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// +build ignore

package main

import (
	"errors"
	"image"
	"image/color"
	"image/png"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"text/template"

	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"

	"github.com/hajimehoshi/ebiten"
	"github.com/hajimehoshi/ebiten/ebitenutil"
	"github.com/hajimehoshi/ebiten/internal"
	"github.com/hajimehoshi/ebiten/text"
)

const (
	arcadeFontSize = 8
)

var (
	arcadeFont font.Face
)

func init() {
	f, err := ebitenutil.OpenFile(filepath.Join("..", "..", "_resources", "fonts", "arcade_n.ttf"))
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	b, err := ioutil.ReadAll(f)
	if err != nil {
		log.Fatal(err)
	}

	tt, err := truetype.Parse(b)
	if err != nil {
		log.Fatal(err)
	}

	const dpi = 72
	arcadeFont = truetype.NewFace(tt, &truetype.Options{
		Size:    arcadeFontSize,
		DPI:     dpi,
		Hinting: font.HintingFull,
	})
}

var keyboardKeys = [][]string{
	{"Esc", "1", "2", "3", "4", "5", "6", "7", "8", "9", "0", "-", "=", "\\", "`", " "},
	{"Tab", "Q", "W", "E", "R", "T", "Y", "U", "I", "O", "P", "[", "]", "BS"},
	{"Ctrl", "A", "S", "D", "F", "G", "H", "J", "K", "L", ";", "'", "Enter"},
	{"Shift", "Z", "X", "C", "V", "B", "N", "M", ",", ".", "/", " "},
	{" ", "Alt", "Space", " ", " "},
	{},
	{"", "Up", ""},
	{"Left", "Down", "Right"},
}

func drawKey(t *ebiten.Image, name string, x, y, width int) {
	const height = 16
	width--
	img, _ := ebiten.NewImage(width, height, ebiten.FilterNearest)
	p := make([]byte, width*height*4)
	for j := 0; j < height; j++ {
		for i := 0; i < width; i++ {
			x := (i + j*width) * 4
			switch j {
			case 0, height - 1:
				if 3 <= i && i <= width-4 {
					p[x] = 0xff
					p[x+1] = 0xff
					p[x+2] = 0xff
					p[x+3] = 0xff
				}
			case 1, height - 2:
				if i == 2 || i == width-3 {
					p[x] = 0xff
					p[x+1] = 0xff
					p[x+2] = 0xff
					p[x+3] = 0xff
				}
			case 2, height - 3:
				if i == 1 || i == width-2 {
					p[x] = 0xff
					p[x+1] = 0xff
					p[x+2] = 0xff
					p[x+3] = 0xff
				}
			default:
				if i == 0 || i == width-1 {
					p[x] = 0xff
					p[x+1] = 0xff
					p[x+2] = 0xff
					p[x+3] = 0xff
				}
			}
		}
	}
	img.ReplacePixels(p)
	const offset = 4
	text.Draw(img, name, arcadeFont, offset, arcadeFontSize+offset, color.White)

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(x), float64(y))
	t.DrawImage(img, op)
}

func outputKeyboardImage() (map[string]image.Rectangle, error) {
	keyMap := map[string]image.Rectangle{}
	img, _ := ebiten.NewImage(320, 240, ebiten.FilterNearest)
	x, y := 0, 0
	for j, line := range keyboardKeys {
		x = 0
		const height = 18
		for i, key := range line {
			width := 16
			switch j {
			default:
				switch i {
				case 0:
					width = 16 + 8*(j+2)
				case len(line) - 1:
					if j > 0 {
						width = 16 + 8*(j+2)
					}
				}
			case 4:
				switch i {
				case 0:
					width = 16 + 8*(j+2)
				case 1:
					width = 16 * 2
				case 2:
					width = 16 * 5
				case 3:
					width = 16 * 2
				case 4:
					width = 16 + 8*(j+2)
				}
			case 6, 7:
				width = 16 * 3
			}
			if key != "" {
				drawKey(img, key, x, y, width)
				if key != " " {
					keyMap[key] = image.Rect(x, y, x+width, y+height)
				}
			}
			x += width
		}
		y += height
	}

	f, err := os.Create(filepath.Join("..", "..", "_resources", "images", "keyboard", "keyboard.png"))
	if err != nil {
		return nil, err
	}
	defer f.Close()

	if err := png.Encode(f, img); err != nil {
		return nil, err
	}
	return keyMap, nil
}

const keyRectTmpl = `{{.License}}

// DO NOT EDIT: This file is auto-generated by gen.go.

{{.BuildTags}}

package keyboard

import (
	"image"
)

var keyboardKeyRects = map[string]image.Rectangle{}

func init() {
{{range $key, $rect := .KeyRectsMap}}	keyboardKeyRects[{{printf "%q" $key}}] = image.Rect({{$rect.Min.X}}, {{$rect.Min.Y}}, {{$rect.Max.X}}, {{$rect.Max.Y}})
{{end}}}

func KeyRect(name string) (image.Rectangle, bool) {
	r, ok := keyboardKeyRects[name]
	return r, ok
}`

func outputKeyRectsGo(k map[string]image.Rectangle) error {
	license, err := internal.LicenseComment()
	if err != nil {
		return err
	}
	path := "keyrects.go"

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	tmpl, err := template.New(path).Parse(keyRectTmpl)
	if err != nil {
		return err
	}
	return tmpl.Execute(f, map[string]interface{}{
		"BuildTags":   "",
		"License":     license,
		"KeyRectsMap": k,
	})
}

var regularTermination = errors.New("regular termination")

func main() {
	var rects map[string]image.Rectangle
	if err := ebiten.Run(func(_ *ebiten.Image) error {
		var err error
		rects, err = outputKeyboardImage()
		if err != nil {
			return err
		}
		return regularTermination
	}, 256, 256, 1, ""); err != regularTermination {
		log.Fatal(err)
	}
	if err := outputKeyRectsGo(rects); err != nil {
		log.Fatal(err)
	}
}
