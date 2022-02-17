import styles from "./Overview.module.css";
import WebsiteCard from "@/components/WebsiteCard";
import { For } from "solid-js";
import DarkModeToggle from "@/components/DarkModeToggle";
import config from "@config";

export default function OverviewPage() {
  return (
    <div class={styles.overview}>
      <div class={styles.overview__header}>
        <h1 class={styles.overview__title}>Overview</h1>
        <DarkModeToggle />
      </div>

      <div class={styles.overview__websites}>
        <For each={config}>
          {({ name, url }) => <WebsiteCard name={name} url={url} />}
        </For>
      </div>
    </div>
  );
}
