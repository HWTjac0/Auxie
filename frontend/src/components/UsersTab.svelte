<script lang="ts">
import type { User } from "../routes/room/[slug]/+page";
import UserAvatar from "./UserAvatar.svelte";

let {
  users = [],
  currentUser,
  slug,
}: { users: User[]; currentUser?: any; slug?: string } = $props();

let isHost = $derived(currentUser?.CurrentRole === "Host");

async function changeRole(username: string, newRole: string) {
  if (!slug || !isHost) return;
  try {
    const res = await fetch(`/api/v1/room/${slug}/user/${username}/role`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ role: newRole }),
    });
    if (!res.ok) console.error("Failed to change role");
  } catch (err) {
    console.error(err);
  }
}

async function kickUser(username: string) {
  if (!slug || !isHost) return;
  if (!confirm(`Are you sure you want to kick ${username}?`)) return;
  try {
    const res = await fetch(`/api/v1/room/${slug}/user/${username}`, {
      method: "DELETE",
    });
    if (!res.ok) console.error("Failed to kick user");
  } catch (err) {
    console.error(err);
  }
}
</script>

<div class="users-tab">
  <div class="users-header">
    <h2 class="onest-500">Users in room</h2>
    <span class="user-count onest-500">{users.length}</span>
  </div>
  
  <ul class="users-list">
    {#each users as user}
      <li class="user-item">
        <UserAvatar username={user.Username} src={user.AvatarUrl} />
        <div class="user-info">
          <span class="username onest-500">{user.Username} {currentUser?.Username === user.Username ? '(You)' : ''}</span>
          <span class="role onest-300">{user.CurrentRole}</span>
        </div>
        
          <div class="user-actions">
            <select 
              class="role-select" 
              value={user.CurrentRole} 
              onchange={(e) => changeRole(user.Username, e.currentTarget.value)}
              disabled={!isHost}
            >
              <option value="Guest">Guest</option>
              <option value="DJ">DJ</option>
              <option value="Host" disabled>Host</option>
            </select>
            {#if isHost}
              <button class="kick-btn" onclick={() => kickUser(user.Username)} title="Kick User">✕</button>
            {/if}
          </div>
      </li>
    {:else}
      <li class="empty-state onest-500">There are no users yet</li>
    {/each}
  </ul>
</div>

<style>
  .users-tab {
    display: flex;
    flex-direction: column;
    gap: 15px;
    padding: 10px 0;
  }

  .users-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 0 5px;
  }

  h2 {
    font-size: 18px;
    color: var(--auxie-cloud-white-50);
    margin: 0;
  }

  .user-count {
    background-color: var(--auxie-deep-navy-600);
    color: var(--auxie-cloud-white-200);
    padding: 4px 12px;
    border-radius: 20px;
    font-size: 14px;
    border: 1px solid var(--auxie-deep-navy-500);
  }

  .users-list {
    list-style: none;
    padding: 0;
    margin: 0;
    display: flex;
    flex-direction: column;
    gap: 10px;
  }

  .user-item {
    display: flex;
    align-items: center;
    gap: 15px;
    background-color: var(--auxie-deep-navy-700);
    border: 1px solid var(--auxie-deep-navy-600);
    padding: 12px 15px;
    border-radius: 16px;
    transition: all 0.2s ease;
  }

  .user-item:hover {
    background-color: var(--auxie-deep-navy-600);
    transform: translateY(-2px);
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.2);
  }

  .user-info {
    display: flex;
    flex-direction: column;
    gap: 4px;
    flex: 1;
  }

  .username {
    color: var(--auxie-cloud-white-50);
    font-size: 16px;
  }

  .role {
    color: var(--auxie-cloud-white-600);
    font-size: 12px;
    text-transform: capitalize;
  }

  .user-actions {
    display: flex;
    align-items: center;
    gap: 10px;
  }

  .role-select {
    background: var(--auxie-deep-navy-800);
    color: white;
    border: 1px solid var(--auxie-deep-navy-500);
    border-radius: 8px;
    padding: 4px 8px;
    font-family: inherit;
    font-size: 13px;
    cursor: pointer;
  }

  .role-select:disabled {
    opacity: 0.5;
    cursor: not-allowed;
    background: transparent;
    border-color: transparent;
    appearance: none;
    -webkit-appearance: none;
    -moz-appearance: none;
  }

  .kick-btn {
    background: rgba(231, 76, 60, 0.1);
    color: #e74c3c;
    border: none;
    border-radius: 50%;
    width: 28px;
    height: 28px;
    display: flex;
    align-items: center;
    justify-content: center;
    cursor: pointer;
    transition: all 0.2s;
  }

  .kick-btn:hover {
    background: rgba(231, 76, 60, 0.3);
    transform: scale(1.1);
  }

  .empty-state {
    color: var(--auxie-cloud-white-600);
    text-align: center;
    padding: 30px;
    background-color: var(--auxie-deep-navy-700);
    border-radius: 16px;
    border: 1px dashed var(--auxie-deep-navy-500);
  }
</style>
