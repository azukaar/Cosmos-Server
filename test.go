package main 

import (
	"fmt"
	"strings"
)

func main() {
	target := ":80";

	destinationArr := strings.Split(target, "://")
	listenProtocol := "tcp"

	if len(destinationArr) > 1 {
			target = destinationArr[1]
			listenProtocol = destinationArr[0]
	} else {
			target = destinationArr[0]
	}

	fmt.Println(listenProtocol)
	fmt.Println(target)
}