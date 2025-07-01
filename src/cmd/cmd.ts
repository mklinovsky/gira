import { Command } from "@cliffy/command";
import { createJiraIssueCommand } from "./commands/create-jira-issue.ts";
import { changeIssueStatusCommand } from "./commands/change-issue-status-command.ts";
import { createMergeRequestCommand } from "./commands/create-merge-request-command.ts";

export async function createCmd() {
  await new Command()
    .name("gira")
    .version("0.0.1")
    .description("CLI tool for managing Gitlab and JIRA")
    .command("create", "Create a new JIRA issue")
    .arguments("<summary:string>")
    .option("-p, --parent <parent:string>", "Parent issue key")
    .option("-t, --type <type:string>", "Issue type", {
      default: "Task",
    })
    .option("-b --branch", "Create git branch")
    .option("-a --assign", "Assign to me")
    .option("-s --start", "Start progress")
    .action((options, ...args) =>
      createJiraIssueCommand({ options, summary: args[0] })
    )
    .command("status", "Change the status of a JIRA issue")
    .arguments("<status:string>")
    .option(
      "-i, --issue <issue:string>",
      "Issue key, if not provided, will parse key from current branch name",
    )
    .action((options, ...args) => {
      changeIssueStatusCommand({ issue: options.issue, status: args[0] });
    })
    .command("mr", "Create a merge request")
    .option(
      "-l, --labels <labels:string>",
      "Comma-separated labels for the merge request",
    )
    .option("-d --draft", "Create a draft merge request")
    .action((options) => {
      createMergeRequestCommand({
        labels: options.labels,
        draft: options.draft,
      });
    })
    .parse(Deno.args);
}
