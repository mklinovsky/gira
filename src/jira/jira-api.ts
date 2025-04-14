import { requireEnv } from "../utils/utils.ts";
import {
  CreateIssuePayload,
  IssueStatus,
  IssueStatusById,
} from "./jira.types.ts";

const API_TOKEN = requireEnv("JIRA_API_TOKEN");
const USER_EMAIL = requireEnv("JIRA_USER_EMAIL");
const USER_ID = requireEnv("JIRA_USER_ID");
const BASE_URL = requireEnv("JIRA_URL");
const PROJECT_KEY = requireEnv("JIRA_PROJECT_KEY");

const API_URL = `${BASE_URL}/rest/api/3`;

export async function createIssue(
  summary: string,
  issueType = "Task",
  parentIssueKey?: string,
  assignToMe?: boolean,
): Promise<string> {
  const payload: CreateIssuePayload = {
    fields: {
      project: { key: PROJECT_KEY },
      summary,
      issuetype: { name: issueType },
      ...(assignToMe ? { assignee: { id: USER_ID } } : {}),
      ...(parentIssueKey ? { parent: { key: parentIssueKey } } : {}),
    },
  };

  const { key } = await postJson<CreateIssuePayload, { key: string }>(
    `${API_URL}/issue`,
    payload,
  );

  console.log(`Created issue: ${BASE_URL}/browse/${key}`);
  return key;
}

export async function changeIssueStatus(
  issueKey: string,
  statusId: IssueStatus,
) {
  const payload = { transition: { id: statusId } };

  const url = `${API_URL}/issue/${issueKey}/transitions`;
  await postJson<typeof payload>(url, payload);

  console.log(
    `Changed status of issue ${issueKey} to ${IssueStatusById[statusId]}`,
  );
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

async function postJson<Payload, Response = unknown>(
  url: string,
  payload: Payload,
): Promise<Response> {
  try {
    const response = await fetch(url, getRequestOptions(payload));

    if (!response.ok) {
      throw new Error(`${response.status} ${response.statusText}`);
    }

    if (response.status === 204) {
      return {} as Response;
    }

    const data = await response.json();

    if (data.errorMessages?.length || data.errors) {
      throw new Error(`${data.errorMessages} ${data.errors}`);
    }

    return data;
  } catch (error) {
    console.error("Request failed:", error);
    throw error;
  }
}
