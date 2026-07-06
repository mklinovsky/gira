import { postJson } from "../utils/post-json.ts";
import { resolveProjectConfig } from "../config/resolve-project-config.ts";
import type {
  CreateIssuePayload,
  JiraAttachment,
  JiraResponse,
} from "./jira.types.ts";

const ERROR_PREFIX = "JIRA API";

type JiraConfig = {
  apiToken: string;
  issueType?: string;
  projectKey: string;
  subtaskIssueType?: string;
  url: string;
  userEmail: string;
  userId: string;
};

type JiraIssue = {
  fields?: {
    attachment?: JiraAttachment[];
    issuetype?: {
      hierarchyLevel?: number;
      subtask?: boolean;
    };
  };
};

export async function createIssue(
  summary: string,
  issueType?: string,
  parentIssueKey?: string,
  assignToMe?: boolean,
  projectKey?: string,
  customField?: Record<string, unknown>,
  description?: string,
): Promise<{ key: string; url: string }> {
  const jiraConfig = await getJiraConfig({
    ...(projectKey ? { projectKey } : {}),
  });
  const resolvedIssueType = await resolveIssueType(
    issueType,
    parentIssueKey,
    jiraConfig,
  );

  const payload: CreateIssuePayload = {
    fields: {
      project: { key: jiraConfig.projectKey },
      summary,
      issuetype: { name: resolvedIssueType },
      ...(assignToMe ? { assignee: { id: jiraConfig.userId } } : {}),
      ...(parentIssueKey ? { parent: { key: parentIssueKey } } : {}),
      ...(customField ? customField : {}),
      ...(description
        ? {
          description: {
            type: "doc",
            version: 1,
            content: [
              {
                type: "paragraph",
                content: [
                  {
                    type: "text",
                    text: description,
                  },
                ],
              },
            ],
          },
        }
        : {}),
    },
  };

  const data = await postJson<JiraResponse<{ key: string }>>(
    `${jiraConfig.url}/rest/api/3/issue`,
    getRequestOptions(payload, jiraConfig),
    ERROR_PREFIX,
  );

  if (data.errorMessages?.length || data.errors) {
    throw new Error(`${data.errorMessages} ${data.errors}`);
  }
  const { key } = data;

  return {
    key,
    url: `${jiraConfig.url}/browse/${key}`,
  };
}

export async function updateIssue(
  issueKey: string,
  customField: Record<string, unknown>,
) {
  const jiraConfig = await getJiraConfig();
  const payload = {
    fields: {
      ...(customField ? customField : {}),
    },
  };

  const data = await postJson<JiraResponse<void>>(
    `${jiraConfig.url}/rest/api/3/issue/${issueKey}`,
    getRequestOptions(payload, jiraConfig, "PUT"),
    ERROR_PREFIX,
  );

  if (data.errorMessages?.length || data.errors) {
    throw new Error(`${data.errorMessages} ${data.errors}`);
  }
}

export async function changeIssueStatus(
  issueKey: string,
  statusName: string,
) {
  const jiraConfig = await getJiraConfig();
  const transitionId = await findTransitionIdByName(issueKey, statusName);
  const payload = { transition: { id: transitionId } };

  const url = `${jiraConfig.url}/rest/api/3/issue/${issueKey}/transitions`;
  const data = await postJson<JiraResponse>(
    url,
    getRequestOptions(payload, jiraConfig),
    ERROR_PREFIX,
  );

  if (data.errorMessages?.length || data.errors) {
    throw new Error(`${data.errorMessages} ${data.errors}`);
  }

  return data;
}

export async function getIssue(issueKey: string) {
  const jiraConfig = await getJiraConfig();

  try {
    return await fetchIssue(issueKey, jiraConfig);
  } catch (error) {
    throw new Error(`${ERROR_PREFIX}: ${error}`);
  }
}

export async function getIssueTransitions(issueKey: string): Promise<
  Array<{ id: string; name: string }>
> {
  const jiraConfig = await getJiraConfig();
  const url = `${jiraConfig.url}/rest/api/3/issue/${issueKey}/transitions`;

  try {
    const response = await fetch(url, {
      method: "GET",
      headers: getHeaders(jiraConfig),
    });

    if (!response.ok) {
      throw new Error(`Failed to fetch transitions: ${response.statusText}`);
    }

    const data = await response.json();

    if (data.errorMessages?.length || data.errors) {
      throw new Error(`${data.errorMessages} ${data.errors}`);
    }

    return data.transitions || [];
  } catch (error) {
    throw new Error(`${ERROR_PREFIX}: ${error}`);
  }
}

async function findTransitionIdByName(
  issueKey: string,
  statusName: string,
): Promise<string> {
  const transitions = await getIssueTransitions(issueKey);

  const transition = transitions.find(
    (t) => t.name.toLowerCase() === statusName.toLowerCase(),
  );

  if (!transition) {
    const availableStatuses = transitions.map((t) => t.name).join(", ");
    throw new Error(
      `Status "${statusName}" not found. Available statuses: ${availableStatuses}`,
    );
  }

  return transition.id;
}

