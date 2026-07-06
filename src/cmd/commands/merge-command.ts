import * as GitlabApi from "../../gitlab/gitlab-api.ts";
import * as JiraApi from "../../jira/jira-api.ts";
import { resolveProjectConfig } from "../../config/resolve-project-config.ts";
import * as Logger from "../../utils/logger.ts";
import { jiraKeyFromBranchName } from "../../utils/jira-from-branch-name.ts";

export async function mergeCommand({
  mergeRequestId,
  closeJira,
  deleteBranch,
}: {
  mergeRequestId: string;
  closeJira?: boolean;
  deleteBranch?: boolean;
}) {
  const mrDetails = await GitlabApi.getMergeRequest(mergeRequestId);

  if (!mrDetails) {
    Logger.error(`Failed to get merge request ${mergeRequestId}.`);
    return;
  }

  await GitlabApi.mergeMergeRequest(mergeRequestId, !!deleteBranch);
  Logger.success(`Merge request ${mergeRequestId} merged successfully.`);

  const jiraKey = jiraKeyFromBranchName(mrDetails.title);
  const { jira } = await resolveProjectConfig();

  if (jira.enabled && jiraKey && closeJira) {
    const statusName = "Done";
    await JiraApi.changeIssueStatus(jiraKey, statusName);
    Logger.success(`Changed status of issue ${jiraKey} to ${statusName}`);
  }
}
