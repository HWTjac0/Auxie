<script lang="ts">
    import { goto } from "$app/navigation";
  import { onMount } from "svelte"
  type User = {
    name: string,
    id: string,
    image: string
  }
  let user: User | null = $state(null)
  let loading = $state(true)

  onMount(async () => {
    try {
      const res = await fetch("/api/v1/auth/me", {
        credentials: "include"
      });

      if(res.ok) {
          user = await res.json();
      }
    } catch (e) { 
      goto("/welcome")
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
        <h1>Witaj, {user?.name}</h1>
    </div>
{/if}
