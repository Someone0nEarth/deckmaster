package main

import (
	"image"
	"image/color"

	"github.com/golang/freetype/truetype"
)

// WidgetElement contains elementSegment. Each elementSegment will be drawn with equally height with interspaces
// between them.
type WidgetElement struct {
	img                *image.RGBA
	segmentHeight      uint
	segmentsInterspace uint
	dpi                uint
	segmentPositionY   uint
	segments           []elementSegment
}

// NewWidgetElement new WidgetElement
func NewWidgetElement(dpi uint) *WidgetElement {
	return &WidgetElement{dpi: dpi}
}

// drawElement draws the element into the target image within the bound at the position.
func (e *WidgetElement) drawElement(targetImage *image.RGBA, position image.Point, bounds image.Rectangle) {
	e.segmentPositionY = 0
	e.img = image.NewRGBA(image.Rect(0, 0, bounds.Dx(), bounds.Dy()))

	e.calculateSegmentsInterspace()
	e.calculateSegmentSize()

	for _, segment := range e.segments {
		segment.draw(e)
		e.incrementSegmentPositionY()
	}

	_ = drawImage(targetImage, e.img, position)
}

// addIconSegment will add the icon into the current elementSegment slot, resize it to fit it and set the pointer to the next elementSegment
func (e *WidgetElement) addIconSegment(icon image.Image) {
	segment := iconSegment{
		img: icon,
	}
	e.segments = append(e.segments, &segment)
}

// addPercentageBarSegment will add a percentage bar into the current elementSegment slot and set the pointer to the next elementSegment
func (e *WidgetElement) addPercentageBarSegment(percentage uint8) {
	segment := percentageBarSegment{
		percentage: percentage,
	}
	e.segments = append(e.segments, &segment)
}

// addTextSegment will add the text into the current elementSegment slot and set the pointer to the next elementSegment
func (e *WidgetElement) addTextSegment(text string, font *truetype.Font, fontColor color.Color, alignmentX int, centerVertically bool) {
	segment := textSegment{
		text:             text,
		font:             font,
		fontColor:        fontColor,
		alignmentX:       alignmentX,
		centerVertically: centerVertically,
	}
	e.segments = append(e.segments, &segment)
}

// addBlankSegment is for lay-outing. It will add a blank elementSegment into the current segment slot and set the pointer to the next elementSegment
func (e *WidgetElement) addBlankSegment() {
	e.segments = append(e.segments, &blankSegment{})
}

func (e *WidgetElement) incrementSegmentPositionY() {
	e.segmentPositionY += e.segmentHeight + e.segmentsInterspace
}

func (e *WidgetElement) calculateSegmentsInterspace() {
	interspace := e.height() / (3 * e.numberOfSegments())
	if interspace < 5 {
		interspace = 5
	}
	e.segmentsInterspace = interspace
}

func (e *WidgetElement) calculateSegmentSize() {
	e.segmentHeight = (e.height() - ((e.numberOfSegments() - 1) * e.segmentsInterspace)) / e.numberOfSegments()
}

func (e *WidgetElement) numberOfSegments() uint {
	return uint(len(e.segments))
}

func (e *WidgetElement) height() uint {
	return uint(e.img.Bounds().Dy())
}

func (e *WidgetElement) width() uint {
	return uint(e.img.Bounds().Dx())
}

type elementSegment interface {
	draw(element *WidgetElement)
}

type iconSegment struct {
	img image.Image
}

func (segment *iconSegment) draw(element *WidgetElement) {
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

func (s *percentageBarSegment) draw(element *WidgetElement) {
	bar := createBar(element.width(), element.segmentHeight, s.percentage)

	_ = drawImage(element.img, bar, image.Pt(0, int(element.segmentPositionY)))
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

type textSegment struct {
	text      string
	font      *truetype.Font
	fontColor color.Color

	alignmentX       int
	centerVertically bool
}

func (s *textSegment) draw(element *WidgetElement) {
	bounds := createRectangle(element.width(), element.segmentHeight)

	fontsize, actualWidth, _, descent := maxFontSize(s.font, element.dpi, bounds.Dx(), bounds.Dy(), s.text)

	var x int
	centerHorizontally := false

	if s.alignmentX == 0 {
		x = 0
		//TODO Already calculate the centered x coor, using actualWidth, here (and not later in drawString)
		centerHorizontally = true
	} else if s.alignmentX == -1 {
		x = 0
	} else {
		x = int(element.width()) - actualWidth
	}

	y := int(element.segmentPositionY) + int(element.segmentHeight) - descent

	drawString2(element.img, bounds, s.font, s.text, element.dpi, fontsize, s.fontColor, image.Pt(x, y), centerHorizontally, s.centerVertically)
}

type blankSegment struct {
}

func (s *blankSegment) draw(_ *WidgetElement) {
	// Nothing to be done here
}
