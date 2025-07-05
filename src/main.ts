import { createCmd } from "./cmd/cmd.ts";
import * as Logger from "./utils/logger.ts";

async function main() {
  try {
    await createCmd();
  } catch (error) {
    Logger.error(error);
  }
}

if (import.meta.main) {
  await main();
}
