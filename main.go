package main

import (
	"fmt"

	"github.com/Leosocy/gipp/pkg/proxy"
)

func main() {
	pxy, _ := proxy.NewProxy("54.39.98.135", "3128")
	if pxy != nil {
		pxy.DetectAnonymity()
	}
	fmt.Printf("%+v", pxy)
}
