export function createBranchName(key: string, summary: string) {
  const dashedSummary = summary
    .replace(/[^a-zA-Z0-9\s{1}]+/g, "")
    .trim()
    .replace(/\s+/g, "-")
    .toLowerCase();

  return `${key}-${dashedSummary}`;
}
