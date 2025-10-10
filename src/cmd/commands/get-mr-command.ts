import * as GitlabApi from "../../gitlab/gitlab-api.ts";
import * as Logger from "../../utils/logger.ts";

export async function getMrCommand({
  mergeRequestId,
}: {
  mergeRequestId: string;
}) {
  const result = await GitlabApi.getMergeRequest(mergeRequestId);

  if (result) {
    console.log(JSON.stringify(result, null, 2));
  } else {
    Logger.error(`Failed to get merge request ${mergeRequestId}.`);
    return;
  }
}
