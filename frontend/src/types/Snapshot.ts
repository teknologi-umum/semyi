export interface Snapshot {
  monitor_id: string;
  status: 0 | 1 | 2 | 3 | 4; // 0: Success, 1: Failure, 2: Degraded, 3: Under Maintenance, 4: Limited
  latency: number;
  timestamp: string;
  additional_message?: string;
  http_protocol?: string;
  tls_version?: string;
  tls_cipher_name?: string;
  tls_expiry_date?: string;
}
