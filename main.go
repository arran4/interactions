// interactions: generate a grid of all basic interaction patterns
// between A and B, with external influences from C and D.
// Project home: https://github.com/arran4/interactions
package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"log"
	"math"
	"os"
	"strings"

	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

type Node struct {
	Name      string
	Y         int  // Vertical layer, 0 is top.
	IsProcess bool // true for rectangle (duration), false for circle (event)
}

type Edge struct {
	From, To      string
	Bidirectional bool
}

type Scenario struct {
	Title    string
	Subtitle string
	Nodes    []Node
	Edges    []Edge
}

func main() {
	if err := run(os.Args[1:]); err != nil {
		log.Fatal(err)
	}
}

func run(args []string) error {
	if len(args) == 0 {
		printGlobalUsage()
		return nil
	}

	switch args[0] {
	case "render":
		return runRender(args[1:])
	case "list":
		return runList(args[1:])
	case "help", "--help", "-h":
		printGlobalUsage()
		return nil
	default:
		printGlobalUsage()
		return fmt.Errorf("unknown subcommand %q", args[0])
	}
}

func runRender(args []string) error {
	fs := flag.NewFlagSet("render", flag.ContinueOnError)
	output := fs.String("output", "interactions.png", "path to write the generated PNG")
	columns := fs.Int("columns", 8, "number of columns in the grid (use 3 for README-friendly long form)")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if *columns < 1 {
		return fmt.Errorf("columns must be at least 1")
	}

	scenarios := generateScenarios()
	renderAllScenarios(*output, scenarios, *columns)
	return nil
}

func runList(args []string) error {
	fs := flag.NewFlagSet("list", flag.ContinueOnError)
	longForm := fs.Bool("long", false, "print subtitles along with scenario titles")
	if err := fs.Parse(args); err != nil {
		return err
	}

	scenarios := generateScenarios()
	for i, s := range scenarios {
		if *longForm {
			fmt.Printf("%02d. %s — %s\n", i+1, s.Title, s.Subtitle)
			continue
		}
		fmt.Printf("%02d. %s\n", i+1, s.Title)
	}
	return nil
}

func printGlobalUsage() {
	fmt.Println("Usage: interactions <command> [options]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  render   Generate the interactions grid PNG (use --output to set the destination)")
	fmt.Println("  list     List scenario titles (use --long to include subtitles)")
	fmt.Println("  help     Show this help text")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  go run main.go render --output interactions.png")
	fmt.Println("  go run main.go render --columns 3 --output interactions-long.png")
	fmt.Println("  go run main.go list --long")
}

// ----------------------------------------------------------------------
// Scenario generation: all combinations
// ----------------------------------------------------------------------
//
// AB pattern codes:
// 0 = no direct link
// 1 = A -> B
// 2 = B -> A
// 3 = A <-> B (mutualism)
//
// External pattern codes for C and D:
// 0 = no edges
// 1 = -> A only
// 2 = -> B only
// 3 = -> A and B
func generateScenarios() []Scenario {
	var scenarios []Scenario
	abPatterns := []string{"none", "A->B", "B->A", "A<->B"}
	cPatterns := []string{"none", "C->A", "C->B", "C->A,B"}
	dPatterns := []string{"none", "D->A", "D->B", "D->A,B"}
	timePatterns := []string{"A,B same", "A before B", "B before A"}
	typePatterns := []string{"A,B events", "A event, B process", "A process, B event", "A,B processes"}

	for _, abPat := range abPatterns {
		for _, cPat := range cPatterns {
			for _, dPat := range dPatterns {
				for _, timePat := range timePatterns {
					for _, typePat := range typePatterns {
						s := Scenario{
							Title:    fmt.Sprintf("AB: %s, Time: %s, Type: %s", abPat, timePat, typePat),
							Subtitle: fmt.Sprintf("C: %s, D: %s", cPat, dPat),
						}

						var aNode, bNode Node
						aNode.Name = "A"
						bNode.Name = "B"

						switch timePat {
						case "A,B same":
							aNode.Y, bNode.Y = 0, 0
						case "A before B":
							aNode.Y, bNode.Y = 0, 1
						case "B before A":
							aNode.Y, bNode.Y = 1, 0
						}

						switch typePat {
						case "A,B events":
							aNode.IsProcess, bNode.IsProcess = false, false
						case "A event, B process":
							aNode.IsProcess, bNode.IsProcess = false, true
						case "A process, B event":
							aNode.IsProcess, bNode.IsProcess = true, false
						case "A,B processes":
							aNode.IsProcess, bNode.IsProcess = true, true
						}
						s.Nodes = append(s.Nodes, aNode, bNode)

						switch abPat {
						case "A->B":
							s.Edges = append(s.Edges, Edge{From: "A", To: "B"})
						case "B->A":
							s.Edges = append(s.Edges, Edge{From: "B", To: "A"})
						case "A<->B":
							s.Edges = append(s.Edges, Edge{From: "A", To: "B", Bidirectional: true})
						}

						if cPat != "none" {
							s.Nodes = append(s.Nodes, Node{Name: "C", Y: 0, IsProcess: false})
							if cPat == "C->A" || cPat == "C->A,B" {
								s.Edges = append(s.Edges, Edge{From: "C", To: "A"})
							}
							if cPat == "C->B" || cPat == "C->A,B" {
								s.Edges = append(s.Edges, Edge{From: "C", To: "B"})
							}
						}
						if dPat != "none" {
							s.Nodes = append(s.Nodes, Node{Name: "D", Y: 0, IsProcess: false})
							if dPat == "D->A" || dPat == "D->A,B" {
								s.Edges = append(s.Edges, Edge{From: "D", To: "A"})
							}
							if dPat == "D->B" || dPat == "D->A,B" {
								s.Edges = append(s.Edges, Edge{From: "D", To: "B"})
							}
						}

						scenarios = append(scenarios, s)

					}
				}
			}
		}
	}
	return scenarios
}


