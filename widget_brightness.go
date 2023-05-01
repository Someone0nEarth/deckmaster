package main

import (
	"fmt"
	"image"
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

// Update renders the widget.
func (w *BrightnessWidget) Update() error {
	w.refreshBrightnessValue()

	var img image.Image

	if w.screenSegmentIndex != nil {
		img = w.updateScreenSegment()
	} else {
		img = w.updateButton()
	}

	return w.render(w.dev, img)
}

func (w *BrightnessWidget) updateButton() image.Image {
	buttonLayout := ButtonLayout{
		dpi: w.dev.DPI,
	}

	if w.icon != nil {
		buttonLayout.SetIcon(w.icon)
	}

	if w.label != "" {
		buttonLayout.AddText(w.label, ttfFont, w.color, 0, true)
	}

	if w.showPercentage {
		percentageLabel := fmt.Sprintf("%d%%", w.brightness)
		buttonLayout.AddText(percentageLabel, ttfFont, w.color, 0, true)
	}

	if w.showBar {
		buttonLayout.AddPercentageBar(w.brightness)
	}

	return buttonLayout.CreateImage(w.getMaxImageSize())
}

func (w *BrightnessWidget) updateScreenSegment() image.Image {
	segmentLayout := ScreenSegmentLayout{
		dpi: w.dev.ScreenVerticalDPI,
	}

	if w.icon != nil {
		segmentLayout.SetIcon(w.icon)
	}

	if w.label != "" {
		segmentLayout.SetLabel(w.label, ttfFont, w.color, -1, true)
	}

	if w.showPercentage {
		percentageLabel := fmt.Sprintf("%d%%", w.brightness)

		if w.showBar {
			segmentLayout.AddBlank()
			segmentLayout.AddText(percentageLabel, ttfFont, w.color, 1, true)
		} else {
			segmentLayout.AddBlank()
			segmentLayout.AddText(percentageLabel, ttfFont, w.color, 0, true)
			segmentLayout.AddBlank()
		}
	}

	if w.showBar {
		if !w.showPercentage {
			segmentLayout.AddBlank()
		}
		segmentLayout.AddPercentageBar(w.brightness)
	}

	return segmentLayout.CreateImage(w.getMaxImageSize())
}

func (w *BrightnessWidget) refreshBrightnessValue() {
	w.brightness = getBrightness()
}

func getBrightness() uint8 {
	return uint8(*brightness)
}
