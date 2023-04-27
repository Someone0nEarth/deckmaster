package main

import (
	"fmt"

	"image"
	"image/color"
	"os"

	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
)

// ImageElement for drawing into vertical equal segments considering the ImageElement bounds and number of Segments.
type ImageElement struct {
	img                *image.RGBA
	segmentHeight      uint
	segmentsInterspace uint
	dpi                uint
	fontColor          color.Color
	segmentPositionY   uint
	segments           []segment
}

type segment interface {
	draw(element *ImageElement)
}

type iconSegment struct {
	img image.Image
}

func (segment iconSegment) draw(element *ImageElement) {
	var iconSize uint
	if element.width() < element.segmentHeight {
		iconSize = element.width()
	} else {
		iconSize = element.segmentHeight
	}

	_ = drawImageWithResizing(element.img, segment.img, int(iconSize), image.Pt(-1, -1))
}

type percentageBarSegment struct {
	percentage uint8
}

func (segment percentageBarSegment) draw(element *ImageElement) {
	bar := createBar(element.width(), element.segmentHeight, segment.percentage)

	_ = drawImage(element.img, bar, image.Pt(0, int(element.segmentPositionY)))
}

type textSegment struct {
	text             string
	alignmentX       int
	centerVertically bool
}

func (segment textSegment) draw(element *ImageElement) {
	bounds := createRectangle(element.width(), element.segmentHeight)
	ttFont := ttfFont

	fontsize, actualWidth, _, descent := maxFontSize(ttFont, element.dpi, bounds.Dx(), bounds.Dy(), segment.text)

	var x int
	centerHorizontally := false

	if segment.alignmentX == 0 {
		x = 0
		centerHorizontally = true
	} else if segment.alignmentX == -1 {
		x = 0
	} else {
		x = int(element.width()) - actualWidth
	}

	y := int(element.segmentPositionY) + int(element.segmentHeight) - descent
	drawString2(element.img, bounds, ttFont, segment.text, element.dpi, fontsize, element.fontColor, image.Pt(x, y), centerHorizontally, segment.centerVertically)
}

type blankSegment struct {
}

func (segment blankSegment) draw(_ *ImageElement) {
	// Nothing to be done here
}

// NewImageElementWithStrings Will "divide" the element bounds into vertical equal segments considering the number of segments.
func NewImageElementWithStrings(bounds image.Rectangle, dpi uint, fontColor color.Color) *ImageElement {
	element := ImageElement{
		dpi:              dpi,
		fontColor:        fontColor,
		segmentPositionY: 0,
		segments:         make([]segment, 0, 3),
	}
	element.img = image.NewRGBA(image.Rect(0, 0, bounds.Dx(), bounds.Dy()))

	return &element
}

// NewImageElement Will "divide" the element bounds into vertical equal segments considering the number of segments.
func NewImageElement(bounds image.Rectangle) *ImageElement {
	element := ImageElement{
		dpi:              0,
		fontColor:        color.Transparent,
		segmentPositionY: 0,
		segments:         make([]segment, 0, 3),
	}
	element.img = image.NewRGBA(image.Rect(0, 0, bounds.Dx(), bounds.Dy()))

	return &element
}

// AddIconSegment will add the icon into the current segment slot, resize it to fit it and set the pointer to the next segment
func (element *ImageElement) AddIconSegment(icon image.Image) {
	element.segments = append(element.segments, iconSegment{
		img: icon,
	})
}

// AddTextSegment will add the text into the current segment slot and set the pointer to the next segment
func (element *ImageElement) AddTextSegment(text string, alignmentX int, centerVertically bool) {
	element.segments = append(element.segments, textSegment{
		text:             text,
		alignmentX:       alignmentX,
		centerVertically: centerVertically,
	})
}

// AddPercentageBarSegment will add a percentage bar into the current segment slot and set the pointer to the next segment
func (element *ImageElement) AddPercentageBarSegment(percentage uint8) {
	element.segments = append(element.segments, percentageBarSegment{
		percentage: percentage,
	})
}

// AddBlankSegment is for lay-outing. It will add a blank segment into the current segment slot and set the pointer to the next segment
func (element *ImageElement) AddBlankSegment() {
	element.segments = append(element.segments, blankSegment{})
}

// DrawElement
func (element *ImageElement) DrawElement(targetImage *image.RGBA, position image.Point) {
	element.calculateSegmentsInterspace()
	element.calculateSegmentSize()

	for _, segment := range element.segments {
		segment.draw(element)
		element.incrementSegmentPositionY()
	}

	_ = drawImage(targetImage, element.img, position)

	element.debugDrawInOnOutLines(targetImage, position)
}

