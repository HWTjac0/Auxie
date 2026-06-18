<script lang="ts">
import LeaveDialog from "../../../components/LeaveDialog.svelte";
import SummaryDialog from "../../../components/SummaryDialog.svelte";
import ArrowLeft from "../../../components/icons/ArrowLeft.svelte";
import EllipsisVert from "../../../components/icons/EllipsisVert.svelte";
import Invite from "../../../components/icons/Invite.svelte";
import InviteDialog from "../../../components/InviteDialog.svelte";
import type { PageProps } from "./$types";
import SettingsPopover from "../../../components/SettingsPopover.svelte";
import UsersTab from "../../../components/UsersTab.svelte";
import QueueTab from "../../../components/QueueTab.svelte";
import NowPlaying from "../../../components/NowPlaying.svelte";
import SearchDialog from "../../../components/SearchDialog.svelte";
import Plus from "../../../components/icons/Plus.svelte";
import List from "../../../components/icons/List.svelte";
import Users from "../../../components/icons/Users.svelte";
import type { Component } from "svelte";
import { onMount } from "svelte";
import { toasts } from "../../../lib/toasts.svelte";

let { data }: PageProps = $props();

let inviteDialog: any = $state(null);
let searchDialog: any = $state(null);
// ref to QueueTab to forward WS messages (likes, skip votes)
let queueTabRef: any = $state(null);
let searchQuery: string = $state("");
let leaveDialog: any = $state(null);
let summaryDialog: any = $state(null);
let roomHistory = $state<any[]>([]);
let activeUsers = $state<any[]>(data.users || []);
let queue = $state<any[]>(data.queue || []);
let proposedQueue = $state<any[]>(data.proposedQueue || []);

let currentUserId = $state(data.currentUserId || 0);
let ws = $state<WebSocket | null>(null);

let currentUser = $derived(activeUsers.find(u => u.ID === currentUserId || u.id === currentUserId));
let isHost = $derived(currentUser?.CurrentRole === "Host");

async function fetchRoomData() {
  const res = await fetch(`/api/v1/room/${data.slug}`);
  if (res.ok) {
    const json = await res.json();
    queue = json.queue || [];
    proposedQueue = json.proposedQueue || [];
    currentUserId = json.current_user_id || 0;
    if (json.users) {
      activeUsers = json.users.map((u: any) => ({
        ID: u.ID || u.id || 0,
        Username: u.Username || u.username || "",
        CurrentRole: u.CurrentRole || u.current_role || "Guest",
        AvatarUrl: u.AvatarUrl || u.avatar_url || ""
      }));
    }
  }
}

type Tab = { label: string; icon: Component<any> };
let tabs: Array<Tab> = [
  { label: "Queue", icon: List },
  { label: "Users", icon: Users },
];
let activeTabIdx = $state(0);

function initiateLeave() {
  leaveDialog?.show();
}

async function loadHistoryAndShowSummary() {
  try {
    const res = await fetch(`/api/v1/room/${data.slug}/history`);
    
    if (res.ok) { 
      roomHistory = await res.json(); 
    } else {
      console.error("Something went wrong. :< ");
    }
    summaryDialog?.show();
  } catch(e) {
    console.error("History fetch error:", e);
  }
}

async function confirmLeaveRoom() {
  if (isHost) {
    const res = await fetch(`/api/v1/room/${data.slug}`, { method: 'DELETE' });
    if (res.ok) {
      toasts.add("Zamknąłeś pokój.", "info");
    } else {
      toasts.add("Nie udało się usunąć pokoju.", "error");
    }
  } else {
    // GOŚĆ: Wychodzi i jako jedyny ręcznie pobiera swoją historię.
    await fetch(`/api/v1/room/${data.slug}/leave`, { method: 'POST' });
    await loadHistoryAndShowSummary();
  }
}

