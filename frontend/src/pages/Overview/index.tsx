import DarkModeToggle from "@/components/DarkModeToggle";
import EndpointStatusCard from "@/components/EndpointStatusCard";
import { BASE_URL } from "@/constants";
import type { Snapshot } from "@/types";
import { fetchAllStaticSnapshots } from "@/utils/fetcher";
import * as Sentry from "@sentry/solid";
import { fromEvent, map } from "rxjs";
import { For, createResource, onCleanup, onMount } from "solid-js";
import styles from "./styles.module.css";

export default function OverviewPage() {
  const abortController = new AbortController();
  const [staticSnapshot, { refetch }] = createResource(() =>
    fetchAllStaticSnapshots(
      [],
      "raw",
      abortController.signal,
    ),
  );
  const source = new EventSource(`${BASE_URL}/api/overview`);
  const snapshotStream$ = fromEvent<MessageEvent<string>>(source, "message").pipe(
    map((event) => {
      return Sentry.startSpan(
        {
          name: "overview.stream_message",
          op: "function",
          attributes: {
            "semyi.page": "overview",
          },
        },
        (span) => {
          try {
            const data = JSON.parse(event.data) as Snapshot;
            return data;
          } catch (err) {
            span.setStatus({
              code: 2,
              message: "parse_error",
            });
            throw err;
          }
        },
      );
    }),
  );

  let fallbackTimeout: NodeJS.Timeout | null = null;

  onMount(() => {
    document.title = "Overview | Semyi";

    // Fallback mechanism in case the event source is not working
    fallbackTimeout = setTimeout(() => {
      refetch();
    }, 2 * 60 * 1000); // 2 minutes
  });

  onCleanup(() => {
    if (source != null) {
      source.close();
    }

    if (fallbackTimeout != null) {
      clearTimeout(fallbackTimeout);
    }
  });

  return (
    <div class={styles.overview}>
      <div class={styles.overview__header}>
        <h1 class={styles.overview__title}>Overview</h1>
        <DarkModeToggle />
      </div>

      <div class={styles.overview__endpoints}>
        <For each={staticSnapshot()}>
          {(snapshot) => {
            return (
              <EndpointStatusCard
                monitorId={snapshot.metadata.id}
                name={snapshot.metadata.name}
                url={snapshot.metadata.public_url ?? ""}
                staticSnapshot={snapshot.historical.reverse().slice(0, 100)}
                snapshotStream$={snapshotStream$}
              />
            );
          }}
        </For>
      </div>
    </div>
  );
}
