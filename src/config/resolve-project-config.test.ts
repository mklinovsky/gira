import { assertEquals } from "@std/assert";
import { join } from "@std/path";
import { resolveProjectConfig } from "./resolve-project-config.ts";

Deno.test("resolveProjectConfig uses longest matching project path", async () => {
  const homeDir = await Deno.makeTempDir();
  const cwd = join(homeDir, "Projects", "alpha", "app", "src");

  try {
    await Deno.mkdir(join(homeDir, ".gira"), { recursive: true });
    await Deno.writeTextFile(
      join(homeDir, ".gira", "config.json"),
      JSON.stringify({
        defaults: {
          gitlab: {
            url: "https://gitlab.example.com",
            targetBranch: "main",
          },
          jira: {
            issueType: "Task",
          },
        },
        projects: [
          {
            path: "~/Projects/alpha",
            gitlab: {
              projectId: "111",
            },
          },
          {
            path: "~/Projects/alpha/app",
            gitlab: {
              projectId: "222",
              targetBranch: "develop",
            },
            jira: {
              enabled: false,
              subtaskIssueType: "Sub-task",
            },
          },
        ],
      }),
    );

    const config = await resolveProjectConfig({ cwd, homeDir });

    assertEquals(config.gitlab.url, "https://gitlab.example.com");
    assertEquals(config.gitlab.projectId, "222");
    assertEquals(config.gitlab.targetBranch, "develop");
    assertEquals(config.jira.enabled, false);
    assertEquals(config.jira.issueType, "Task");
    assertEquals(config.jira.subtaskIssueType, "Sub-task");
    assertEquals(config.matchedProject?.path, "~/Projects/alpha/app");
  } finally {
    await Deno.remove(homeDir, { recursive: true });
  }
});

Deno.test("resolveProjectConfig keeps default gitlab targetBranch when project leaves it unset", async () => {
  const homeDir = await Deno.makeTempDir();
  const cwd = join(homeDir, "Projects", "app");

  try {
    await Deno.mkdir(join(homeDir, ".gira"), { recursive: true });
    await Deno.writeTextFile(
      join(homeDir, ".gira", "config.json"),
      JSON.stringify({
        defaults: {
          gitlab: {
            targetBranch: "main",
          },
        },
        projects: [
          {
            path: "~/Projects/app",
            gitlab: {
              projectId: "111",
            },
          },
        ],
      }),
    );

    const config = await resolveProjectConfig({ cwd, homeDir });

    assertEquals(config.gitlab.projectId, "111");
    assertEquals(config.gitlab.targetBranch, "main");
  } finally {
    await Deno.remove(homeDir, { recursive: true });
  }
});

Deno.test("resolveProjectConfig matches project by worktreeBasePath", async () => {
  const homeDir = await Deno.makeTempDir();
  const cwd = join(homeDir, "Projects", "foo-worktrees", "branch-a");

  try {
    await Deno.mkdir(join(homeDir, ".gira"), { recursive: true });
    await Deno.writeTextFile(
      join(homeDir, ".gira", "config.json"),
      JSON.stringify({
        projects: [
          {
            path: "~/Projects/foo",
            worktreeBasePath: "~/Projects/foo-worktrees",
            gitlab: {
              projectId: "111",
            },
            jira: {
              projectKey: "FOO",
            },
          },
        ],
      }),
    );

    const config = await resolveProjectConfig({ cwd, homeDir });

    assertEquals(config.matchedProject?.path, "~/Projects/foo");
    assertEquals(
      config.matchedProject?.worktreeBasePath,
      "~/Projects/foo-worktrees",
    );
    assertEquals(config.gitlab.projectId, "111");
    assertEquals(config.jira.projectKey, "FOO");
  } finally {
    await Deno.remove(homeDir, { recursive: true });
  }
});

Deno.test("resolveProjectConfig keeps env values when config values empty", async () => {
  const homeDir = await Deno.makeTempDir();
  const previousGitlabUrl = Deno.env.get("GITLAB_URL");
  const previousGitlabProjectId = Deno.env.get("GITLAB_PROJECT_ID");

  try {
    Deno.env.set("GITLAB_URL", "https://gitlab.example.com");
    Deno.env.set("GITLAB_PROJECT_ID", "999");
    await Deno.mkdir(join(homeDir, ".gira"), { recursive: true });
    await Deno.writeTextFile(
      join(homeDir, ".gira", "config.json"),
      JSON.stringify({
        defaults: {
          gitlab: {
            url: "",
          },
        },
        projects: [
          {
            path: "~/Projects/app",
            gitlab: {
              projectId: "",
            },
          },
        ],
      }),
    );

    const config = await resolveProjectConfig({
      cwd: join(homeDir, "Projects", "app"),
      homeDir,
    });

    assertEquals(config.gitlab.url, "https://gitlab.example.com");
    assertEquals(config.gitlab.projectId, "999");
  } finally {
    if (previousGitlabUrl === undefined) {
      Deno.env.delete("GITLAB_URL");
    } else {
      Deno.env.set("GITLAB_URL", previousGitlabUrl);
    }

    if (previousGitlabProjectId === undefined) {
      Deno.env.delete("GITLAB_PROJECT_ID");
    } else {
      Deno.env.set("GITLAB_PROJECT_ID", previousGitlabProjectId);
    }

    await Deno.remove(homeDir, { recursive: true });
  }
});
