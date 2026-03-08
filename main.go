package main

import (
	"embed"
	"log/slog"
	"os"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/mac"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	setupLogger()

	app := NewApp()

	err := wails.Run(&options.App{
		Title:     "Vinyl",
		Width:     380,
		Height:    520,
		MinWidth:  380,
		MinHeight: 520,
		MaxWidth:  380,
		MaxHeight: 520,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 26, G: 20, B: 16, A: 255},
		OnStartup:        app.startup,
		OnShutdown:       app.shutdown,
		Frameless:        true,
		Mac: &mac.Options{
			TitleBar:             mac.TitleBarHiddenInset(),
			WebviewIsTransparent: true,
			WindowIsTranslucent:  true,
			About: &mac.AboutInfo{
				Title:   "Vinyl",
				Message: "Every song is a memory. Vinyl remembers them all.",
			},
		},
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		slog.Error("application error", "err", err)
		os.Exit(1)
	}
}

func setupLogger() {
	handler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	slog.SetDefault(slog.New(handler))
}
