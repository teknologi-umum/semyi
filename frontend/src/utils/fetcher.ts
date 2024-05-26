import { BASE_URL } from "@/constants";
import { Response } from "@/types/Response";

export async function fetchAllStaticSnapshots(urls: string[]) {
  try {
    const response: Response[][] = await Promise.all(
      urls.map((u) =>
        fetch(BASE_URL + "/api/static?url=" + u).then((r) => r.json())
      )
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
    const response: Response[] = await fetch(
      BASE_URL + "/api/static?url=" + url
    ).then((r) => r.json());

    return response;
  } catch (err) {
    // eslint-disable-next-line no-console
    console.error(err);
    throw err;
  }
}