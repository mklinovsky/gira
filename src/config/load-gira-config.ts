import { dirname, join } from "@std/path";
import type {
  GiraConfig,
  GiraConfigDefaults,
  GiraGitlabConfig,
  GiraJiraConfig,
  GiraProjectConfig,
} from "./gira-config.types.ts";

export const GIRA_CONFIG_TEMPLATE = {
  defaults: {
    gitlab: {
      url: "",
      apiToken: "",
      projectId: "",
      userId: "",
      targetBranch: "",
    },
    jira: {
      enabled: true,
      url: "",
      apiToken: "",
      userEmail: "",
      userId: "",
      projectKey: "",
      issueType: "",
      subtaskIssueType: "",
    },
  },
  projects: [],
} as const;

type LoadGiraConfigOptions = {
  homeDir?: string;
};

export async function loadGiraConfig(
  { homeDir = getHomeDirectory() }: LoadGiraConfigOptions = {},
): Promise<GiraConfig> {
  const configPath = getGiraConfigPath(homeDir);

  try {
    await Deno.stat(configPath);
  } catch (error) {
    if (!(error instanceof Deno.errors.NotFound)) {
      throw error;
    }

    throw new Error(getMissingGiraConfigMessage(configPath));
  }

  const fileContent = await Deno.readTextFile(configPath);

  try {
    return parseGiraConfig(JSON.parse(fileContent));
  } catch (error) {
    if (error instanceof SyntaxError) {
      throw new Error(`Invalid JSON in ${configPath}: ${error.message}`);
    }

    throw error;
  }
}

export function getHomeDirectory(): string {
  const homeDir = Deno.env.get("HOME") ?? Deno.env.get("USERPROFILE");

  if (!homeDir) {
    throw new Error("Could not determine home directory.");
  }

  return homeDir;
}

export function getGiraConfigPath(homeDir = getHomeDirectory()): string {
  return join(homeDir, ".gira", "config.json");
}

export function getMissingGiraConfigMessage(
  configPath = getGiraConfigPath(),
): string {
  return `Config file ${configPath} not found. Run gira init.`;
}

export async function initializeGiraConfig(
  { force = false, homeDir = getHomeDirectory() }: {
    force?: boolean;
    homeDir?: string;
  } = {},
): Promise<string> {
  const configPath = getGiraConfigPath(homeDir);

  try {
    await Deno.stat(configPath);

    if (!force) {
      throw new Error(
        `Config file ${configPath} already exists. Run gira init --force to overwrite it.`,
      );
    }
  } catch (error) {
    if (!(error instanceof Deno.errors.NotFound)) {
      throw error;
    }
  }

  await writeGiraConfig(configPath);

  return configPath;
}

async function writeGiraConfig(configPath: string): Promise<void> {
  await Deno.mkdir(dirname(configPath), { recursive: true });
  await Deno.writeTextFile(
    configPath,
    `${JSON.stringify(GIRA_CONFIG_TEMPLATE, null, 2)}\n`,
  );
}

function parseGiraConfig(rawConfig: unknown): GiraConfig {
  if (!isRecord(rawConfig)) {
    return { projects: [] };
  }

  const defaults = parseDefaults(rawConfig.defaults);
  const projects = Array.isArray(rawConfig.projects)
    ? rawConfig.projects.flatMap((project) => {
      const parsedProject = parseProjectConfig(project);

      if (!parsedProject) {
        return [];
      }

      return [parsedProject];
    })
    : [];

  return {
    ...(defaults ? { defaults } : {}),
    projects,
  };
}

function parseDefaults(rawDefaults: unknown): GiraConfigDefaults | undefined {
  if (!isRecord(rawDefaults)) {
    return undefined;
  }

  const gitlab = parseGitlabConfig(rawDefaults.gitlab);
  const jira = parseJiraConfig(rawDefaults.jira);

  if (!gitlab && !jira) {
    return undefined;
  }

  return {
    ...(gitlab ? { gitlab } : {}),
    ...(jira ? { jira } : {}),
  };
}

function parseProjectConfig(
  rawProject: unknown,
): GiraProjectConfig | undefined {
  if (!isRecord(rawProject)) {
    return undefined;
  }

  const path = normalizeString(rawProject.path);
  const worktreeBasePath = normalizeString(rawProject.worktreeBasePath);

  if (!path) {
    return undefined;
  }

  const gitlab = parseGitlabConfig(rawProject.gitlab);
  const jira = parseJiraConfig(rawProject.jira);

  return {
    path,
    ...(worktreeBasePath ? { worktreeBasePath } : {}),
    ...(gitlab ? { gitlab } : {}),
    ...(jira ? { jira } : {}),
  };
}

function parseGitlabConfig(rawGitlab: unknown): GiraGitlabConfig | undefined {
  if (!isRecord(rawGitlab)) {
    return undefined;
  }

  const gitlab: GiraGitlabConfig = {
    url: normalizeString(rawGitlab.url),
    apiToken: normalizeString(rawGitlab.apiToken),
    projectId: normalizeString(rawGitlab.projectId),
    userId: normalizeString(rawGitlab.userId),
    targetBranch: normalizeString(rawGitlab.targetBranch),
  };

  if (Object.values(gitlab).every((value) => value === undefined)) {
    return undefined;
  }

  return gitlab;
}

function parseJiraConfig(rawJira: unknown): GiraJiraConfig | undefined {
  if (!isRecord(rawJira)) {
    return undefined;
  }

  const jira: GiraJiraConfig = {
    ...(typeof rawJira.enabled === "boolean"
      ? { enabled: rawJira.enabled }
      : {}),
    url: normalizeString(rawJira.url),
    apiToken: normalizeString(rawJira.apiToken),
    userEmail: normalizeString(rawJira.userEmail),
    userId: normalizeString(rawJira.userId),
    projectKey: normalizeString(rawJira.projectKey),
    issueType: normalizeString(rawJira.issueType),
    subtaskIssueType: normalizeString(rawJira.subtaskIssueType),
  };

  if (Object.keys(jira).length === 0) {
    return undefined;
  }

  return jira;
}

function normalizeString(value: unknown): string | undefined {
  if (typeof value !== "string") {
    return undefined;
  }

  const trimmedValue = value.trim();

  if (!trimmedValue) {
    return undefined;
  }

  return trimmedValue;
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null;
}
