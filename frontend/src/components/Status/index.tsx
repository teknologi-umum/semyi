import { createMemo, createSignal, For, onMount } from "solid-js";
import Tooltip from "@/components/StatusTooltip";
import type { Response } from "@/types";
import styles from "./styles.module.css";

interface StatusProps {
  snapshots: Response[];
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

    const idx = target.getAttribute("data-index");
    if (idx === null) return;

    setHoveredSnapshotIndex(parseInt(idx));
  }

  function hideTooltip() {
    setHoveredSnapshotIndex(null);
  }

  return (
    <>
      <Tooltip
        isVisible={hoveredSnapshotIndex() !== null}
        snapshotIndex={hoveredSnapshotIndex()}
        snapshot={props.snapshots[hoveredSnapshotIndex()!]}
        left={containerWidth() - hoveredSnapshotIndex() * (barWidth() + GAP)}
      />
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
                  props.snapshots?.[i]?.statusCode !== undefined
                    ? props.snapshots[i].success
                      ? "var(--color-emerald)"
                      : "var(--color-red)"
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
