package interfaces

import "github.com/urfave/cli/v2"

type MarketBot interface {
	Run(c *cli.Context)
}
