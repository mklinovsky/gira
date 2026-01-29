import { $ } from "zx";
import * as Logger from "../utils/logger.ts";
import { join } from "@std/path";

export async function createBranch(branchName: string) {
  const currentBranch = await getCurrentBranch();

  if (currentBranch === branchName) {
    Logger.info(`Already on branch ${branchName}`);
    return;
  }

  try {
    await $`git checkout -b ${branchName}`.quiet();
    Logger.info(`Switched to new branch: ${branchName}`);
  } catch (error) {
    Logger.error(error);
  }
}

export async function createWorktree(branchName: string, baseDir: string) {
  const worktreePath = join(baseDir, branchName);

  try {
    await $`git worktree add ${worktreePath} -b ${branchName}`.quiet();
    Logger.success(`Created worktree at ${worktreePath}`);
    Logger.info(`Run: cd ${worktreePath}`);
  } catch (error) {
    Logger.error(error);
  }
}

export async function getCurrentBranch() {
  const { stdout } = await $`git rev-parse --abbrev-ref HEAD`;
  return stdout.trim();
}
