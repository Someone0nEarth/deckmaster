package main

import (
	"fmt"
	"image"
	"image/color"
	"math"

	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
)

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

func (w *BrightnessWidget) updateScreenSegment() error {

	if !w.showBar && !w.showPercentage {
		return w.updateButton()
	}

	showLabel := w.label != ""
	showIcon := w.icon != nil

	segmentSize := w.getMaxImageSize()
	segmentWidth := segmentSize.Dx()
	segmentHeight := segmentSize.Dy()

	segmentImage := image.NewRGBA(image.Rect(0, 0, segmentWidth, segmentHeight))

	var barLength uint
	barThickness := uint(10)

	iconSize := 50
	if showIcon {
		drawImageWithResizing(segmentImage, w.icon, iconSize, image.Pt(0, 30))
	}

	if w.showBar {
		barLength = 120
		bar := createBar(barLength, barThickness, w.brightness)

		x := 60
		y := 75
		drawImage(segmentImage, bar, image.Pt(x, y))
	}

	if w.showPercentage {
		var fontsize float64
		var xOffset int

		percentageString := fmt.Sprintf("%d%%", w.brightness)

		if w.showBar {
			fontsize = 12.0
			actualWidth := w.calculateWidth(segmentImage, ttfFont, percentageString, fontsize)
			xOffset = actualWidth + 20
		} else {
			fontsize = 20.0
			actualWidth := w.calculateWidth(segmentImage, ttfFont, percentageString, fontsize)
			xOffset = actualWidth + 50
		}

		x := segmentWidth - xOffset
		y := 65
		drawString(segmentImage, image.Rectangle{}, ttfFont, percentageString, w.dev.DPI, fontsize, w.color, image.Pt(x, y))
	}

	if showLabel {
		labelBounds := createRectangle(100, 20)

		x := 0
		y := 20
		drawString(segmentImage, labelBounds, ttfFont, w.label, w.dev.DPI, 0, w.color, image.Pt(x, y))
	}

	return w.render(w.dev, segmentImage)
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

	buttonImage := createButtonImage(w.getMaxImageSize(), w.dev.DPI, w.color, w.icon, w.label, percentageLabel, percentage)

	return w.render(w.dev, buttonImage)
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

	margin := uint(imageHeight / 18)
	height := uint(imageHeight - (margin * 2))
	width := uint(imageWidth - (margin * 2))

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

		drawImage(img, iconElement.img, image.Pt(int(margin), int(margin)))
	}

	if numberOfSegments > 0 {
		elementBounds := createRectangle(width, height-(iconSize+interspaceElements))

		elementImage := createImageElement(elementBounds, dpi, fontColor, numberOfSegments, label, percentageLabel, percentage)

		y := int(margin + iconSize + interspaceElements)
		drawImage(img, elementImage, image.Pt(int(margin), y))
	}

	return img
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

func createImageElement(bounds image.Rectangle, dpi uint, fontColor color.Color, numberOfSegments uint, label string, percentageLabel string, percentage *uint8) image.Image {
	element := NewImageElementWithStrings(bounds, numberOfSegments, dpi, fontColor)

	if label != "" {
		element.DrawStringSegment(label)
	}

	if percentageLabel != "" {
		element.DrawStringSegment(percentageLabel)
	}

	if percentage != nil {
		element.DrawBarSegment(percentage)
	}

	return element.img
}
