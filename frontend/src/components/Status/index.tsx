import Tooltip from "@/components/StatusTooltip";
import type { Response, Snapshot } from "@/types";
import { For, createMemo, createSignal, onMount } from "solid-js";
import styles from "./styles.module.css";

interface StatusProps {
  snapshots: Snapshot[];
}

export default function Status(props: StatusProps) {
  let container: HTMLDivElement | undefined;
  const [containerWidth, setContainerWidth] = createSignal(0);
  const [hoveredSnapshotIndex, setHoveredSnapshotIndex] = createSignal<number | null>(null);

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

    const idx = target.getAttribute("data-index");
    if (idx === null) return;

    setHoveredSnapshotIndex(Number.parseInt(idx));
  }

  function hideTooltip() {
    setHoveredSnapshotIndex(null);
  }

  return (
    <>
      <Tooltip
        isVisible={hoveredSnapshotIndex() !== null}
        snapshotIndex={hoveredSnapshotIndex() ?? 0}
        snapshot={props.snapshots[hoveredSnapshotIndex() ?? 0]}
        left={containerWidth() - (hoveredSnapshotIndex() ?? 0) * (barWidth() + GAP)}
      />
      <div
        class={styles.status}
        ref={container}
        onMouseOver={showTooltip}
        onMouseLeave={hideTooltip}
        onFocus={() => {}}
        onBlur={hideTooltip}
      >
        <svg
          width={containerWidth()}
          height={CONTAINER_HEIGHT}
          viewBox={`0 0 ${containerWidth()} ${CONTAINER_HEIGHT}`}
          role="img"
          aria-label="Status timeline"
        >
          <title>Status timeline</title>
          <For
            each={Array(BAR_AMOUNT)
              .fill(0)
              .map((_, i) => i)
              .slice()
              .reverse()}
          >
            {(i) => (
              <rect
                data-index={i}
                class={styles.status__bar}
                width={barWidth()}
                height={CONTAINER_HEIGHT}
                x={containerWidth() - i * (barWidth() + GAP)}
                y="0"
                fill={
                  props.snapshots?.[i]?.status !== undefined
                    ? props.snapshots[i].status === 0
                      ? "var(--color-emerald)"
                      : props.snapshots[i].status === 1
                        ? "var(--color-red)"
                        : props.snapshots[i].status === 2
                          ? "var(--color-amber)"
                          : props.snapshots[i].status === 3
                            ? "var(--color-blue)"
                            : "var(--color-purple)"
                    : "var(--color-lighter-gray)"
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
