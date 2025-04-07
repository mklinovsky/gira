import { $ } from "zx";

export const Git = {
  createBranch,
  getCurrentBranch,
};

async function createBranch(branchName: string) {
  const { stdout } = await $`git rev-parse --abbrev-ref HEAD`;
  const currentBranch = stdout.trim();

  if (currentBranch === branchName) {
    console.log(`Already on branch ${branchName}`);
    return;
  }

  await $`git checkout -b ${branchName}`;
  console.log(`Created and switched to branch ${branchName}`);
}

async function getCurrentBranch() {
  const { stdout } = await $`git rev-parse --abbrev-ref HEAD`;
  return stdout.trim();
}
