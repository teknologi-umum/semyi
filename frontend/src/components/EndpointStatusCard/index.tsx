import type { Snapshot } from "@/types/Snapshot";
import Status from "@/components/Status";
import styles from "./EndpointStatusCard.module.css";
import { createSignal, onMount } from "solid-js";
import { fromEvent, map, take } from "rxjs";
import { Link } from "solid-app-router";

interface EndpointStatusCardProps {
  name: string;
  url: string;
}

export default function EndpointStatusCard(props: EndpointStatusCardProps) {
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
    <div class={styles["endpoint-card"]}>
      <div class={styles["endpoint-card__header"]}>
        <Link
          class={styles["endpoint-card__title"]}
          href={"/by?name=" + encodeURIComponent(props.name)}
        >
          {props.name}
        </Link>
        <a class={styles["endpoint-card__url"]} href={props.url}>
          {props.url}
        </a>
      </div>
      <Status snapshots={snapshot()}></Status>
    </div>
  );
}
