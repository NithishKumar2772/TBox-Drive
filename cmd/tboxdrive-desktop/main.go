//go:build desktop

package main

import (
	"log"

	"github.com/NithishKumar2772/TBox-Drive/internal/app/web"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

func main(){
	desktop:=NewDesktopApp()
	err:=wails.Run(&options.App{
		Title:"TBox Drive",
		Width:1360,
		Height:900,
		MinWidth:1000,
		MinHeight:680,
		AssetServer:&assetserver.Options{Assets:web.Assets},
		OnStartup:desktop.startup,
		OnShutdown:desktop.shutdown,
		Bind:[]interface{}{desktop},
	})
	if err!=nil{log.Fatal(err)}
}
