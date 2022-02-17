import type { Snapshot } from "@/types/Snapshot";
import { createMemo, createSignal, For, onMount } from "solid-js";
import styles from "./Status.module.css";

interface StatusProps {
  snapshots: Snapshot[];
}

export default function Status(props: StatusProps) {
  let container: HTMLDivElement | undefined;
  const [containerWidth, setContainerWidth] = createSignal(0);
  const [hoveredSnapshotIndex, setHoveredSnapshotIndex] = createSignal<
    number | null
  >(null);

  const CONTAINER_HEIGHT = 30;
  const BAR_AMOUNT = 100;
  const GAP = 4;

  const barWidth = createMemo(() => {
    const value = (containerWidth() - GAP * (BAR_AMOUNT - 1)) / BAR_AMOUNT;
    return value < 1 ? 0 : value;
  });

  const barRadius = createMemo(() => {
    const value = barWidth() / 2;
    return value < 1 ? 0 : value;
  });

  onMount(() => {
    if (container !== undefined) {
      setContainerWidth(container.clientWidth);
    }
  });

  function showTooltip(e: MouseEvent) {
    const target = e.target as Element;
    if (target.tagName !== "rect") return;

    const isValid = target.getAttribute("data-isValid");
    if (isValid === null || isValid === "false") return;

    const idx = target.getAttribute("data-index");
    if (idx === null) return;

    setHoveredSnapshotIndex(parseInt(idx));
  }

  function hideTooltip() {
    setHoveredSnapshotIndex(null);
  }

  return (
    <>
      <div
        class={styles.overlay}
        style={{
          left:
            hoveredSnapshotIndex() !== null &&
            (hoveredSnapshotIndex() as number) * (barWidth() + GAP) + "px",
          transform: hoveredSnapshotIndex() !== null ? "scale(1)" : "scale(0)",
          visiblity: hoveredSnapshotIndex() !== null ? "visible" : "hidden",
          opacity: hoveredSnapshotIndex() !== null ? 1 : 0
        }}
      >
        <div class={styles.overlay__datetime}>
          <span class={styles.overlay__date}>
            {new Date(
              props.snapshots[hoveredSnapshotIndex()!]?.timestamp
            ).toLocaleDateString("en-GB", {
              day: "numeric",
              month: "short",
              year: "numeric"
            })}
          </span>
          <span class={styles.overlay__time}>
            {new Date(
              props.snapshots[hoveredSnapshotIndex()!]?.timestamp
            ).toLocaleTimeString("en-GB", {
              hour: "numeric",
              minute: "numeric",
              second: "numeric"
            })}
          </span>
        </div>
        <span class={styles["overlay__response-time"]}>
          Duration: {props.snapshots[hoveredSnapshotIndex()!]?.requestDuration}ms
        </span>
        <span class={styles["overlay__response-time"]}>
          Status Code: {props.snapshots[hoveredSnapshotIndex()!]?.statusCode}
        </span>
      </div>
      <div
        class={styles.status}
        ref={container}
        onMouseOver={showTooltip}
        onMouseLeave={hideTooltip}
      >
        <svg
          width={containerWidth()}
          height={30}
          xmlns="http://www.w3.org/2000/svg"
          version="1.1"
          viewBox={`0 0 ${containerWidth()} ${CONTAINER_HEIGHT}`}
        >
          <For
            each={Array(BAR_AMOUNT)
              .fill(0)
              .map((_, i) => i)
              .reverse()}
          >
            {(i) => (
              <rect
                data-index={i}
                data-isValid={props.snapshots[i]?.statusCode !== undefined}
                class={styles.status__bar}
                width={barWidth()}
                height={CONTAINER_HEIGHT}
                x={i * (barWidth() + GAP)}
                y="0"
                fill={
                  props.snapshots?.[i]?.statusCode !== undefined &&
                  props.snapshots?.[i]?.statusCode !== null
                    ? props.snapshots[i].statusCode === 200
                      ? "var(--color-emerald)"
                      : "var(--color-red)"
                    : "var(--color-light-gray)"
                }
                fill-opacity="1"
                rx={barRadius()}
                ry={barRadius()}
              />
            )}
          </For>
        </svg>
      </div>
    </>
  );
}
