// internal/cli/progress/spinner_style.go

package progress

import "github.com/vbauerster/mpb/v8"

// defaultBouncingBarStyle returns a default style for bouncing progress bars, which can be used for indefinite tasks where total progress is unknown.
func defaultBouncingBarStyle() mpb.BarFillerBuilder {
	return BouncingBarStyle().Lbound("[").Filler("█").Padding("░").Rbound("]")
}

// defaultBarStyle returns a default style for progress bars, which can be used if no custom
// style is provided.
func defaultBarStyle() mpb.BarFillerBuilder {
	return mpb.BarStyle().Lbound("[").Filler("█").Tip("█").Padding("░").Rbound("]")
}

// defaultSpinnerStyle returns a default style for spinner bars.
// The spinner will cycle through the specified characters to indicate progress.
func defaultSpinnerStyle() mpb.BarFillerBuilder {
	return mpb.SpinnerStyle(
		">==============<",
		"=>============<=",
		"==>==========<==",
		"===>========<===",
		"====>======<====",
		"=====>====<=====",
		"======>==<======",
		"=======><=======",
		"=======<>=======",
		"======<==>======",
		"=====<====>=====",
		"====<======>====",
		"===<========>===",
		"==<==========>==",
		"=<============>=",
		"<==============>",
	)
}
