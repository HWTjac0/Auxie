<script lang="ts">
    import type { PageProps } from "./$types";
    let { data }: PageProps = $props();
</script>

<h1>Room: {data.slug}</h1>
<p>Welcome to the party! This is where the music happens.</p>
