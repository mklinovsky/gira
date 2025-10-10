import * as JiraApi from "../../jira/jira-api.ts";
import * as Logger from "../../utils/logger.ts";
import * as path from "@std/path";

export async function getJiraFilesCommand({
  issueKey,
  outputDir,
}: {
  issueKey: string;
  outputDir?: string;
}) {
  const attachments = await JiraApi.getIssueAttachments(issueKey);

  if (attachments.length === 0) {
    Logger.info(`No attachments found for issue ${issueKey}`);
    return;
  }

  const baseDir = outputDir || path.join(Deno.cwd(), issueKey);
  await Deno.mkdir(baseDir, { recursive: true });

  Logger.info(
    `Downloading ${attachments.length} attachment(s) for ${issueKey}...`,
  );

  const results = [];
  for (const attachment of attachments) {
    const outputPath = path.join(baseDir, attachment.filename);

    try {
      await JiraApi.downloadAttachment(attachment.content, outputPath);
      results.push({
        filename: attachment.filename,
        path: outputPath,
        size: attachment.size,
        mimeType: attachment.mimeType,
        success: true,
      });
      Logger.info(`✓ Downloaded: ${attachment.filename}`);
    } catch (error) {
      results.push({
        filename: attachment.filename,
        path: outputPath,
        error: String(error),
        success: false,
      });
      Logger.error(`✗ Failed to download: ${attachment.filename} - ${error}`);
    }
  }

  console.log(
    JSON.stringify(
      {
        issueKey,
        outputDirectory: baseDir,
        totalAttachments: attachments.length,
        results,
      },
      null,
      2,
    ),
  );
}
