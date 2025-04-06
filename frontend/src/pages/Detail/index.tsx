import DarkModeToggle from "@/components/DarkModeToggle";
import EndpointOverviewCard from "@/components/EndpointOverviewCard";
import EndpointStatusCard from "@/components/EndpointStatusCard";
import Notice from "@/components/Notice";
import { BASE_URL } from "@/constants";
import { LeftArrowIcon } from "@/icons";
import type { Snapshot } from "@/types";
import { fetchSingleStaticSnapshot } from "@/utils/fetcher";
import * as Sentry from "@sentry/solid";
import { A, Navigate, useSearchParams } from "@solidjs/router";
import { fromEvent, map } from "rxjs";
import { Match, Switch, createResource, onCleanup, onMount } from "solid-js";
import styles from "./styles.module.css";

export default function DetailPage() {
  const abortController = new AbortController();
  const [searchParams] = useSearchParams();
  const uniqueId = (Array.isArray(searchParams.id) ? searchParams.id.at(0) : searchParams.id) ?? "";
  if (searchParams.id === undefined || searchParams.id === null || searchParams.id === "") {
    return <Navigate href="/" />;
  }

  const [staticSnapshot, { refetch }] = createResource(() =>
    fetchSingleStaticSnapshot(uniqueId, "raw", abortController.signal),
  );

  const source = new EventSource(`${BASE_URL}/api/by?ids=${uniqueId}`);
  const snapshotStream$ = fromEvent<MessageEvent<string>>(source, "message").pipe(
    map((event) => {
      return Sentry.startSpan(
        {
          name: "detail.stream_message",
          op: "function",
          attributes: {
            "semyi.page": "detail",
            "semyi.monitor.id": uniqueId,
          },
        },
        (span) => {
          try {
            const data = JSON.parse(event.data) as Snapshot;
            return data;
          } catch (err) {
            span.setStatus({
              code: 2,
              message: "parse_error",
            });
            throw err;
          }
        },
      );
    }),
  );

  let fallbackTimeout: NodeJS.Timeout | null = null;

  onMount(() => {
    document.title = `Status for ${staticSnapshot()?.metadata.name} | Semyi`;

    // Fallback mechanism in case the event source is not working
    fallbackTimeout = setTimeout(
      () => {
        refetch();
      },
      2 * 60 * 1000,
    ); // 2 minutes
  });

  onCleanup(() => {
    if (source != null) {
      source.close();
    }

    if (fallbackTimeout != null) {
      clearTimeout(fallbackTimeout);
    }
  });

  return (
    <div class={styles.detail}>
      <div class={styles.detail__header}>
        <div class={styles["detail__header-left"]}>
          <h1 class={styles.detail__title}>Status for {staticSnapshot()?.metadata.name}</h1>
          <A href="/" class={styles.detail__back}>
            <LeftArrowIcon /> Back to Home
          </A>
        </div>
        <DarkModeToggle />
      </div>
      <Switch fallback={<Notice text="Loading..." />}>
        <Match when={!staticSnapshot.loading}>
          <div class={styles.detail__body}>
            <table class={styles.detail__metadata}>
              <tbody>
                <tr>
                  <td class={styles["detail__metadata-title"]}>Name:</td>
                  <td class={styles["detail__metadata-value"]}>{staticSnapshot()?.metadata.name}</td>
                </tr>
                <tr>
                  <td class={styles["detail__metadata-title"]}>URL:</td>
                  <td class={styles["detail__metadata-value"]}>{staticSnapshot()?.metadata.public_url}</td>
                </tr>
                <tr>
                  <td class={styles["detail__metadata-title"]}>Description:</td>
                  <td class={styles["detail__metadata-value"]}>{staticSnapshot()?.metadata.description}</td>
                </tr>
              </tbody>
            </table>
            <EndpointStatusCard
              monitorId={staticSnapshot()?.metadata.id ?? ""}
              name={staticSnapshot()?.metadata.name ?? ""}
              url={staticSnapshot()?.metadata.public_url ?? ""}
              staticSnapshot={staticSnapshot()?.historical ?? []}
              snapshotStream$={snapshotStream$}
            />
            <EndpointOverviewCard
              name={staticSnapshot()?.metadata.name ?? ""}
              staticSnapshot={staticSnapshot()?.historical ?? []}
              snapshotStream$={snapshotStream$}
            />
          </div>
        </Match>
        <Match when={staticSnapshot.error !== undefined}>
          <Notice text="Error while fetching. Try checking the console." />
        </Match>
      </Switch>
    </div>
  );
}
