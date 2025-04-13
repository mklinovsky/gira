import { Command } from "@cliffy/command";
import { createJiraIssueCommand } from "./commands/create-jira-issue.ts";

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
    .action((options, ...args) =>
      createJiraIssueCommand({ options, summary: args[0] }),
    )
    .parse(Deno.args);
}
