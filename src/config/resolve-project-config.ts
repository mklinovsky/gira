import { resolve } from "@std/path";
import type {
  GiraGitlabConfig,
  GiraJiraConfig,
  GiraProjectConfig,
  ResolvedGitlabConfig,
  ResolvedJiraConfig,
} from "./gira-config.types.ts";
import {
  getGiraConfigPath,
  getHomeDirectory,
  loadGiraConfig,
} from "./load-gira-config.ts";

type ResolveProjectConfigOptions = {
  cwd?: string;
  homeDir?: string;
  gitlab?: GiraGitlabConfig;
  jira?: GiraJiraConfig;
};

export async function resolveProjectConfig(
  {
    cwd = Deno.cwd(),
    homeDir = getHomeDirectory(),
    gitlab: gitlabOverrides,
    jira: jiraOverrides,
  }: ResolveProjectConfigOptions = {},
): Promise<{
  configPath: string;
  matchedProject?: GiraProjectConfig;
  gitlab: ResolvedGitlabConfig;
  jira: ResolvedJiraConfig;
}> {
  const config = await loadGiraConfig({ homeDir });
  const matchedProject = findMatchingProjectConfig(
    config.projects,
    cwd,
    homeDir,
  );

  return {
    configPath: getGiraConfigPath(homeDir),
    matchedProject,
    gitlab: mergeGitlabConfig(
      getGitlabEnvConfig(),
      config.defaults?.gitlab,
      matchedProject?.gitlab,
      gitlabOverrides,
    ),
    jira: mergeJiraConfig(
      getJiraEnvConfig(),
      config.defaults?.jira,
      matchedProject?.jira,
      jiraOverrides,
    ),
  };
}

export function findMatchingProjectConfig(
  projects: GiraProjectConfig[],
  cwd: string,
  homeDir = getHomeDirectory(),
): GiraProjectConfig | undefined {
  const normalizedCwd = resolve(cwd);

  return projects
    .map((project) => ({
      project,
      matchLength: getProjectMatchLength(project, normalizedCwd, homeDir),
    }))
    .filter(({ matchLength }) => matchLength > 0)
    .sort((left, right) => right.matchLength - left.matchLength)[0]
    ?.project;
}

function getProjectMatchLength(
  project: GiraProjectConfig,
  cwd: string,
  homeDir: string,
): number {
  const candidatePaths = [
    project.path,
    ...(project.worktreeBasePath ? [project.worktreeBasePath] : []),
  ].map((projectPath) => normalizeConfiguredProjectPath(projectPath, homeDir));

  const matchLengths = candidatePaths
    .filter((projectPath) => pathMatches(projectPath, cwd))
    .map((projectPath) => projectPath.length);

  return Math.max(0, ...matchLengths);
}

function getGitlabEnvConfig(): GiraGitlabConfig {
  return {
    url: getEnvValue("GITLAB_URL"),
    apiToken: getEnvValue("GITLAB_API_TOKEN"),
    projectId: getEnvValue("GITLAB_PROJECT_ID"),
    userId: getEnvValue("GITLAB_USER_ID"),
  };
}

function getJiraEnvConfig(): GiraJiraConfig {
  return {
    enabled: true,
    url: getEnvValue("JIRA_URL"),
    apiToken: getEnvValue("JIRA_API_TOKEN"),
    userEmail: getEnvValue("JIRA_USER_EMAIL"),
    userId: getEnvValue("JIRA_USER_ID"),
    projectKey: getEnvValue("JIRA_PROJECT_KEY"),
  };
}

function mergeGitlabConfig(
  ...sources: Array<GiraGitlabConfig | undefined>
): ResolvedGitlabConfig {
  const mergedConfig: ResolvedGitlabConfig = {};

  for (const source of sources) {
    if (!source) {
      continue;
    }

    if (source.url !== undefined) {
      mergedConfig.url = source.url;
    }

    if (source.apiToken !== undefined) {
      mergedConfig.apiToken = source.apiToken;
    }

    if (source.projectId !== undefined) {
      mergedConfig.projectId = source.projectId;
    }

    if (source.userId !== undefined) {
      mergedConfig.userId = source.userId;
    }

    if (source.targetBranch !== undefined) {
      mergedConfig.targetBranch = source.targetBranch;
    }
  }

  return mergedConfig;
}

function mergeJiraConfig(
  ...sources: Array<GiraJiraConfig | undefined>
): ResolvedJiraConfig {
  const mergedConfig: ResolvedJiraConfig = { enabled: true };

  for (const source of sources) {
    if (!source) {
      continue;
    }

    if (source.enabled !== undefined) {
      mergedConfig.enabled = source.enabled;
    }

    if (source.url !== undefined) {
      mergedConfig.url = source.url;
    }

    if (source.apiToken !== undefined) {
      mergedConfig.apiToken = source.apiToken;
    }

    if (source.userEmail !== undefined) {
      mergedConfig.userEmail = source.userEmail;
    }

    if (source.userId !== undefined) {
      mergedConfig.userId = source.userId;
    }

    if (source.projectKey !== undefined) {
      mergedConfig.projectKey = source.projectKey;
    }

    if (source.issueType !== undefined) {
      mergedConfig.issueType = source.issueType;
    }

    if (source.subtaskIssueType !== undefined) {
      mergedConfig.subtaskIssueType = source.subtaskIssueType;
    }
  }

  return mergedConfig;
}

function normalizeConfiguredProjectPath(
  projectPath: string,
  homeDir: string,
): string {
  return resolve(expandHomeDirectoryPath(projectPath, homeDir));
}

function expandHomeDirectoryPath(projectPath: string, homeDir: string): string {
  if (projectPath === "~") {
    return homeDir;
  }

  if (projectPath.startsWith("~/")) {
    return resolve(homeDir, projectPath.slice(2));
  }

  return projectPath;
}

function pathMatches(projectPath: string, cwd: string): boolean {
  if (projectPath === cwd) {
    return true;
  }

  const separator = projectPath.includes("\\") ? "\\" : "/";
  const normalizedProjectPath = projectPath.endsWith(separator)
    ? projectPath
    : `${projectPath}${separator}`;

  return cwd.startsWith(normalizedProjectPath);
}

function getEnvValue(key: string): string | undefined {
  const value = Deno.env.get(key)?.trim();

  if (!value) {
    return undefined;
  }

  return value;
}
