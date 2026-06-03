<script lang="ts">
import ArrowLeft from "../../../components/icons/ArrowLeft.svelte";
import type { PageProps } from "./$types";
let { data }: PageProps = $props();
</script>

<div>
  <nav>
    <a href="/">
      <ArrowLeft color="white" />
    </a>
    <h1>Room: {data.slug}</h1>
    <details>
      <summary>Users: {data.users ? data.users.length : 0}</summary>
      {#each data.users as user}
        <li>{user.Username} - {user.CurrentRole}</li>
      {/each}
    </details>
  </nav>
  <main></main>
</div>
<p>Welcome to the party! This is where the music happens.</p>

<style>
</style>
