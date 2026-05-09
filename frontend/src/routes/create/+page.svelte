<script lang="ts">
    import type { PageProps } from "./$types";


let {data}: PageProps = $props();
console.log(data.room.name);
</script>
<h2>Host party</h2>
<form method="POST"> 
  <input type="text" placeholder="Username" value={data.room.name}>
  <input type="text" placeholder="Name of the party">
  <button>Create</button>
</form>
