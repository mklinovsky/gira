import * as JiraApi from "../../jira/jira-api.ts";
import * as Logger from "../../utils/logger.ts";
import { parseCustomField } from "../../utils/parse-custom-field.ts";

type UpdateJiraIssueCommand = {
  issueKey: string;
  options: {
    customField?: string;
  };
};

export async function updateJiraIssueCommand(args: UpdateJiraIssueCommand) {
  const { customField } = args.options;

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
    args.issueKey,
    customFieldObject,
  );

  Logger.success(`Issue ${args.issueKey} updated.`);
}
