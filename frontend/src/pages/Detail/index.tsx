import DarkModeToggle from "@/components/DarkModeToggle";
import EndpointOverviewCard from "@/components/EndpointOverviewCard";
import EndpointStatusCard from "@/components/EndpointStatusCard";
import Notice from "@/components/Notice";
import { BASE_URL } from "@/constants";
import { LeftArrowIcon } from "@/icons";
import type { Monitor, Response, Snapshot } from "@/types";
import { fetchSingleStaticSnapshot } from "@/utils/fetcher";
import config from "@config";
import * as Sentry from "@sentry/solid";
import { A, Navigate, useSearchParams } from "@solidjs/router";
import { fromEvent, map } from "rxjs";
import { Match, Switch, createResource, onMount } from "solid-js";
import styles from "./styles.module.css";

export default function DetailPage() {
  const abortController = new AbortController();
  const [searchParams] = useSearchParams();
  if (searchParams.id === "") {
    return <Navigate href="/" />;
  }

  const endpoint = config.monitors.find(
    ({ unique_id }) =>
      unique_id ===
      decodeURIComponent((Array.isArray(searchParams.id) ? searchParams.id.at(0) : searchParams.id) ?? ""),
  );
  if (endpoint === undefined) {
    return <Navigate href="/" />;
  }

  const [staticSnapshot] = createResource(() =>
    fetchSingleStaticSnapshot(endpoint.unique_id, "raw", abortController.signal),
  );

  const source = new EventSource(`${BASE_URL}/api/by?ids=${endpoint.unique_id}`);
  const snapshotStream$ = fromEvent<MessageEvent<string>>(source, "message").pipe(
    map((event) => {
      return Sentry.startSpan(
        {
          name: "detail.stream_message",
          op: "function",
          attributes: {
            "semyi.page": "detail",
            "semyi.monitor.id": endpoint.unique_id,
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

  onMount(() => {
    document.title = `Status for ${endpoint.name} | Semyi`;
  });

  return (
    <div class={styles.detail}>
      <div class={styles.detail__header}>
        <div class={styles["detail__header-left"]}>
          <h1 class={styles.detail__title}>Status for {endpoint.name}</h1>
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
                  <td class={styles["detail__metadata-value"]}>{endpoint.name}</td>
                </tr>
                <tr>
                  <td class={styles["detail__metadata-title"]}>URL:</td>
                  <td class={styles["detail__metadata-value"]}>{endpoint.public_url}</td>
                </tr>
                <tr>
                  <td class={styles["detail__metadata-title"]}>Description:</td>
                  <td class={styles["detail__metadata-value"]}>{endpoint.description}</td>
                </tr>
                <tr>
                  <td class={styles["detail__metadata-title"]}>Method:</td>
                  <td class={styles["detail__metadata-value"]}>{endpoint.http_method || "GET"}</td>
                </tr>
              </tbody>
            </table>
            <EndpointStatusCard
              monitorId={endpoint.unique_id}
              name={endpoint.name}
              url={endpoint.public_url ?? ""}
              staticSnapshot={staticSnapshot()?.historical ?? []}
              snapshotStream$={snapshotStream$}
            />
            <EndpointOverviewCard
              name={endpoint.name}
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
