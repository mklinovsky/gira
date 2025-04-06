export function requireEnv(key: string): string {
  const value = Deno.env.get(key);
  if (!value) {
    throw new Error(`Environment variable ${key} is required`);
  }
  return value;
}
