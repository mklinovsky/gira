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
  errorMessages?: string[];
  errors?: Record<string, string>;
};

export type JiraAttachment = {
  id: string;
  filename: string;
  size: number;
  mimeType: string;
  content: string;
  created: string;
};
