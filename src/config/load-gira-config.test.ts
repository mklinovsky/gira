import { assertEquals, assertRejects } from "@std/assert";
import { join } from "@std/path";
import {
  getGiraConfigPath,
  getMissingGiraConfigMessage,
  GIRA_CONFIG_TEMPLATE,
  initializeGiraConfig,
  loadGiraConfig,
} from "./load-gira-config.ts";

Deno.test("loadGiraConfig throws clear error when config missing", async () => {
  const homeDir = await Deno.makeTempDir();

  try {
    const configPath = getGiraConfigPath(homeDir);

    await assertRejects(
      () => loadGiraConfig({ homeDir }),
      Error,
      getMissingGiraConfigMessage(configPath),
    );
  } finally {
    await Deno.remove(homeDir, { recursive: true });
  }
});

Deno.test("initializeGiraConfig creates template config", async () => {
  const homeDir = await Deno.makeTempDir();

  try {
    const configPath = await initializeGiraConfig({ homeDir });
    const config = await loadGiraConfig({ homeDir });
    const fileContent = await Deno.readTextFile(configPath);

    assertEquals(
      fileContent,
      `${JSON.stringify(GIRA_CONFIG_TEMPLATE, null, 2)}\n`,
    );
    assertEquals(config.projects, []);
    assertEquals(config.defaults?.jira?.enabled, true);
    assertEquals(config.defaults?.gitlab?.url, undefined);
    assertEquals(config.defaults?.gitlab?.targetBranch, undefined);
    assertEquals(config.defaults?.jira?.issueType, undefined);
    assertEquals(config.defaults?.jira?.subtaskIssueType, undefined);
  } finally {
    await Deno.remove(homeDir, { recursive: true });
  }
});

Deno.test("loadGiraConfig keeps empty values unset", async () => {
  const homeDir = await Deno.makeTempDir();

  try {
    const configPath = getGiraConfigPath(homeDir);
    await Deno.mkdir(join(homeDir, ".gira"), { recursive: true });
    await Deno.writeTextFile(
      configPath,
      JSON.stringify({
        defaults: {
          gitlab: {
            url: "   ",
            targetBranch: "",
          },
          jira: {
            enabled: false,
            projectKey: "",
            issueType: "",
            subtaskIssueType: "   ",
          },
        },
        projects: [],
      }),
    );

    const config = await loadGiraConfig({ homeDir });

    assertEquals(config.defaults?.gitlab?.url, undefined);
    assertEquals(config.defaults?.gitlab?.targetBranch, undefined);
    assertEquals(config.defaults?.jira?.projectKey, undefined);
    assertEquals(config.defaults?.jira?.issueType, undefined);
    assertEquals(config.defaults?.jira?.subtaskIssueType, undefined);
    assertEquals(config.defaults?.jira?.enabled, false);
  } finally {
    await Deno.remove(homeDir, { recursive: true });
  }
});

Deno.test("loadGiraConfig throws clear error for invalid JSON", async () => {
  const homeDir = await Deno.makeTempDir();

  try {
    const configPath = getGiraConfigPath(homeDir);
    await Deno.mkdir(join(homeDir, ".gira"), { recursive: true });
    await Deno.writeTextFile(configPath, "{");

    await assertRejects(
      () => loadGiraConfig({ homeDir }),
      Error,
      `Invalid JSON in ${configPath}`,
    );
  } finally {
    await Deno.remove(homeDir, { recursive: true });
  }
});

Deno.test("initializeGiraConfig requires force to overwrite", async () => {
  const homeDir = await Deno.makeTempDir();

  try {
    const configPath = await initializeGiraConfig({ homeDir });
    await Deno.writeTextFile(configPath, '{"projects":[{"path":"/tmp/app"}]}');

    await assertRejects(
      () => initializeGiraConfig({ homeDir }),
      Error,
      `Config file ${configPath} already exists. Run gira init --force to overwrite it.`,
    );

    await initializeGiraConfig({ force: true, homeDir });

    const fileContent = await Deno.readTextFile(configPath);

    assertEquals(
      fileContent,
      `${JSON.stringify(GIRA_CONFIG_TEMPLATE, null, 2)}\n`,
    );
  } finally {
    await Deno.remove(homeDir, { recursive: true });
  }
});
