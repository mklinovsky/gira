import { assertEquals } from "@std/assert";
import {
  jiraKeyFromBranchName,
  jiraSummaryFromBranchName,
} from "./jira-from-branch-name.ts";

Deno.test("jira-from-branch-name", () => {
  assertEquals(
    jiraKeyFromBranchName("JIRA-123-this-is-a-test-summary"),
    "JIRA-123",
  );

  assertEquals(
    jiraKeyFromBranchName("JIRA-123-this-is-a-test-summary-with-123-numbers"),
    "JIRA-123",
  );

  assertEquals(jiraKeyFromBranchName("JIRA-123"), "JIRA-123");
  assertEquals(jiraKeyFromBranchName("this-is-a-test-summary"), undefined);
  assertEquals(jiraKeyFromBranchName("test-summary-JIRA-123"), undefined);
});

Deno.test("jira summary from branch name", () => {
  assertEquals(
    jiraSummaryFromBranchName("JIRA-123-this-is-a-test-summary"),
    "This is a test summary",
  );

  assertEquals(
    jiraSummaryFromBranchName(
      "JIRA-123-this-is-a-test-summary-with-123-numbers",
    ),
    "This is a test summary with 123 numbers",
  );

  assertEquals(jiraSummaryFromBranchName("JIRA-123"), undefined);
  assertEquals(jiraSummaryFromBranchName("this-is-a-test-summary"), undefined);
  assertEquals(jiraSummaryFromBranchName("test-summary-JIRA-123"), undefined);
});