onMount(() => {
  fetchRoomData();
  const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
  const wsUrl = `${protocol}//${window.location.hostname}:8080/api/v1/room/${data.slug}/ws`;
  const socket = new WebSocket(wsUrl);
  ws = socket;

  socket.onmessage = (event) => {
    try {
      const msg = JSON.parse(event.data);

      if (msg.type === "USER_JOINED") {
        const joinedUser = msg.payload;
        if (!activeUsers.some((u) => u.Username === joinedUser.username)) {
          activeUsers = [...activeUsers, {
            ID: joinedUser.id || joinedUser.user_id || 0,
            Username: joinedUser.username,
            CurrentRole: joinedUser.role || "Guest",
            AvatarUrl: "",
          }];
        }
        toasts.add(`${joinedUser.username} joined the room`, "joined");
      } else if (msg.type === "USER_LEFT") {
        activeUsers = activeUsers.filter((u) => u.Username !== msg.payload.username);
        toasts.add(`${msg.payload.username} left the room`, "left");
      } else if (msg.type === "USER_ROLE_CHANGED") {
        const { username, role } = msg.payload;
        activeUsers = activeUsers.map((u) => u.Username === username ? { ...u, CurrentRole: role } : u);
        toasts.add(`${username} is now ${role}`, "joined");
      } else if (msg.type === "USER_KICKED") {
        const { username } = msg.payload;
        if (currentUser && currentUser.Username === username) {
          toasts.add(`You were kicked from the room`, "left");
          window.location.href = "/";
        } else {
          activeUsers = activeUsers.filter((u) => u.Username !== username);
          toasts.add(`${username} was kicked`, "left");
        }
      } else if (msg.type === "TRACK_ADDED") {
        toasts.add(`"${msg.payload.title}" by ${msg.payload.artist} was added to the queue`, "track");
        fetchRoomData();
      } else if (msg.type === "TRACK_PROPOSED") {
        toasts.add(`"${msg.payload.title}" by ${msg.payload.artist} was proposed`, "track");
        fetchRoomData();
      } else if (msg.type === "TRACK_APPROVED") {
        toasts.add(`A proposed track was approved`, "track");
        fetchRoomData();
      } else if (msg.type === "TRACK_REJECTED") {
        toasts.add(`A proposed track was rejected`, "track");
        fetchRoomData();
      } else if (msg.type === "TRACK_SKIPPED") {
        const p = msg.payload;
        if (p?.by_vote) {
          toasts.add(`"${p.title}" was skipped by vote`, "track");
        } else if (p?.title) {
          toasts.add(`"${p.title}" by ${p.artist} was skipped`, "track");
        }
        fetchRoomData();
      } else if (msg.type === "playback:start") {
        fetchRoomData();
      } else if (msg.type === "playback:queue_empty") {
        queue = [];
        toasts.add("Queue is now empty", "track");
      } else if (msg.type === "TRACK_LIKED" || msg.type === "SKIP_VOTE") {
        // Forward to QueueTab for optimistic counter updates
        queueTabRef?.onWsMessage(msg);
      }else if (msg.type === "ROOM_CLOSED") {
        toasts.add("Host has ended the sesion.", "left");
        if (msg.payload && msg.payload.history) {
          roomHistory = msg.payload.history;
        } else {
          roomHistory = [];
        }
        summaryDialog?.show();
      }
    } catch (e) {
      console.error("Failed to parse WS message:", e);
    }
  };

  return () => { socket.close(); };
});
</script>

<div class="room_wrapper">
  <div class="background_gradient background_gradient_top"></div>
  <div class="background_gradient background_gradient_bottom"></div>
  <div class="content_wrapper">
    <nav>
      <div class="nav_container">
        <div class="back-link">
          <a href="/"><ArrowLeft color="white" /></a>
        </div>
        <h1 class="room_name onest-600">{data?.room?.Name}</h1>
        <div class="room_actions">
          <button class="nav_button nav_invite" onclick={() => inviteDialog?.show()}>
            <Invite color="var(--auxie-deep-navy-900)" />
          </button>
          <button class="nav_button nav_actions" id="nav_actions" popovertarget="actions_popover">
            <EllipsisVert color="white" />
          </button>
          <SettingsPopover 
  onInvite={() => inviteDialog?.show()} 
  onLeave={initiateLeave} 
