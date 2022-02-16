import styles from "./Overview.module.css";
import WebsiteCard from "@/components/WebsiteCard";
import FAKE_SNAPSHOTS from "@/fake/fakeSnapshots";
import { For, onMount } from "solid-js";
import DarkModeToggle from "@/components/DarkModeToggle";
import config from "@config";

export default function OverviewPage() {
  onMount(() => {
    console.log(config);
  });

  return (
    <div class={styles.overview}>
      <div class={styles.overview__display}>
        <h1 class={styles.overview__title}>Overview</h1>
        <DarkModeToggle />
      </div>

      <div class={styles.overview__websites}>
        <For each={config}>
          {({ name, url }) => (
            <WebsiteCard
              name={name}
              url={url}
              snapshots={FAKE_SNAPSHOTS[name]}
            />
          )}
        </For>
      </div>
    </div>
  );
}