// ----------------------------------------------------------------------
// Rendering
// ----------------------------------------------------------------------

func renderAllScenarios(filename string, scenarios []Scenario, columns int) {
	const (
		panelW       = 360
		panelH       = 220
		margin       = 20
		titleHeight  = 50
		legendHeight = 120
	)

	cols := columns
	rows := (len(scenarios) + cols - 1) / cols

	imgW := cols*panelW + (cols+1)*margin
	imgH := titleHeight + legendHeight + rows*panelH + (rows+2)*margin

	canvas := image.NewRGBA(image.Rect(0, 0, imgW, imgH))
	fillRect(canvas, canvas.Bounds(), color.RGBA{240, 240, 240, 255})

	// Global title and repo URL
	mainTitle := "Interaction patterns of A and B with C and D (all basic combinations)"
	drawCenteredLabel(canvas, mainTitle, imgW/2, margin+18, color.RGBA{10, 10, 10, 255})
	drawCenteredLabel(canvas, "Source: github.com/arran4/interactions", imgW/2, margin+36, color.RGBA{60, 60, 60, 255})

	// Legend area under the title
	legendTop := margin + titleHeight
	legendRect := image.Rect(margin, legendTop, imgW-margin, legendTop+legendHeight)
	drawLegend(canvas, legendRect)

	// Panels below legend
	for i, s := range scenarios {
		colIndex := i % cols
		rowIndex := i / cols

		x := margin + colIndex*(panelW+margin)
		y := legendTop + legendHeight + margin + rowIndex*(panelH+margin)

		panel := image.Rect(x, y, x+panelW, y+panelH)
		drawScenario(canvas, panel, s)
	}

	f, err := os.Create(filename)
	if err != nil {
		log.Fatalf("failed to create output file: %v", err)
	}
	defer f.Close()

	if err := png.Encode(f, canvas); err != nil {
		log.Fatalf("failed to encode PNG: %v", err)
	}

	log.Println("Generated:", filename)
}

