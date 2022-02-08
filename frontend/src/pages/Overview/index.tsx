import styles from "./Overview.module.css";
import WebsiteCard from "@/components/WebsiteCard";
import FAKE_WEBSITES from "@/fake/fakeWebsites";
import FAKE_SNAPSHOTS from "@/fake/fakeSnapshots";
import { For } from "solid-js";
import DarkModeToggle from "@/components/DarkModeToggle";

export default function OverviewPage() {
  return (
    <div class={styles.overview}>
      <div class={styles.overview__display}>
        <h1 class={styles.overview__title}>Overview</h1>
        <DarkModeToggle />
      </div>

      <div class={styles.overview__websites}>
        <For each={FAKE_WEBSITES}>
          {({ name, url }) =>
            <WebsiteCard
              name={name}
              url={url}
              snapshots={FAKE_SNAPSHOTS[name]}
            />
          }
        </For>
      </div>
    </div>
  );
}
