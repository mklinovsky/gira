import * as JiraApi from "../../jira/jira-api.ts";
import * as Git from "../../gitlab/git-branch.ts";
import { createBranchName } from "../../utils/branch-name-from-jira.ts";
import { IssueStatus, IssueStatusById } from "../../jira/jira.types.ts";

type CreateJiraIssueCommand = {
  summary: string;
  options: {
    parent?: string;
    type?: string;
    assign?: boolean;
    branch?: boolean;
    start?: boolean;
  };
};

export async function createJiraIssueCommand(args: CreateJiraIssueCommand) {
  const {
    parent,
    type,
    assign: assignToMe,
    branch: createBranch,
    start: startProgress,
  } = args.options;

  const createdIssue = await JiraApi.createIssue(
    args.summary,
    type,
    parent,
    assignToMe,
  );

  console.log(`✅ Issue created: ${createdIssue.url}`);

  if (!createBranch) {
    return;
  }

  const branchName = createBranchName(createdIssue.key, args.summary);
  await Git.createBranch(branchName);

  if (!startProgress) {
    return;
  }

  await JiraApi.changeIssueStatus(
    createdIssue.key ?? "",
    IssueStatus.InProgress,
  );

  console.log(
    `✅ Changed status of issue ${createdIssue.key} to ${
      IssueStatusById[IssueStatus.InProgress]
    }`,
  );
}
