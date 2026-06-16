import type { PageLoad } from "./$types";

export type User = {
  ID: number;
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
      users: (data.users || []).map((u: any) => ({
        ID: u.ID || u.id || 0,
        Username: u.Username || u.username || "",
        CurrentRole: u.CurrentRole || u.current_role || "Guest",
        AvatarUrl: u.AvatarUrl || u.avatar_url || ""
      })) as User[],
      queue: data.queue || [],
      proposedQueue: data.proposedQueue || [],
      currentUserId: data.current_user_id
    }
  } else {
    return { slug: params.slug }
  }
};
