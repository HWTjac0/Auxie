<script lang="ts">
  import TextInput from "../../../components/TextInput.svelte";
  import Button from "../../../components/Button.svelte";
  import CodeInput from "../../../components/CodeInput.svelte";
  import { goto } from "$app/navigation";
  import { onMount } from "svelte";

  let username = $state("");
  let joinCode = $state("");
  let isLoggedIn = $state(false);

  onMount(async () => {
    try {
      const res = await fetch("/api/v1/auth/me");
      if (res.ok) {
        const user = await res.json();
        username = user.spotify_name || user.name;
        isLoggedIn = true;
      }
    } catch (e) {
      console.error("Failed to check auth", e);
    }
  });

  async function handleJoin(e: Event) {
    e.preventDefault();

    try {
      const response = await fetch("/api/v1/room/join", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          join_code: joinCode,
          username: username,
        }),
      });

      if (response.ok) {
        const result = await response.json();
        goto(`/room/${result.room.Slug}`);
      } else {
        const error = await response.json();
        alert(`Error: ${error.error}`);
      }
    } catch (err) {
      console.error(err);
      alert("Failed to connect to the server");
    }
  }
</script>

<h2 class="sora-800">Join room</h2>
<form onsubmit={handleJoin}>
  {#if !isLoggedIn}
    <TextInput bind:value={username} name="username" placeholder="Username" />
  {:else}
    <div class="logged-in-label onest-500">
      Joining as: <span class="username">{username}</span>
    </div>
  {/if}
  <CodeInput bind:value={joinCode} name="joinCode" />
  <Button
    bgColor="var(--auxie-warm-orange-500)"
    shadowColor="var(--auxie-warm-orange-700)"
    fontSize="14px"
  >
    Enter
  </Button>
</form>

<style>
  h2 {
    font-size: 32px;
    color: var(--auxie-cloud-white-50);
    text-align: center;
  }
  .logged-in-label {
    background-color: var(--auxie-deep-navy-700);
    border: 2px dashed var(--auxie-deep-navy-500);
    border-radius: 20px;
    padding: 11px 16px;
    color: var(--auxie-cloud-white-600);
    font-size: 15px;
    text-align: center;
  }
  .logged-in-label .username {
    color: var(--auxie-cloud-white-100);
    font-weight: bold;
  }
</style>
