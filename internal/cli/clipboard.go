package cli

import "context"

func copyToClipboard(ctx context.Context, runner Runner, goos, text string) bool {
	payload := text + "\n"

	switch goos {
	case "darwin":
		return runner.RunWithInput(ctx, payload, "pbcopy") == nil
	case "linux":
		if runner.RunWithInput(ctx, payload, "xclip", "-selection", "clipboard") == nil {
			return true
		}

		return runner.RunWithInput(ctx, payload, "xsel", "--clipboard", "--input") == nil
	case "windows":
		return runner.RunWithInput(ctx, payload, "clip") == nil
	}

	return false
}
