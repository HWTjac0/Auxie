<script lang="ts">
    import { goto } from "$app/navigation";
    import type { PageProps } from "./$types";
    let { data }: PageProps = $props();

    let username = $derived(data.user.name);
    let roomName = $derived(data.room.name);

    async function handleCreate(e: Event) {
        e.preventDefault();

        try {
            const response = await fetch("/api/v1/room/create", {
                method: "POST",
                headers: {
                    "Content-Type": "application/json"
                },
                body: JSON.stringify({
                    room_name: roomName,
                    username: username
                })
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

<h2>Host party</h2>
<form onsubmit={handleCreate}> 
  <input type="text" placeholder="Username" bind:value={username}>
  <input type="text" placeholder="Name of the party" bind:value={roomName}>
  <button type="submit">Create</button>
</form>
