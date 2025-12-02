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

type Edge struct {
	From, To      string
	Bidirectional bool
}

type Scenario struct {
	Title    string
	Subtitle string
	Nodes    []string
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

	for ab := 0; ab < 4; ab++ {
		for cPat := 0; cPat < 4; cPat++ {
			for dPat := 0; dPat < 4; dPat++ {
				title := abTitle(ab)
				subtitle := externalSubtitle(cPat, dPat)

				nodesSet := map[string]bool{
					"A": true,
					"B": true,
				}
				var edges []Edge

				// A-B edges
				switch ab {
				case 0:
					// none
				case 1:
					edges = append(edges, Edge{"A", "B", false})
				case 2:
					edges = append(edges, Edge{"B", "A", false})
				case 3:
					edges = append(edges, Edge{"A", "B", true}) // mutualism
				}

				// C edges
				if cPat != 0 {
					nodesSet["C"] = true
					if cPat == 1 || cPat == 3 {
						edges = append(edges, Edge{"C", "A", false})
					}
					if cPat == 2 || cPat == 3 {
						edges = append(edges, Edge{"C", "B", false})
					}
				}

				// D edges
				if dPat != 0 {
					nodesSet["D"] = true
					if dPat == 1 || dPat == 3 {
						edges = append(edges, Edge{"D", "A", false})
					}
					if dPat == 2 || dPat == 3 {
						edges = append(edges, Edge{"D", "B", false})
					}
				}

				// Stable ordering for nicer layouts
				order := []string{"C", "D", "A", "B"}
				var nodes []string
				for _, name := range order {
					if nodesSet[name] {
						nodes = append(nodes, name)
					}
				}

				scenarios = append(scenarios, Scenario{
					Title:    title,
					Subtitle: subtitle,
					Nodes:    nodes,
					Edges:    edges,
				})
			}
		}
	}
	return scenarios
}

func abTitle(ab int) string {
	switch ab {
	case 0:
		return "A & B: no direct link"
	case 1:
		return "A → B"
	case 2:
		return "B → A"
	case 3:
		return "A ↔ B (mutualism)"
	default:
		return "A/B pattern ?"
	}
}

func externalSubtitle(cPat, dPat int) string {
	return fmt.Sprintf("C %s; D %s",
		externalSentenceFragment("C", cPat),
		externalSentenceFragment("D", dPat),
	)
}

func externalSentenceFragment(role string, p int) string {
	switch p {
	case 0:
		return "has no effect on A or B"
	case 1:
		return "influences A only"
	case 2:
		return "influences B only"
	case 3:
		return "influences both A and B"
	default:
		return "?"
	}
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
	drawArrow(img, sx1, sy1, sx2, sy2, color.Black)
	drawLabel(img, "Single arrow: influence (e.g. C → A)", sx2+10, sy1+4, color.Black)

	// --- Section 2: mutualism ---
	s2x := x0 + sectionW
	s2y := s1y
	drawLabel(img, "Mutualism", s2x, s2y-8, color.RGBA{40, 40, 40, 255})

	mx1, my1 := s2x+10, s2y
	mx2, my2 := mx1+60, my1
	drawArrow(img, mx1, my1-3, mx2, my2-3, color.Black)
	drawArrow(img, mx2, my2+3, mx1, my1+3, color.Black)
	drawLabel(img, "Double arrow: mutualism (A ↔ B)", mx2+10, my1+4, color.Black)

	// --- Section 3: chronology ---
	s3x := x0 + 2*sectionW
	s3y := s1y
	drawLabel(img, "Chronology", s3x, s3y-8, color.RGBA{40, 40, 40, 255})
	drawLabel(img, "Within each panel:", s3x+10, s3y+10, color.Black)
	drawLabel(img, "Upper row = earlier (no incoming arrows)", s3x+10, s3y+30, color.RGBA{60, 60, 60, 255})
	drawLabel(img, "Lower row = later (influenced by others)", s3x+10, s3y+46, color.RGBA{60, 60, 60, 255})
}

