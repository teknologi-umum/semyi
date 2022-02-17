import { Response } from "@/types/Response";

export async function fetchAllStaticSnapshots(urls: string[]) {
  try {
    const response: Response[][] = await Promise.all(
      urls.map((u) =>
        fetch(import.meta.env.VITE_BASE_URL + "/api/static?url=" + u).then(
          (r) => r.json()
        )
      )
    );

    return response;
  } catch (err) {
    console.error(err);
  }
}

export async function fetchSingleStaticSnapshot(url: string) {
  try {
    const response: Response[] = await fetch(
      import.meta.env.VITE_BASE_URL + "/api/static?url=" + url
    ).then((r) => r.json());

    return response;
  } catch (err) {
    console.error(err);
  }
}
