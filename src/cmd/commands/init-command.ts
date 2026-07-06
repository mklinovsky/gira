import { initializeGiraConfig } from "../../config/load-gira-config.ts";
import * as Logger from "../../utils/logger.ts";

export async function initCommand({ force }: { force?: boolean }) {
  const configPath = await initializeGiraConfig({ force });

  Logger.success(`Created config at ${configPath}`);
}
