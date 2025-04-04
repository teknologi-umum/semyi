import DarkModeToggle from "@/components/DarkModeToggle";
import EndpointStatusCard from "@/components/EndpointStatusCard";
import Notice from "@/components/Notice";
import { BASE_URL } from "@/constants";
import type { Response, Snapshot } from "@/types";
import { fetchAllStaticSnapshots } from "@/utils/fetcher";
import config from "@config";
import { fromEvent, map } from "rxjs";
import { For, Match, Switch, createResource, onCleanup, onMount } from "solid-js";
import styles from "./styles.module.css";

export default function OverviewPage() {
  const abortController = new AbortController();
  const [staticSnapshot, { refetch }] = createResource(() =>
    fetchAllStaticSnapshots(
      config.monitors.map((c) => c.unique_id),
      "raw",
      abortController.signal,
    ),
  );
  const source = new EventSource(`${BASE_URL}/api/overview`);
  const snapshotStream$ = fromEvent<MessageEvent<string>>(source, "message").pipe(
    map((event) => JSON.parse(event.data) as Snapshot),
  );

  onMount(() => {
    document.title = "Overview | Semyi";
  });

  onCleanup(() => {
    if (source != null) {
      source.close();
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
