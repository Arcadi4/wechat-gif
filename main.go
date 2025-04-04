package main

import (
	"context"
	"os"
)

func main() {
	err := command.Run(context.Background(), os.Args)
	if err != nil {
		return
	}
}
