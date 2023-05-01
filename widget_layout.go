package main

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"os"

	"github.com/golang/freetype"
	"golang.org/x/image/font"

	"github.com/golang/freetype/truetype"
)

// ButtonLayout A layout for a stream deck button. The sizes and the position of the elements depends on which and how
// many will be added to the elements.
//
// The conceptional layout of ButtonLayout is:
//
// +----------+
// |   icon   |
// |  element |
// |----------|
// |  info    |
// | element  |
// +----------+
type ButtonLayout struct {
	iconElement *WidgetElement
	infoElement *WidgetElement
	dpi         uint
}

// SetIcon Sets the icon in the icon element.
func (l *ButtonLayout) SetIcon(icon image.Image) {
	l.iconElement = &WidgetElement{}
	l.iconElement.addIconSegment(icon)
}

// AddPercentageBar Add a percentage bar with the given value to the info element.
func (l *ButtonLayout) AddPercentageBar(percentage uint8) {
	if l.infoElement == nil {
		l.infoElement = &WidgetElement{}
	}
	l.infoElement.addPercentageBarSegment(percentage)
}

// AddText Add the text to the info element.
func (l *ButtonLayout) AddText(text string, font *truetype.Font, fontColor color.Color, alignmentX int, centerVertically bool) {
	if l.infoElement == nil {
		l.infoElement = &WidgetElement{dpi: l.dpi}
	}
	l.infoElement.addTextSegment(text, font, fontColor, alignmentX, centerVertically)
}

// CreateImage creates and return the image for the layout.
func (l *ButtonLayout) CreateImage(bounds image.Rectangle) image.Image {
	imageHeight := uint(bounds.Dy())
	imageWidth := uint(bounds.Dx())
	img := image.NewRGBA(image.Rect(0, 0, int(imageWidth), int(imageHeight)))

	margin := imageHeight / 18
	height := imageHeight - (margin * 2)
	width := imageWidth - (margin * 2)

	elementPosition := image.Pt(int(margin), int(margin))

	if l.iconElement != nil {
		elementHeight := l.calculateIconSize(height, width)
		l.iconElement.drawElement(img, elementPosition, createRectangle(width, elementHeight))

		elementPosition.Y += int(l.iconElement.height() + margin)
	}

	if l.infoElement != nil {
		elementHeight := height - uint(elementPosition.Y)
		l.infoElement.drawElement(img, elementPosition, createRectangle(width, elementHeight))
	}

	return img
}

func (l *ButtonLayout) calculateIconSize(maxHeight uint, maxWidth uint) uint {
	var iconHeightRatio float64

	if l.infoElement != nil {
		iconHeightRatio = calculateIconSizeRatio(l.infoElement.numberOfSegments())
	} else {
		iconHeightRatio = calculateIconSizeRatio(0)
	}

	iconHeight := uint(float64(maxHeight) * iconHeightRatio)
	if iconHeight > maxWidth {
		return maxWidth
	}
	return iconHeight
}

func calculateIconSizeRatio(numberOfInfoSegments uint) float64 {
	if numberOfInfoSegments == 0 {
		return 1.0
	} else if numberOfInfoSegments == 1 {
		return 0.66
	} else {
		return math.Round(1.0/float64(numberOfInfoSegments)*100) / 100
	}
}

// ScreenSegmentLayout A layout for a stream deck touch screen segment. The sizes and the position of the elements
// depends on which and how many will be added to the elements.
//
// The conceptional layout of ScreenSegmentLayout is:
//
// +---------------------+
// | label element       |
// |---------------------|
// |   icon   |  info    |
// | element  | element  |
// |          |          |
// +---------------------+
type ScreenSegmentLayout struct {
	iconElement  *WidgetElement
	labelElement *WidgetElement
	infoElement  *WidgetElement
	dpi          uint
}

// SetIcon Sets the icon in the icon element.
func (l *ScreenSegmentLayout) SetIcon(icon image.Image) {
	l.iconElement = &WidgetElement{}
	l.iconElement.addIconSegment(icon)
}

// AddPercentageBar Add a percentage bar with the given value to the info element.
func (l *ScreenSegmentLayout) AddPercentageBar(percentage uint8) {
	if l.infoElement == nil {
		l.infoElement = &WidgetElement{}
	}
	l.infoElement.addPercentageBarSegment(percentage)
}

// AddText Add the text to the info element.
func (l *ScreenSegmentLayout) AddText(text string, font *truetype.Font, fontColor color.Color, alignmentX int, centerVertically bool) {
	if l.infoElement == nil {
		l.infoElement = &WidgetElement{dpi: l.dpi}
	}
	l.infoElement.addTextSegment(text, font, fontColor, alignmentX, centerVertically)
}

