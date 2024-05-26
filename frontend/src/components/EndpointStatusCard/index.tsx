import { map, Observable, take } from "rxjs";
import { createSignal, onMount } from "solid-js";
import Status from "@/components/Status";
import type { Response } from "@/types";
import styles from "./styles.module.css";
import {A} from "@solidjs/router";

interface EndpointStatusCardProps {
  name: string;
  url: string;
  staticSnapshot: Response[] | undefined;
  snapshotStream$: Observable<Response>;
}

export default function EndpointStatusCard(props: EndpointStatusCardProps) {
  const [snapshot, setSnapshot] = createSignal<Response[]>(
    // this doesn't need to be reactive
    // eslint-disable-next-line solid/reactivity
    props.staticSnapshot || []
  );

  onMount(() => {
    props.snapshotStream$
      .pipe(
        // this doesn't need to be reactive
        // eslint-disable-next-line solid/reactivity
        map((newSnapshot) => snapshot().concat(newSnapshot)),
        take(100)
      )
      .subscribe((snapshots) => setSnapshot(snapshots));
  });

  return (
    <div class={styles["endpoint-card"]}>
      <div class={styles["endpoint-card__header"]}>
        <A
          class={styles["endpoint-card__title"]}
          href={"/by?name=" + encodeURIComponent(props.name)}
        >
          {props.name}
        </A>
        <a class={styles["endpoint-card__url"]} href={props.url}>
          {props.url}
        </a>
      </div>
      <Status snapshots={snapshot()}></Status>
    </div>
  );
}
