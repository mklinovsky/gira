export async function postJson<Response = unknown>(
  url: string,
  payload: RequestInit,
  errorPrefix: string,
): Promise<Response> {
  try {
    const response = await fetch(url, payload);

    if (!response.ok) {
      throw new Error(`${response.status} ${response.statusText} ${url}`);
    }

    if (response.status === 204) {
      return {} as Response;
    }

    return await response.json();
  } catch (error) {
    const message = error instanceof Error ? error.message : error;
    throw new Error(`${errorPrefix}: ${message}`);
  }
}
