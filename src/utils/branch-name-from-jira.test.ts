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

  assertEquals(
    createBranchName("JIRA-123", "Summary with-dash"),
    "JIRA-123-summary-with-dash",
  );

  assertEquals(
    createBranchName("JIRA-123", "Summary with_underscore"),
    "JIRA-123-summary-with-underscore",
  );
});
