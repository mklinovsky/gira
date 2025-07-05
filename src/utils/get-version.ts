import config from "../../deno.json" with { type: "json" };

export function getVersion() {
  return config.version;
}
