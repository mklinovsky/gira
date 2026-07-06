import { assertEquals } from "@std/assert";
import { createTitleFromBranchName } from "./create-title-from-branch-name.ts";

Deno.test("createTitleFromBranchName keeps conventional prefix in title", () => {
  assertEquals(
    createTitleFromBranchName("feat/add-user-avatar"),
    "feat: Add user avatar",
  );
  assertEquals(
    createTitleFromBranchName("feat/fix/mobile/login-flow"),
    "feat: Mobile login flow",
  );
  assertEquals(
    createTitleFromBranchName("fix/crash_on_start"),
    "fix: Crash on start",
  );
});

Deno.test("createTitleFromBranchName keeps non-prefix path segments", () => {
  assertEquals(
    createTitleFromBranchName("mobile/feat/login-flow"),
    "Mobile feat login flow",
  );
  assertEquals(createTitleFromBranchName("release/1.2.0"), "release: 1.2.0");
});
