package main

import (
	"fmt"
	"image"
	"image/color"

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

	var barLength int
	barThickness := 10

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
		labelBounds := segmentImage.Bounds()

		labelBounds.Min.X = 0
		labelBounds.Max.X = 100

		labelBounds.Min.Y = 0
		labelBounds.Max.Y = 20

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

func createButtonImage(bounds image.Rectangle, dpi uint, fontColor color.Color, icon image.Image, label string, percentageLabel string, percentage *uint8) image.Image {
	imageHeight := bounds.Dy()
	imageWidth := bounds.Dx()

	margin := imageHeight / 18
	height := imageHeight - (margin * 2)
	width := imageWidth - (margin * 2)
	img := image.NewRGBA(image.Rect(0, 0, imageWidth, imageHeight))

	elementCount := 0

	if label != "" {
		elementCount++
	}
	if percentageLabel != "" {
		elementCount++
	}
	if percentage != nil {
		elementCount++
	}

	iconRatio := 0.0
	if icon != nil {
		switch elementCount {
		case 0:
			iconRatio = 1.0
		case 1:
			iconRatio = 0.66
		case 2:
			iconRatio = 0.5
		case 3:
			iconRatio = 0.33
		}
	}

	iconSize := int(float64(height) * iconRatio)

	if icon != nil {
		drawImageWithResizing(img,
			icon,
			iconSize,
			image.Pt(-1, margin))
	}

	if elementCount > 0 {
		elementsBounds := image.Rectangle{
			Min: image.Point{
				X: 0,
				Y: 0,
			},
			Max: image.Point{
				X: width,
				Y: height - (iconSize + margin),
			},
		}

		elements := createElementsImage(elementsBounds, dpi, fontColor, elementCount, label, percentageLabel, percentage)

		drawImage(img, elements, image.Pt(margin, iconSize+margin))
	}

	return img
}

func createElementsImage(bounds image.Rectangle, dpi uint, fontColor color.Color, elementCount int, label string, percentageLabel string, percentage *uint8) image.Image {
	width := bounds.Dx()
	height := bounds.Dy()

	margin := height / (3 * elementCount)
	if margin < 5 {
		margin = 5
	}

	img := image.NewRGBA(image.Rect(0, 0, width, height))
	segmentsSize := (height - ((elementCount - 1) * margin)) / elementCount

	segmentPositionY := 0
	if label != "" {
		labelBounds := image.Rectangle{}

		labelBounds.Min.X = 0
		labelBounds.Max.X = width

		labelBounds.Min.Y = 0
		labelBounds.Max.Y = segmentsSize

		var y int
		if elementCount == 1 {
			y = -1
		} else {
			y = segmentPositionY + segmentsSize

		}
		drawString(img, labelBounds, ttfFont, label, dpi, 0, fontColor, image.Pt(-1, y))

		segmentPositionY += (segmentsSize + margin)
	}

	if percentageLabel != "" {

		labelBounds := image.Rectangle{}

		labelBounds.Min.X = 0
		labelBounds.Max.X = width

		labelBounds.Min.Y = 0
		labelBounds.Max.Y = segmentsSize

		var y int
		if elementCount == 1 {
			y = -1
		} else {
			y = segmentPositionY + segmentsSize

		}
		drawString(img, labelBounds, ttfFont, percentageLabel, dpi, 0, fontColor, image.Pt(-1, y))

		segmentPositionY += segmentsSize + margin
	}

	if percentage != nil {
		bar := createBar(width, segmentsSize-2, *percentage)

		drawImage(img, bar, image.Pt(0, segmentPositionY))
	}

	return img
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

func (w *BrightnessWidget) calculateWidth(img *image.RGBA, font *truetype.Font, text string, fontsize float64) int {
	extent, _ := ftContext(img, font, w.dev.DPI, fontsize).DrawString(text, freetype.Pt(0, 0))
	return extent.X.Floor()
}

func (w *BrightnessWidget) refreshBrightnessValue() {
	w.brightness = getBrightness()
}
