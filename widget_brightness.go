package main

import (
	"fmt"
	"image"
	"image/color"
	"math"

	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
)

// TODO move this to device
const TOUCHSCREEN_DPI = uint(200)

// BrightnessWidget is a widget for controlling the device brightness.
type BrightnessWidget struct {
	*ButtonWidget
	brightness     uint8
	showBar        bool
	showPercentage bool
}

// NewBrightnessWidget returns a new BrightnessWidget.
func NewBrightnessWidget(bw *BaseWidget, opts WidgetConfig) (*BrightnessWidget, error) {

	showBar := true
	if opts.Config["bar"] != nil {
		_ = ConfigValue(opts.Config["bar"], &showBar)
	}

	showPercentage := true
	if opts.Config["percentage"] != nil {
		_ = ConfigValue(opts.Config["percentage"], &showPercentage)
	}

	buttonWidget, err := NewButtonWidget(bw, opts)
	if err != nil {
		return nil, err
	}

	return &BrightnessWidget{
		ButtonWidget:   buttonWidget,
		brightness:     getBrightness(),
		showBar:        showBar,
		showPercentage: showPercentage,
	}, nil
}

// RequiresUpdate returns true when the widget wants to be repainted.
func (w *BrightnessWidget) RequiresUpdate() bool {
	if (w.showBar || w.showPercentage) && w.brightness != getBrightness() {
		return true
	}

	return w.BaseWidget.RequiresUpdate()
}

func getBrightness() uint8 {
	return uint8(*brightness)
}

// Update renders the widget.
func (w *BrightnessWidget) Update() error {
	w.refreshBrightnessValue()

	if w.screenSegmentIndex != nil {
		return w.updateScreenSegment()
	}
	return w.updateButton()
}

func (w *BrightnessWidget) updateButton() error {
	var percentageLabel string

	if w.showPercentage {
		percentageLabel = fmt.Sprintf("%d%%", w.brightness)
	}

	var percentage *uint8
	if w.showBar {
		percentageValue := w.brightness
		percentage = &percentageValue
	}

	img := createButtonImage(w.getMaxImageSize(), w.dev.DPI, w.color, w.icon, w.label, percentageLabel, percentage)

	return w.render(w.dev, img)
}

func (w *BrightnessWidget) updateScreenSegment() error {
	var percentageLabel string

	if w.showPercentage {
		percentageLabel = fmt.Sprintf("%d%%", w.brightness)
	}

	var percentage *uint8
	if w.showBar {
		percentageValue := w.brightness
		percentage = &percentageValue
	}

	dpi := TOUCHSCREEN_DPI
	img := createSegmentImage(w.getMaxImageSize(), dpi, w.color, w.icon, w.label, percentageLabel, percentage)

	return w.render(w.dev, img)
}

func (w *BrightnessWidget) calculateWidth(img *image.RGBA, font *truetype.Font, text string, fontsize float64) int {
	extent, _ := ftContext(img, font, w.dev.DPI, fontsize).DrawString(text, freetype.Pt(0, 0))
	return extent.X.Floor()
}

func (w *BrightnessWidget) refreshBrightnessValue() {
	w.brightness = getBrightness()
}

func createButtonImage(bounds image.Rectangle, dpi uint, fontColor color.Color, icon image.Image, label string, percentageLabel string, percentage *uint8) image.Image {
	imageHeight := uint(bounds.Dy())
	imageWidth := uint(bounds.Dx())

	margin := imageHeight / 18
	height := imageHeight - (margin * 2)
	width := imageWidth - (margin * 2)

	img := image.NewRGBA(image.Rect(0, 0, int(imageWidth), int(imageHeight)))

	numberOfSegments := countSegments(label, percentageLabel, percentage)

	interspaceElements := uint(0)
	if icon != nil && numberOfSegments > 0 {
		interspaceElements = margin
	}

	iconSize := uint(0)
	if icon != nil {
		iconSizeRatio := calculateIconSizeRatio(numberOfSegments)
		iconSize = uint(float64(height) * iconSizeRatio)

		elementBounds := createRectangle(width, iconSize)
		iconElement := NewImageElement(elementBounds, 1)
		iconElement.DrawIconSegment(icon)

		elementPosition := image.Pt(int(margin), int(margin))
		drawImage(img, iconElement.img, elementPosition)
		drawInOrOutLine(img, elementPosition, elementBounds)
	}

	if numberOfSegments > 0 {
		elementBounds := createRectangle(width, height-(iconSize+interspaceElements))

		elementImage := createButtonImageElement(elementBounds, dpi, fontColor, numberOfSegments, label, percentageLabel, percentage)

		y := int(margin + iconSize + interspaceElements)
		elementPosition := image.Pt(int(margin), y)
		drawImage(img, elementImage, elementPosition)
		drawInOrOutLine(img, elementPosition, elementBounds)
	}

	return img
}

