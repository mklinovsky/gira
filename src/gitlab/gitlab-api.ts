import { postJson } from "../utils/post-json.ts";
import { resolveProjectConfig } from "../config/resolve-project-config.ts";

export async function createMergeRequest(
  sourceBranch: string,
  targetBranch: string,
  title: string,
  labels?: string,
) {
  const gitlabConfig = await getCreateMergeRequestConfig();
  const payload = {
    source_branch: sourceBranch,
    target_branch: targetBranch,
    title,
    assignee_id: gitlabConfig.userId,
    remove_source_branch: true,
    ...(labels ? { labels } : {}),
  };

  const url =
    `${gitlabConfig.url}/api/v4/projects/${gitlabConfig.projectId}/merge_requests`;
  const data = await postJson<{ web_url: string }>(
    url,
    {
      method: "POST",
      headers: getHeaders(gitlabConfig.apiToken),
      body: JSON.stringify(payload),
    },
    "GitLab API",
  );

  return data.web_url;
}

export async function getMergeRequest(
  mrId: string,
): Promise<{ title: string }> {
  const gitlabConfig = await getGitlabConfig();
  const url =
    `${gitlabConfig.url}/api/v4/projects/${gitlabConfig.projectId}/merge_requests/${mrId}`;

  return postJson(
    url,
    {
      method: "GET",
      headers: getHeaders(gitlabConfig.apiToken),
    },
    "GitLab API",
  );
}

export async function mergeMergeRequest(
  mrId: string,
  deleteSourceBranch = false,
) {
  const gitlabConfig = await getGitlabConfig();
  const url =
    `${gitlabConfig.url}/api/v4/projects/${gitlabConfig.projectId}/merge_requests/${mrId}/merge`;
  const payload = {
    should_remove_source_branch: deleteSourceBranch,
  };

  return postJson(
    url,
    {
      method: "PUT",
      headers: getHeaders(gitlabConfig.apiToken),
      body: JSON.stringify(payload),
    },
    "GitLab API",
  );
}

async function getCreateMergeRequestConfig(): Promise<{
  apiToken: string;
  projectId: string;
  url: string;
  userId: string;
}> {
  const gitlabConfig = await getGitlabConfig();

  if (!gitlabConfig.userId) {
    throw new Error("GitLab setting userId is required for current folder.");
  }

  return {
    ...gitlabConfig,
    userId: gitlabConfig.userId,
  };
}

async function getGitlabConfig(): Promise<{
  apiToken: string;
  projectId: string;
  url: string;
  userId?: string;
}> {
  const { gitlab } = await resolveProjectConfig();

  return {
    url: requireGitlabValue(gitlab.url, "url"),
    apiToken: requireGitlabValue(gitlab.apiToken, "apiToken"),
    projectId: requireGitlabValue(gitlab.projectId, "projectId"),
    userId: gitlab.userId,
  };
}

function requireGitlabValue(
  value: string | undefined,
  key: "apiToken" | "projectId" | "url",
): string {
  if (!value) {
    throw new Error(`GitLab setting ${key} is required for current folder.`);
  }

  return value;
}

function getHeaders(apiToken: string) {
  return {
    "PRIVATE-TOKEN": apiToken,
    "Content-Type": "application/json",
  };
}
