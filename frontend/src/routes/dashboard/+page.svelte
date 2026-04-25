<script lang="ts">
  import { onMount } from "svelte"
  let user = $state(null)
  let loading = $state(true)

  onMount(async () => {
    try {
      const res = await fetch("http://127.0.0.1:8080/api/v1/auth/me", {
        credentials: "include"
      });

      if(res.ok) {
          user = await res.json();
      }
    } catch (e) {
      console.error(`Session error: ${e}`);
    } finally {
      loading = false;
    }
  })
</script>
{#if loading}
    <p>Sprawdzanie sesji...</p>
{:else if user}
    <div class="user-info">
        <img src={user.image} alt="Avatar" />
        <h1>Witaj, {user.name}</h1>
    </div>
{:else}
    <button on:click={loginWithSpotify}>Zaloguj przez Spotify</button>
{/if}
