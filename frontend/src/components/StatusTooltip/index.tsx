import { Snapshot } from "@/types/Snapshot";
import { Show } from "solid-js";
import styles from "./Tooltip.module.css";

interface TooltipProps {
  isVisible: boolean;
  snapshotIndex: number;
  snapshot: Snapshot | undefined;
  left: number;
}

export default function Tooltip(props: TooltipProps) {
  return (
    <div
      class={styles.tooltip}
      style={{
        left: props.snapshotIndex !== null && props.left + "px",
        transform:
          (props.snapshotIndex !== null ? "scale(1)" : "scale(0)") +
          " transformX(-50%)",
        visibility: props.snapshotIndex !== null ? "visible" : "hidden",
        opacity: props.snapshotIndex !== null ? 1 : 0
      }}
    >
      <div class={styles.tooltip__datetime}>
        <Show when={props.snapshot !== undefined}>
          <span class={styles.tooltip__date}>
            {new Date(props.snapshot!.timestamp).toLocaleDateString("en-GB", {
              day: "numeric",
              month: "short",
              year: "numeric"
            })}
          </span>
          <span class={styles.tooltip__time}>
            {new Date(props.snapshot!.timestamp).toLocaleTimeString("en-GB", {
              hour: "numeric",
              minute: "numeric",
              second: "numeric"
            })}
          </span>
        </Show>
      </div>
      <span class={styles["tooltip__response-time"]}>
        Duration: {props.snapshot?.requestDuration}ms
      </span>
      <span class={styles["tooltip__response-time"]}>
        Status Code: {props.snapshot?.statusCode}
      </span>
    </div>
  );
}
