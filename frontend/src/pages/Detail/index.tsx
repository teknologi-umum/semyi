import { Link, Navigate, useSearchParams } from "solid-app-router";
import EndpointStatusCard from "@/components/EndpointStatusCard";
import DarkModeToggle from "@/components/DarkModeToggle";
import EndpointOverviewCard from "@/components/EndpointOverviewCard";
import styles from "./Detail.module.css";
import config from "@config";
import type { Endpoint } from "@/types/Endpoint";

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
      <div class={styles.detail__body}>
        <EndpointStatusCard name={endpoint.name} url={endpoint.url} />
        <EndpointOverviewCard name={endpoint.name} />
      </div>
    </div>
  );
}
