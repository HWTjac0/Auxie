<script lang="ts">
  import TextInput from "../../../components/TextInput.svelte";
  import Button from "../../../components/Button.svelte";
  import CodeInput from "../../../components/CodeInput.svelte";
  import { goto } from "$app/navigation";
  let username = $state("");
  let joinCode = $state("");

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
  <TextInput bind:value={username} name="username" placeholder="Username" />
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
</style>
