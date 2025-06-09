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
			Value:   false,
		},
		&cli.BoolFlag{
			Name:    "autoplay",
			Usage:   "Compress further so it can autoplay (<1MiB)",
			Aliases: []string{"a"},
			Value:   false,
		},
		&cli.IntFlag{
			Name:    "workers",
			Usage:   "Number of concurrent workers (default: 4)",
			Aliases: []string{"w"},
			Value:   4,
		},
	},
	Action: action,
}

func action(ctx context.Context, c *cli.Command) (err error) {
	args := c.Args().Slice()
	if len(args) == 0 {
		cli.ShowAppHelpAndExit(c, 0)
	}

	var gifs []*gifImg
	if c.Bool("dir") {
		gifs = loadGifsFromDirectory(args)
	} else {
		gifs = loadGifsFromPaths(args)
	}

	processGifs(gifs, c.Bool("autoplay"), int(c.Int("workers")))
	return nil
}
