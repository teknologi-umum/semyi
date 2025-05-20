import type { Response, Snapshot } from "@/types";
import { type Observable, map, take } from "rxjs";
import { createMemo, createSignal, onMount, Show } from "solid-js";
import styles from "./styles.module.css";

interface EndpointOverviewCardProps {
  name: string;
  staticSnapshot: Snapshot[] | undefined;
  snapshotStream$: Observable<Snapshot>;
}

export default function EndpointOverviewCard(props: EndpointOverviewCardProps) {
  const [snapshot, setSnapshot] = createSignal<Snapshot[]>(
    // this doesn't need to be reactive
    // eslint-disable-next-line solid/reactivity
    props.staticSnapshot || [],
  );
  const uptimeRate = createMemo(() => {
    const total = snapshot().length;
    if (total === 0) return 0;
    const upCount = snapshot().filter((s) => s.status === 0).length;
    return Math.round((upCount / total) * 100);
  });
  const avgRespTime = createMemo(() => {
    const total = snapshot().reduce((acc, r) => acc + (r.latency ?? 0), 0);
    return (total / snapshot().length).toFixed(2);
  });
  const maxRespTime = createMemo(() => {
    const max = snapshot().reduce((acc, r) => (r.latency > acc ? r.latency : acc), 0);
    return max.toFixed(1);
  });
  const minRespTime = createMemo(() => {
    const min = snapshot().reduce((acc, r) => (r.latency < acc ? r.latency : acc), Number.POSITIVE_INFINITY);
    return min.toFixed(1);
  });

  const lastAdditionalMessage = createMemo(() => {
    // The value of "additional_message" that exists within the last hour.
    // If the value does not exists for snapshot bar within the hour, we return `null`.
    const lastHour = snapshot().filter((s) => new Date(s.timestamp).getTime() > Date.now() - 3600 * 1000);
    let additionalMessage: string | null = null;
    for (const s of lastHour) {
      if (s.additional_message != null) {
        additionalMessage = s.additional_message;
        break;
      }
    }

    return additionalMessage;
  });

  const lastHttpProtocol = createMemo(() => {
    // The value of "http_protocol" that exists within the last hour.
    // If the value does not exists for snapshot bar within the hour, we return `null`.
    const lastHour = snapshot().filter((s) => new Date(s.timestamp).getTime() > Date.now() - 3600 * 1000);
    let httpProtocol: string | null = null;
    for (const s of lastHour) {
      if (s.http_protocol != null) {
        httpProtocol = s.http_protocol;
        break;
      }
    }

    return httpProtocol;
  });

  const lastTLSInformation = createMemo(() => {
    // The value of "tls_version", "tls_cipher_name", and "tls_expiry_date" that exists within the last hour.
    // If the value does not exists for snapshot bar within the hour, we return `null`.
    const lastHour = snapshot().filter((s) => new Date(s.timestamp).getTime() > Date.now() - 3600 * 1000);
    let tlsVersion: string | null = null;
    let tlsCipherName: string | null = null;
    let tlsExpiryDate: string | null = null;
    for (const s of lastHour) {
      if (s.tls_version != null) {
        tlsVersion = s.tls_version;
      }
      if (s.tls_cipher_name != null) {
        tlsCipherName = s.tls_cipher_name;
      }
      if (s.tls_expiry_date != null) {
        tlsExpiryDate = s.tls_expiry_date;
      }
    }
    return {
      tlsVersion,
      tlsCipherName,
      tlsExpiryDate,
    };
  });

  onMount(() => {
    props.snapshotStream$
      .pipe(
        // this doesn't need to be reactive
        // eslint-disable-next-line solid/reactivity
        map((newSnapshot) => snapshot().concat(newSnapshot)),
        take(100),
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

      <div class={styles.overview__content}>
          <div class={styles["overview__content-item"]}>
            <span class={styles["overview__item-label"]}>Message</span>
            <span class={styles["overview__item-value"]}>{lastAdditionalMessage() ?? ""}</span>
          </div>
        <Show when={lastHttpProtocol() != null}>
          <div class={styles["overview__content-item"]}>
            <span class={styles["overview__item-label"]}>HTTP Protocol</span>
            <span class={styles["overview__item-value"]}>{lastHttpProtocol()}</span>
          </div>
        </Show>
      </div>
      <Show when={lastTLSInformation().tlsVersion != null && lastTLSInformation().tlsCipherName != null && lastTLSInformation().tlsExpiryDate != null}>
        <div class={styles.overview__content}>
            <div class={styles["overview__content-item"]}>
              <span class={styles["overview__item-label"]}>TLS Version</span>
              <span class={styles["overview__item-value"]}>{lastTLSInformation().tlsVersion}</span>
            </div>
            <div class={styles["overview__content-item"]}>
              <span class={styles["overview__item-label"]}>TLS Cipher</span>
              <span class={styles["overview__item-value"]}>{lastTLSInformation().tlsCipherName?.replaceAll("_", " ")}</span>
            </div>
            <div class={styles["overview__content-item"]}>
              <span class={styles["overview__item-label"]}>TLS Expiry Date</span>
              <span class={styles["overview__item-value"]}>{new Date(lastTLSInformation().tlsExpiryDate as string).toLocaleString(undefined, { dateStyle: "long", timeStyle: "short" })}</span>
            </div>
        </div>
      </Show>
    </div>
  );
}
