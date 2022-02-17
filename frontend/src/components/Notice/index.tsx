import styles from "./styles.module.css";

interface NoticeProps {
  text: string;
}

export default function Notice(props: NoticeProps) {
  return (
    <div class={styles.notice}>
      <h1 class={styles.notice__text}>{props.text}</h1>
    </div>
  );
}
