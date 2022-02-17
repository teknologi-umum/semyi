import styles from "./styles.module.css";

export default function Loading() {
  return (
    <div class={styles.loading}>
      <h1 class={styles.loading__text}>Loading...</h1>
    </div>
  );
}
