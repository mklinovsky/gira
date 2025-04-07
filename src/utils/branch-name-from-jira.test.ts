import { createBranchName } from "./branch-name-from-jira.ts";
import { assertEquals } from "@std/assert";

Deno.test("branch-name-from-jira", () => {
  assertEquals(
    createBranchName("JIRA-123", "This is a test summary"),
    "JIRA-123-this-is-a-test-summary",
  );

  assertEquals(
    createBranchName(
      "JIRA-123",
      "This is a test summ$ry with speci@l chars $%^&*()",
    ),
    "JIRA-123-this-is-a-test-summry-with-specil-chars",
  );
});
