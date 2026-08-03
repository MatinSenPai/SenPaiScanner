package main

import (
	"embed"

	"github.com/matinsenpai/senpaiscanner/pkg/version"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/linux"
	"github.com/wailsapp/wails/v2/pkg/options/mac"
)

//go:embed all:frontend/dist
var assets embed.FS

//go:embed build/appicon.png
var appIcon []byte

func main() {
	app := NewApp()

	err := wails.Run(&options.App{
		Title:     "SenPai Scanner",
		Width:     1180,
		Height:    780,
		MinWidth:  960,
		MinHeight: 640,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 10, G: 12, B: 24, A: 255},
		Linux:            &linux.Options{Icon: appIcon},
		Mac: &mac.Options{About: &mac.AboutInfo{
			Title:   "SenPai Scanner",
			Message: "Signal Desk " + version.Version,
			Icon:    appIcon,
		}},
		OnStartup: app.startup,
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
