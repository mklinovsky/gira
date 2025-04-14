import * as JiraApi from "../../jira/jira-api.ts";
import * as Git from "../../gitlab/git-branch.ts";
import { createBranchName } from "../../utils/branch-name-from-jira.ts";

type CreateJiraIssueCommand = {
  summary: string;
  options: {
    parent?: string;
    type?: string;
    assign?: boolean;
    branch?: boolean;
  };
};

export async function createJiraIssueCommand(args: CreateJiraIssueCommand) {
  const {
    parent,
    type,
    assign: assignToMe,
    branch: createBranch,
  } = args.options;
  const createdIssue = await JiraApi.createIssue(
    args.summary,
    type,
    parent,
    assignToMe,
  );

  if (!createBranch) {
    return;
  }

  const branchName = createBranchName(createdIssue, args.summary);
  await Git.createBranch(branchName);
}
