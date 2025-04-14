export function jiraKeyFromBranchName(branchName: string) {
  const jiraKeyPattern = /(^[A-Z]+-\d+)/;
  const match = branchName.match(jiraKeyPattern);
  return match ? match[1] : undefined;
}

export function jiraSummaryFromBranchName(branchName: string) {
  const jiraKeyPattern = /(^[A-Z]+-\d+)-(.+)/;
  const match = branchName.match(jiraKeyPattern);
  const summary = match ? match[2].replace(/-/g, " ") : undefined;

  return summary
    ? summary.charAt(0).toUpperCase() + summary.slice(1)
    : undefined;
}
