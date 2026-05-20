<script lang="ts">
import { goto } from "$app/navigation";
import type { PageProps } from "./$types";
import Button from "../../../components/Button.svelte";
import TextInput from "../../../components/TextInput.svelte";
let { data }: PageProps = $props();

let username = $derived(data.user.name);
let roomName = $derived(data.room.name);

async function handleCreate(e: Event) {
  e.preventDefault();

  try {
    const response = await fetch("/api/v1/room/create", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({
        room_name: roomName,
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

<h2 class="sora-800">Host party</h2>
<form onsubmit={handleCreate}> 
  <TextInput placeholder="Username" bind:value={username} />
  <TextInput placeholder="Name of the party" bind:value={roomName} />
  <Button 
    bgColor="var(--auxie-warm-orange-500)" 
    shadowColor="var(--auxie-warm-orange-700)" 
    fontSize="14px"
  >
    Create
  </Button>
</form>

<style>
  h2 {
    font-size: 32px;
    color: var(--auxie-cloud-white-50);
    text-align: center;
    margin-bottom: 20px;
  }
</style>
