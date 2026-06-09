<script lang="ts">
import ArrowLeft from "../../../components/icons/ArrowLeft.svelte";
import EllipsisVert from "../../../components/icons/EllipsisVert.svelte";
import Invite from "../../../components/icons/Invite.svelte";
import InviteDialog from "../../../components/InviteDialog.svelte";
import type { PageProps } from "./$types";
let { data }: PageProps = $props();

let inviteDialog: any = $state(null);
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
      <button class="nav_button nav_invite" onclick={() => inviteDialog?.show()}>
        <Invite color="white"/>
      </button>
      <button class="nav_button nav_actions" popovertarget="actions_popover">
        <EllipsisVert color="white"/>
      </button>

      <div id="actions_popover" popover>
        <div class="popover_menu">
          <button class="popover_item" onclick={() => { document.getElementById('actions_popover')?.hidePopover(); inviteDialog?.show(); }}>
            Invite friends
          </button>
          <button class="popover_item" onclick={() => console.log('Settings')}>
            Room settings
          </button>
          <div class="popover_divider"></div>
          <a href="/" class="popover_item popover_leave">
            Leave room
          </a>
        </div>
      </div>

      </div>
    </div>
  </nav>
  <main>
    <h1>Room: {data.slug}</h1>
    <details>
      <summary>Users: {data.users ? data.users.length : 0}</summary>
      {#each data.users as user}
        <li>{user.Username} - {user.CurrentRole}</li>
      {/each}
    </details>
  </main>
  <InviteDialog bind:this={inviteDialog} slug={data.slug} />
</div>
<p>Welcome to the party! This is where the music happens.</p>

<style>
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
    justify-content:center;
    border-radius: 10px;
    corner-shape: squircle;
    background:none;
    padding: 5px;
    border: 2px solid var(--auxie-deep-navy-600);
    }
  .nav_invite {
    background-color: var(--auxie-intense-mint-500);
    border-color: var(--auxie-intense-mint-700);
    box-shadow: 
      inset 0 -2px 3px 0 color-mix(in srgb, var(--auxie-intense-mint-500), black 25%),
      inset 0 2px 4.5px 0 color-mix(in srgb, var(--auxie-intense-mint-500), white 25%),
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
</style>