// SetLabel sets the label in the label element.
func (l *ScreenSegmentLayout) SetLabel(text string, font *truetype.Font, fontColor color.Color, alignmentX int, centerVertically bool) {
	l.labelElement = &WidgetElement{dpi: l.dpi}
	l.labelElement.addTextSegment(text, font, fontColor, alignmentX, centerVertically)
}

// AddBlank adds a blank segment to the info element.
func (l *ScreenSegmentLayout) AddBlank() {
	if l.infoElement == nil {
		l.infoElement = &WidgetElement{dpi: l.dpi}
	}

	l.infoElement.addBlankSegment()
}

// CreateImage creates and return the image for the layout.
func (l *ScreenSegmentLayout) CreateImage(bounds image.Rectangle) image.Image {
	imageHeight := uint(bounds.Dy())
	imageWidth := uint(bounds.Dx())
	img := image.NewRGBA(image.Rect(0, 0, int(imageWidth), int(imageHeight)))

	margin := imageHeight / 18
	height := imageHeight - (margin * 2)
	width := imageWidth - (margin * 2)

	initialPosition := image.Pt(int(margin), int(margin))
	iconPosition := initialPosition
	labelPosition := initialPosition
	infoPosition := initialPosition

	var iconBounds = createRectangle(width, height)
	var labelBounds = createRectangle(width, height)
	var infoBounds = createRectangle(width, height)

	interspaceElements := margin

	if l.labelElement != nil {
		labelHeight := uint(float64(height)/4.5) - interspaceElements/2

		labelBounds.Max.Y = int(labelHeight)
		iconBounds.Max.Y -= int(labelHeight + interspaceElements)
		infoBounds.Max.Y -= int(labelHeight + interspaceElements)

		iconPosition.Y += int(labelHeight + interspaceElements)
		infoPosition.Y += int(labelHeight + interspaceElements)
	}

	if l.iconElement != nil && l.infoElement != nil {
		iconBounds.Max.X = int(uint(float64(width)/3.0) - interspaceElements/2)
		infoBounds.Max.X = int(width - uint(iconBounds.Dx()) - interspaceElements/2)

		infoPosition.X += iconBounds.Dx() + int(interspaceElements)
	}

	if l.labelElement != nil {
		l.labelElement.drawElement(img, labelPosition, labelBounds)
	}

	if l.iconElement != nil {
		l.iconElement.drawElement(img, iconPosition, iconBounds)
	}

	if l.infoElement != nil {
		l.infoElement.drawElement(img, infoPosition, infoBounds)
	}

	return img
}

func maxFontSize(ttFont *truetype.Font, dpi uint, width, height int, text string) (float64, int, int, int) {
	startingFontsize := float64(height) / (float64(dpi) / (72.0))
	initialFontsize := startingFontsize * 1.4

	_, _, heightFittingFontsize := determineHeightFittingFontsize(dpi, ttFont, initialFontsize, height, text)

	actualWidth, fontsize := determineWidthFittingFontsize(dpi, ttFont, heightFittingFontsize, width, text)

	_, ascent, descent := actualStringHeight(float64(dpi), ttFont, fontsize, text)

	return fontsize, actualWidth, ascent, descent
}

func determineHeightFittingFontsize(dpi uint, ttFont *truetype.Font, startingFontsize float64, maxHeight int, text string) (int, int, float64) {
	var ascent int
	var descent int
	fontsize := startingFontsize

	actualHeight := 0
	for {
		actualHeight, ascent, descent = actualStringHeight(float64(dpi), ttFont, fontsize, text)

		sizeDecrease := 0.7
		if actualHeight > maxHeight && fontsize > sizeDecrease {
			fontsize -= sizeDecrease
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

		sizeDecrease := 0.7
		if actualWidth > maxWidth && fontsize > sizeDecrease {
			fontsize -= sizeDecrease
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

// fast, but not accurate
//
//nolint:golint,unused // TODO Not using atm. But could be part of "image element" lib?
func approximatedMaxFontHeight(dpi uint, ttFont *truetype.Font, fontsize float64) int {
	context := freetype.NewContext()
	context.SetDPI(float64(dpi))
	context.SetFontSize(fontsize)
	context.SetFont(ttFont)
	return context.PointToFixed(fontsize).Ceil()
}

// slow, but accurate
//
//nolint:golint,unused // TODO Not using atm. But could be part of "image element" lib?
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
		pt = image.Pt(pt.X+int(xCenter), pt.Y)
	}
	if centerVertically {
		actualHeight, _, _ := actualStringHeight(float64(dpi), ttf, fontsize, text)
		yCenter := float64(bounds.Dy()/2.0) + (float64(actualHeight) / 2.0)
		pt = image.Pt(pt.X, (pt.Y-bounds.Dy())+int(yCenter))
	}

	c.SetSrc(image.NewUniform(color))
	if _, err := c.DrawString(text, freetype.Pt(pt.X, pt.Y)); err != nil {
		fmt.Fprintf(os.Stderr, "Can't render string: %s\n", err)
		return
	}
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