export async function getIssueAttachments(issueKey: string) {
  const issue = await getIssue(issueKey);
  const attachments = issue?.fields?.attachment || [];

  return attachments.map((att: JiraAttachment) => ({
    id: att.id,
    filename: att.filename,
    size: att.size,
    mimeType: att.mimeType,
    content: att.content,
    created: att.created,
  }));
}

export async function downloadAttachment(
  contentUrl: string,
  outputPath: string,
) {
  const jiraConfig = await getJiraConfig();

  try {
    const response = await fetch(contentUrl, {
      method: "GET",
      headers: getHeaders(jiraConfig),
    });

    if (!response.ok) {
      throw new Error(`Failed to download attachment: ${response.statusText}`);
    }

    const arrayBuffer = await response.arrayBuffer();
    await Deno.writeFile(outputPath, new Uint8Array(arrayBuffer));

    return true;
  } catch (error) {
    throw new Error(`${ERROR_PREFIX}: ${error}`);
  }
}

async function getJiraConfig(
  overrides?: {
    projectKey?: string;
  },
): Promise<JiraConfig> {
  const { jira } = await resolveProjectConfig({
    jira: overrides,
  });

  if (!jira.enabled) {
    throw new Error("Jira is disabled for current folder.");
  }

  return {
    url: requireJiraValue(jira.url, "url"),
    apiToken: requireJiraValue(jira.apiToken, "apiToken"),
    userEmail: requireJiraValue(jira.userEmail, "userEmail"),
    userId: requireJiraValue(jira.userId, "userId"),
    projectKey: requireJiraValue(jira.projectKey, "projectKey"),
    issueType: jira.issueType,
    subtaskIssueType: jira.subtaskIssueType,
  };
}

async function resolveIssueType(
  issueType: string | undefined,
  parentIssueKey: string | undefined,
  jiraConfig: JiraConfig,
): Promise<string> {
  if (issueType) {
    return issueType;
  }

  if (!parentIssueKey) {
    return jiraConfig.issueType ?? "Task";
  }

  const parentIssue = await fetchIssue(parentIssueKey, jiraConfig);
  const parentIssueType = parentIssue.fields?.issuetype;

  if (!parentIssueType) {
    throw new Error(
      `Could not determine issue type for parent ${parentIssueKey}.`,
    );
  }

  if (parentIssueType.subtask) {
    throw new Error(
      `Cannot create child issue under subtask ${parentIssueKey}.`,
    );
  }

  if (
    parentIssueType.hierarchyLevel !== undefined &&
    parentIssueType.hierarchyLevel > 0
  ) {
    return jiraConfig.issueType ?? "Task";
  }

  if (!jiraConfig.subtaskIssueType) {
    throw new Error(
      "Jira setting subtaskIssueType is required when parent issue requires subtask creation.",
    );
  }

  return jiraConfig.subtaskIssueType;
}

async function fetchIssue(
  issueKey: string,
  jiraConfig: JiraConfig,
): Promise<JiraIssue> {
  const url = `${jiraConfig.url}/rest/api/3/issue/${issueKey}`;
  const response = await fetch(url, {
    method: "GET",
    headers: getHeaders(jiraConfig),
  });

  if (!response.ok) {
    throw new Error(`Failed to fetch issue: ${response.statusText}`);
  }

  const data = await response.json();

  if (hasJiraErrors(data)) {
    throw new Error(`${data.errorMessages} ${data.errors}`);
  }

  return data;
}

function requireJiraValue(
  value: string | undefined,
  key: "apiToken" | "projectKey" | "url" | "userEmail" | "userId",
): string {
  if (!value) {
    throw new Error(`Jira setting ${key} is required for current folder.`);
  }

  return value;
}

function getHeaders(
  jiraConfig: { apiToken: string; userEmail: string },
): HeadersInit {
  return {
    Authorization: `Basic ${
      btoa(`${jiraConfig.userEmail}:${jiraConfig.apiToken}`)
    }`,
    Accept: "application/json",
    "Content-Type": "application/json",
  };
}

function getRequestOptions<Payload>(
  payload: Payload,
  jiraConfig: { apiToken: string; userEmail: string },
  method: "POST" | "PUT" = "POST",
): RequestInit {
  return {
    method,
    headers: getHeaders(jiraConfig),
    body: JSON.stringify(payload),
  };
}

function hasJiraErrors(
  value: unknown,
): value is JiraResponse<Record<string, unknown>> {
  if (typeof value !== "object" || value === null) {
    return false;
  }

  const jiraValue = value as JiraResponse<Record<string, unknown>>;

  return Boolean(jiraValue.errorMessages?.length || jiraValue.errors);
}
