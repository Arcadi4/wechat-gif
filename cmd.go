package main

import (
	"context"

	"github.com/urfave/cli/v3"
)

var command = &cli.Command{
	Name:  "wechat-gif",
	Usage: "Compress gif so it can be sent via WeChat",
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:    "dir",
			Usage:   "Process all gif files in given directory",
			Aliases: []string{"d"},
		},
		&cli.BoolFlag{
			Name:    "autoplay",
			Usage:   "Compress further so it can autoplay (<1MiB)",
			Aliases: []string{"a"},
		},
	},
	Action: action,
}

func action(ctx context.Context, c *cli.Command) (err error) {
	args := c.Args().Slice()
	if len(args) == 0 {
		cli.ShowAppHelpAndExit(c, 0)
	}

	var objs []*gifImg
	if c.Bool("dir") {
		objs = processDirectory(args)
	} else {
		objs = processFiles(args)
	}

	processGifs(objs, c)
	return nil
}