func (element *ImageElement) height() uint {
	return uint(element.img.Bounds().Dy())
}

func (element *ImageElement) width() uint {
	return uint(element.img.Bounds().Dx())
}

func (element *ImageElement) incrementSegmentPositionY() {
	element.segmentPositionY += element.segmentHeight + element.segmentsInterspace
}

func (element *ImageElement) calculateSegmentsInterspace() {
	interspace := element.height() / (3 * element.numberOfSegments())
	if interspace < 5 {
		interspace = 5
	}
	element.segmentsInterspace = interspace
}

func (element *ImageElement) calculateSegmentSize() {
	element.segmentHeight = (element.height() - ((element.numberOfSegments() - 1) * element.segmentsInterspace)) / element.numberOfSegments()
}

func (element *ImageElement) debugDrawInOnOutLines(targetImage *image.RGBA, elementPosition image.Point) {
	//TODO Implementing a debug log routine
	if true {
		if false {
			bounds := element.img.Bounds()
			elementBounds := bounds.Bounds().Add(elementPosition)

			debugDrawOutline(targetImage, elementBounds)
			debugDrawOnLine(targetImage, elementBounds)
			debugDrawInline(targetImage, elementBounds)
		}

		if false {
			segmentPosition := uint(0)
			for segment := uint(0); segment < element.numberOfSegments(); segment++ {
				segmentBounds := image.Rectangle{
					Min: image.Point{
						X: 0,
						Y: int(segmentPosition),
					},
					Max: image.Point{
						X: int(element.width()),
						Y: int(segmentPosition) + int(element.segmentHeight),
					},
				}

				debugDrawOutline(targetImage, segmentBounds.Add(elementPosition))
				debugDrawOnLine(targetImage, segmentBounds.Add(elementPosition))
				debugDrawInline(targetImage, segmentBounds.Add(elementPosition))

				segmentPosition += element.segmentHeight + element.segmentsInterspace
			}
		}
	}
}

func (element *ImageElement) numberOfSegments() uint {
	return uint(len(element.segments))
}

func maxFontSize(ttFont *truetype.Font, dpi uint, width, height int, text string) (float64, int, int, int) {
	startingFontsize := float64(height) / (float64(dpi) / (72.0))
	initialFontsize := startingFontsize * 1.4

	_, _, heightFittingFontsize := determineHeightFittingFontsize(dpi, ttFont, initialFontsize, height, text)

	actualWidth, fontsize := determineWidthFittingFontsize(dpi, ttFont, heightFittingFontsize, width, text)

	actualHeight, ascent, descent := actualStringHeight(float64(dpi), ttFont, fontsize, text)

	//TODO Implementing a debug log routine
	if false {
		approximateHeight := approximatedMaxFontHeight(dpi, ttFont, fontsize)
		maxHeight, _, _ := maxFontHeight(float64(dpi), ttFont, fontsize)

		fmt.Printf("String '%s'\n\n", text)

		fmt.Printf(
			"Given %dw x %dh\n"+
				"String %dw x %dh\n"+
				"start %.1f, ini %.1f, fit H %.1f, fit W %.1f\n"+
				"asc %d, desc %d\n"+
				"approxi: %dh\n"+
				"max: %dh\n"+
				"\n",
			width,
			height,
			actualWidth,
			actualHeight,
			startingFontsize,
			initialFontsize,
			heightFittingFontsize,
			fontsize,
			ascent,
			descent,
			approximateHeight,
			maxHeight,
		)

		maxFontHeight2, maxAscent, maxDescent := maxFontHeight(float64(dpi), ttFont, initialFontsize)
		actualHeight2, ascent2, descent2 := actualStringHeight(float64(dpi), ttFont, initialFontsize, text)
		actualWidth2 := actualStringWidth(float64(dpi), ttFont, initialFontsize, text)

		fmt.Printf(
			"Font/string metrics with initial fontsize\n"+
				"String %dw x %dh\n"+
				"asc %d, desc %d\n"+
				"max: %dh, asc %d, desc %d\n"+
				"-----------\n",
			actualWidth2,
			actualHeight2,
			ascent2,
			descent2,
			maxFontHeight2,
			maxAscent,
			maxDescent,
		)
	}

	//TODO Implementing a debug log routine
	if false {
		context := freetype.NewContext()
		context.SetDPI(float64(dpi))
		context.SetFontSize(fontsize)
		context.SetFont(ttFont)
		contextHeight := context.PointToFixed(fontsize).Ceil()

		print("Maxfontsize: ")
		fmt.Printf("%.2f", fontsize)
		print(", height: ")
		print(height)
		print(", actualHeight ")
		print(actualHeight)
		print(", contextHeight ")
		print(contextHeight)
		print(", asc ")
		fmt.Printf("%d ", ascent)
		print(", des ")
		fmt.Printf("%d ", descent)
		println()
		println()
	}

	return fontsize, actualWidth, ascent, descent
}

