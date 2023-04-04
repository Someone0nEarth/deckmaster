package main

import (
	"image"
	"image/color"
)

// ImageElement for drawing into vertical equal segments considering the ImageElement bounds and number of Segments.
type ImageElement struct {
	img                *image.RGBA
	segmentHeight      uint
	segmentsInterspace uint
	numberOfSegments   uint
	dpi                uint
	fontColor          color.Color
	segmentPositionY   uint
}

// NewImageElementWithStrings Will "divide" the element bounds into vertical equal segments considering the number of segments.
func NewImageElementWithStrings(bounds image.Rectangle, numberOfSegments uint, dpi uint, fontColor color.Color) *ImageElement {
	element := ImageElement{
		numberOfSegments: numberOfSegments,
		dpi:              dpi,
		fontColor:        fontColor,
		segmentPositionY: 0,
	}
	element.img = image.NewRGBA(image.Rect(0, 0, bounds.Dx(), bounds.Dy()))
	element.calculateSegmentsInterspace()
	element.calculateSegmentSize()

	return &element
}

// NewImageElement Will "divide" the element bounds into vertical equal segments considering the number of segments.
func NewImageElement(bounds image.Rectangle, numberOfSegments uint) *ImageElement {
	element := ImageElement{
		numberOfSegments: numberOfSegments,
		dpi:              0,
		fontColor:        color.Transparent,
		segmentPositionY: 0,
	}
	element.img = image.NewRGBA(image.Rect(0, 0, bounds.Dx(), bounds.Dy()))
	element.calculateSegmentsInterspace()
	element.calculateSegmentSize()

	return &element
}

// DrawBarSegment will draw a percentage bar into the current segment slot and set the pointer to the next segment
func (element *ImageElement) DrawBarSegment(percentage *uint8) {
	bar := createBar(element.width(), element.segmentHeight, *percentage)

	drawImage(element.img, bar, image.Pt(0, int(element.segmentPositionY)))

	element.incrementSegmentPositionY()
}

// DrawStringSegment will draw the text into the current segment slot and set the pointer to the next segment
func (element *ImageElement) DrawStringSegment(text string) {
	bounds := createRectangle(element.width(), element.segmentHeight)

	y := element.calculateStringSegmentY(element.segmentPositionY)

	drawString(element.img, bounds, ttfFont, text, element.dpi, 0, element.fontColor, image.Pt(-1, y))

	element.incrementSegmentPositionY()
}

func (element *ImageElement) DrawIconSegment(icon image.Image) {
	drawImageWithResizing(element.img, icon, int(element.segmentHeight), image.Pt(-1, int(element.segmentPositionY)))

	element.incrementSegmentPositionY()
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

func (element *ImageElement) calculateStringSegmentY(segmentPositionY uint) int {
	if element.numberOfSegments == 1 {
		// y < 0 will center the string vertically (see widget.go drawString())
		return -1
	} else {
		return int(segmentPositionY + element.segmentHeight)
	}
}

func (element *ImageElement) calculateSegmentsInterspace() {
	interspace := element.height() / (3 * element.numberOfSegments)
	if interspace < 5 {
		interspace = 5
	}
	element.segmentsInterspace = interspace
}

func (element *ImageElement) calculateSegmentSize() {
	element.segmentHeight = (element.height() - ((element.numberOfSegments - 1) * element.segmentsInterspace)) / element.numberOfSegments
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
