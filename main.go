package main

import (
	"context"
	"os"

	"wechat-gif/cmd"
)

func main() {
	err := cmd.Cmd.Run(context.Background(), os.Args)
	if err != nil {
		return
	}
}
