import { postJson } from "../utils/post-json.ts";
import { requireEnv } from "../utils/utils.ts";
import type {
  CreateIssuePayload,
  IssueStatus,
  JiraResponse,
} from "./jira.types.ts";

const API_TOKEN = requireEnv("JIRA_API_TOKEN");
const USER_EMAIL = requireEnv("JIRA_USER_EMAIL");
const USER_ID = requireEnv("JIRA_USER_ID");
const BASE_URL = requireEnv("JIRA_URL");
const PROJECT_KEY = requireEnv("JIRA_PROJECT_KEY");

const API_URL = `${BASE_URL}/rest/api/3`;
const ERROR_PREFIX = "JIRA API";

export async function createIssue(
  summary: string,
  issueType = "Task",
  parentIssueKey?: string,
  assignToMe?: boolean,
): Promise<{ key: string; url: string }> {
  const payload: CreateIssuePayload = {
    fields: {
      project: { key: PROJECT_KEY },
      summary,
      issuetype: { name: issueType },
      ...(assignToMe ? { assignee: { id: USER_ID } } : {}),
      ...(parentIssueKey ? { parent: { key: parentIssueKey } } : {}),
    },
  };

  const data = await postJson<JiraResponse<{ key: string }>>(
    `${API_URL}/issue`,
    getRequestOptions(payload),
    ERROR_PREFIX,
  );

  if (data.errorMessages?.length || data.errors) {
    throw new Error(`${data.errorMessages} ${data.errors}`);
  }
  const { key } = data;

  return {
    key,
    url: `${BASE_URL}/browse/${key}`,
  };
}

export async function changeIssueStatus(
  issueKey: string,
  statusId: IssueStatus,
) {
  const payload = { transition: { id: statusId } };

  const url = `${API_URL}/issue/${issueKey}/transitions`;
  const data = await postJson<JiraResponse>(
    url,
    getRequestOptions(payload),
    ERROR_PREFIX,
  );

  if (data.errorMessages?.length || data.errors) {
    throw new Error(`${data.errorMessages} ${data.errors}`);
  }

  return data;
}

function getHeaders(): HeadersInit {
  return {
    Authorization: `Basic ${btoa(`${USER_EMAIL}:${API_TOKEN}`)}`,
    Accept: "application/json",
    "Content-Type": "application/json",
  };
}

function getRequestOptions<Payload>(payload: Payload): RequestInit {
  return {
    method: "POST",
    headers: getHeaders(),
    body: JSON.stringify(payload),
  };
}
