import type { Snapshot } from "@/types/Snapshot";
import Status from "@/components/Status";
import styles from "./WebsiteCard.module.css";
import { createSignal, onMount } from "solid-js";
import { fromEvent, map, take } from "rxjs";

interface WebsiteCardProps {
  name: string;
  url: string;
}

export default function WebsiteCard(props: WebsiteCardProps) {
  const [snapshot, setSnapshot] = createSignal<Snapshot[]>([]);

  onMount(async () => {
    const staticSnapshot: Snapshot[] = await fetch(
      import.meta.env.VITE_BASE_URL + "/api/static?url=" + props.url
    ).then((r) => r.json());
    setSnapshot(staticSnapshot);

    const source = new EventSource(
      import.meta.env.VITE_BASE_URL + "/api/by?url=" + props.url
    );
    fromEvent<MessageEvent<string>>(source, "message")
      .pipe(
        map((event) => JSON.parse(event.data) as Snapshot),
        // eslint-disable-next-line solid/reactivity
        map((s) => snapshot().concat(s)),
        take(100)
      )
      .subscribe((snapshots) => setSnapshot(snapshots));
  });

  return (
    <div class={styles["website-card"]}>
      <div class={styles["website-card__content"]}>
        <div class={styles["website-card__header"]}>
          <span class={styles["website-card__title"]}>{props.name}</span>
          <a class={styles["website-card__url"]} href={props.url}>
            {props.url}
          </a>
        </div>
        <Status snapshots={snapshot()}></Status>
      </div>
    </div>
  );
}