// Legend describing arrows, mutualism, chronology
// Laid out horizontally in three sections.
func drawLegend(img *image.RGBA, rect image.Rectangle) {
	bg := color.RGBA{255, 255, 255, 255}
	border := color.RGBA{120, 120, 120, 255}
	fillRect(img, rect, bg)
	drawRectBorder(img, rect, border)

	padding := 10
	x0 := rect.Min.X + padding
	y0 := rect.Min.Y + padding
	w := rect.Dx() - 2*padding
	sectionW := w / 3

	drawLabel(img, "Legend", x0, y0+12, color.RGBA{20, 20, 20, 255})

	// --- Section 1: single arrow ---
	s1x := x0
	s1y := y0 + 30
	drawLabel(img, "Influence", s1x, s1y-8, color.RGBA{40, 40, 40, 255})

	sx1, sy1 := s1x+10, s1y
	sx2, sy2 := sx1+60, sy1
	drawArrow(img, sx1, sy1, sx2, sy2, color.Black, false, false)
	drawLabel(img, "Single arrow: influence (e.g. C → A)", sx2+10, sy1+4, color.Black)

	// --- Section 2: mutualism ---
	s2x := x0 + sectionW
	s2y := s1y
	drawLabel(img, "Mutualism", s2x, s2y-8, color.RGBA{40, 40, 40, 255})

	mx1, my1 := s2x+10, s2y
	mx2, my2 := mx1+60, my1
	drawArrow(img, mx1, my1-3, mx2, my2-3, color.Black, false, false)
	drawArrow(img, mx2, my2+3, mx1, my1+3, color.Black, false, false)
	drawLabel(img, "Double arrow: mutualism (A ↔ B)", mx2+10, my1+4, color.Black)

	// --- Section 3: chronology ---
	s3x := x0 + 2*sectionW
	s3y := s1y
	drawLabel(img, "Chronology & Type", s3x, s3y-8, color.RGBA{40, 40, 40, 255})

	// Event vs Process
	drawNode(img, s3x+20, s3y+15, 10, color.White, color.Black)
	drawLabel(img, "Circle: event (instantaneous)", s3x+40, s3y+20, color.Black)
	drawProcess(img, s3x+20, s3y+45, 20, 10, color.White, color.Black)
	drawLabel(img, "Rectangle: process (duration)", s3x+40, s3y+50, color.Black)

	// Vertical position
	drawLabel(img, "Vertical position: sequence", s3x+10, s3y+70, color.Black)
	drawLabel(img, "(top = earlier, bottom = later)", s3x+10, s3y+85, color.RGBA{60, 60, 60, 255})
}

func drawScenario(img *image.RGBA, rect image.Rectangle, s Scenario) {
	bg := color.RGBA{255, 255, 255, 255}
	border := color.RGBA{180, 180, 180, 255}
	fillRect(img, rect, bg)
	drawRectBorder(img, rect, border)

	// Title & subtitle
	textX := rect.Min.X + 10
	maxTextWidth := rect.Dx() - 20
	drawWrappedLabel(img, s.Title, textX, rect.Min.Y+22, maxTextWidth, color.RGBA{20, 20, 20, 255})

	// Group nodes by Y level
	levels := make(map[int][]Node)
	maxY := 0
	for _, n := range s.Nodes {
		levels[n.Y] = append(levels[n.Y], n)
		if n.Y > maxY {
			maxY = n.Y
		}
	}

	// Vertical positioning
	yStep := (rect.Max.Y - rect.Min.Y - 80) / (maxY + 1)
	yOffset := rect.Min.Y + 80

	positions := make(map[string]image.Point)
	for yLevel, nodes := range levels {
		xStep := (rect.Max.X - rect.Min.X - 80) / len(nodes)
		xOffset := rect.Min.X + 40 + xStep/2
		for i, n := range nodes {
			positions[n.Name] = image.Point{
				X: xOffset + i*xStep,
				Y: yOffset + yLevel*yStep,
			}
		}
	}

	// Draw edges first
	nodesByName := make(map[string]Node)
	for _, n := range s.Nodes {
		nodesByName[n.Name] = n
	}
	for _, e := range s.Edges {
		from := positions[e.From]
		to := positions[e.To]
		fromNode := nodesByName[e.From]
		toNode := nodesByName[e.To]
		if e.Bidirectional {
			drawBidirectionalArrow(img, from.X, from.Y, to.X, to.Y, color.RGBA{0, 0, 0, 255}, fromNode.IsProcess, toNode.IsProcess)
		} else {
			drawArrow(img, from.X, from.Y, to.X, to.Y, color.RGBA{0, 0, 0, 255}, fromNode.IsProcess, toNode.IsProcess)
		}
	}

	// Draw nodes on top
	nodeFill := color.RGBA{220, 235, 250, 255}
	nodeBorder := color.RGBA{20, 40, 120, 255}
	for _, n := range s.Nodes {
		pt := positions[n.Name]
		if n.IsProcess {
			drawProcess(img, pt.X, pt.Y, 40, 20, nodeFill, nodeBorder)
		} else {
			drawNode(img, pt.X, pt.Y, 20, nodeFill, nodeBorder)
		}
		drawLabel(img, n.Name, pt.X-5, pt.Y+5, color.RGBA{0, 0, 0, 255})
	}
}

func drawProcess(img *image.RGBA, cx, cy, w, h int, fill, border color.Color) {
	rect := image.Rect(cx-w/2, cy-h/2, cx+w/2, cy+h/2)
	fillRect(img, rect, fill)
	drawRectBorder(img, rect, border)
}

// ----------------------------------------------------------------------
// Drawing helpers
// ----------------------------------------------------------------------

