export interface Monitor {
  /** Unique identifier for the monitor. This should remain constant to maintain historical data when monitor configuration changes. */
  id: string;
  /** Display name shown in the dashboard. */
  name: string;
  /** Optional friendly description of what is being monitored. */
  description?: string;
  /** Optional public URL shown in dashboard, different from the actual check URL. */
  public_url?: string;
}
