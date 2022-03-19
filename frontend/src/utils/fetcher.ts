import { Response } from "@/types/Response";

export async function fetchAllStaticSnapshots(urls: string[]) {
  try {
    const response: Response[][] = await Promise.all(
      urls.map((u) => fetch("/api/static?url=" + u).then((r) => r.json()))
    );

    return response;
  } catch (err) {
    // eslint-disable-next-line no-console
    console.error(err);
    throw err;
  }
}

export async function fetchSingleStaticSnapshot(url: string) {
  try {
    const response: Response[] = await fetch("/api/static?url=" + url).then(
      (r) => r.json()
    );

    return response;
  } catch (err) {
    // eslint-disable-next-line no-console
    console.error(err);
    throw err;
  }
}
