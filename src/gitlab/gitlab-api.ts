import { requireEnv } from "../utils/utils.ts";

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

  const response = await fetch(`${API_URL}/merge_requests`, {
    method: "POST",
    headers: getHeaders(),
    body: JSON.stringify(payload),
  });

  if (!response.ok) {
    throw new Error(`Failed to create merge request: ${response.statusText}`);
  }

  const data = await response.json();
  return data.web_url;
}

function getHeaders() {
  return {
    "PRIVATE-TOKEN": API_TOKEN,
    "Content-Type": "application/json",
  };
}
