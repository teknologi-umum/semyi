import { BASE_URL } from "@/constants";
import type { Response } from "@/types/Response";

export async function fetchAllStaticSnapshots(
  urls: string[],
  interval: "raw" | "hourly" | "daily" = "hourly",
  signal?: AbortSignal,
) {
  try {
    const response: Response[] = await Promise.all(
      urls.map((u) => fetch(`${BASE_URL}/api/static?id=${u}&interval=${interval}`, { signal }).then((r) => r.json())),
    );

    return response;
  } catch (err) {
    // eslint-disable-next-line no-console
    console.error(err);
    throw err;
  }
}

export async function fetchSingleStaticSnapshot(
  id: string,
  interval: "raw" | "hourly" | "daily" = "hourly",
  signal?: AbortSignal,
) {
  try {
    const response: Response = await fetch(`${BASE_URL}/api/static?id=${id}&interval=${interval}`, { signal }).then(
      (r) => r.json(),
    );

    return response;
  } catch (err) {
    // eslint-disable-next-line no-console
    console.error(err);
    throw err;
  }
}
