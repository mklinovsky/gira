import * as GitlabApi from "../../gitlab/gitlab-api.ts";
import * as JiraApi from "../../jira/jira-api.ts";
import * as Logger from "../../utils/logger.ts";
import { $ } from "zx";
import { getCurrentBranch } from "../../gitlab/git-branch.ts";
import {
  jiraKeyFromBranchName,
  jiraSummaryFromBranchName,
} from "../../utils/jira-from-branch-name.ts";
import { IssueStatus, IssueStatusById } from "../../jira/jira.types.ts";

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

  await $`echo ${url} | pbcopy`;
  Logger.success(`MR created, link copied to clipboard: ${url}`);
  await JiraApi.changeIssueStatus(jiraKey ?? "", IssueStatus.InReview);

  Logger.success(
    `Changed status of issue ${jiraKey} to ${
      IssueStatusById[IssueStatus.InReview]
    }`,
  );
}
