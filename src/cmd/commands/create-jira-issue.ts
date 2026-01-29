import * as JiraApi from "../../jira/jira-api.ts";
import * as Git from "../../gitlab/git-branch.ts";
import * as Logger from "../../utils/logger.ts";
import { createBranchName } from "../../utils/branch-name-from-jira.ts";
import { parseCustomField } from "../../utils/parse-custom-field.ts";

type CreateJiraIssueCommand = {
  summary: string;
  options: {
    parent?: string;
    type?: string;
    key?: string;
    assign?: boolean;
    branch?: boolean;
    worktree?: string;
    start?: boolean | string;
    customField?: string;
    description?: string;
  };
};

export async function createJiraIssueCommand(args: CreateJiraIssueCommand) {
  const {
    parent,
    type,
    key,
    assign: assignToMe,
    branch: createBranch,
    worktree: worktreeBaseDir,
    start: startProgress,
    customField,
    description,
  } = args.options;

  if (createBranch && worktreeBaseDir) {
    Logger.error("Cannot use both --branch and --worktree options");
    Deno.exit(1);
  }

  const customFieldObject = parseCustomField(customField);

  const createdIssue = await JiraApi.createIssue(
    args.summary,
    type,
    parent,
    assignToMe,
    key,
    customFieldObject,
    description,
  );

  Logger.success(`Issue created: ${createdIssue.url}`);

  if (createBranch || worktreeBaseDir) {
    const branchName = createBranchName(createdIssue.key, args.summary);

    if (worktreeBaseDir) {
      await Git.createWorktree(branchName, worktreeBaseDir);
    } else {
      await Git.createBranch(branchName);
    }
  }

  if (!startProgress) {
    return;
  }

  const statusName = typeof startProgress === "string"
    ? startProgress
    : "In Progress";
  await JiraApi.changeIssueStatus(
    createdIssue.key ?? "",
    statusName,
  );

  Logger.success(
    `Changed status of issue ${createdIssue.key} to ${statusName}`,
  );
}
