import {
  getGiraConfigPath,
  getMissingGiraConfigMessage,
  initializeGiraConfig,
  loadGiraConfig,
} from "../src/config/load-gira-config.ts";
import type {
  GiraConfig,
  GiraGitlabConfig,
  GiraJiraConfig,
} from "../src/config/gira-config.types.ts";

type GitlabKey = keyof GiraGitlabConfig;
type JiraKey = Exclude<keyof GiraJiraConfig, "enabled">;

const GITLAB_ENV_MAPPINGS: Array<{ env: string; key: GitlabKey }> = [
  { env: "GITLAB_URL", key: "url" },
  { env: "GITLAB_API_TOKEN", key: "apiToken" },
  { env: "GITLAB_PROJECT_ID", key: "projectId" },
  { env: "GITLAB_USER_ID", key: "userId" },
];

const JIRA_ENV_MAPPINGS: Array<{ env: string; key: JiraKey }> = [
  { env: "JIRA_URL", key: "url" },
  { env: "JIRA_API_TOKEN", key: "apiToken" },
  { env: "JIRA_USER_EMAIL", key: "userEmail" },
  { env: "JIRA_USER_ID", key: "userId" },
  { env: "JIRA_PROJECT_KEY", key: "projectKey" },
];

async function main() {
  const configPath = getGiraConfigPath();
  let configCreated = false;

  try {
    await loadGiraConfig();
  } catch (error) {
    if (
      !(error instanceof Error) ||
      error.message !== getMissingGiraConfigMessage(configPath)
    ) {
      throw error;
    }

    await initializeGiraConfig();
    configCreated = true;
  }

  const config = await loadGiraConfig();
  const nextConfig = createMutableConfig(config);
  const migratedFields: string[] = [];
  const skippedFields: string[] = [];

  for (const mapping of GITLAB_ENV_MAPPINGS) {
    const envValue = getEnvValue(mapping.env);

    if (!envValue) {
      continue;
    }

    if (nextConfig.defaults.gitlab[mapping.key]) {
      skippedFields.push(`defaults.gitlab.${mapping.key}`);
      continue;
    }

    nextConfig.defaults.gitlab[mapping.key] = envValue;
    migratedFields.push(`defaults.gitlab.${mapping.key}`);
  }

  for (const mapping of JIRA_ENV_MAPPINGS) {
    const envValue = getEnvValue(mapping.env);

    if (!envValue) {
      continue;
    }

    if (nextConfig.defaults.jira[mapping.key]) {
      skippedFields.push(`defaults.jira.${mapping.key}`);
      continue;
    }

    nextConfig.defaults.jira[mapping.key] = envValue;
    migratedFields.push(`defaults.jira.${mapping.key}`);
  }

  await Deno.writeTextFile(
    configPath,
    `${JSON.stringify(nextConfig, null, 2)}\n`,
  );

  if (configCreated) {
    console.log(`Created config at ${configPath}`);
  }

  if (migratedFields.length === 0 && skippedFields.length === 0) {
    console.log("No matching env vars found. Nothing migrated.");
    return;
  }

  if (migratedFields.length > 0) {
    console.log(`Migrated: ${migratedFields.join(", ")}`);
  }

  if (skippedFields.length > 0) {
    console.log(`Skipped existing values: ${skippedFields.join(", ")}`);
  }
}

function createMutableConfig(config: GiraConfig): {
  defaults: {
    gitlab: GiraGitlabConfig;
    jira: GiraJiraConfig;
  };
  projects: GiraConfig["projects"];
} {
  return {
    defaults: {
      gitlab: { ...(config.defaults?.gitlab ?? {}) },
      jira: { ...(config.defaults?.jira ?? {}) },
    },
    projects: config.projects,
  };
}

function getEnvValue(key: string): string | undefined {
  const value = Deno.env.get(key)?.trim();

  if (!value) {
    return undefined;
  }

  return value;
}

if (import.meta.main) {
  await main();
}
