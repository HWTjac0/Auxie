<script lang="ts">
import { goto } from "$app/navigation";
import { onMount } from "svelte";
import InviteDialog from "../components/InviteDialog.svelte";
import Button from "../components/Button.svelte";

type User = {
  name: string;
  id: string;
  image: string;
  spotify_name?: string;
};

type Room = {
  name: string;
  code: string;
  slug: string;
};

let activeSlug = $state("");
let inviteDialog: any = $state(null);

const loginWithSpotify = () => {
  window.location.href = "http://127.0.0.1:8080/api/v1/auth/spotify/login";
};

function showQRCode(slug: string) {
  activeSlug = slug;
  inviteDialog?.show();
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

<div class="page-wrapper">
  <div class="background_gradient background_gradient_top"></div>
  <div class="background_gradient background_gradient_bottom"></div>
  <div class="overlay"></div>

  {#if loading}
    <p class="subtitle" style="text-align: center; margin-top: 40px;">
      Sprawdzanie sesji...
    </p>
  {:else if user}
    <div class="content_wrapper">
      <div class="login_container">
        <div class="user-info">
          {#if user.image}
            <img src={user.image} alt="Avatar" class="avatar" />
          {/if}
          <h2 class="sora-800">
            Witaj, {user.spotify_name ? user.spotify_name : user?.name}
          </h2>
          <a href="/api/v1/user/logout" class="logout-link onest-500">Log out</a>
        </div>

        <div class="rooms-section">
          <p
            class="onest-500 subtitle"
            style="text-align: left; margin-bottom: 10px;"
          >
            Your rooms:
          </p>
          {#await roomsRes}
            <p class="subtitle">Loading rooms...</p>
          {:then roomsList}
            {#if roomsList && roomsList.length > 0}
              <ul class="rooms-list">
                {#each roomsList as room}
                  <li class="room-item">
                    <span class="room-name">{room.name}</span>
                    <span class="room-code">#{room.code}</span>
                    <div class="room-actions">
                      <a href="/room/{room.slug}" class="room-btn enter-btn"
                        >Enter</a
                      >
                      <button
                        class="room-btn qr-btn"
                        onclick={() => showQRCode(room.slug)}>QR</button
                      >
                    </div>
                  </li>
                {/each}
              </ul>
            {:else}
              <p class="subtitle" style="text-align: left;">
              You haven't created any room yet.
              </p>
            {/if}
          {/await}
        </div>

        <div class="button_group">
          <Button href="/create" class="onest-800">Host party</Button>
          <Button
            href="/join"
            bgColor="var(--auxie-electric-purple-500)"
            shadowColor="var(--auxie-electric-purple-700)"
          >
            Join party
          </Button>
        </div>

        {#if !user.spotify_name}
          <div class="separator_row">
            <div class="separator"></div>
            <p>Or login with</p>
            <div class="separator"></div>
          </div>
          <div class="button_group">
            <Button
              onclick={loginWithSpotify}
              bgColor="var(--auxie-intense-mint-500)"
              shadowColor="var(--auxie-intense-mint-700)"
            >
              Login with Spotify
            </Button>
          </div>
        {/if}
      </div>
    </div>
    <InviteDialog bind:this={inviteDialog} slug={activeSlug} />
  {/if}
</div>

<style>
  /* Skopiowane style z ekranu Welcome w celu zintegrowania UI */
  .page-wrapper {
    background-color: var(--auxie-deep-navy-900);
    min-height: 100vh;
    position: relative;
    overflow: hidden;
    display: flex;
    align-items: center;
    justify-content: center;
  }
  .overlay {
    background-color: var(--auxie-deep-navy-900);
    opacity: 0.56;
    z-index: -10;
    position: fixed;
    top: 0;
    left: 0;
    width: 100%;
    height: 100%;
    pointer-events: none;
  }

  .background_gradient {
    position: fixed;
    z-index: 0;
    width: 60%;
    height: 60%;
    filter: blur(150px);
    border-radius: 100%;
  }

  .background_gradient_top {
    top: -30%;
    left: -30%;
    background: linear-gradient(
      to top right,
      var(--auxie-electric-purple-600),
      var(--auxie-intense-mint-500)
    );
    animation: float-top 15s ease-in-out infinite alternate;
  }

  .background_gradient_bottom {
    bottom: -30%;
    right: -30%;
    background: linear-gradient(
      to top right,
      var(--auxie-razzmatazz-600),
      var(--auxie-electric-purple-300)
    );
    animation: float-bottom 20s ease-in-out infinite alternate;
  }

  @keyframes float-top {
    0% {
      transform: translate(0, 0) scale(1) rotate(0deg);
      filter: blur(150px) hue-rotate(0deg);
    }
    50% {
      transform: translate(5%, 5%) scale(1.1) rotate(15deg);
      filter: blur(150px) hue-rotate(30deg);
    }
    100% {
      transform: translate(-5%, 5%) scale(0.9) rotate(-10deg);
      filter: blur(150px) hue-rotate(-15deg);
    }
  }

  @keyframes float-bottom {
    0% {
      transform: translate(0, 0) scale(1) rotate(0deg);
      filter: blur(150px) hue-rotate(0deg);
    }
    50% {
      transform: translate(-5%, -5%) scale(1.1) rotate(-20deg);
      filter: blur(150px) hue-rotate(-30deg);
    }
    100% {
      transform: translate(5%, -5%) scale(0.95) rotate(10deg);
      filter: blur(150px) hue-rotate(15deg);
    }
  }

  .content_wrapper {
    position: relative;
    z-index: 10;
    width: 100%;
    max-width: 300px;
    padding: 20px 0;
  }

  .login_container {
    width: 100%;
    display: grid;
    row-gap: 40px;
  }

  .button_group {
    width: 100%;
    display: grid;
    row-gap: 18px;
  }
  .separator_row {
    font-family: "Onest";
    font-weight: bold;
    font-size: 14px;
    color: var(--auxie-cloud-white-200);
    display: grid;
    grid-template-columns: 30% 1fr 30%;
    justify-items: center;
    align-items: center;
  }
  .separator {
    width: 100%;
    height: 3px;
    border-radius: 4px;
    background-color: var(--auxie-deep-navy-500);
  }
  h2 {
    font-size: 32px;
    color: var(--auxie-cloud-white-50);
    text-align: center;
  }
  .subtitle {
    color: var(--auxie-cloud-white-600);
    font-size: 16px;
    margin-bottom: 32px;
  }
  .user-info {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 10px;
  }
  .avatar {
    width: 60px;
    height: 60px;
    border-radius: 50%;
  }
  .logout-link {
    color: var(--auxie-cloud-white-600);
    font-size: 14px;
    text-decoration: underline;
  }
  .rooms-section {
    font-family: "Onest";
    }
  .rooms-list {
    list-style: none;
    padding: 0;
    margin: 0;
    display: grid;
    row-gap: 10px;
  }
  .room-item {
    background-color: var(--auxie-deep-navy-700);
    border: 1px solid rgba(255, 255, 255, 0.06);
    border-radius: 14px;
    padding: 12px 14px;
    display: grid;
    grid-template-columns: 1fr auto;
    grid-template-rows: auto auto;
    gap: 6px 8px;
    align-items: center;
  }
  .room-name {
    color: var(--auxie-cloud-white-50);
    font-weight: 700;
    font-size: 15px;
    grid-column: 1;
    grid-row: 1;
  }
  .room-code {
    color: var(--auxie-cloud-white-600);
    font-size: 12px;
    font-family: monospace;
    grid-column: 1;
    grid-row: 2;
  }

  .room-actions {
    grid-column: 2;
    grid-row: 1 / 3;
    display: flex;
    gap: 6px;
    align-items: center;
  }

  .room-btn {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    padding: 6px 12px;
    border-radius: 8px;
    font-size: 13px;
    font-weight: 600;
    cursor: pointer;
    border: none;
    text-decoration: none;
    transition: opacity 0.2s ease;
  }

  .room-btn:hover {
    opacity: 0.8;
  }

  .enter-btn {
    background-color: var(--auxie-razzmatazz-600, #e91e8c);
    color: white;
  }

  .qr-btn {
    background-color: var(--auxie-deep-navy-500);
    color: var(--auxie-cloud-white-200);
    border: 1px solid rgba(255, 255, 255, 0.15);
  }
</style>
