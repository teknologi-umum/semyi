import { fromEvent, map } from "rxjs";
import { A, Navigate, useSearchParams } from "@solidjs/router";
import { createResource, Match, onMount, Switch } from "solid-js";
import DarkModeToggle from "@/components/DarkModeToggle";
import EndpointOverviewCard from "@/components/EndpointOverviewCard";
import EndpointStatusCard from "@/components/EndpointStatusCard";
import Notice from "@/components/Notice";
import { fetchSingleStaticSnapshot } from "@/utils/fetcher";
import type { Response, Endpoint } from "@/types";
import config from "@config";
import styles from "./styles.module.css";
import { LeftArrowIcon } from "@/icons";
import { BASE_URL } from "@/constants";

export default function DetailPage() {
  const [searchParams] = useSearchParams();
  if (searchParams.name === "") {
    return <Navigate href="/" />;
  }

  const endpoint: Endpoint | undefined = config.endpoints.find(
    ({ name }) => name === decodeURIComponent(searchParams.name)
  );
  if (endpoint === undefined) {
    return <Navigate href="/" />;
  }

  const [staticSnapshot] = createResource(() =>
    fetchSingleStaticSnapshot(endpoint.url)
  );

  const source = new EventSource(BASE_URL + "/api/by?url=" + endpoint.url);
  const snapshotStream$ = fromEvent<MessageEvent<string>>(
    source,
    "message"
  ).pipe(map((event) => JSON.parse(event.data) as Response));

  onMount(() => {
    document.title = `Status for ${endpoint.name} | Semyi`;
  });

  return (
    <div class={styles.detail}>
      <div class={styles.detail__header}>
        <div class={styles["detail__header-left"]}>
          <h1 class={styles.detail__title}>Status for {searchParams.name}</h1>
          <A href="/" class={styles.detail__back}>
            <LeftArrowIcon /> Back to Home
          </A>
        </div>
        <DarkModeToggle />
      </div>
      <Switch fallback={<Notice text="Loading..." />}>
        <Match when={!staticSnapshot.loading}>
          <div class={styles.detail__body}>
            <table class={styles["detail__metadata"]}>
              <tbody>
                <tr>
                  <td class={styles["detail__metadata-title"]}>Name:</td>
                  <td class={styles["detail__metadata-value"]}>
                    {endpoint.name}
                  </td>
                </tr>
                <tr>
                  <td class={styles["detail__metadata-title"]}>URL:</td>
                  <td class={styles["detail__metadata-value"]}>
                    {endpoint.url}
                  </td>
                </tr>
                <tr>
                  <td class={styles["detail__metadata-title"]}>Description:</td>
                  <td class={styles["detail__metadata-value"]}>
                    {endpoint.description}
                  </td>
                </tr>
                <tr>
                  <td class={styles["detail__metadata-title"]}>Method:</td>
                  <td class={styles["detail__metadata-value"]}>
                    {endpoint.method || "GET"}
                  </td>
                </tr>
              </tbody>
            </table>
            <EndpointStatusCard
              name={endpoint.name}
              url={endpoint.url}
              staticSnapshot={staticSnapshot()}
              snapshotStream$={snapshotStream$}
            />
            <EndpointOverviewCard
              name={endpoint.name}
              staticSnapshot={staticSnapshot()}
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