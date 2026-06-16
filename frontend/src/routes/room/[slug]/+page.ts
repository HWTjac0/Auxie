import type { PageLoad } from "./$types";

export type User = {
  Username: string;
  CurrentRole: string;
  AvatarUrl?: string;
}

export type Room = {
  ID: number;
  Name: string;
  JoinCode: string;
  Slug: string;
  HostID: number;
  LastPlayedPosition: { Int64: number; Valid: boolean } | null;
  CreatedAt: string;
}

export const load: PageLoad = async ({ params, fetch }) => {
  const res = await fetch(`/api/v1/room/${params.slug}`)
  const data = await res.json()
  if (res.ok) {
    return {
      slug: params.slug,
      room: data.room as Room,
      users: data.users as Array<User>,
      queue: data.queue || []
    }
  } else {
    return { slug: params.slug }
  }
};
