package cli

import (
	"context"

	"github.com/urfave/cli/v3"
)

func newGetMrCommand(a *app) *cli.Command {
	return &cli.Command{
		Name:         "get-mr",
		Usage:        "Get details of a merge request",
		ArgsUsage:    "<mergeRequestId>",
		OnUsageError: onUsageError,
		Action: func(ctx context.Context, c *cli.Command) error {
			if err := requireArgs(c, 1); err != nil {
				return err
			}

			gitlab, _, err := a.gitlab(false)
			if err != nil {
				return err
			}

			raw, err := gitlab.MergeRequestRaw(ctx, c.Args().Get(0))
			if err != nil {
				return err
			}

			return printJSON(a.e.stdout, raw)
		},
	}
}
