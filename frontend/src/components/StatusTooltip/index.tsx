import type { Snapshot } from "@/types";
import { Match, Switch } from "solid-js";
import styles from "./styles.module.css";

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
        // minus 7 rem because we want to make it centered
        left: props.isVisible ? `${props.left - 5 * 16}px` : "0px",
        transform: props.isVisible ? "scale(1)" : "scale(0)",
        opacity: props.isVisible ? 1 : 0,
        visibility: props.isVisible ? "visible" : "hidden",
      }}
    >
      <Switch>
        <Match when={props.snapshot === undefined}>
          <span class={styles.tooltip__placeholder}>No Data Available Yet</span>
        </Match>
        <Match when={props.snapshot !== undefined}>
          {props.snapshot?.timestamp != null && props.snapshot?.timestamp !== "" ? (
            <>
              <div class={styles.tooltip__datetime}>
                <span class={styles.tooltip__date}>
                  {new Date(props.snapshot.timestamp).toLocaleDateString("en-GB", {
                    day: "numeric",
                    month: "short",
                    year: "numeric",
                  })}
                </span>
                <span class={styles.tooltip__time}>
                  {new Date(props.snapshot.timestamp).toLocaleTimeString("en-GB", {
                    hour: "numeric",
                    minute: "numeric",
                    second: "numeric",
                  })}
                </span>
              </div>
            </>
          ) : (
            <></>
          )}
          <span class={styles["tooltip__response-time"]}>Duration: {props.snapshot?.latency}ms</span>
          <span class={styles["tooltip__response-time"]}>
            Status:{" "}
            {props.snapshot?.status === 0
              ? "UP"
              : props.snapshot?.status === 1
                ? "DOWN"
                : props.snapshot?.status === 2
                  ? "DEGRADED"
                  : props.snapshot?.status === 3
                    ? "MAINTENANCE"
                    : "LIMITED"}
          </span>
        </Match>
      </Switch>
    </div>
  );
}
