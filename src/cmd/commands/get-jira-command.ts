import * as JiraApi from "../../jira/jira-api.ts";
import * as Logger from "../../utils/logger.ts";

export async function getJiraCommand({
  issueKey,
}: {
  issueKey: string;
}) {
  const result = await JiraApi.getIssue(issueKey);

  if (result) {
    console.log(JSON.stringify(result, null, 2));
  } else {
    Logger.error(`Failed to get Jira issue ${issueKey}.`);
    return;
  }
}
