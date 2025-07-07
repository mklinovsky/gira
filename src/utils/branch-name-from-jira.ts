export function createBranchName(key: string, summary: string) {
  const dashedSummary = summary
    .replace(/[^a-zA-Z0-9\s{1}_-]+/g, "")
    .trim()
    .replace(/[\s_]+/g, "-")
    .toLowerCase();

  return `${key}-${dashedSummary}`;
}
