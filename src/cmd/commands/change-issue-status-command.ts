import * as JiraApi from "../../jira/jira-api.ts";
import * as Git from "../../gitlab/git-branch.ts";
import * as Logger from "../../utils/logger.ts";
import { jiraKeyFromBranchName } from "../../utils/jira-from-branch-name.ts";
import { IssueStatus } from "../../jira/jira.types.ts";

export async function changeIssueStatusCommand({
  issue,
  status,
}: {
  issue?: string;
  status: string;
}) {
  let issueKey = issue;
  if (!issueKey) {
    const currentBranch = await Git.getCurrentBranch();
    if (!currentBranch) {
      throw new Error("No current branch found.");
    }

    issueKey = jiraKeyFromBranchName(currentBranch) ?? "";
  }

  if (!issueKey) {
    throw new Error("No issue key found.");
  }

  await JiraApi.changeIssueStatus(
    issueKey,
    IssueStatus[status as keyof typeof IssueStatus],
  );

  Logger.success(`Changed status of issue ${issueKey} to ${status}`);
}
