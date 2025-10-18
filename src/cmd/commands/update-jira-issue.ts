import * as JiraApi from "../../jira/jira-api.ts";
import * as Logger from "../../utils/logger.ts";
import { parseCustomField } from "../../utils/parse-custom-field.ts";
import * as Git from "../../gitlab/git-branch.ts";
import { jiraKeyFromBranchName } from "../../utils/jira-from-branch-name.ts";

type UpdateJiraIssueCommand = {
  issue?: string;
  customField?: string;
};

export async function updateJiraIssueCommand({
  issue,
  customField,
}: UpdateJiraIssueCommand) {
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

  if (!customField) {
    Logger.error("No custom field provided.");
    return;
  }

  const customFieldObject = parseCustomField(customField);

  if (!customFieldObject) {
    Logger.error("Invalid custom field format.");
    return;
  }

  await JiraApi.updateIssue(
    issueKey,
    customFieldObject,
  );

  Logger.success(`Issue ${issueKey} updated.`);
}
