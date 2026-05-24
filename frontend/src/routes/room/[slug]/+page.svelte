<script lang="ts">
import type { PageProps } from "./$types";
let { data }: PageProps = $props();
</script>

<a href="/">Go to profile</a>
<h1>Room: {data.slug}</h1>
<details>
  <summary>Users: {data.users ? data.users.length : 0}</summary>
  {#each data.users as user}
    <li>{user.Username} - {user.CurrentRole}</li>
  {/each}
</details>
<p>Welcome to the party! This is where the music happens.</p>
