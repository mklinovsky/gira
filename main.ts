import * as JiraClient from "./src/jira/jira-api.ts";
import { Command } from "@cliffy/command";

async function main() {
  await new Command()
    .name("jirant")
    .version("0.0.1")
    .description("Jira CLI tool, for creating and managing issues")
    .command("create", "Create a new issue")
    .arguments("<summary:string>")
    .option("-p, --parent <parent:string>", "Parent issue key")
    .option("-i, --issue <issue:string>", "Issue type", {
      default: "Task",
    })
    .action(
      async (options: { issue: string; parent: string }, summary: string) => {
        const issueKey = await JiraClient.createIssue(
          summary,
          options.issue,
          options.parent,
        );
        console.log(`Created issue: ${issueKey}`);
      },
    )
    .parse(Deno.args);
}

if (import.meta.main) {
  await main();
}
