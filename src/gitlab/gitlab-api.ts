import { requireEnv } from "../utils/utils.ts";
import { postJson } from "../utils/post-json.ts";

const API_TOKEN = requireEnv("GITLAB_API_TOKEN");
const URL = requireEnv("GITLAB_URL");
const PROJECT_ID = requireEnv("GITLAB_PROJECT_ID");
const USER_ID = requireEnv("GITLAB_USER_ID");

const API_URL = `${URL}/api/v4/projects/${PROJECT_ID}`;

export async function createMergeRequest(
  sourceBranch: string,
  targetBranch: string,
  title: string,
  labels?: string,
) {
  const payload = {
    source_branch: sourceBranch,
    target_branch: targetBranch,
    title,
    assignee_id: USER_ID,
    ...(labels ? { labels } : {}),
  };

  const url = `${API_URL}/merge_requests`;
  const data = await postJson<{ web_url: string }>(
    url,
    {
      method: "POST",
      headers: getHeaders(),
      body: JSON.stringify(payload),
    },
    "GitLab API",
  );

  return data.web_url;
}

export function getMergeRequest(mrId: string): Promise<{ title: string }> {
  const url = `${API_URL}/merge_requests/${mrId}`;

  return postJson(
    url,
    {
      method: "GET",
      headers: getHeaders(),
    },
    "GitLab API",
  );
}

export function mergeMergeRequest(mrId: string, deleteSourceBranch = false) {
  const url = `${API_URL}/merge_requests/${mrId}/merge`;
  const payload = {
    should_remove_source_branch: deleteSourceBranch,
  };

  return postJson(
    url,
    {
      method: "PUT",
      headers: getHeaders(),
      body: JSON.stringify(payload),
    },
    "GitLab API",
  );
}

function getHeaders() {
  return {
    "PRIVATE-TOKEN": API_TOKEN,
    "Content-Type": "application/json",
  };
}
