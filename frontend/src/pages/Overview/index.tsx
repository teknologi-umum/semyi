import { fromEvent, map } from "rxjs";
import { createResource, For, Match, onMount, Switch } from "solid-js";
import DarkModeToggle from "@/components/DarkModeToggle";
import EndpointStatusCard from "@/components/EndpointStatusCard";
import { fetchAllStaticSnapshots } from "@/utils/fetcher";
import Notice from "@/components/Notice";
import type { Response } from "@/types";
import config from "@config";
import styles from "./styles.module.css";
import { BASE_URL } from "@/constants";

export default function OverviewPage() {
  const [staticSnapshot] = createResource(() =>
    fetchAllStaticSnapshots(config.endpoints.map((c) => c.url))
  );

  onMount(() => {
    document.title = "Overview | Semyi";
  });

  return (
    <div class={styles.overview}>
      <div class={styles.overview__header}>
        <h1 class={styles.overview__title}>Overview</h1>
        <DarkModeToggle />
      </div>

      <Switch fallback={<Notice text="Loading..." />}>
        <Match when={!staticSnapshot.loading}>
          <div class={styles.overview__endpoints}>
            <For each={staticSnapshot()}>
              {(snapshot) => {
                const source = new EventSource(
                  BASE_URL + "/api/by?url=" + snapshot[0].url
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
          <Notice text="Error while fetching. Try checking the console." />
        </Match>
      </Switch>
    </div>
  );
}
