import { Link, Navigate, useSearchParams } from "solid-app-router";
import { createMemo, createSignal } from "solid-js";
import EndpointCard from "@/components/EndpointCard";
import styles from "./Detail.module.css";
import type { Endpoint } from "@/types/Endpoint";
import config from "@config";

export default function DetailPage() {
  const [searchParams] = useSearchParams();

  if (searchParams.name === "") {
    return <Navigate href="/" />;
  }

  const endpoint = config.find(({ name }) => name === searchParams.name);

  if (endpoint === undefined) {
    return <Navigate href="/" />;
  }

  return (
    <div class={styles.detail}>
      <div class={styles.detail__header}>
        <h1 class={styles.detail__title}>Status for {searchParams.name}</h1>
        <Link href="/" class={styles.detail__back}>
          Back to Home
        </Link>
      </div>
      <div class={styles.detail__body}>
        <EndpointCard name={endpoint.name} url={endpoint.url} />
      </div>
    </div>
  );
}
