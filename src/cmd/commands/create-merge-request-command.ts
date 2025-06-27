import * as GitlabApi from "../../gitlab/gitlab-api.ts";
import * as JiraApi from "../../jira/jira-api.ts";
import { $ } from "zx";
import { getCurrentBranch } from "../../gitlab/git-branch.ts";
import {
  jiraKeyFromBranchName,
  jiraSummaryFromBranchName,
} from "../../utils/jira-from-branch-name.ts";
import { IssueStatus, IssueStatusById } from "../../jira/jira.types.ts";

export async function createMergeRequestCommand() {
  const targetBranch = "master";
  const sourceBranch = await getCurrentBranch();

  if (!sourceBranch) {
    throw new Error("No current branch found.");
  }

  const jiraKey = jiraKeyFromBranchName(sourceBranch);
  const title = `${jiraKey} ${jiraSummaryFromBranchName(sourceBranch)}`;

  if (!title) {
    throw new Error("No title found.");
  }

  const url = await GitlabApi.createMergeRequest(
    sourceBranch,
    targetBranch,
    title,
  );

  await $`echo ${url} | pbcopy`;
  console.log("✅ MR created, link copied to clipboard");
  console.log(url);

  await JiraApi.changeIssueStatus(jiraKey ?? "", IssueStatus.InReview);

  console.log(
    `✅ Changed status of issue ${jiraKey} to ${
      IssueStatusById[IssueStatus.InReview]
    }`,
  );
}
