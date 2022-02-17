import { fromEvent, map } from "rxjs";
import { createResource, For, Match, Switch } from "solid-js";
import DarkModeToggle from "@/components/DarkModeToggle";
import EndpointStatusCard from "@/components/EndpointStatusCard";
import { fetchAllStaticSnapshots } from "@/utils/fetcher";
import type { Response } from "@/types";
import config from "@config";
import styles from "./styles.module.css";

export default function OverviewPage() {
  const [staticSnapshot] = createResource(() =>
    fetchAllStaticSnapshots(config.map((c) => c.url))
  );

  return (
    <div class={styles.overview}>
      <div class={styles.overview__header}>
        <h1 class={styles.overview__title}>Overview</h1>
        <DarkModeToggle />
      </div>

      <Switch fallback={<div>Loading...</div>}>
        <Match when={!staticSnapshot.loading}>
          <div class={styles.overview__endpoints}>
            <For each={staticSnapshot()}>
              {(snapshot) => {
                const source = new EventSource(
                  import.meta.env.VITE_BASE_URL +
                    "/api/by?url=" +
                    snapshot[0].url
                );
                const snapshotStream$ = fromEvent<MessageEvent<string>>(
                  source,
                  "message"
                ).pipe(map((event) => JSON.parse(event.data) as Response));

                return (
                  <EndpointStatusCard
                    name={snapshot[0].name}
                    url={snapshot[0].url}
                    staticSnapshot={snapshot}
                    snapshotStream$={snapshotStream$}
                  />
                );
              }}
            </For>
          </div>
        </Match>
        <Match when={staticSnapshot.error !== undefined}>
          <h1>Error while fetching</h1>
        </Match>
      </Switch>
    </div>
  );
}
