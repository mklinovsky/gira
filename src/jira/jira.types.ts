export type CreateIssuePayload = {
  fields: {
    project: {
      key: string;
    };
    assignee?: {
      id: string;
    };
    summary: string;
    issuetype: {
      name: string;
    };
  };
};

// TODO: make this configurable
export const IssueStatus = {
  InProgress: "141",
  InReview: "101",
  Done: "201",
} as const;

export type IssueStatus = (typeof IssueStatus)[keyof typeof IssueStatus];

export const IssueStatusById = Object.fromEntries(
  Object.entries(IssueStatus).map(([key, value]) => [value, key]),
) as Record<IssueStatus, string>;

export type JiraResponse<T = unknown> = T & {
  errorMessages?: unknown[];
  errors?: unknown[];
};
