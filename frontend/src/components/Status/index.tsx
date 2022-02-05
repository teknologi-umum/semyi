import type { Snapshot } from "@/types/Snapshot";
import { createMemo, createSignal, For, onMount } from "solid-js";
import styles from "./Status.module.css";

interface StatusProps {
  snapshots: Snapshot[];
}

export default function Status(props: StatusProps) {
  let container: HTMLDivElement | undefined;
  const [containerWidth, setContainerWidth] = createSignal(0);

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

  return (
    <div class={styles.status} ref={container}>
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
              class="status__bar"
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
  );
}