func createSegmentImage(bounds image.Rectangle, dpi uint, fontColor color.Color, icon image.Image, label string, percentageLabel string, percentage *uint8) image.Image {
	imageHeight := uint(bounds.Dy())
	imageWidth := uint(bounds.Dx())

	margin := int(imageHeight / 18)
	height := imageHeight - uint(margin*2)
	width := imageWidth - uint(margin*2)

	interspaceElements := uint(margin)

	img := image.NewRGBA(image.Rect(0, 0, int(imageWidth), int(imageHeight)))

	boundsWithoutMargin := image.Rectangle{}
	boundsWithoutMargin.Max.X = int(width)
	boundsWithoutMargin.Max.Y = int(height)

	iconBounds := boundsWithoutMargin
	labelBounds := boundsWithoutMargin
	valueBounds := boundsWithoutMargin

	initialPosition := image.Pt(margin, margin)
	iconPosition := initialPosition
	labelPosition := initialPosition
	valuePosition := initialPosition

	hasIcon, hasLabel, hasPercentageValue, hasPercentageBar, numberOfItems := countScreenSegmentElements(icon, label, percentageLabel, percentage)

	if numberOfItems <= 2 {
		return createButtonImage(bounds, dpi, fontColor, icon, label, percentageLabel, percentage)
	}

	if hasLabel {
		labelHeight := uint(float64(height)/4.5) - interspaceElements/2
		labelBounds.Max.Y = int(labelHeight)

		if hasIcon {
			iconHeight := height - labelHeight - interspaceElements
			iconBounds.Max.Y = int(iconHeight)

			iconPosition.Y = margin + int(labelHeight+interspaceElements)
		}

		if hasPercentageValue || hasPercentageBar {
			valueHeight := height - labelHeight - interspaceElements
			valueBounds.Max.Y = int(valueHeight)

			valuePosition.Y = margin + int(labelHeight+interspaceElements)
		}
	}

	if hasIcon {
		iconWidth := uint(float64(width)/3.0) - interspaceElements/2
		iconBounds.Max.X = int(iconWidth)

		valueWidth := width - iconWidth - interspaceElements
		valueBounds.Max.X = int(valueWidth)

		iconPosition.X = margin
		valuePosition.X = margin + int(iconWidth+interspaceElements)
	}

	if hasLabel {
		labelElement := NewImageElementWithStrings(labelBounds, 1, dpi, fontColor)
		labelElement.DrawStringSegment(label, -1, false)

		drawImage(img, labelElement.img, labelPosition)

		//drawInOrOutLine(img, labelPosition, labelBounds)
		drawInOrOutLine(img, labelPosition, labelElement.img.Bounds())

	}

	if hasIcon {
		iconElement := NewImageElement(iconBounds, 1)
		iconElement.DrawIconSegment(icon)

		drawImage(img, iconElement.img, iconPosition)
		drawInOrOutLine(img, iconPosition, iconBounds)
	}

	if hasPercentageValue && hasPercentageBar {
		valueElement := NewImageElementWithStrings(valueBounds, 3, dpi, fontColor)

		valueElement.DrawBlankSegment()
		valueElement.DrawStringSegment(percentageLabel, 1, false)
		valueElement.DrawBarSegment(percentage)

		drawImage(img, valueElement.img, valuePosition)
		drawInOrOutLine(img, valuePosition, valueBounds)
	} else if hasPercentageValue {
		valueElement := NewImageElementWithStrings(valueBounds, 1, dpi, fontColor)

		valueElement.DrawStringSegment(percentageLabel, 0, true)

		drawImage(img, valueElement.img, valuePosition)
		drawInOrOutLine(img, valuePosition, valueBounds)
	} else if hasPercentageBar {
		valueElement := NewImageElement(valueBounds, 2)

		valueElement.DrawBlankSegment()
		valueElement.DrawBarSegment(percentage)

		drawImage(img, valueElement.img, valuePosition)
		drawInOrOutLine(img, valuePosition, valueBounds)
	}

	return img
}