func determineHeightFittingFontsize(dpi uint, ttFont *truetype.Font, startingFontsize float64, maxHeight int, text string) (int, int, float64) {
	var ascent int
	var descent int
	fontsize := startingFontsize

	actualHeight := 0
	for {
		actualHeight, ascent, descent = actualStringHeight(float64(dpi), ttFont, fontsize, text)

		if actualHeight > maxHeight && fontsize > 0.25 {
			fontsize -= 0.25
		} else {
			break
		}
	}
	//TODO ascent & descent arent used yet
	return ascent, descent, fontsize
}

func determineWidthFittingFontsize(dpi uint, ttFont *truetype.Font, startingFontsize float64, maxWidth int, text string) (int, float64) {
	var actualWidth int
	fontsize := startingFontsize

	for {
		actualWidth = actualStringWidth(float64(dpi), ttFont, fontsize, text)

		if actualWidth > maxWidth && fontsize > 0.25 {
			fontsize -= 0.25
		} else {
			break
		}
	}
	return actualWidth, fontsize
}

func actualStringWidth(dpi float64, ttFont *truetype.Font, fontsize float64, text string) (width int) {
	opts := truetype.Options{}
	opts.Size = fontsize
	opts.DPI = dpi

	face := truetype.NewFace(ttFont, &opts)

	return font.MeasureString(face, text).Ceil()
}

// TODO Not using atm. But could be part of "image element" lib?
// fast, but not accurate
func approximatedMaxFontHeight(dpi uint, ttFont *truetype.Font, fontsize float64) int {
	context := freetype.NewContext()
	context.SetDPI(float64(dpi))
	context.SetFontSize(fontsize)
	context.SetFont(ttFont)
	return context.PointToFixed(fontsize).Ceil()
}

// TODO Not using atm. But could be part of "image element" lib?
// slow, but accurate
func maxFontHeight(dpi float64, ttFont *truetype.Font, fontsize float64) (height int, ascent int, descent int) {
	opts := truetype.Options{}
	opts.Size = fontsize
	opts.DPI = dpi
	face := truetype.NewFace(ttFont, &opts)

	metrics := face.Metrics()
	return metrics.Ascent.Ceil() + metrics.Descent.Ceil(), metrics.Ascent.Ceil(), metrics.Descent.Ceil()
}

func actualStringHeight(dpi float64, ttFont *truetype.Font, fontsize float64, text string) (height int, ascent int, descent int) {
	opts := truetype.Options{}
	opts.Size = fontsize
	opts.DPI = dpi
	face := truetype.NewFace(ttFont, &opts)

	bounds, _ := font.BoundString(face, text)

	ascent = (-bounds.Min.Y).Ceil()
	descent = bounds.Max.Y.Ceil()

	height = ascent + descent

	//TODO Implementing a debug log routine
	if false {
		heightDebug := (+bounds.Max.Y + (-bounds.Min.Y)).Ceil()

		heightFloat := float64(bounds.Max.Y-bounds.Min.Y) / 64
		width := (bounds.Max.X >> 6) - (bounds.Min.X >> 6)
		print(width)
		print(" x ")
		fmt.Printf("%d ", height)
		print(" height float ")
		fmt.Printf(" %.2f ", heightFloat)
		print(" height ceil ")
		fmt.Printf("%d ", heightDebug)
		print(" pt ")
		fmt.Printf(" %.2f ", fontsize)

		ascentFloat := float64(-bounds.Min.Y) / 64.0
		descentFloat := float64(+bounds.Max.Y) / 64.0

		print(" asc ")
		fmt.Printf("%.2f ", ascentFloat)
		print(" des ")
		fmt.Printf("%.2f ", descentFloat)

		print(" asc ")
		fmt.Printf("%d ", ascent)
		print(" des ")
		fmt.Printf("%d ", descent)

		print(" asc2 ")
		fmt.Printf("%d ", (-bounds.Min.Y).Ceil())
		print(" des2 ")
		fmt.Printf("%d ", +bounds.Max.Y.Ceil())
		println()
	}

	return height, ascent, descent
}

