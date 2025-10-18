import { Command } from "@cliffy/command";
import { createJiraIssueCommand } from "./commands/create-jira-issue.ts";
import { updateJiraIssueCommand } from "./commands/update-jira-issue.ts";
import { changeIssueStatusCommand } from "./commands/change-issue-status-command.ts";
import { createMergeRequestCommand } from "./commands/create-merge-request-command.ts";
import { getVersion } from "../utils/get-version.ts";
import { getMrCommand } from "./commands/get-mr-command.ts";
import { getJiraCommand } from "./commands/get-jira-command.ts";
import { getJiraFilesCommand } from "./commands/get-jira-files-command.ts";
import { mergeCommand } from "./commands/merge-command.ts";

export async function createCmd() {
  const program = new Command()
    .name("gira")
    .version(getVersion())
    .description("CLI tool for managing Gitlab and JIRA");

  program
    .command("create", "Create a new JIRA issue")
    .arguments("<summary:string>")
    .option("-p, --parent <parent:string>", "Parent issue key")
    .option("-t, --type <type:string>", "Issue type", {
      default: "Task",
    })
    .option(
      "-k, --key <key:string>",
      "Project key (overrides JIRA_PROJECT_KEY env variable)",
    )
    .option("-b --branch", "Create git branch")
    .option("-a --assign", "Assign to me")
    .option("-s --start", "Start progress")
    .option(
      "--custom-field <customField:string>",
      "Custom field in key=value format",
    )
    .action(
      async (options, ...args) =>
        await createJiraIssueCommand({ options, summary: args[0] }),
    );

  program
    .command("update", "Update a JIRA issue")
    .arguments("<issueKey:string>")
    .option(
      "--custom-field <customField:string>",
      "Custom field in key=value format",
    )
    .action(
      async (options, ...args) =>
        await updateJiraIssueCommand({ options, issueKey: args[0] }),
    );

  program
    .command("status", "Change the status of a JIRA issue")
    .arguments("<status:string>")
    .option(
      "-i, --issue <issue:string>",
      "Issue key, if not provided, will parse key from current branch name",
    )
    .action(async (options, ...args) => {
      await changeIssueStatusCommand({ issue: options.issue, status: args[0] });
    });

  program
    .command("mr", "Create a merge request")
    .option("-t, --target <target:string>", "Target branch")
    .option(
      "-l, --labels <labels:string>",
      "Comma-separated labels for the merge request",
    )
    .option("-d --draft", "Create a draft merge request")
    .action(async (options) => {
      await createMergeRequestCommand({
        labels: options.labels,
        draft: options.draft,
        targetBranch: options.target,
      });
    });

  program
    .command("get-mr", "Get details of a merge request")
    .arguments("<mergeRequestId:string>")
    .action(async (_options, ...args) => {
      await getMrCommand({ mergeRequestId: args[0] });
    });

  program
    .command("get-jira", "Get details of a Jira issue")
    .arguments("<issueKey:string>")
    .action(async (_options, ...args) => {
      await getJiraCommand({ issueKey: args[0] });
    });

  program
    .command("get-jira-files", "Download attachments from a Jira issue")
    .arguments("<issueKey:string>")
    .option("-o, --output <output:string>", "Output directory for attachments")
    .action(async (options, ...args) => {
      await getJiraFilesCommand({
        issueKey: args[0],
        outputDir: options.output,
      });
    });

  program
    .command("merge", "Merge the current merge request")
    .arguments("<mergeRequestId:string>")
    .option("-c --close", "Close the JIRA issue after merging")
    .option("-d --delete", "Delete the source branch after merging")
    .action(async (options, ...args) => {
      await mergeCommand({
        mergeRequestId: args[0],
        closeJira: options.close,
        deleteBranch: options.delete,
      });
    });

  await program.parse(Deno.args);
}
