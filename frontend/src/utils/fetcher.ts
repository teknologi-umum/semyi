import { BASE_URL } from "@/constants";
import type { Response } from "@/types/Response";
import * as Sentry from "@sentry/solid";

export async function fetchAllStaticSnapshots(
  urls: string[],
  interval: "raw" | "hourly" | "daily" = "hourly",
  signal?: AbortSignal,
) {
  return Sentry.startSpan({
    name: "fetchAllStaticSnapshots",
    op: "function",
    attributes: {
      "semyi.snapshot.monitor_ids": urls,
      "semyi.snapshot.interval": interval,
    },
  }, async (span) => {
    try {
      Sentry.addBreadcrumb({
        category: "api",
        message: "Fetching all static snapshots",
        level: "info",
        data: {
        urls,
        interval,
      },
    });

    const response: Response[] = await Promise.all(
      urls.map((u) => fetch(`${BASE_URL}/api/static?id=${u}&interval=${interval}`, { signal }).then((r) => r.json())),
    );

    return response;
  } catch (err) {
    span.setStatus({
      code: 2,
      message: "internal_error",
    });
    // eslint-disable-next-line no-console
    console.error(err);
    Sentry.captureException(err);
    throw err;
  }
});
}

export async function fetchSingleStaticSnapshot(
  id: string,
  interval: "raw" | "hourly" | "daily" = "hourly",
  signal?: AbortSignal,
) {
  return Sentry.startSpan({
    name: "fetchSingleStaticSnapshot",
    op: "function",
    attributes: {
      "semyi.snapshot.monitor_id": id,
      "semyi.snapshot.interval": interval,
    },
  }, async (span) => {
    try {
      Sentry.addBreadcrumb({
        category: "api",
        message: "Fetching single static snapshot",
        level: "info",
        data: {
        id,
        interval,
      },
    });

    const response: Response = await fetch(`${BASE_URL}/api/static?id=${id}&interval=${interval}`, { signal }).then(
      (r) => r.json(),
    );

    return response;
  } catch (err) {
    span.setStatus({
      code: 2,
      message: "internal_error",
    });
    // eslint-disable-next-line no-console
    console.error(err);
    Sentry.captureException(err);
    throw err;
  }
});
}
