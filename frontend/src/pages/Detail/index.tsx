import { fromEvent, map } from "rxjs";
import { Link, Navigate, useSearchParams } from "solid-app-router";
import { createResource, Match, Switch } from "solid-js";
import DarkModeToggle from "@/components/DarkModeToggle";
import EndpointOverviewCard from "@/components/EndpointOverviewCard";
import EndpointStatusCard from "@/components/EndpointStatusCard";
import { fetchSingleStaticSnapshot } from "@/utils/fetcher";
import type { Response, Endpoint } from "@/types";
import config from "@config";
import styles from "./styles.module.css";

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
