import { assertEquals, assertThrows } from "@std/assert";
import { parseCustomField } from "./parse-custom-field.ts";

Deno.test("parseCustomField should return undefined for undefined input", () => {
  assertEquals(parseCustomField(undefined), undefined);
});

Deno.test("parseCustomField should parse a simple string value", () => {
  const result = parseCustomField("key=value");
  assertEquals(result, { key: "value" });
});

Deno.test("parseCustomField should parse a JSON object value", () => {
  const result = parseCustomField('key={"id":"123"}');
  assertEquals(result, { key: { id: "123" } });
});

Deno.test("parseCustomField should handle equals signs in the value", () => {
  const result = parseCustomField("key=value1=value2");
  assertEquals(result, { key: "value1=value2" });
});

Deno.test("parseCustomField should throw an error for invalid format (key only)", () => {
  assertThrows(
    () => parseCustomField("keyonly"),
    Error,
    "Invalid custom field format. Expected key=value.",
  );
});

Deno.test("parseCustomField should return undefined for empty string input", () => {
  assertEquals(parseCustomField(""), undefined);
});