func drawInOrOutLine(img *image.RGBA, boundsPosition image.Point, bounds image.Rectangle) {

	//TODO Implementing a debug log routine
	if false {
		drawOutlineLine(img, boundsPosition, bounds)
		drawInline(img, boundsPosition, bounds)
	}
}

func drawInline(img *image.RGBA, boundsPosition image.Point, boundss image.Rectangle) {

	rgbaGreenYellow := color.RGBA{
		R: 0xad,
		G: 0xff,
		B: 0x2f,
		A: 0,
	}

	translatedBounds := boundss.Add(boundsPosition)

	for x := translatedBounds.Min.X; x < translatedBounds.Max.X; x++ {

		img.Set(x, translatedBounds.Min.Y, rgbaGreenYellow)
		img.Set(x, translatedBounds.Max.Y, rgbaGreenYellow)
	}

	for y := translatedBounds.Min.Y; y < translatedBounds.Max.Y; y++ {
		img.Set(translatedBounds.Min.X, y, rgbaGreenYellow)
		img.Set(translatedBounds.Max.X, y, rgbaGreenYellow)
	}
}

func drawOutlineLine(img *image.RGBA, boundsPosition image.Point, bounds image.Rectangle) {
	rgbaRed := color.RGBA{
		R: 0xff,
		G: 0x0,
		B: 0x0,
		A: 0,
	}

	translatedBounds := bounds.Add(boundsPosition)
	translatedBounds.Min.X -= 1
	translatedBounds.Min.Y -= 1
	translatedBounds.Max.X += 1
	translatedBounds.Max.Y += 1

	for x := translatedBounds.Min.X; x < translatedBounds.Max.X; x++ {

		img.Set(x, translatedBounds.Min.Y, rgbaRed)
		img.Set(x, translatedBounds.Max.Y, rgbaRed)
	}

	for y := translatedBounds.Min.Y; y < translatedBounds.Max.Y; y++ {
		img.Set(translatedBounds.Min.X, y, rgbaRed)
		img.Set(translatedBounds.Max.X, y, rgbaRed)
	}
}

func countScreenSegmentElements(icon image.Image, label string, percentageLabel string, percentage *uint8) (bool, bool, bool, bool, uint) {
	count := uint(0)

	hasIcon := false
	if icon != nil {
		count++
		hasIcon = true
	}

	hasLabel := false
	if label != "" {
		count++
		hasLabel = true
	}

	hasPercentage := false
	if percentageLabel != "" {
		hasPercentage = true
		count++
	}

	hasBar := false
	if percentage != nil {
		hasBar = true
		count++
	}

	return hasIcon, hasLabel, hasPercentage, hasBar, count
}

func countSegments(label string, percentageLabel string, percentage *uint8) uint {
	count := uint(0)

	if label != "" {
		count++
	}
	if percentageLabel != "" {
		count++
	}
	if percentage != nil {
		count++
	}
	return count
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

func calculateIconSizeRatio(numberOfSegments uint) float64 {
	if numberOfSegments == 0 {
		return 1.0
	} else if numberOfSegments == 1 {
		return 0.66
	} else {
		return math.Round(1.0/float64(numberOfSegments)*100) / 100
	}
}

func createButtonImageElement(bounds image.Rectangle, dpi uint, fontColor color.Color, numberOfSegments uint, label string, percentageLabel string, percentage *uint8) image.Image {
	element := NewImageElementWithStrings(bounds, numberOfSegments, dpi, fontColor)

	centerVertically := false
	if numberOfSegments == 1 {
		centerVertically = true
	}

	if label != "" {

		element.DrawStringSegment(label, 0, centerVertically)
	}

	if percentageLabel != "" {
		element.DrawStringSegment(percentageLabel, 0, centerVertically)
	}

	if percentage != nil {
		element.DrawBarSegment(percentage)
	}

	return element.img
}
