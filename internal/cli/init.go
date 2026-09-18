package cli

import (
	"context"

	"github.com/mklinovsky/gira/internal/gira"
	"github.com/urfave/cli/v3"
)

func newInitCommand(a *app) *cli.Command {
	return &cli.Command{
		Name:         "init",
		Usage:        "Create default Gira config file",
		OnUsageError: onUsageError,
		Flags: []cli.Flag{
			&cli.BoolFlag{Name: "force", Usage: "Overwrite existing config file"},
		},
		Action: func(_ context.Context, c *cli.Command) error {
			if err := requireArgs(c, 0); err != nil {
				return err
			}

			homeDir, err := gira.HomeDir()
			if err != nil {
				return err
			}

			configPath, err := gira.InitConfig(homeDir, c.Bool("force"))
			if err != nil {
				return err
			}

			a.e.success("Created config at %s", configPath)

			return nil
		},
	}
}
