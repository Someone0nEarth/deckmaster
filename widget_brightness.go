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

	showBar := false
	if opts.Config["bar"] != nil {
		_ = ConfigValue(opts.Config["bar"], &showBar)
	} else if bw.screenSegmentIndex != nil {
		showBar = true
	}

	showPercentage := false
	if opts.Config["percentage"] != nil {
		_ = ConfigValue(opts.Config["percentage"], &showPercentage)
	} else if bw.screenSegmentIndex != nil {
		showPercentage = true
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
		//brightness: uint8(*brightness),
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
	if w.screenSegmentIndex != nil {
		return w.updateScreenSegment()
	}

	return w.updateButton()
}

func (w *BrightnessWidget) updateScreenSegment() error {

	if !w.showBar && !w.showPercentage {
		return w.ButtonWidget.Update()
	}

	w.refreshBrightnessValue()

	showLabel := w.label != ""
	showIcon := w.icon != nil

	percentageString := fmt.Sprintf("%d%%", w.brightness)

	if !showLabel && !w.showBar {
		return w.updateButtonUsingLabel(percentageString)
	}

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
		bar := w.createBar(barLength, barThickness, int(w.brightness))

		x := 60
		y := 75
		drawImage(segmentImage, bar, image.Pt(x, y))
	}

	if w.showPercentage {

		var fontsize float64
		var xOffset int

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

	if !w.showBar && !w.showPercentage {
		return w.ButtonWidget.Update()
	}

	if w.showBar {
		//TODO
		fatal("Not supported yet: Showing brightness bar on normal buttons.")
	}

	if w.label != "" {
		//TODO Atm w.label is "used" for supporting percentage drawing on normal buttons
		fatal("Not supported yet: Showing brightness percentage on normal buttons when label is set.")
	}

	w.refreshBrightnessValue()
	brightnessString := fmt.Sprintf("%d%%", w.brightness)

	if w.label == "" && w.showPercentage {
		return w.updateButtonUsingLabel(brightnessString)
	}

	//TODO
	fatal("Unexpected state in widget_brightness updateButton.")
	return nil
}

func (w *ButtonWidget) updateButtonUsingLabel(label string) error {
	imageSize := w.getMaxImageSize()
	imageHeight := imageSize.Dy()
	imageWidth := imageSize.Dx()

	margin := imageHeight / 18
	height := imageHeight - (margin * 2)
	img := image.NewRGBA(image.Rect(0, 0, imageWidth, imageHeight))

	iconSize := int((float64(height) / 3.0) * 2.0)
	bounds := img.Bounds()

	if w.icon != nil {
		err := drawImageWithResizing(img,
			w.icon,
			iconSize,
			image.Pt(-1, margin))

		if err != nil {
			return err
		}

		bounds.Min.Y += iconSize + margin
		bounds.Max.Y -= margin
	}

	drawString(img,
		bounds,
		ttfFont,
		label,
		w.dev.DPI,
		w.fontsize,
		w.color,
		image.Pt(-1, -1))

	return w.render(w.dev, img)
}

func (w *BrightnessWidget) createBar(length int, thickness int, percentage int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, length, thickness))

	thickPart := int((float64(length) / 100.0) * float64(percentage))

	for x := 0; x < thickPart; x++ {
		for y := 0; y < thickness; y++ {
			img.Set(x, y, color.White)
		}
	}

	yOffset := thickness / 10 * 2
	for x := thickPart; x < length; x++ {
		for y := 0; y < thickness-5; y++ {
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
