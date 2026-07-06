import * as GitlabApi from "../../gitlab/gitlab-api.ts";
import * as JiraApi from "../../jira/jira-api.ts";
import { resolveProjectConfig } from "../../config/resolve-project-config.ts";
import * as Logger from "../../utils/logger.ts";
import { getCurrentBranch } from "../../gitlab/git-branch.ts";
import {
  jiraKeyFromBranchName,
  jiraSummaryFromBranchName,
} from "../../utils/jira-from-branch-name.ts";
import { copyToClipboard } from "../../utils/clipboard.ts";
import { createTitleFromBranchName } from "../../utils/create-title-from-branch-name.ts";

export async function createMergeRequestCommand({
  labels,
  draft,
  targetBranch,
  title,
}: {
  labels?: string;
  draft?: boolean;
  targetBranch?: string;
  title?: string;
}) {
  const sourceBranch = await getCurrentBranch();

  if (!sourceBranch) {
    throw new Error("No current branch found.");
  }

  const jiraKey = jiraKeyFromBranchName(sourceBranch);
  const jiraSummary = jiraSummaryFromBranchName(sourceBranch);
  let mergeRequestTitle = title;

  if (!mergeRequestTitle) {
    mergeRequestTitle = jiraKey && jiraSummary
      ? `${jiraKey} ${jiraSummary}`
      : createTitleFromBranchName(sourceBranch);
  }

  if (draft) {
    mergeRequestTitle = `Draft: ${mergeRequestTitle}`;
  }

  const { gitlab, jira } = await resolveProjectConfig();
  const resolvedTargetBranch = targetBranch ?? gitlab.targetBranch ?? "master";

  const url = await GitlabApi.createMergeRequest(
    sourceBranch,
    resolvedTargetBranch,
    mergeRequestTitle,
    labels,
  );

  const clipboardSuccess = await copyToClipboard(url);
  if (clipboardSuccess) {
    Logger.success(`MR created, link copied to clipboard: ${url}`);
  } else {
    Logger.success(`MR created: ${url}`);
    Logger.info("Could not copy link to clipboard");
  }

  if (!jira.enabled || !jiraKey) {
    return;
  }

  const statusName = "In Review";
  await JiraApi.changeIssueStatus(jiraKey, statusName);

  Logger.success(`Changed status of issue ${jiraKey} to ${statusName}`);
}
