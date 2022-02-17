import styles from "./EndpointOverviewCard.module.css";
import type { Response } from "@/types/Response";
import { Observable } from "rxjs";

interface EndpointOverviewCard {
  name: string;
  staticSnapshot: Response[] | undefined;
  snapshotStream$: Observable<Response>;
}

export default function EndpointOverviewCard(props: EndpointOverviewCard) {
  return (
    <div class={styles.overview}>
      <h2 class={styles.overview__title}>
        Overview for {props.name || "Digital Ocean"}
      </h2>
      <div class={styles.overview__content}>
        <div class={styles["overview__content-item"]}>
          <span class={styles["overview__item-label"]}>Uptime Rate</span>
          <span class={styles["overview__item-value"]}>100%</span>
        </div>
        <div class={styles["overview__content-item"]}>
          <span class={styles["overview__item-label"]}>Avg. Response Time</span>
          <span class={styles["overview__item-value"]}>120ms</span>
        </div>
        <div class={styles["overview__content-item"]}>
          <span class={styles["overview__item-label"]}>Max. Response Time</span>
          <span class={styles["overview__item-value"]}>200ms</span>
        </div>
        <div class={styles["overview__content-item"]}>
          <span class={styles["overview__item-label"]}>Min. Response Time</span>
          <span class={styles["overview__item-value"]}>20ms</span>
        </div>
      </div>
    </div>
  );
}
