export type GiraGitlabConfig = {
  url?: string;
  apiToken?: string;
  projectId?: string;
  userId?: string;
  targetBranch?: string;
};

export type GiraJiraConfig = {
  enabled?: boolean;
  url?: string;
  apiToken?: string;
  userEmail?: string;
  userId?: string;
  projectKey?: string;
  issueType?: string;
  subtaskIssueType?: string;
};

export type GiraConfigDefaults = {
  gitlab?: GiraGitlabConfig;
  jira?: GiraJiraConfig;
};

export type GiraProjectConfig = {
  path: string;
  worktreeBasePath?: string;
  gitlab?: GiraGitlabConfig;
  jira?: GiraJiraConfig;
};

export type GiraConfig = {
  defaults?: GiraConfigDefaults;
  projects: GiraProjectConfig[];
};

export type ResolvedGitlabConfig = GiraGitlabConfig;

export type ResolvedJiraConfig = {
  enabled: boolean;
  url?: string;
  apiToken?: string;
  userEmail?: string;
  userId?: string;
  projectKey?: string;
  issueType?: string;
  subtaskIssueType?: string;
};
