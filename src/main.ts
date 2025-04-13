import { createCmd } from "./cmd/cmd.ts";

async function main() {
  await createCmd();
}

if (import.meta.main) {
  await main();
}
