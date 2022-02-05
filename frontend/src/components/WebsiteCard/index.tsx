import type { Snapshot } from "@/types/Snapshot";
import Status from "@/components/Status";
import styles from "./WebsiteCard.module.css";

interface WebsiteCardProps {
  name: string;
  url: string;
  snapshots: Snapshot[];
}

export default function WebsiteCard(props: WebsiteCardProps) {
  return (
    <div class={styles["website-card"]}>
      <div class={styles["website-card__content"]}>
        <div class={styles["website-card__header"]}>
          <span class={styles["website-card__title"]}>{props.name}</span>
          <a class={styles["website-card__url"]} href={props.url}>
            {props.url}
          </a>
        </div>
        <Status snapshots={props.snapshots}></Status>
      </div>
    </div>
  );
}
