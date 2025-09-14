import { $ } from "zx";

export async function copyToClipboard(text: string): Promise<boolean> {
  try {
    switch (Deno.build.os) {
      case "darwin":
        await $`echo ${text} | pbcopy`.quiet();
        break;
      case "linux":
        // Try xclip first, fall back to xsel
        try {
          await $`echo ${text} | xclip -selection clipboard`.quiet();
        } catch {
          await $`echo ${text} | xsel --clipboard --input`.quiet();
        }
        break;
      case "windows":
        await $`echo ${text} | clip`.quiet();
        break;
      default:
        return false;
    }
    return true;
  } catch {
    return false;
  }
}
