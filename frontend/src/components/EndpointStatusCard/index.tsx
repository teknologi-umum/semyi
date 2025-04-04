import Status from "@/components/Status";
import type { Response, Snapshot } from "@/types";
import { A } from "@solidjs/router";
import { type Observable, filter, take } from "rxjs";
import { createSignal, onMount } from "solid-js";
import styles from "./styles.module.css";

interface EndpointStatusCardProps {
  monitorId: string;
  name: string;
  url: string;
  staticSnapshot: Snapshot[];
  snapshotStream$: Observable<Snapshot>;
}

export default function EndpointStatusCard(props: EndpointStatusCardProps) {
  const [snapshot, setSnapshot] = createSignal<Snapshot[]>(
    // this doesn't need to be reactive
    // eslint-disable-next-line solid/reactivity
    props.staticSnapshot.sort((a, b) => new Date(b.timestamp).getTime() - new Date(a.timestamp).getTime()),
  );

  const [averageLatency, setAverageLatency] = createSignal<number>(
    Number.parseFloat((snapshot()?.reduce((acc, curr) => acc + curr.latency, 0) / snapshot()?.length).toFixed(1)),
  );

  onMount(() => {
    props.snapshotStream$
      .pipe(
        // this doesn't need to be reactive
        // eslint-disable-next-line solid/reactivity
        filter((newSnapshot) => newSnapshot.monitor_id === props.monitorId),
        take(100),
      )
      .subscribe((snapshots) => {
        // should concat from the start of the array
        // and remove the last item of the array
        // make sure the array length is 100
        setSnapshot(snapshot()?.concat(snapshots).slice(0, 100));
        setAverageLatency(
          Number.parseFloat((snapshot()?.reduce((acc, curr) => acc + curr.latency, 0) / snapshot()?.length).toFixed(1)),
        );
      });
  });

  return (
    <div class={styles["endpoint-card"]}>
      <div class={styles["endpoint-card__header"]}>
        <div class={styles["endpoint-card__header-left"]}>
          <A class={styles["endpoint-card__title"]} href={`/by?id=${encodeURIComponent(props.monitorId)}`}>
            {props.name}
          </A>
          <a class={styles["endpoint-card__url"]} href={props.url}>
            {props.url}
          </a>
        </div>
        <div class={styles["endpoint-card__latency"]}>Average Latency {averageLatency()}ms</div>
      </div>
      <Status snapshots={snapshot()} />
    </div>
  );
}
