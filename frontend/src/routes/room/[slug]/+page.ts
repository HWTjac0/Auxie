import type { PageLoad } from "./$types";

type User = {
  Username: string;
  CurrentRole: string;
}

export const load: PageLoad = async ({ params, fetch }) => {
  const res = await fetch(`/api/v1/room/${params.slug}`)
  const data = await res.json()
  if (res.ok) {
    return {
      slug: params.slug,
      room: data.room,
      users: data.users as Array<User>,
    }
  } else {
    return { slug: params.slug }
  }
};
