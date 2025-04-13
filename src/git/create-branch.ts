import { $ } from "zx";

export async function createBranch(branchName: string) {
  const currentBranch = await getCurrentBranch();

  if (currentBranch === branchName) {
    console.log(`Already on branch ${branchName}`);
    return;
  }

  await $`git checkout -b ${branchName}`;
  console.log(`Switched to branch ${branchName}`);
}

export async function getCurrentBranch() {
  const { stdout } = await $`git rev-parse --abbrev-ref HEAD`;
  return stdout.trim();
}