func intersectionPoint(x, y, otherX, otherY int, isProcess bool) image.Point {
	if !isProcess {
		const nodeRadius = 20.0
		dx := float64(otherX - x)
		dy := float64(otherY - y)
		dist := math.Hypot(dx, dy)
		if dist == 0 {
			return image.Point{x, y}
		}
		return image.Point{
			X: x + int(dx/dist*nodeRadius),
			Y: y + int(dy/dist*nodeRadius),
		}
	}

	const (
		w = 40
		h = 20
	)
	dx := float64(otherX - x)
	dy := float64(otherY - y)

	// Check intersection with vertical edges
	if dx != 0 {
		t := float64(w/2) / math.Abs(dx)
		candY := float64(y) + t*dy
		if candY >= float64(y-h/2) && candY <= float64(y+h/2) {
			return image.Point{x + int(math.Copysign(float64(w/2), dx)), int(candY)}
		}
	}

	// Check intersection with horizontal edges
	if dy != 0 {
		t := float64(h/2) / math.Abs(dy)
		candX := float64(x) + t*dx
		if candX >= float64(x-w/2) && candX <= float64(x+w/2) {
			return image.Point{int(candX), y + int(math.Copysign(float64(h/2), dy))}
		}
	}

	return image.Point{x, y}
}

func fillRect(img *image.RGBA, r image.Rectangle, c color.Color) {
	draw.Draw(img, r, &image.Uniform{c}, image.Point{}, draw.Src)
}

func drawRectBorder(img *image.RGBA, r image.Rectangle, c color.Color) {
	for x := r.Min.X; x < r.Max.X; x++ {
		img.Set(x, r.Min.Y, c)
		img.Set(x, r.Max.Y-1, c)
	}
	for y := r.Min.Y; y < r.Max.Y; y++ {
		img.Set(r.Min.X, y, c)
		img.Set(r.Max.X-1, y, c)
	}
}

func drawLabel(img *image.RGBA, text string, x, y int, col color.Color) {
	d := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(col),
		Face: basicfont.Face7x13,
		Dot:  fixed.P(x, y),
	}
	d.DrawString(text)
}

const (
	approxCharWidth = 7
	lineHeight      = 14
)

// drawWrappedLabel renders text within a maximum width, wrapping at word
// boundaries. It returns the total height used so callers can adjust layouts.
func drawWrappedLabel(img *image.RGBA, text string, x, y, maxWidth int, col color.Color) int {
	text = strings.TrimSpace(text)
	if text == "" {
		return 0
	}

	words := strings.Fields(text)
	if len(words) == 0 {
		return 0
	}

	var lines []string
	line := words[0]
	for _, w := range words[1:] {
		if (len(line)+1+len(w))*approxCharWidth <= maxWidth {
			line += " " + w
			continue
		}
		lines = append(lines, line)
		line = w
	}
	lines = append(lines, line)

	for i, l := range lines {
		drawLabel(img, l, x, y+i*lineHeight, col)
	}

	return len(lines) * lineHeight
}

func drawCenteredLabel(img *image.RGBA, text string, centerX, y int, col color.Color) {
	// Approximate text width: ~7px per char for Face7x13
	width := len(text) * 7
	x := centerX - width/2
	drawLabel(img, text, x, y, col)
}

func drawNode(img *image.RGBA, cx, cy, r int, fill, border color.Color) {
	r2 := r * r
	for y := -r; y <= r; y++ {
		for x := -r; x <= r; x++ {
			if x*x+y*y <= r2 {
				img.Set(cx+x, cy+y, fill)
			}
		}
	}
	// outline
	for y := -r; y <= r; y++ {
		for x := -r; x <= r; x++ {
			d := x*x + y*y
			if d >= r2-2 && d <= r2+2 {
				img.Set(cx+x, cy+y, border)
			}
		}
	}
}

func drawArrow(img *image.RGBA, x0, y0, x1, y1 int, col color.Color, fromProcess, toProcess bool) {
	dx := float64(x1 - x0)
	dy := float64(y1 - y0)
	dist := math.Hypot(dx, dy)
	if dist == 0 {
		return
	}

	ux := dx / dist
	uy := dy / dist

	// shorten line so it meets node edges
	tail := intersectionPoint(x0, y0, x1, y1, fromProcess)
	head := intersectionPoint(x1, y1, x0, y0, toProcess)

	drawLine(img, tail.X, tail.Y, head.X, head.Y, col)

	// arrowhead
	arrowLen := 10.0
	perpX := -uy
	perpY := ux

	hx := float64(head.X)
	hy := float64(head.Y)

	p2x := hx - ux*arrowLen + perpX*(arrowLen/2)
	p2y := hy - uy*arrowLen + perpY*(arrowLen/2)
	p3x := hx - ux*arrowLen - perpX*(arrowLen/2)
	p3y := hy - uy*arrowLen - perpY*(arrowLen/2)

	fillTriangle(img,
		int(hx), int(hy),
		int(p2x), int(p2y),
		int(p3x), int(p3y),
		col,
	)
}

