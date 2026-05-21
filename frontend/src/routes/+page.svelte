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

// Everything here will need to be cleaned up lateer :P
let dialog: HTMLDialogElement;
let QRCode: any = $state(null);

function showQRCode(slug: string) {
  dialog?.showModal();
  if (QRCode) {
    const qrcode_container = dialog.querySelector("#dialog_qrcode");
    const qrcode_code = dialog.querySelector("#dialog_code");
    if (qrcode_container) {
      qrcode_container.innerHTML = "";
      const url = `${window.location.origin}/room/${slug}`;
      new QRCode(qrcode_container, url);
    }
    if (qrcode_code) {
      const code = slug.split("-").at(-1);
      qrcode_code.textContent = code ?? "";
    }
  }
}

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
  const module = await import("$lib/vendor/qrcode.min.js");
  QRCode = module.default;
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
                 <strong>{room.name}</strong> (Code: {room.code}) - <a href="/room/{room.slug}">Enter</a>
                 <button onclick={() => showQRCode(room.slug)}>Show QR Code</button>
               </li>
             {/each }
            </ul>
          {:else}
            <p>You haven't created any room yet.</p>
          {/if}
        {/await}
    </div>
    <dialog bind:this={dialog}>
      <div id="dialog_qrcode"></div>
      <p id="dialog_code"></p>
      <button class="close-btn" onclick={() => dialog?.close()}>Zamknij</button>
    </dialog>
{/if}

<style>
  dialog {
    border: none;
    border-radius: 16px;
    padding: 24px;
    box-shadow: 0 10px 30px rgba(0, 0, 0, 0.2);
    background: #ffffff;
    max-width: 90vw;
    width: 300px;
    box-sizing: border-box;

    /* Wycentrowanie dialogu na ekranie */
    position: fixed;
    top: 50%;
    left: 50%;
    transform: translate(-50%, -50%);
    margin: 0;
  }

  dialog[open] {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 16px;
  }

  dialog::backdrop {
    background: rgba(0, 0, 0, 0.5);
    backdrop-filter: blur(5px);
  }

  #dialog_qrcode {
    background: #ffffff;
    padding: 8px;
    border-radius: 8px;
  }

  #dialog_qrcode :global(img), #dialog_qrcode :global(canvas) {
    max-width: 100%;
    height: auto;
    display: block;
  }

  #dialog_code {
    font-size: 1.25rem;
    font-weight: 700;
    color: #1a1a1a;
    margin: 0;
  }

  .close-btn {
    background: #f3f4f6;
    border: none;
    border-radius: 8px;
    padding: 8px 16px;
    font-size: 0.875rem;
    font-weight: 600;
    color: #4b5563;
    cursor: pointer;
    transition: background 0.2s, color 0.2s;
    width: 100%;
  }

  .close-btn:hover {
    background: #e5e7eb;
    color: #1f2937;
  }
</style>


