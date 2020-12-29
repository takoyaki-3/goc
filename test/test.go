package main

import (
	"fmt"
	"github.com/takoyaki-3/goc"
)

func main(){
	fmt.Println("start")
	err := goc.ZipArchive("zip.zip",[]string{"test.go"})
	if err != nil{
		fmt.Println(err)
	}
}