func drawBidirectionalArrow(img *image.RGBA, x0, y0, x1, y1 int, col color.Color, fromProcess, toProcess bool) {
	dx := float64(x1 - x0)
	dy := float64(y1 - y0)
	dist := math.Hypot(dx, dy)
	if dist == 0 {
		return
	}

	ux := dx / dist
	uy := dy / dist

	// shorten line so it meets node edges
	tail := intersectionPoint(x0, y0, x1, y1, fromProcess)
	head := intersectionPoint(x1, y1, x0, y0, toProcess)

	drawLine(img, tail.X, tail.Y, head.X, head.Y, col)

	// arrowhead setup
	arrowLen := 10.0
	perpX := -uy
	perpY := ux

	// arrowhead at (x1, y1) end
	hx1 := float64(head.X)
	hy1 := float64(head.Y)
	p2x1 := hx1 - ux*arrowLen + perpX*(arrowLen/2)
	p2y1 := hy1 - uy*arrowLen + perpY*(arrowLen/2)
	p3x1 := hx1 - ux*arrowLen - perpX*(arrowLen/2)
	p3y1 := hy1 - uy*arrowLen - perpY*(arrowLen/2)
	fillTriangle(img, int(hx1), int(hy1), int(p2x1), int(p2y1), int(p3x1), int(p3y1), col)

	// arrowhead at (x0, y0) end
	hx2 := float64(tail.X)
	hy2 := float64(tail.Y)
	p2x2 := hx2 + ux*arrowLen + perpX*(arrowLen/2)
	p2y2 := hy2 + uy*arrowLen + perpY*(arrowLen/2)
	p3x2 := hx2 + ux*arrowLen - perpX*(arrowLen/2)
	p3y2 := hy2 + uy*arrowLen - perpY*(arrowLen/2)
	fillTriangle(img, int(hx2), int(hy2), int(p2x2), int(p2y2), int(p3x2), int(p3y2), col)
}

func drawLine(img *image.RGBA, x0, y0, x1, y1 int, col color.Color) {
	dx := abs(x1 - x0)
	sx := 1
	if x0 > x1 {
		sx = -1
	}
	dy := -abs(y1 - y0)
	sy := 1
	if y0 > y1 {
		sy = -1
	}
	err := dx + dy

	for {
		img.Set(x0, y0, col)
		if x0 == x1 && y0 == y1 {
			break
		}
		e2 := 2 * err
		if e2 >= dy {
			err += dy
			x0 += sx
		}
		if e2 <= dx {
			err += dx
			y0 += sy
		}
	}
}

func fillTriangle(img *image.RGBA, x1, y1, x2, y2, x3, y3 int, col color.Color) {
	minX := min(x1, min(x2, x3))
	maxX := max(x1, max(x2, x3))
	minY := min(y1, min(y2, y3))
	maxY := max(y1, max(y2, y3))

	for y := minY; y <= maxY; y++ {
		for x := minX; x <= maxX; x++ {
			if pointInTriangle(x, y, x1, y1, x2, y2, x3, y3) {
				img.Set(x, y, col)
			}
		}
	}
}

func pointInTriangle(px, py, x1, y1, x2, y2, x3, y3 int) bool {
	dx := float64(px)
	dy := float64(py)

	ax := float64(x1)
	ay := float64(y1)
	bx := float64(x2)
	by := float64(y2)
	cx := float64(x3)
	cy := float64(y3)

	v0x := cx - ax
	v0y := cy - ay
	v1x := bx - ax
	v1y := by - ay
	v2x := dx - ax
	v2y := dy - ay

	dot00 := v0x*v0x + v0y*v0y
	dot01 := v0x*v1x + v0y*v1y
	dot02 := v0x*v2x + v0y*v2y
	dot11 := v1x*v1x + v1y*v1y
	dot12 := v1x*v2x + v1y*v2y

	denom := dot00*dot11 - dot01*dot01
	if denom == 0 {
		return false
	}
	u := (dot11*dot02 - dot01*dot12) / denom
	v := (dot00*dot12 - dot01*dot02) / denom

	return u >= 0 && v >= 0 && u+v <= 1
}

// ----------------------------------------------------------------------
// small helpers
// ----------------------------------------------------------------------

func abs(a int) int {
	if a < 0 {
		return -a
	}
	return a
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
