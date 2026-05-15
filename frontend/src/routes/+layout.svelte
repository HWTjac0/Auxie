<script lang="ts">
import { onMount } from 'svelte';
import { goto } from '$app/navigation';
import { page } from '$app/state';
import favicon from '$lib/assets/favicon.svg';
import "../styles/global.css"

let { children } = $props();

onMount(async () => {
   const publicPaths = ['/welcome', '/create', '/host'];
   const isPublicPath = publicPaths.includes(page.url.pathname);
    try {
        const res = await fetch("http://127.0.0.1:8080/api/v1/auth/me", {
            credentials: "include"
        });
        if (!res.ok) {
          if(!isPublicPath)
            goto('/welcome');
        } else if (res.ok && page.url.pathname === '/welcome') {
            goto("/");
        }
    } catch (e) {
        goto('/welcome');
    }
});
</script>

<svelte:head>
	<link rel="icon" href={favicon} />
</svelte:head>

{@render children()}
