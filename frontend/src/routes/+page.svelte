<script lang="ts">
import { goto } from "$app/navigation";
import { onMount } from "svelte";
import InviteDialog from "../components/InviteDialog.svelte";
import Button from "../components/Button.svelte";
import UserAvatar from "../components/UserAvatar.svelte";

type User = {
  name: string;
  id: string;
  image: string;
  spotify_name?: string;
  soundcloud_name?: string;
  tidal_name?: string;
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

const loginWithSoundCloud = () => {
  window.location.href = "http://127.0.0.1:8080/api/v1/auth/soundcloud/login";
};

const loginWithTidal = () => {
  window.location.href = "http://127.0.0.1:8080/api/v1/auth/tidal/login";
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
let refreshing = $state(false);

async function refreshStatus() {
  if (refreshing) return;
  refreshing = true;
  try {
    const res = await fetch("/api/v1/user/refresh", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
    });
    if (res.ok) {
      user = await res.json();
    }
  } catch (err) {
    console.error("Failed to refresh user status:", err);
  } finally {
    refreshing = false;
  }
}

async function disconnectService(service: "spotify" | "tidal" | "soundcloud") {
  if (confirm(`Are you sure you want to disconnect your ${service} account?`)) {
    try {
      const res = await fetch(`/api/v1/user/disconnect/${service}`, {
        method: "POST",
      });
      if (res.ok) {
        await refreshStatus();
      } else {
        alert("Failed to disconnect service");
      }
    } catch (err) {
      console.error(err);
      alert("Error occurred while disconnecting service");
    }
  }
}

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
          <h2 class="sora-800">
            Welcome,<br> {user.spotify_name ? user.spotify_name : user?.name}
          </h2>
          <div>
            <UserAvatar 
              username={user.spotify_name ? user.spotify_name : user?.name}
              src={user.image || ""}
              size={60}
            />
            <a href="/api/v1/user/logout" class="logout-link onest-500">Log out</a>
          </div>
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

        <div class="button_group_row">
          <Button href="/create" class="onest-800 smaller-btn">Host party</Button>
          <Button
            href="/join"
            class="smaller-btn"
            bgColor="var(--auxie-electric-purple-500)"
            shadowColor="var(--auxie-electric-purple-700)"
          >
            Join party
          </Button>
        </div>
        <div class="accounts_group">
          <div class="separator_row">
            <div class="separator"></div>
            <div class="header-with-action">
              <span class="header-title">Connected accounts</span>
              <button onclick={refreshStatus} class="refresh-btn" title="Refresh connections" disabled={refreshing}>
                {#if refreshing}
                  <span class="spinner">⏳</span>
                {:else}
                  <span class="refresh-icon">🔄</span>
                {/if}
              </button>
            </div>
            <div class="separator"></div>
          </div>

            <div class="accounts-list">
          <!-- Spotify -->
          <div class="account-item">
            <div class="account-details">
              <span class="platform-name onest-800">Spotify</span>
              <span class="platform-status onest-500 {user.spotify_name ? 'status-connected' : 'status-disconnected'}">
                {user.spotify_name ? `Connected (${user.spotify_name})` : 'Not connected'}
              </span>
            </div>
            {#if user.spotify_name}
              <Button
                onclick={() => disconnectService("spotify")}
                class="account-btn smaller-btn"
                bgColor="var(--auxie-soft-crimson-500)"
                shadowColor="var(--auxie-soft-crimson-700)"
              >
                Disconnect
              </Button>
            {:else}
              <Button
                onclick={loginWithSpotify}
                class="account-btn smaller-btn"
                bgColor="var(--auxie-intense-mint-500)"
                shadowColor="var(--auxie-intense-mint-700)"
              >
                Login
              </Button>
            {/if}
          </div>

          <!-- SoundCloud -->
          <div class="account-item">
            <div class="account-details">
              <span class="platform-name onest-800">SoundCloud</span>
              <span class="platform-status onest-500 {user.soundcloud_name ? 'status-connected' : 'status-disconnected'}">
                {user.soundcloud_name ? `Connected (${user.soundcloud_name})` : 'Not connected'}
              </span>
            </div>
            {#if user.soundcloud_name}
              <Button
                onclick={() => disconnectService("soundcloud")}
                class="account-btn smaller-btn"
                bgColor="var(--auxie-soft-crimson-500)"
                shadowColor="var(--auxie-soft-crimson-700)"
              >
                Disconnect
              </Button>
            {:else}
              <Button
                onclick={loginWithSoundCloud}
                class="account-btn smaller-btn"
                bgColor="var(--auxie-warm-orange-500)"
                shadowColor="var(--auxie-warm-orange-700)"
              >
                Login
              </Button>
            {/if}
          </div>

          <!-- Tidal -->
          <div class="account-item">
            <div class="account-details">
              <span class="platform-name onest-800">Tidal</span>
              <span class="platform-status onest-500 {user.tidal_name ? 'status-connected' : 'status-disconnected'}">
                {user.tidal_name ? `Connected (${user.tidal_name})` : 'Not connected'}
              </span>
            </div>
            {#if user.tidal_name}
              <Button
                onclick={() => disconnectService("tidal")}
                class="account-btn smaller-btn"
                bgColor="var(--auxie-soft-crimson-500)"
                shadowColor="var(--auxie-soft-crimson-700)"
              >
                Disconnect
              </Button>
            {:else}
              <Button
                onclick={loginWithTidal}
                class="account-btn smaller-btn"
                bgColor="var(--auxie-vivid-blue-500)"
                shadowColor="var(--auxie-vivid-blue-700)"
              >
                Login
              </Button>
            {/if}
          </div>
        </div>

        {#if !user.spotify_name && !user.soundcloud_name && !user.tidal_name}
          <p class="guest-info onest-500" style="margin-top: -6px; margin-bottom: 0;">
            You're in a guest session. Link a streaming service to search songs.
          </p>
        {/if}

        </div>
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

  .accounts_group {
    display: grid;
    gap: 10px;
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

  .button_group_row {
    width: 100%;
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 12px;
  }

  :global(.smaller-btn) {
    padding: 8px 12px !important;
    font-size: 14px !important;
  }

  .separator_row {
    font-family: "Onest";
    font-weight: bold;
    font-size: 14px;
    color: var(--auxie-cloud-white-200);
    display: flex;
    align-items: center;
    width: 100%;
    gap: 12px;
  }
  .header-with-action {
    display: flex;
    align-items: center;
    gap: 8px;
  }
  .header-title {
    white-space: nowrap;
  }
  .refresh-btn {
    background: transparent;
    border: none;
    cursor: pointer;
    font-size: 14px;
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 4px;
    border-radius: 50%;
    transition: background-color 0.2s, transform 0.2s;
  }
  .refresh-btn:hover:not(:disabled) {
    background-color: rgba(255, 255, 255, 0.1);
    transform: rotate(45deg);
  }
  .refresh-btn:disabled {
    cursor: not-allowed;
    opacity: 0.5;
  }
  .separator_row p {
    margin: 0;
    white-space: nowrap;
  }
  .separator {
    flex: 1;
    height: 3px;
    border-radius: 4px;
    background-color: var(--auxie-deep-navy-500);
  }
  .guest-info {
    text-align: center;
    color: var(--auxie-cloud-white-600);
    font-size: 11px;
    opacity: 0.65;
  }

  .accounts-list {
    display: grid;
    row-gap: 12px;
    width: 100%;
    background-color: var(--auxie-deep-navy-700);
    border: 1px solid var(--auxie-deep-navy-600);
    border-radius: 16px;
    padding: 14px;
    margin-top: 4px;
    margin-bottom: 12px;
  }

  .account-item {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding-bottom: 10px;
    border-bottom: 1px solid var(--auxie-deep-navy-600);
  }

  .account-item:last-child {
    padding-bottom: 0;
    border-bottom: none;
  }

  .account-details {
    display: flex;
    flex-direction: column;
    gap: 3px;
  }

  .platform-name {
    color: var(--auxie-cloud-white-50);
    font-size: 14px;
  }

  .platform-status {
    font-size: 11px;
  }

  .status-connected {
    color: var(--auxie-intense-mint-500);
  }

  .status-disconnected {
    color: var(--auxie-cloud-white-600);
    opacity: 0.6;
  }

  :global(.account-btn) {
    width: 72px !important;
    padding: 6px 0 !important;
    font-size: 12px !important;
    border-radius: 14px !important;
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
    justify-content: space-between;
    align-items: center;
    gap: 10px;
    h2 {
      text-align: left;
      text-overflow: ellipsis;
    }
    div {
      display: flex;
      flex-direction: column;
      align-items: center;
      gap: 10px;
      }
  }
  .logout-link {
    color: var(--auxie-soft-crimson-400);
    border: 1px solid rgba(244, 124, 124, 0.35);
    background-color: transparent;
    font-size: 12px;
    text-decoration: none;
    padding: 4px 10px;
    border-radius: 12px;
    transition: all 0.2s ease;
  }
  .logout-link:hover {
    color: var(--auxie-soft-crimson-500);
    border-color: var(--auxie-soft-crimson-500);
    background-color: rgba(239, 68, 68, 0.06);
  }
  .logout-link:active {
    transform: scale(0.97);
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
