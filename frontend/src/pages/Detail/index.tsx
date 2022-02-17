import { Link, Navigate, useSearchParams } from "solid-app-router";
import EndpointStatusCard from "@/components/EndpointStatusCard";
import DarkModeToggle from "@/components/DarkModeToggle";
import EndpointOverviewCard from "@/components/EndpointOverviewCard";
import styles from "./Detail.module.css";
import config from "@config";
import type { Endpoint } from "@/types/Endpoint";
import type { Response } from "@/types/Response";
import { fetchSingleStaticSnapshot } from "@/utils/fetchStaticSnapshots";
import { createResource, Match, Switch } from "solid-js";
import { fromEvent, map } from "rxjs";

export default function DetailPage() {
  const [searchParams] = useSearchParams();
  if (searchParams.name === "") {
    return <Navigate href="/" />;
  }

  const endpoint: Endpoint | undefined = config.find(
    ({ name }) => name === decodeURIComponent(searchParams.name)
  );
  if (endpoint === undefined) {
    return <Navigate href="/" />;
  }

  const [staticSnapshot] = createResource(() =>
    fetchSingleStaticSnapshot(endpoint.url)
  );

  const source = new EventSource(
    import.meta.env.VITE_BASE_URL + "/api/by?url=" + endpoint.url
  );
  const snapshotStream$ = fromEvent<MessageEvent<string>>(
    source,
    "message"
  ).pipe(map((event) => JSON.parse(event.data) as Response));

  return (
    <div class={styles.detail}>
      <div class={styles.detail__header}>
        <div class={styles["detail__header-left"]}>
          <h1 class={styles.detail__title}>Status for {searchParams.name}</h1>
          <Link href="/" class={styles.detail__back}>
            Back to Home
          </Link>
        </div>
        <DarkModeToggle />
      </div>
      <Switch fallback={<div>Loading...</div>}>
        <Match when={!staticSnapshot.loading}>
          <div class={styles.detail__body}>
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
          <h1>Error while fetching</h1>
        </Match>
      </Switch>
    </div>
  );
}
