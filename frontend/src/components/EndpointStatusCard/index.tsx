import { map, Observable, take } from "rxjs";
import { Link } from "solid-app-router";
import { createSignal, onMount } from "solid-js";
import Status from "@/components/Status";
import type { Response } from "@/types";
import styles from "./styles.module.css";

interface EndpointStatusCardProps {
  name: string;
  url: string;
  staticSnapshot: Response[] | undefined;
  snapshotStream$: Observable<Response>;
}

export default function EndpointStatusCard(props: EndpointStatusCardProps) {
  const [snapshot, setSnapshot] = createSignal<Response[]>(
    props.staticSnapshot || []
  );

  onMount(() => {
    props.snapshotStream$
      .pipe(
        map((newSnapshot) => snapshot().concat(newSnapshot)),
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
