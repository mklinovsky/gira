const STRIPPABLE_BRANCH_PREFIXES = new Set([
  "feat",
  "feature",
  "fix",
  "bugfix",
  "chore",
  "refactor",
  "docs",
  "test",
  "ci",
  "build",
  "perf",
  "style",
  "hotfix",
  "release",
  "wip",
]);

export function createTitleFromBranchName(branchName: string): string {
  const branchSegments = branchName.trim().split("/").filter(Boolean);
  let titlePrefix: string | undefined;

  while (branchSegments.length > 1) {
    const firstSegment = branchSegments[0]?.toLowerCase();

    if (!firstSegment || !STRIPPABLE_BRANCH_PREFIXES.has(firstSegment)) {
      break;
    }

    if (!titlePrefix) {
      titlePrefix = firstSegment;
    }

    branchSegments.shift();
  }

  const title = branchSegments
    .join("/")
    .split(/[\/_-]+/)
    .filter(Boolean)
    .join(" ");

  if (!title) {
    throw new Error("No title found.");
  }

  const formattedTitle = title.charAt(0).toUpperCase() + title.slice(1);

  if (titlePrefix) {
    return `${titlePrefix}: ${formattedTitle}`;
  }

  return formattedTitle;
}