// TODO replace main.drawString() with this?
func drawString2(img *image.RGBA, bounds image.Rectangle, ttf *truetype.Font, text string, dpi uint, fontsize float64, color color.Color, pt image.Point, centerHorizontally bool, centerVertically bool) {
	c := ftContext(img, ttf, dpi, fontsize)

	if fontsize <= 0 {
		fontsize, _, _, _ = maxFontSize(ttf, dpi, bounds.Dx(), bounds.Dy(), text)
		c.SetFontSize(fontsize)
	}

	//TODO calculating actual string width/height is very expensive. Providing as parameter, for the cases if they were already calculated
	if centerHorizontally {
		actualWidth := actualStringWidth(float64(dpi), ttf, fontsize, text)
		xCenter := float64(bounds.Dx())/2.0 - (float64(actualWidth) / 2.0)
		pt = image.Pt(bounds.Min.X+int(xCenter), pt.Y)
	}
	if centerVertically {
		actualHeight, _, _ := actualStringHeight(float64(dpi), ttf, fontsize, text)
		yCenter := float64(bounds.Dy()/2.0) + (float64(actualHeight) / 2.0)
		pt = image.Pt(pt.X, bounds.Min.Y+int(yCenter))
	}

	c.SetSrc(image.NewUniform(color))
	if _, err := c.DrawString(text, freetype.Pt(pt.X, pt.Y)); err != nil {
		fmt.Fprintf(os.Stderr, "Can't render string: %s\n", err)
		return
	}
}

func createBar(length uint, thickness uint, percentage uint8) image.Image {
	thicknessInt := int(thickness)
	lengthInt := int(length)
	img := image.NewRGBA(image.Rect(0, 0, lengthInt, thicknessInt))

	thickPartLength := int((float64(lengthInt) / 100.0) * float64(percentage))

	for x := 0; x < thickPartLength; x++ {
		for y := 0; y < thicknessInt; y++ {
			img.Set(x, y, color.White)
		}
	}

	thinThickness := thicknessInt - (thicknessInt / 3)
	yOffset := (thicknessInt - thinThickness) / 2
	for x := thickPartLength; x < lengthInt; x++ {
		for y := 0; y < thinThickness; y++ {
			img.Set(x, y+yOffset, color.Gray16{0x7FFF})
		}
	}

	return img
}

func createRectangle(width uint, height uint) image.Rectangle {
	return image.Rectangle{
		Min: image.Point{
			X: 0,
			Y: 0,
		},
		Max: image.Point{
			X: int(width),
			Y: int(height),
		},
	}
}

func debugDrawInline(img *image.RGBA, rectangle image.Rectangle) {
	rgbaGreenBlue := color.RGBA{
		R: 0x0,
		G: 0xff,
		B: 0xff,
		A: 0,
	}

	innerLine := rectangle.Bounds()
	innerLine.Min.X++
	innerLine.Min.Y++
	innerLine.Max.X--
	innerLine.Max.Y--

	drawRectangle(img, rgbaGreenBlue, innerLine)
}

func debugDrawOutline(img *image.RGBA, rectangle image.Rectangle) {
	colour := color.RGBA{
		R: 0xff,
		G: 0x0,
		B: 0xff,
		A: 0,
	}

	outLine := rectangle.Bounds()
	outLine.Min.X--
	outLine.Min.Y--
	outLine.Max.X++
	outLine.Max.Y++

	drawRectangle(img, colour, outLine)
}

func debugDrawOnLine(img *image.RGBA, rectangle image.Rectangle) {
	colour := color.RGBA{
		R: 0xff,
		G: 0x0,
		B: 0x0,
		A: 0,
	}

	drawRectangle(img, colour, rectangle)
}

func drawRectangle(img *image.RGBA, colour color.Color, rectangle image.Rectangle) {
	for x := rectangle.Min.X; x < rectangle.Max.X; x++ {
		img.Set(x, rectangle.Min.Y, colour)
		img.Set(x, rectangle.Max.Y-1, colour)
	}

	for y := rectangle.Min.Y + 1; y < rectangle.Max.Y-1; y++ {
		img.Set(rectangle.Min.X, y, colour)
		img.Set(rectangle.Max.X-1, y, colour)
	}
}
