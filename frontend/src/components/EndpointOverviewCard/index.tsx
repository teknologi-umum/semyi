import { map, Observable, take } from "rxjs";
import type { Response } from "@/types";
import styles from "./styles.module.css";
import { createMemo, createSignal, onMount } from "solid-js";

interface EndpointOverviewCardProps {
  name: string;
  staticSnapshot: Response[] | undefined;
  snapshotStream$: Observable<Response>;
}

export default function EndpointOverviewCard(props: EndpointOverviewCardProps) {
  const [snapshot, setSnapshot] = createSignal<Response[]>(
    // this doesn't need to be reactive
    // eslint-disable-next-line solid/reactivity
    props.staticSnapshot || []
  );
  const uptimeRate = createMemo(() => {
    const uptime = snapshot().filter((r) => r.success).length;
    const total = snapshot().length;
    return (uptime / total) * 100;
  });
  const avgRespTime = createMemo(() => {
    const total = snapshot().reduce((acc, r) => acc + r.requestDuration, 0);
    return (total / snapshot().length).toFixed(2);
  });
  const maxRespTime = createMemo(() => {
    const max = snapshot().reduce(
      (acc, r) => (r.requestDuration > acc ? r.requestDuration : acc),
      0
    );
    return max;
  });
  const minRespTime = createMemo(() => {
    const min = snapshot().reduce(
      (acc, r) => (r.requestDuration < acc ? r.requestDuration : acc),
      Infinity
    );
    return min;
  });

  onMount(() => {
    props.snapshotStream$
      .pipe(
        // this doesn't need to be reactive
        // eslint-disable-next-line solid/reactivity
        map((newSnapshot) => snapshot().concat(newSnapshot)),
        take(100)
      )
      .subscribe((s) => setSnapshot(s));
  });

  return (
    <div class={styles.overview}>
      <h2 class={styles.overview__title}>Overview</h2>
      <div class={styles.overview__content}>
        <div class={styles["overview__content-item"]}>
          <span class={styles["overview__item-label"]}>Avg. Response Time</span>
          <span class={styles["overview__item-value"]}>{avgRespTime()}ms</span>
        </div>
        <div class={styles["overview__content-item"]}>
          <span class={styles["overview__item-label"]}>Max. Response Time</span>
          <span class={styles["overview__item-value"]}>{maxRespTime()}ms</span>
        </div>
        <div class={styles["overview__content-item"]}>
          <span class={styles["overview__item-label"]}>Min. Response Time</span>
          <span class={styles["overview__item-value"]}>{minRespTime()}ms</span>
        </div>
        <div class={styles["overview__content-item"]}>
          <span class={styles["overview__item-label"]}>Uptime Rate</span>
          <span class={styles["overview__item-value"]}>{uptimeRate()}%</span>
        </div>
      </div>
    </div>
  );
}
