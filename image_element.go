package main

import (
	"image"
	"image/color"
)

// ImageElement for drawing into vertical equal segments considering the ImageElement bounds and number of Segments.
type ImageElement struct {
	img              *image.RGBA
	segmentsSize     int
	verticalMargin   int
	numberOfSegments uint8
	dpi              uint
	fontColor        color.Color
	segmentPositionY int
}

// NewImageElement Will "divide" the element bounds into vertical equal segments considering the number of segments.
func NewImageElement(bounds image.Rectangle, numberOfSegments uint8, dpi uint, fontColor color.Color) *ImageElement {
	element := ImageElement{
		numberOfSegments: numberOfSegments,
		dpi:              dpi,
		fontColor:        fontColor,
		segmentPositionY: 0,
	}
	element.img = image.NewRGBA(image.Rect(0, 0, bounds.Dx(), bounds.Dy()))
	element.calculateVerticalMargin()
	element.calculateSegmentSize()

	return &element
}

// DrawBarSegment will draw a percentage bar into the current segment slot and set the pointer to the next segment
func (element *ImageElement) DrawBarSegment(percentage *uint8) {
	bar := createBar(element.width(), element.segmentsSize-element.verticalMargin, *percentage)

	drawImage(element.img, bar, image.Pt(0, element.segmentPositionY+(element.verticalMargin/2)))

	element.incrementSegmentPositionY()
}

// DrawStringSegment will draw the text into the current segment slot and set the pointer to the next segment
func (element *ImageElement) DrawStringSegment(text string) {
	bounds := createRectangle(0, 0, element.width(), element.segmentsSize)

	y := element.calculateStringSegmentY(element.segmentPositionY)

	drawString(element.img, bounds, ttfFont, text, element.dpi, 0, element.fontColor, image.Pt(-1, y))

	element.incrementSegmentPositionY()
}

func (element *ImageElement) height() int {
	return element.img.Bounds().Dy()
}

func (element *ImageElement) width() int {
	return element.img.Bounds().Dx()
}

func (element *ImageElement) incrementSegmentPositionY() {
	element.segmentPositionY += element.segmentsSize + element.verticalMargin
}

func (element *ImageElement) calculateStringSegmentY(segmentPositionY int) int {
	if element.numberOfSegments == 1 {
		// y < 0 will center the string vertically (see widget.go drawString())
		return -1
	} else {
		return segmentPositionY + element.segmentsSize
	}
}

func (element *ImageElement) calculateVerticalMargin() {
	margin := element.height() / (3 * int(element.numberOfSegments))
	if margin < 5 {
		margin = 5
	}
	element.verticalMargin = margin
}

func (element *ImageElement) calculateSegmentSize() {
	element.segmentsSize = (element.height() - ((int(element.numberOfSegments) - 1) * element.verticalMargin)) / int(element.numberOfSegments)
}

func createBar(length int, thickness int, percentage uint8) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, length, thickness))

	thickPartLength := int((float64(length) / 100.0) * float64(percentage))

	for x := 0; x < thickPartLength; x++ {
		for y := 0; y < thickness; y++ {
			img.Set(x, y, color.White)
		}
	}

	thinThickness := thickness - (thickness / 3)
	yOffset := (thickness - thinThickness) / 2
	for x := thickPartLength; x < length; x++ {
		for y := 0; y < thinThickness; y++ {
			img.Set(x, y+yOffset, color.Gray16{0x7FFF})
		}
	}

	return img
}
