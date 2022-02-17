import styles from "./Overview.module.css";
import EndpointStatusCard from "@/components/EndpointStatusCard";
import { createResource, For } from "solid-js";
import DarkModeToggle from "@/components/DarkModeToggle";
import config from "@config";
import { fetchAllStaticSnapshots } from "@/utils/fetchStaticSnapshots";

export default function OverviewPage() {
  const [staticSnapshot] = createResource(async () =>
    fetchAllStaticSnapshots(config.map((c) => c.url))
  );

  return (
    <div class={styles.overview}>
      <div class={styles.overview__header}>
        <h1 class={styles.overview__title}>Overview</h1>
        <DarkModeToggle />
      </div>

      <div class={styles.overview__endpoints}>
        <For each={staticSnapshot()}>
          {(snapshot) => (
            <EndpointStatusCard
              name={snapshot[0].name}
              url={snapshot[0].url}
              staticSnapshot={snapshot}
            />
          )}
        </For>
      </div>
    </div>
  );
}
