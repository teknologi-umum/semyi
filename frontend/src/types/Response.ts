import type { Monitor } from "./Monitor";
import type { Snapshot } from "./Snapshot";

export type Response = {
  metadata: Monitor;
  historical: Snapshot[];
};
