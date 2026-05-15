<script lang="ts">
    import type { PageProps } from "./$types";
let {data}: PageProps = $props();
</script>
<h2>Host party</h2>
<form method="POST"> 
  <input type="text" placeholder="Username" value={data.user.name}>
  <input type="text" placeholder="Name of the party" value={data.room.name}>
  <button>Create</button>
</form>
