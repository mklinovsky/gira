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

export type JiraResponse<T = unknown> = T & {
  errorMessages?: unknown[];
  errors?: unknown[];
};
