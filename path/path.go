// Copyright 2010 The draw2d Authors. All rights reserved.
// created: 21/11/2010 by Laurent Le Goff

// Package path implements function to build path
package path

import (
	"fmt"
	"github.com/llgcode/draw2d/curve"
	"math"
)

// PathBuilder define method that create path
type PathBuilder interface {
	// Return the current point of the current path
	LastPoint() (x, y float64)

	// MoveTo start a new path at (x, y) position
	MoveTo(x, y float64)

	// LineTo add a line to the current path
	LineTo(x, y float64)

	// QuadCurveTo add a quadratic curve to the current path
	QuadCurveTo(cx, cy, x, y float64)

	// CubicCurveTo add a cubic bezier curve to the current path
	CubicCurveTo(cx1, cy1, cx2, cy2, x, y float64)

	// ArcTo add an arc to the path
	ArcTo(cx, cy, rx, ry, startAngle, angle float64)

	// Close the current path
	Close()
}

// Component represent components of a path
type Component int

const (
	MoveToCmp Component = iota
	LineToCmp
	QuadCurveToCmp
	CubicCurveToCmp
	ArcToCmp
	CloseCmp
)

// Type Path store path
type Path struct {
	Components []Component
	Points     []float64
	x, y       float64
}

func (p *Path) Clear() {
	p.Components = p.Components[0:0]
	p.Points = p.Points[0:0]
	return
}

func (p *Path) appendToPath(cmd Component, points ...float64) {
	p.Components = append(p.Components, cmd)
	p.Points = append(p.Points, points...)
}

// Copy make a clone of the current path and return it
func (src *Path) Copy() (dest *Path) {
	dest = new(Path)
	dest.Components = make([]Component, len(src.Components))
	copy(dest.Components, src.Components)
	dest.Points = make([]float64, len(src.Points))
	copy(dest.Points, src.Points)
	dest.x, dest.y = src.x, src.y
	return dest
}

func (p *Path) LastPoint() (x, y float64) {
	return p.x, p.y
}

func (p *Path) IsEmpty() bool {
	return len(p.Components) == 0
}

func (p *Path) Close() {
	p.appendToPath(CloseCmp)
}

func (p *Path) MoveTo(x, y float64) {
	p.appendToPath(MoveToCmp, x, y)

	p.x = x
	p.y = y
}

func (p *Path) LineTo(x, y float64) {
	if len(p.Components) == 0 { //special case when no move has been done
		p.MoveTo(0, 0)
	}
	p.appendToPath(LineToCmp, x, y)
	p.x = x
	p.y = y
}

func (p *Path) QuadCurveTo(cx, cy, x, y float64) {
	if len(p.Components) == 0 { //special case when no move has been done
		p.MoveTo(0, 0)
	}
	p.appendToPath(QuadCurveToCmp, cx, cy, x, y)
	p.x = x
	p.y = y
}

func (p *Path) CubicCurveTo(cx1, cy1, cx2, cy2, x, y float64) {
	if len(p.Components) == 0 { //special case when no move has been done
		p.MoveTo(0, 0)
	}
	p.appendToPath(CubicCurveToCmp, cx1, cy1, cx2, cy2, x, y)
	p.x = x
	p.y = y
}

func (p *Path) ArcTo(cx, cy, rx, ry, startAngle, angle float64) {
	endAngle := startAngle + angle
	clockWise := true
	if angle < 0 {
		clockWise = false
	}
	// normalize
	if clockWise {
		for endAngle < startAngle {
			endAngle += math.Pi * 2.0
		}
	} else {
		for startAngle < endAngle {
			startAngle += math.Pi * 2.0
		}
	}
	startX := cx + math.Cos(startAngle)*rx
	startY := cy + math.Sin(startAngle)*ry
	if len(p.Components) > 0 {
		p.LineTo(startX, startY)
	} else {
		p.MoveTo(startX, startY)
	}
	p.appendToPath(ArcToCmp, cx, cy, rx, ry, startAngle, angle)
	p.x = cx + math.Cos(endAngle)*rx
	p.y = cy + math.Sin(endAngle)*ry
}

func (p *Path) String() string {
	s := ""
	j := 0
	for _, cmd := range p.Components {
		switch cmd {
		case MoveToCmp:
			s += fmt.Sprintf("MoveTo: %f, %f\n", p.Points[j], p.Points[j+1])
			j = j + 2
		case LineToCmp:
			s += fmt.Sprintf("LineTo: %f, %f\n", p.Points[j], p.Points[j+1])
			j = j + 2
		case QuadCurveToCmp:
			s += fmt.Sprintf("QuadCurveTo: %f, %f, %f, %f\n", p.Points[j], p.Points[j+1], p.Points[j+2], p.Points[j+3])
			j = j + 4
		case CubicCurveToCmp:
			s += fmt.Sprintf("CubicCurveTo: %f, %f, %f, %f, %f, %f\n", p.Points[j], p.Points[j+1], p.Points[j+2], p.Points[j+3], p.Points[j+4], p.Points[j+5])
			j = j + 6
		case ArcToCmp:
			s += fmt.Sprintf("ArcTo: %f, %f, %f, %f, %f, %f\n", p.Points[j], p.Points[j+1], p.Points[j+2], p.Points[j+3], p.Points[j+4], p.Points[j+5])
			j = j + 6
		case CloseCmp:
			s += "Close\n"
		}
	}
	return s
}

// Flatten convert curves in straight segments keeping join segements
func (path *Path) Flatten(liner LineBuilder, scale float64) {
	// First Point
	var startX, startY float64 = 0, 0
	// Current Point
	var x, y float64 = 0, 0
	i := 0
	for _, cmd := range path.Components {
		switch cmd {
		case MoveToCmp:
			x, y = path.Points[i], path.Points[i+1]
			startX, startY = x, y
			if i != 0 {
				liner.End()
			}
			liner.MoveTo(x, y)
			i += 2
		case LineToCmp:
			x, y = path.Points[i], path.Points[i+1]
			liner.LineTo(x, y)
			liner.LineJoin()
			i += 2
		case QuadCurveToCmp:
			curve.TraceQuad(liner, path.Points[i-2:], 0.5)
			x, y = path.Points[i+2], path.Points[i+3]
			liner.LineTo(x, y)
			i += 4
		case CubicCurveToCmp:
			curve.TraceCubic(liner, path.Points[i-2:], 0.5)
			x, y = path.Points[i+4], path.Points[i+5]
			liner.LineTo(x, y)
			i += 6
		case ArcToCmp:
			x, y = curve.TraceArc(liner, path.Points[i], path.Points[i+1], path.Points[i+2], path.Points[i+3], path.Points[i+4], path.Points[i+5], scale)
			liner.LineTo(x, y)
			i += 6
		case CloseCmp:
			liner.LineTo(startX, startY)
			liner.Close()
		}
	}
	liner.End()
}