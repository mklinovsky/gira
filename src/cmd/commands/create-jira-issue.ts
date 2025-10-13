import * as JiraApi from "../../jira/jira-api.ts";
import * as Git from "../../gitlab/git-branch.ts";
import * as Logger from "../../utils/logger.ts";
import { createBranchName } from "../../utils/branch-name-from-jira.ts";

type CreateJiraIssueCommand = {
  summary: string;
  options: {
    parent?: string;
    type?: string;
    key?: string;
    assign?: boolean;
    branch?: boolean;
    start?: boolean;
  };
};

export async function createJiraIssueCommand(args: CreateJiraIssueCommand) {
  const {
    parent,
    type,
    key,
    assign: assignToMe,
    branch: createBranch,
    start: startProgress,
  } = args.options;

  const createdIssue = await JiraApi.createIssue(
    args.summary,
    type,
    parent,
    assignToMe,
    key,
  );

  Logger.success(`Issue created: ${createdIssue.url}`);

  if (!createBranch) {
    return;
  }

  const branchName = createBranchName(createdIssue.key, args.summary);
  await Git.createBranch(branchName);

  if (!startProgress) {
    return;
  }

  const statusName = "In Progress";
  await JiraApi.changeIssueStatus(
    createdIssue.key ?? "",
    statusName,
  );

  Logger.success(
    `Changed status of issue ${createdIssue.key} to ${statusName}`,
  );
}
