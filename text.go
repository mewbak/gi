// Copyright (c) 2018, Randall C. O'Reilly. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gi

import (
	// "fmt"
	"github.com/rcoreilly/goki/ki"
	"image"
	"log"
	"strings"
	"unicode"
)

type TextAlign int

const (
	TextAlignLeft TextAlign = iota
	TextAlignRight
	TextAlignCenter
	TextAlignJustify
)

//go:generate stringer -type=TextAlign

// all the style information associated with how to render text
type TextStyle struct {
	Align         TextAlign `xml:"text-align",desc:"how to align text"`
	LineHeight    float64   `xml:"line-height",desc:"specified height of a line of text 0 = normal"`
	LineSpacing   float64   `xml:"line-spacing",desc:"spacing between lines -- NOTE: line height is used instead of spacing in css / html"`
	Indent        float64   `xml:"text-indent",desc:"how much to indent the first line in a paragraph"`
	LetterSpacing float64   `xml:"letter-spacing",desc:"spacing between characters and lines"`
	WordSpacing   float64   `xml:"word-spacing",desc:"extra space to add between words"`
	WordWrap      bool      `xml:"word-wrap",desc:"wrap text within a given size"`
	// todo:
	// page-break options
	// text-decoration-line -- underline, overline, line-through, -style, -color
	// text-justify -- how to justify text
	// text-overflow -- clip, ellipsis, string..
	// text-shadow
	// text-transform -- uppercase, lowercase, capitalize
	// user-select -- can user select text?
	// white-space -- what to do with white-space
	// word-break
}

func (p *TextStyle) Defaults() {
	p.WordWrap = false
	p.Align = TextAlignLeft
	p.LineSpacing = 1.0
	p.LetterSpacing = 1.0
}

// update the font settings from the style info on the node
func (pt *TextStyle) SetFromNode(g *Node2DBase) {
	// always check if property has been set before setting -- otherwise defaults to empty -- true = inherit props

	if wr, got := g.GiPropBool("word-wrap"); got { // gi version
		pt.WordWrap = wr
	}
	if sz, got := g.PropNumber("line-spacing"); got {
		pt.LineSpacing = sz
	}
	if es, got := g.PropEnum("text-align"); got {
		var al TextAlign = -1
		switch es { // first go through short-hand codes
		case "left":
			al = TextAlignLeft
		case "start":
			al = TextAlignLeft
		case "center":
			al = TextAlignCenter
		case "right":
			al = TextAlignRight
		case "end":
			al = TextAlignRight
		}
		if al == -1 {
			i, err := StringToTextAlign(es) // stringer gen
			if err != nil {
				pt.Align = i
			} else {
				log.Print(err)
			}
		} else {
			pt.Align = al
		}
	}
}

////////////////////////////////////////////////////////////////////////////////////////
// Text2D Node

// todo: lots of work likely needed on laying-out text in proper way
// https://www.w3.org/TR/SVG2/text.html#GlyphsMetrics
// todo: tspan element

// 2D Text
type Text2D struct {
	Node2DBase
	Text        string   `xml:"text",desc:"text string to render"`
	WrappedText []string `json:"-","desc:word-wrapped version of the string"`
}

// must register all new types so type names can be looked up by name -- e.g., for json
var KiT_Text2D = ki.KiTypes.AddType(&Text2D{})

func (g *Text2D) GiNode2D() *Node2DBase {
	return &g.Node2DBase
}

func (g *Text2D) GiViewport2D() *Viewport2D {
	return nil
}

func (g *Text2D) InitNode2D() {
	g.Layout.Defaults()
}

func (g *Text2D) PaintProps2D() {
	// pc := &g.MyPaint
	// if pc.HasNoStrokeOrFill() || len(g.Text) == 0 {
	// 	pc.Off = true
	// }
}

func (g *Text2D) Layout2D(iter int) {
	if iter == 0 {
		pc := &g.MyPaint
		var w, h float64
		// pre-wrap the text
		if pc.Text.WordWrap { // todo: switch to LineHeight
			g.WrappedText, h = pc.MeasureStringWrapped(g.Text, g.Size.X, pc.Text.LineSpacing)
		} else {
			w, h = pc.MeasureString(g.Text)
		}
		g.Layout.AllocSize = Size2D{w, h}
	}
}

func (g *Text2D) Node2DBBox() image.Rectangle {
	return g.MyPaint.BoundingBox(g.Pos.X, g.Pos.Y, g.Pos.X+g.Layout.AllocSize.X, g.Pos.Y+g.Layout.AllocSize.Y)
}

func (g *Text2D) Render2D() {
	g.SetWinBBox(g.Node2DBBox())
	// fmt.Printf("rendering text %v\n", g.Text)
	pc := &g.MyPaint
	rs := &g.Viewport.Render
	if pc.Text.WordWrap {
		pc.DrawStringLines(rs, g.WrappedText, g.Pos.X, g.Pos.Y, g.Layout.AllocSize.X,
			g.Layout.AllocSize.Y)
	} else {
		pc.DrawString(rs, g.Text, g.Pos.X, g.Pos.Y, g.Layout.AllocSize.X)
	}
}

func (g *Text2D) CanReRender2D() bool {
	// todo: could optimize by checking for an opaque fill, and same bbox
	return false
}

// check for interface implementation
var _ Node2D = &Text2D{}

//////////////////////////////////////////////////////////////////////////////////
//  Utilities

type measureStringer interface {
	MeasureString(s string) (w, h float64)
}

func splitOnSpace(x string) []string {
	var result []string
	pi := 0
	ps := false
	for i, c := range x {
		s := unicode.IsSpace(c)
		if s != ps && i > 0 {
			result = append(result, x[pi:i])
			pi = i
		}
		ps = s
	}
	result = append(result, x[pi:])
	return result
}

func wordWrap(m measureStringer, s string, width float64) []string {
	var result []string
	for _, line := range strings.Split(s, "\n") {
		fields := splitOnSpace(line)
		if len(fields)%2 == 1 {
			fields = append(fields, "")
		}
		x := ""
		for i := 0; i < len(fields); i += 2 {
			w, _ := m.MeasureString(x + fields[i])
			if w > width {
				if x == "" {
					result = append(result, fields[i])
					x = ""
					continue
				} else {
					result = append(result, x)
					x = ""
				}
			}
			x += fields[i] + fields[i+1]
		}
		if x != "" {
			result = append(result, x)
		}
	}
	for i, line := range result {
		result[i] = strings.TrimSpace(line)
	}
	return result
}