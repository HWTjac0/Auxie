import type { PageLoad } from "./$types";

export const load: PageLoad = async ({ fetch }) => {
  const roomName = await fetch("/api/v1/room/random_name")
  const userName = await fetch("/api/v1/user/random_name")

  return {
    room: { name: await roomName.json().then(j => j.name) },
    user: { name: await userName.json().then(j => j.name) }
  };
}
