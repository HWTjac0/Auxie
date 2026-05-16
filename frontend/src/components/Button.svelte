<script lang="ts">
  import type { Snippet } from 'svelte';

  interface Props {
    href?: string;
    onclick?: () => void;
    bgColor?: string;
    shadowColor?: string;
    innerShadow?: boolean;
    outerShadow?: boolean;
    class?: string;
    children: Snippet;
  }

  let {
    href,
    onclick,
    bgColor = 'var(--auxie-razzmatazz-500)',
    shadowColor = 'var(--auxie-razzmatazz-700)',
    innerShadow = true,
    outerShadow = true,
    class: className = '',
    children
  }: Props = $props();

  const isLink = !!href;
</script>

{#if isLink}
  <a 
    {href} 
    class="btn {className}" 
    style:--bg={bgColor} 
    style:--shadow={shadowColor}
    class:has-inner={innerShadow}
    class:has-outer={outerShadow}
  >
    {@render children()}
  </a>
{:else}
  <button 
    {onclick} 
    class="btn {className}" 
    style:--bg={bgColor} 
    style:--shadow={shadowColor}
    class:has-inner={innerShadow}
    class:has-outer={outerShadow}
  >
    {@render children()}
  </button>
{/if}

<style>
  .btn {
    width: 100%;
    font-family: "Onest", sans-serif;
    border-radius: 25px;
    background-color: var(--bg);
    text-align: center;
    font-size: 16px;
    font-weight: bold;
    color: var(--auxie-cloud-white-50);
    text-decoration: none;
    padding: 11px;
    border: 2px solid var(--shadow);
    cursor: pointer;
    display: inline-block;
    transition: transform 0.1s ease, box-shadow 0.2s ease;
    border-style: solid;
    corner-shape: squircle;
  }

  .btn:active {
    transform: scale(0.98);
  }

  .has-inner {
    box-shadow: 
      inset 0 -4px 6.1px 0 color-mix(in srgb, var(--bg), black 40%),
      inset 0 4px 9.1px 0 color-mix(in srgb, var(--bg), white 40%);
  }

  .has-outer {
    box-shadow: 0 0 18.5px -3px var(--shadow);
  }

  .has-inner.has-outer {
    box-shadow: 
      inset 0 -4px 6.1px 0 color-mix(in srgb, var(--bg), black 40%),
      inset 0 4px 9.1px 0 color-mix(in srgb, var(--bg), white 40%),
      0 0 18.5px -3px var(--shadow);
  }
</style>
