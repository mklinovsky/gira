import * as GitlabApi from "../../gitlab/gitlab-api.ts";
import * as JiraApi from "../../jira/jira-api.ts";
import * as Logger from "../../utils/logger.ts";
import { getCurrentBranch } from "../../gitlab/git-branch.ts";
import {
  jiraKeyFromBranchName,
  jiraSummaryFromBranchName,
} from "../../utils/jira-from-branch-name.ts";
import { copyToClipboard } from "../../utils/clipboard.ts";

export async function createMergeRequestCommand({
  labels,
  draft,
  targetBranch = "master",
}: {
  labels?: string;
  draft?: boolean;
  targetBranch?: string;
}) {
  const sourceBranch = await getCurrentBranch();

  if (!sourceBranch) {
    throw new Error("No current branch found.");
  }

  const jiraKey = jiraKeyFromBranchName(sourceBranch);
  let title = `${jiraKey} ${jiraSummaryFromBranchName(sourceBranch)}`;

  if (!title) {
    throw new Error("No title found.");
  }

  if (draft) {
    title = `Draft: ${title}`;
  }

  const url = await GitlabApi.createMergeRequest(
    sourceBranch,
    targetBranch,
    title,
    labels,
  );

  const clipboardSuccess = await copyToClipboard(url);
  if (clipboardSuccess) {
    Logger.success(`MR created, link copied to clipboard: ${url}`);
  } else {
    Logger.success(`MR created: ${url}`);
    Logger.info("Could not copy link to clipboard");
  }

  const statusName = "In Review";
  await JiraApi.changeIssueStatus(jiraKey ?? "", statusName);

  Logger.success(`Changed status of issue ${jiraKey} to ${statusName}`);
}
