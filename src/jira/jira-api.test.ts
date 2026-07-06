import { assertEquals, assertRejects } from "@std/assert";
import { join } from "@std/path";
import { createIssue } from "./jira-api.ts";

type FetchCall = {
  init?: RequestInit;
  url: string;
};

Deno.test("createIssue uses configured subtask type for standard parent", async () => {
  const calls: FetchCall[] = [];

  await withJiraConfig(
    {
      defaults: {
        jira: {
          url: "https://example.atlassian.net",
          apiToken: "token",
          userEmail: "me@example.com",
          userId: "user-1",
          projectKey: "APP",
          issueType: "Task",
          subtaskIssueType: "Sub-task",
        },
      },
      projects: [],
    },
    async () => {
      await withFetchMock(async (input, init) => {
        const url = getUrl(input);
        calls.push({ url, init });

        if (url.endsWith("/rest/api/3/issue/APP-1")) {
          return jsonResponse({
            fields: {
              issuetype: {
                subtask: false,
                hierarchyLevel: 0,
              },
            },
          });
        }

        if (url.endsWith("/rest/api/3/issue")) {
          return jsonResponse({ key: "APP-2" }, 201);
        }

        throw new Error(`Unexpected request ${url}`);
      }, async () => {
        const result = await createIssue("Child issue", undefined, "APP-1");

        assertEquals(result.key, "APP-2");
      });
    },
  );

  assertEquals(calls.length, 2);
  assertEquals(
    calls[0].url,
    "https://example.atlassian.net/rest/api/3/issue/APP-1",
  );

  const createPayload = getJsonBody(calls[1].init);
  assertEquals(createPayload.fields.issuetype.name, "Sub-task");
  assertEquals(createPayload.fields.parent?.key, "APP-1");
});

Deno.test("createIssue uses configured standard issue type for epic parent", async () => {
  const calls: FetchCall[] = [];

  await withJiraConfig(
    {
      defaults: {
        jira: {
          url: "https://example.atlassian.net",
          apiToken: "token",
          userEmail: "me@example.com",
          userId: "user-1",
          projectKey: "APP",
          issueType: "Engineering Task",
          subtaskIssueType: "Sub-task",
        },
      },
      projects: [],
    },
    async () => {
      await withFetchMock(async (input, init) => {
        const url = getUrl(input);
        calls.push({ url, init });

        if (url.endsWith("/rest/api/3/issue/APP-1")) {
          return jsonResponse({
            fields: {
              issuetype: {
                subtask: false,
                hierarchyLevel: 1,
              },
            },
          });
        }

        if (url.endsWith("/rest/api/3/issue")) {
          return jsonResponse({ key: "APP-2" }, 201);
        }

        throw new Error(`Unexpected request ${url}`);
      }, async () => {
        await createIssue("Child issue", undefined, "APP-1");
      });
    },
  );

  const createPayload = getJsonBody(calls[1].init);
  assertEquals(createPayload.fields.issuetype.name, "Engineering Task");
});

Deno.test("createIssue fails when subtask type missing for standard parent", async () => {
  await withJiraConfig(
    {
      defaults: {
        jira: {
          url: "https://example.atlassian.net",
          apiToken: "token",
          userEmail: "me@example.com",
          userId: "user-1",
          projectKey: "APP",
          issueType: "Task",
        },
      },
      projects: [],
    },
    async () => {
      await withFetchMock(async (input) => {
        const url = getUrl(input);

        if (url.endsWith("/rest/api/3/issue/APP-1")) {
          return jsonResponse({
            fields: {
              issuetype: {
                subtask: false,
                hierarchyLevel: 0,
              },
            },
          });
        }

        throw new Error(`Unexpected request ${url}`);
      }, async () => {
        await assertRejects(
          () => createIssue("Child issue", undefined, "APP-1"),
          Error,
          "Jira setting subtaskIssueType is required when parent issue requires subtask creation.",
        );
      });
    },
  );
});

Deno.test("createIssue fails when parent is subtask", async () => {
  await withJiraConfig(
    {
      defaults: {
        jira: {
          url: "https://example.atlassian.net",
          apiToken: "token",
          userEmail: "me@example.com",
          userId: "user-1",
          projectKey: "APP",
          issueType: "Task",
          subtaskIssueType: "Sub-task",
        },
      },
      projects: [],
    },
    async () => {
      await withFetchMock(async (input) => {
        const url = getUrl(input);

        if (url.endsWith("/rest/api/3/issue/APP-1")) {
          return jsonResponse({
            fields: {
              issuetype: {
                subtask: true,
                hierarchyLevel: -1,
              },
            },
          });
        }

        throw new Error(`Unexpected request ${url}`);
      }, async () => {
        await assertRejects(
          () => createIssue("Child issue", undefined, "APP-1"),
          Error,
          "Cannot create child issue under subtask APP-1.",
        );
      });
    },
  );
});

async function withJiraConfig(
  config: Record<string, unknown>,
  callback: () => Promise<void>,
) {
  const previousHome = Deno.env.get("HOME");
  const homeDir = await Deno.makeTempDir();
  const configPath = join(homeDir, ".gira", "config.json");

  try {
    await Deno.mkdir(join(homeDir, ".gira"), { recursive: true });
    await Deno.writeTextFile(configPath, JSON.stringify(config));
    Deno.env.set("HOME", homeDir);
    await callback();
  } finally {
    if (previousHome === undefined) {
      Deno.env.delete("HOME");
    } else {
      Deno.env.set("HOME", previousHome);
    }

    await Deno.remove(homeDir, { recursive: true });
  }
}

async function withFetchMock(
  mockFetch: typeof fetch,
  callback: () => Promise<void>,
) {
  const originalFetch = globalThis.fetch;
  globalThis.fetch = mockFetch;

  try {
    await callback();
  } finally {
    globalThis.fetch = originalFetch;
  }
}

function getUrl(input: RequestInfo | URL): string {
  if (input instanceof URL) {
    return input.toString();
  }

  if (typeof input === "string") {
    return input;
  }

  return input.url;
}

function getJsonBody(init?: RequestInit): {
  fields: {
    issuetype: {
      name: string;
    };
    parent?: {
      key: string;
    };
  };
} {
  if (typeof init?.body !== "string") {
    throw new Error("Expected JSON string request body.");
  }

  return JSON.parse(init.body) as {
    fields: {
      issuetype: {
        name: string;
      };
      parent?: {
        key: string;
      };
    };
  };
}

function jsonResponse(body: unknown, status = 200): Response {
  return new Response(JSON.stringify(body), {
    status,
    headers: {
      "Content-Type": "application/json",
    },
  });
}
