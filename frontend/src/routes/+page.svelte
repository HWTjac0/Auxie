<script lang="ts">
import { goto } from "$app/navigation";
import { onMount } from "svelte";

type User = {
  name: string;
  id: string;
  image: string;
};

type Room = {
  name: string;
  code: string;
  slug: string;
};

async function getRooms() {
  try {
    const res = await fetch("/api/v1/user/rooms");
    if (res.ok) {
      const data = await res.json();
      const rawRooms = data?.rooms;
      if (rawRooms && Array.isArray(rawRooms)) {
        return rawRooms.map((room: any) => {
          return {
            name: room.Name || room.name,
            slug: room.Slug || room.slug,
            code: room.JoinCode || room.join_code || room.code,
          };
        });
      }
    }
  } catch (err) {
    console.error("Error fetching rooms:", err);
  }
  return null;
}

let user: User | null = $state(null);
let loading = $state(true);
let roomsRes = $state<Promise<Room[] | null>>(Promise.resolve(null));

onMount(async () => {
  try {
    const meRes = await fetch("/api/v1/auth/me", {
      credentials: "include",
    });

    if (meRes.ok) {
      user = await meRes.json();
      roomsRes = getRooms();
    } else {
      goto("/welcome");
    }
  } catch (e) {
    goto("/welcome");
  } finally {
    loading = false;
  }
});
</script>

{#if loading}
    <p>Sprawdzanie sesji...</p>
{:else if user}
    <div class="user-info">
        {#if user.image}
            <img src={user.image} alt="Avatar" />
        {/if}
        <h1>Witaj, {user?.name}</h1>
        <a href="/api/v1/user/logout">Logout</a>
        {#await roomsRes}
         <p>Loading rooms hosted by user...</p> 
        {:then roomsList} 
          {#if roomsList && roomsList.length > 0}
            <ul>
             {#each roomsList as room}
               <li>
                 <strong>{room.name}</strong> (Kod: {room.code}) - <a href="/room/{room.slug}">Wejdź</a>
               </li>
             {/each }
            </ul>
          {:else}
            <p>Nie masz jeszcze żadnego pokoju.</p>
          {/if}
        {/await}
    </div>
{/if}

