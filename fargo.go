package main

import (
    "os"

    "github.com/urfave/cli"
)

var (
    app = cli.NewApp()
    cfg Config
)

func main() {
    app.Name = "fargo"
    app.Usage = "Static Site Generator"
    app.Version = "0.0.1"
    app.Run(os.Args)
}