// Within a panel, we infer simple chronology from the graph:
// - nodes with no incoming arrows are "earlier" (upper row)
// - nodes with at least one incoming arrow are "later" (lower row)
// This means A and B don't have to be simultaneous or last, and in
// mutualism-only cases (A ↔ B) they appear on the same row.
func drawScenario(img *image.RGBA, rect image.Rectangle, s Scenario) {
	bg := color.RGBA{255, 255, 255, 255}
	border := color.RGBA{180, 180, 180, 255}
	fillRect(img, rect, bg)
	drawRectBorder(img, rect, border)

	// Title & subtitle
	textX := rect.Min.X + 10
	maxTextWidth := rect.Dx() - 20
	titleHeight := drawWrappedLabel(img, s.Title, textX, rect.Min.Y+22, maxTextWidth, color.RGBA{20, 20, 20, 255})
	subtitleY := rect.Min.Y + 22 + titleHeight + 6
	subtitleHeight := drawWrappedLabel(img, s.Subtitle, textX, subtitleY, maxTextWidth, color.RGBA{80, 80, 80, 255})
	extraTextHeight := (titleHeight - lineHeight) + (subtitleHeight - lineHeight)
	if extraTextHeight < 0 {
		extraTextHeight = 0
	}

	// Layout rows
	left := rect.Min.X + 40
	right := rect.Max.X - 40
	topY := rect.Min.Y + 90 + extraTextHeight  // more recent
	botY := rect.Min.Y + 170 + extraTextHeight // later

	// Compute incoming edge counts
	incoming := map[string]int{}
	for _, n := range s.Nodes {
		incoming[n] = 0
	}
	for _, e := range s.Edges {
		incoming[e.To]++
		if e.Bidirectional {
			// mutualism: treat as two directed edges for layering
			incoming[e.From]++
		}
	}

	var early, late []string
	for _, n := range s.Nodes {
		if incoming[n] == 0 {
			early = append(early, n)
		} else {
			late = append(late, n)
		}
	}

	// Fallbacks: if graph is fully cyclic or fully independent,
	// put everything in the upper row.
	if len(early) == 0 {
		early = s.Nodes
		late = nil
	}

	positions := map[string]image.Point{}

	// Position early nodes
	if len(early) == 1 {
		positions[early[0]] = image.Point{(left + right) / 2, topY}
	} else if len(early) > 1 {
		for i, name := range early {
			x := left + (right-left)*i/(len(early)-1)
			positions[name] = image.Point{x, topY}
		}
	}

	// Position late nodes
	if len(late) == 1 {
		positions[late[0]] = image.Point{(left + right) / 2, botY}
	} else if len(late) > 1 {
		for i, name := range late {
			x := left + (right-left)*i/(len(late)-1)
			positions[name] = image.Point{x, botY}
		}
	}

	// Fallback for any missing position
	for _, name := range s.Nodes {
		if _, ok := positions[name]; !ok {
			positions[name] = image.Point{(left + right) / 2, (topY + botY) / 2}
		}
	}

	// Draw edges first
	for _, e := range s.Edges {
		from := positions[e.From]
		to := positions[e.To]
		if e.Bidirectional {
			drawBidirectionalArrow(img, from.X, from.Y, to.X, to.Y, color.RGBA{0, 0, 0, 255})
		} else {
			// Single arrow for unidirectional influence
			drawArrow(img, from.X, from.Y, to.X, to.Y, color.RGBA{0, 0, 0, 255})
		}
	}

	// Draw nodes on top
	nodeFill := color.RGBA{220, 235, 250, 255}
	nodeBorder := color.RGBA{20, 40, 120, 255}
	for _, name := range s.Nodes {
		pt := positions[name]
		drawNode(img, pt.X, pt.Y, 20, nodeFill, nodeBorder)
		drawLabel(img, name, pt.X-5, pt.Y+5, color.RGBA{0, 0, 0, 255})
	}
}

// ----------------------------------------------------------------------
// Drawing helpers
// ----------------------------------------------------------------------

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

func drawArrow(img *image.RGBA, x0, y0, x1, y1 int, col color.Color) {
	const nodeRadius = 20.0

	dx := float64(x1 - x0)
	dy := float64(y1 - y0)
	dist := math.Hypot(dx, dy)
	if dist == 0 {
		return
	}

	ux := dx / dist
	uy := dy / dist

	// shorten line so it meets node edges
	tailX := float64(x0) + ux*nodeRadius
	tailY := float64(y0) + uy*nodeRadius
	headX := float64(x1) - ux*nodeRadius
	headY := float64(y1) - uy*nodeRadius

	drawLine(img, int(tailX), int(tailY), int(headX), int(headY), col)

	// arrowhead
	arrowLen := 10.0
	perpX := -uy
	perpY := ux

	hx := headX
	hy := headY

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

func drawBidirectionalArrow(img *image.RGBA, x0, y0, x1, y1 int, col color.Color) {
	const nodeRadius = 20.0

	dx := float64(x1 - x0)
	dy := float64(y1 - y0)
	dist := math.Hypot(dx, dy)
	if dist == 0 {
		return
	}

	ux := dx / dist
	uy := dy / dist

	// shorten line so it meets node edges
	tailX := float64(x0) + ux*nodeRadius
	tailY := float64(y0) + uy*nodeRadius
	headX := float64(x1) - ux*nodeRadius
	headY := float64(y1) - uy*nodeRadius

	drawLine(img, int(tailX), int(tailY), int(headX), int(headY), col)

	// arrowhead setup
	arrowLen := 10.0
	perpX := -uy
	perpY := ux

	// arrowhead at (x1, y1) end
	hx1 := headX
	hy1 := headY
	p2x1 := hx1 - ux*arrowLen + perpX*(arrowLen/2)
	p2y1 := hy1 - uy*arrowLen + perpY*(arrowLen/2)
	p3x1 := hx1 - ux*arrowLen - perpX*(arrowLen/2)
	p3y1 := hy1 - uy*arrowLen - perpY*(arrowLen/2)
	fillTriangle(img, int(hx1), int(hy1), int(p2x1), int(p2y1), int(p3x1), int(p3y1), col)

	// arrowhead at (x0, y0) end
	hx2 := tailX
	hy2 := tailY
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
