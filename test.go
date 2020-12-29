package main

import (
	"fmt"
	"github.com/takoyaki-3/goc"
)

func main(){
	fmt.Println("start")
	goc.ZipArchive("zip.zip",[]string{"./goc.go"})
}