package web

import (
	"embed"
	"io/fs"
)

//go:embed dist
var embedded embed.FS

var Assets fs.FS

func init(){
	var err error
	Assets,err=fs.Sub(embedded,"dist")
	if err!=nil{panic(err)}
}
