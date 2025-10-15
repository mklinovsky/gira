export function parseCustomField(
  customField: string | undefined,
): Record<string, unknown> | undefined {
  if (!customField) {
    return undefined;
  }

  const [key, ...valueParts] = customField.split("=");
  if (!key || valueParts.length === 0) {
    throw new Error("Invalid custom field format. Expected key=value.");
  }
  const value = valueParts.join("=");

  try {
    return { [key]: JSON.parse(value) };
  } catch {
    return { [key]: value };
  }
}
