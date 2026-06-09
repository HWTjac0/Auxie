<script lang="ts">
import ArrowLeft from "../../../components/icons/ArrowLeft.svelte";
import EllipsisVert from "../../../components/icons/EllipsisVert.svelte";
import Invite from "../../../components/icons/Invite.svelte";
import InviteDialog from "../../../components/InviteDialog.svelte";
import type { PageProps } from "./$types";
import TextInput from "../../../components/TextInput.svelte";
import SettingsPopover from "../../../components/SettingsPopover.svelte";
import UsersTab from "../../../components/UsersTab.svelte";
import QueueTab from "../../../components/QueueTab.svelte";

let { data }: PageProps = $props();

let inviteDialog: any = $state(null);
let searchQuery: string = $state("");

type Tab = {
  id: number;
  label: string;
};

let tabs = ["Queue", "Users"];
let activeTabIdx = $state(0);
</script>

<div>
  <nav>
    <div class="nav_container">
      <a href="/" class="back-link">
        <ArrowLeft color="white" />
      </a>
      <h1 class="room_name">
      {data?.room?.Name ?? "Name"}
      </h1>
      <div class="room_actions">
        <button
          class="nav_button nav_invite"
          onclick={() => inviteDialog?.show()}
        >
          <Invite color="white" />
        </button>
        <button class="nav_button nav_actions" popovertarget="actions_popover">
          <EllipsisVert color="white" />
        </button>

        <SettingsPopover onInvite={() => inviteDialog?.show()} />
      </div>
    </div>
  </nav>
  <main>
    <div class="tabs-container">
      <div class="tabs-header">
        {#each tabs as tabLabel, tabIdx}
          <button 
            class="tab-button onest-500 {tabIdx === activeTabIdx ? 'active' : ''}" 
            onclick={() => activeTabIdx = tabIdx}
          >
            {tabLabel}
          </button>
        {/each}
      </div>
      
      <div class="tab-content">
        {#if activeTabIdx === 0}
          <QueueTab />
        {:else if activeTabIdx === 1}
          <UsersTab users={data.users} />
        {/if}
      </div>
    </div>
  </main>
  <InviteDialog bind:this={inviteDialog} slug={data.slug} />
</div>
<div class="search-section">
  <TextInput bind:value={searchQuery} placeholder="Wpisz nazwę utworu..." />
  <div class="search-results">
    <p class="empty-state">Brak dopasowanych utworów</p>
  </div>
</div>

<style>
  .search-section {
    margin: 20px 0;
    display: flex;
    flex-direction: column;
    gap: 10px;
  }
  .search-results {
    background-color: var(--auxie-deep-navy-700);
    border-radius: 10px;
    padding: 15px;
    min-height: 100px;
    display: flex;
    align-items: center;
    justify-content: center;
  }
  .empty-state {
    color: var(--auxie-cloud-white-600);
    font-family: "Onest", sans-serif;
  }
  nav {
    padding: 5px;
    display: flex;
    justify-content: center;
    .nav_container {
      display: flex;
      justify-content: space-around;
      gap: 20px;
      padding: 15px;
      align-items: center;
      background-color: var(--auxie-deep-navy-700);
      border-radius: 20px;
      corner-shape: squircle;
      border: 2px solid var(--auxie-deep-navy-600);
    }
    .room_name {
      font-size: 20px;
    }
    .room_actions {
      display: flex;
      align-items: center;
      anchor-scope: --actions-anchor;
      gap: 10px;
    }
    .nav_button {
      display: flex;
      align-items: center;
      justify-content: center;
      border-radius: 10px;
      corner-shape: squircle;
      background: none;
      padding: 5px;
      border: 2px solid var(--auxie-deep-navy-600);
    }
    .nav_invite {
      background-color: var(--auxie-intense-mint-500);
      border-color: var(--auxie-intense-mint-700);
      box-shadow:
        inset 0 -2px 3px 0 color-mix(in srgb, var(--auxie-intense-mint-500), black
              25%),
        inset 0 2px 4.5px 0
          color-mix(in srgb, var(--auxie-intense-mint-500), white 25%),
        0 0 10px -2px var(--auxie-intense-mint-700);
    }
    .nav_actions {
      background-color: var(--auxie-deep-navy-500);
    }
  }
  .back-link {
    display: inline-flex;
    align-items: center;
    justify-content: center;
  }
  main {
    padding: 15px;
  }
  .tabs-container {
    display: flex;
    flex-direction: column;
    gap: 20px;
    margin-top: 10px;
  }
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
  .tab-content {
    animation: fadeIn 0.3s ease;
  }
  @keyframes fadeIn {
    from { opacity: 0; transform: translateY(5px); }
    to { opacity: 1; transform: translateY(0); }
  }
</style>
