<script lang="ts">
interface Props {
  username: string;
  size?: number;
  gradient?: string;
  src?: string;
}

let {
  username,
  size = 40,
  gradient = "linear-gradient(135deg, var(--auxie-electric-purple-500), var(--auxie-razzmatazz-500))",
  src = "",
}: Props = $props();

const initial = $derived(username ? username.charAt(0).toUpperCase() : "");
</script>

<div 
  class="user-avatar" 
  style:width="{size}px" 
  style:height="{size}px" 
  style:background={src ? 'none' : gradient}
  style:font-size="{size * 0.45}px"
>
  {#if src}
    <img 
      {src} 
      alt="{username}'s avatar" 
      class="avatar-img"
      style:width="{size}px" 
      style:height="{size}px" 
    />
  {:else}
    {initial}
  {/if}
</div>

<style>
  .user-avatar {
    border-radius: 50%;
    color: white;
    display: flex;
    align-items: center;
    justify-content: center;
    font-weight: bold;
    font-family: "Sora", sans-serif;
    user-select: none;
    overflow: hidden;
    flex-shrink: 0;
  }
  
  .avatar-img {
    object-fit: cover;
    border-radius: 50%;
    display: block;
  }
</style>
