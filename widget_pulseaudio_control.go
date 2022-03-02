package main

import (
	"fmt"
	"image"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"sync"
)

const (
	regexExpression     = `index: ([0-9]+)[\s\S]*?muted: (no|yes)[\s\S]*?media.name = \"(.*?)\"[\s\S]*?application.name = \"(.*?)\"[\s\S]*?application.process.id = \"(.*?)\"`
	regexGroupClientId  = 1
	regexGroupMuted     = 2
	regexGroupMediaName = 3
	regexGroupAppName   = 4
	regexGroupAppPid    = 5

	listInputSinksCommand = "pacmd list-sink-inputs"
)

// PulseAudioControlWidget is a widget displaying a recently activated window.
type PulseAudioControlWidget struct {
	*ButtonWidget

	appName    string
	mode       string
	showTitle  bool
	useAppIcon bool

	update      bool
	updateMutex sync.RWMutex

	failingAppIcon bool
}

type sinkInputData struct {
	muted  bool
	title  string
	index  string
	appPid uint64
}

// NewPulseAudioControlWidget returns a new PulseAudioControlWidget.
func NewPulseAudioControlWidget(bw *BaseWidget, opts WidgetConfig) (*PulseAudioControlWidget, error) {
	var appName string
	if err := ConfigValue(opts.Config["appName"], &appName); err != nil {
		return nil, err
	}

	var mode string
	if err := ConfigValue(opts.Config["mode"], &mode); err != nil {
		return nil, err
	}

	var showTitle bool
	_ = ConfigValue(opts.Config["showTitle"], &showTitle)

	var useAppIcon bool
	_ = ConfigValue(opts.Config["useAppIcon"], &useAppIcon)

	widget, err := NewButtonWidget(bw, opts)

	if err != nil {
		return nil, err
	}

	widget.assetDir = "pulseaudio_control"

	return &PulseAudioControlWidget{
		ButtonWidget: widget,
		appName:      appName,
		mode:         mode,
		showTitle:    showTitle,
		useAppIcon:   useAppIcon,

		failingAppIcon: false,
	}, nil
}

// RequiresUpdate returns true when the widget wants to be repainted.
func (w *PulseAudioControlWidget) RequiresUpdate() bool {
	return w.updateRequired() || w.BaseWidget.RequiresUpdate()
}

func (w *PulseAudioControlWidget) updateRequired() bool {
	w.updateMutex.RLock()
	defer w.updateMutex.RUnlock()

	return w.update
}

// Update renders the widget.
func (w *PulseAudioControlWidget) Update() error {
	sinkInputData, err := getSinkInputDataForApp(w.appName)
	if err != nil {
		return err
	}

	iconImage := w.getIcon(sinkInputData)
	w.SetImage(iconImage)

	w.label = stripTextTo(10, w.getLabel(sinkInputData))

	return w.ButtonWidget.Update()
}

func (w *PulseAudioControlWidget) getIcon(sinkInputData *sinkInputData) image.Image {
	var appImage image.Image
	if w.showAppIcon(sinkInputData) {
		appImage, _ = xorg.getIconFromWindow(uint(sinkInputData.appPid))

		if appImage != nil {
			return appImage
		} else {
			w.failingAppIcon = true
			fmt.Fprintf(os.Stderr, "not able to receive icon from application '%s'. Using default icon as fallback.\n", w.appName)
		}
	}

	return w.loadThemeOrWidgetAssetIcon(w.getIconName(sinkInputData))
}

func (w *PulseAudioControlWidget) showAppIcon(sinkInputData *sinkInputData) bool {
	return w.useAppIcon && !w.failingAppIcon && sinkInputData != nil && !sinkInputData.muted
}

func (w *PulseAudioControlWidget) getLabel(sinkInputData *sinkInputData) string {
	if w.showTitle && sinkInputData != nil && sinkInputData.title != "" {
		return sinkInputData.title
	}
	return w.appName
}

func (w *PulseAudioControlWidget) getIconName(sinkInputData *sinkInputData) string {
	if sinkInputData == nil {
		return "not_playing"
	}
	if sinkInputData.muted {
		return "muted"
	} else {
		return "playing"
	}
}

// TriggerAction gets called when a button is pressed.
func (w *PulseAudioControlWidget) TriggerAction(hold bool) {
	w.updateMutex.Lock()
	defer w.updateMutex.Unlock()
	w.update = true

	sinkInputData, err := getSinkInputDataForApp(w.appName)

	if err != nil {
		fmt.Fprintln(os.Stderr, "can't get sink date for pulseaudio app "+w.appName, err)
		return
	}

	if sinkInputData == nil {
		fmt.Fprintln(os.Stderr, "no running sink found for pulseaudio app "+w.appName, err)
		return
	}

	switch w.mode {
	case "mute":
		toggleMute(sinkInputData.index)
	case "up":
		volumeUp(sinkInputData.index)
	case "down":
		volumeDown(sinkInputData.index)
	default:
		fmt.Fprintln(os.Stderr, "unkown pulseaudio control mode: "+w.appName)
	}
}

func toggleMute(sinkIndex string) {
	err := exec.Command("sh", "-c", "pactl set-sink-input-mute "+sinkIndex+" toggle").Run()

	if err != nil {
		fmt.Fprintln(os.Stderr, "can't toggle mute for pulseaudio sink index: "+sinkIndex, err)
	}
}

func volumeUp(sinkIndex string) {
	err := exec.Command("sh", "-c", "pactl set-sink-input-volume "+sinkIndex+" +6554").Run()

	if err != nil {
		fmt.Fprintln(os.Stderr, "can't volume up for pulseaudio sink index: "+sinkIndex, err)
	}
}

func volumeDown(sinkIndex string) {
	err := exec.Command("sh", "-c", "pactl set-sink-input-volume "+sinkIndex+" -6554").Run()

	if err != nil {
		fmt.Fprintln(os.Stderr, "can't volume down for pulseaudio sink index: "+sinkIndex, err)
	}
}

func stripTextTo(maxLength int, text string) string {
	runes := []rune(text)
	if len(runes) > maxLength {
		return string(runes[:maxLength])
	}
	return text
}

func getSinkInputDataForApp(appName string) (*sinkInputData, error) {
	output, err := exec.Command("sh", "-c", listInputSinksCommand).Output()
	if err != nil {
		return nil, fmt.Errorf("can't get pulseaudio sinks. 'pacmd' missing? %s", err)
	}

	var regex = regexp.MustCompile(regexExpression)
	matches := regex.FindAllStringSubmatch(string(output), -1)
	var sinkData *sinkInputData
	for match := range matches {
		if appName == matches[match][regexGroupAppName] {
			sinkData = &sinkInputData{}
			sinkData.index = matches[match][regexGroupClientId]
			sinkData.muted = yesOrNoToBool(matches[match][regexGroupMuted])
			sinkData.title = matches[match][regexGroupMediaName]
			sinkData.appPid, _ = strconv.ParseUint(matches[match][regexGroupAppPid], 10, 64)
		}
	}

	return sinkData, nil
}

func yesOrNoToBool(yesOrNo string) bool {
	switch yesOrNo {
	case "yes":
		return true
	case "no":
		return false
	}
	fmt.Fprintln(os.Stderr, "can't convert yes|no to bool: "+yesOrNo)
	return false
}
