import { Snapshot } from "@/types/Snapshot";

const random = (min: number, max: number) => {
  return Math.floor(Math.random() * (max - min + 1)) + min;
};

const STATUS_CODE = [200, 200, 200, 200, 302, 404, 500];

const mainSnapshot: Snapshot[] = Array(60)
  .fill(0)
  .map(() => ({
    timeout: random(0, 5000),
    requestDuration: random(20, 5000),
    statusCode: 200,
    timestamp: Date.now(),
  }));

const supportSnapshot: Snapshot[] = Array(90)
  .fill(0)
  .map(() => ({
    timeout: random(0, 5000),
    requestDuration: random(20, 5000),
    statusCode: STATUS_CODE[random(0, STATUS_CODE.length - 1)],
    timestamp: Date.now(),
  }));

const documentationSnapshot: Snapshot[] = Array(80)
  .fill(0)
  .map(() => ({
    timeout: random(0, 5000),
    requestDuration: random(20, 5000),
    statusCode: 200,
    timestamp: Date.now(),
  }));

const blogSnapshot: Snapshot[] = Array(70)
  .fill(0)
  .map(() => ({
    timeout: random(0, 5000),
    requestDuration: random(20, 5000),
    statusCode: STATUS_CODE[random(0, STATUS_CODE.length - 1)],
    timestamp: Date.now(),
  }));

const fakeSnapshots: Record<string, Snapshot[]> = {
  "Main Website": mainSnapshot,
  "Support Website": supportSnapshot,
  "Documentation Website": documentationSnapshot,
  "Blog Website": blogSnapshot,
};

export default fakeSnapshots;
