import type { Snapshot } from "@/types/Snapshot";
import type { Response } from "@/types/Response";
import Status from "@/components/Status";
import styles from "./WebsiteCard.module.css";
import { createSignal, onMount } from "solid-js";
import { fromEvent, map } from "rxjs";

interface WebsiteCardProps {
  name: string;
  url: string;
}

export default function WebsiteCard(props: WebsiteCardProps) {
  const [snapshot, setSnapshot] = createSignal<Snapshot[]>([]);

  onMount(() => {
    const snapshot$ = fromEvent<MessageEvent<string>>(
      new EventSource("http://localhost:5000/api/by?url=" + props.url),
      "message"
    ).pipe(map((event) => JSON.parse(event.data) as Snapshot));
    snapshot$.subscribe((m) => setSnapshot((s) => [m, ...s]));
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
