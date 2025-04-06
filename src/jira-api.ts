import {
  CreateIssuePayload,
  IssueStatus,
  IssueStatusById,
} from "./jira.types.ts";
import { requireEnv } from "./utils.ts";

export class JiraClient {
  private readonly apiToken: string;
  private readonly userEmail: string;
  private readonly userId: string;
  private readonly baseUrl: string;
  private readonly projectKey: string;
  private readonly apiUrl: string;

  constructor() {
    this.apiToken = requireEnv("JIRA_API_TOKEN");
    this.userEmail = requireEnv("JIRA_USER_EMAIL");
    this.userId = requireEnv("JIRA_USER_ID");
    this.baseUrl = requireEnv("JIRA_URL");
    this.projectKey = requireEnv("JIRA_PROJECT_KEY");

    this.apiUrl = `${this.baseUrl}/rest/api/3`;
  }

  async createIssue(summary: string, issueName = "Task"): Promise<string> {
    const payload: CreateIssuePayload = {
      fields: {
        project: { key: this.projectKey },
        assignee: { id: this.userId },
        summary,
        issuetype: { name: issueName },
      },
    };

    const { key } = await this.postJson<CreateIssuePayload, { key: string }>(
      `${this.apiUrl}/issue`,
      payload,
    );

    console.log(`Created issue: ${this.baseUrl}/${key}`);
    return key;
  }

  async changeIssueStatus(issueKey: string, statusId: IssueStatus) {
    const payload = { transition: { id: statusId } };

    const url = `${this.apiUrl}/issue/${issueKey}/transitions`;
    await this.postJson<typeof payload>(url, payload);

    console.log(
      `Changed status of issue ${issueKey} to ${IssueStatusById[statusId]}`,
    );
  }

  private async postJson<Payload, Response = unknown>(
    url: string,
    payload: Payload,
  ): Promise<Response> {
    try {
      const response = await fetch(url, this.getRequestOptions(payload));

      if (!response.ok) {
        throw new Error(`${response.status} ${response.statusText}`);
      }

      if (response.status === 204) {
        return {} as Response; // No content
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

  private getRequestOptions<Payload>(payload: Payload): RequestInit {
    return {
      method: "POST",
      headers: this.getHeaders(),
      body: JSON.stringify(payload),
    };
  }

  private getHeaders(): HeadersInit {
    return {
      Authorization: `Basic ${btoa(`${this.userEmail}:${this.apiToken}`)}`,
      Accept: "application/json",
      "Content-Type": "application/json",
    };
  }
}