/>
        </div>
      </div>
    </nav>

    <main>
      <div class="main_content">
        <div class="tabs-container">
          <div class="tabs-header">
            {#each tabs as tab, tabIdx}
              {@const TabIcon = tab.icon}
              <button
                class="tab-button onest-500 {tabIdx === activeTabIdx ? 'active' : ''}"
                onclick={() => activeTabIdx = tabIdx}
              >
                <TabIcon color={tabIdx === activeTabIdx ? "white" : "currentColor"} />
                {tab.label}
              </button>
            {/each}
          </div>

          <div class="tab-content">
            <NowPlaying queue={queue} currentUser={currentUser} slug={data.slug} {ws} />
            {#if activeTabIdx === 0}
              <QueueTab
                bind:this={queueTabRef}
                queue={queue}
                proposedQueue={proposedQueue}
                currentUser={currentUser}
                slug={data.slug}
              />
            {:else if activeTabIdx === 1}
              <UsersTab users={activeUsers} currentUser={currentUser} slug={data.slug} />
            {/if}
          </div>
        </div>
      </div>
    </main>

    <InviteDialog bind:this={inviteDialog} slug={data.slug} />
    <SearchDialog bind:this={searchDialog} bind:searchQuery={searchQuery} slug={data.slug} />
    <LeaveDialog bind:this={leaveDialog} {isHost} onConfirm={confirmLeaveRoom} />
    <SummaryDialog bind:this={summaryDialog} history={roomHistory} />
    <button class="fab-add-song onest-500" onclick={() => searchDialog?.show()}>
      <Plus /> Add Song
    </button>
  </div>
</div>

<style>
  #nav_actions { anchor-name: --actions-anchor; }
  .room_wrapper {
    background-color: var(--auxie-deep-navy-900);
    min-height: 100vh;
    position: relative;
    overflow: hidden;
    display: flex;
    flex-direction: column;
  }
  .content_wrapper {
    position: relative;
    z-index: 10;
    flex: 1;
    display: flex;
    flex-direction: column;
  }
  .background_gradient {
    position: absolute;
    z-index: 0;
    width: 60%;
    height: 60%;
    filter: blur(150px);
    border-radius: 100%;
    opacity: 0.15;
    pointer-events: none;
  }
  .background_gradient_top {
    top: -20%;
    left: -20%;
    background: linear-gradient(to top right, var(--auxie-electric-purple-600), var(--auxie-intense-mint-500));
    animation: float-top 20s ease-in-out infinite alternate;
  }
  .background_gradient_bottom {
    bottom: -20%;
    right: -20%;
    background: linear-gradient(to top right, var(--auxie-razzmatazz-600), var(--auxie-vivid-blue-500));
    animation: float-bottom 25s ease-in-out infinite alternate;
  }
  @keyframes float-top {
    0% { transform: translate(0,0) scale(1) rotate(0deg); filter: blur(150px) hue-rotate(0deg); }
    50% { transform: translate(5%,5%) scale(1.1) rotate(15deg); filter: blur(150px) hue-rotate(30deg); }
    100% { transform: translate(-5%,5%) scale(0.9) rotate(-10deg); filter: blur(150px) hue-rotate(-15deg); }
  }
  @keyframes float-bottom {
    0% { transform: translate(0,0) scale(1) rotate(0deg); filter: blur(150px) hue-rotate(0deg); }
    50% { transform: translate(-5%,-5%) scale(1.1) rotate(-20deg); filter: blur(150px) hue-rotate(-30deg); }
    100% { transform: translate(5%,-5%) scale(0.95) rotate(10deg); filter: blur(150px) hue-rotate(15deg); }
  }
  nav {
    padding: 5px;
    display: flex;
    justify-content: center;
  }
  nav .nav_container {
    display: flex;
    justify-content: space-around;
    gap: 20px;
    padding: 15px;
    align-items: center;
    background-color: var(--auxie-deep-navy-700);
    border-radius: 20px;
    border: 2px solid var(--auxie-deep-navy-600);
  }
  nav .room_name { font-size: 20px; }
  nav .room_actions {
    display: flex;
    align-items: center;
    anchor-scope: --actions-anchor;
    gap: 10px;
  }
  nav .nav_button {
    display: flex;
    align-items: center;
    justify-content: center;
    border-radius: 10px;
    background: none;
    padding: 5px;
    border: 2px solid var(--auxie-deep-navy-600);
  }
  nav .nav_invite {
    background-color: var(--auxie-intense-mint-500);
    border-color: var(--auxie-intense-mint-700);
    box-shadow:
      inset 0 -2px 3px 0 color-mix(in srgb, var(--auxie-intense-mint-500), black 25%),
      inset 0 2px 4.5px 0 color-mix(in srgb, var(--auxie-intense-mint-500), white 25%),
      0 0 10px -2px var(--auxie-intense-mint-700);
  }
  nav .nav_actions { background-color: var(--auxie-deep-navy-500); }
  .back-link { display: inline-flex; align-items: center; justify-content: center; }
  main { padding: 15px; }
  .main_content {
    margin: 0 auto;
    display: grid;
    width: 100%;
    max-width: 600px;
  }
  .tabs-container { display: flex; flex-direction: column; gap: 20px; margin-top: 10px; }
  .tabs-header {
    display: flex;
    gap: 8px;
    background-color: var(--auxie-deep-navy-700);
    padding: 8px;
    border-radius: 16px;
    border: 1px solid var(--auxie-deep-navy-600);
  }
  .tab-button {
    flex: 1;
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 8px;
    background: transparent;
    border: none;
    padding: 12px 0;
    border-radius: 12px;
    color: var(--auxie-cloud-white-600);
    font-size: 15px;
    cursor: pointer;
    transition: all 0.3s ease;
  }
  .tab-button:hover:not(.active) {
    color: var(--auxie-cloud-white-300);
    background-color: rgba(255, 255, 255, 0.05);
  }
  .tab-button.active {
    background-color: var(--auxie-deep-navy-500);
    color: var(--auxie-cloud-white-50);
    box-shadow: 0 4px 10px rgba(0, 0, 0, 0.2);
  }
  .tab-content { animation: fadeIn 0.3s ease; }
  @keyframes fadeIn {
    from { opacity: 0; transform: translateY(5px); }
    to { opacity: 1; transform: translateY(0); }
  }
  .fab-add-song {
    position: fixed;
    bottom: 30px;
    left: 50%;
    display: flex;
    justify-content: center;
    align-items: center;
    gap: 5px;
    transform: translateX(-50%);
    background: linear-gradient(135deg, var(--auxie-intense-mint-500), var(--auxie-vivid-blue-500));
    color: var(--auxie-deep-navy-900);
    border: none;
    padding: 14px 28px;
    border-radius: 30px;
    font-size: 16px;
    font-weight: 600;
    box-shadow: 0 8px 20px rgba(0, 0, 0, 0.4);
    cursor: pointer;
    transition: all 0.3s ease;
    z-index: 100;
  }
  .fab-add-song:hover {
    transform: translateX(-50%) translateY(-3px);
    box-shadow: 0 12px 25px rgba(0, 0, 0, 0.5);
  }
</style>
