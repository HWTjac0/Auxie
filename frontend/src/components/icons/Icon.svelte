<script lang="ts">
  import type { Snippet } from "svelte";

  interface Props {
    size?: number | string;
    color?: string;
    class?: string;
    children?: Snippet;
    [key: string]: any;
  }

  let {
    size = 24,
    color = "currentColor",
    class: className = "",
    children,
    ...restProps
  }: Props = $props();

  let cssSize = $derived(typeof size === "number" ? `${size}px` : size);
</script>

<span
  class="inline-icon {className}"
  style:width={cssSize}
  style:height={cssSize}
  style:color={color}
  {...restProps}
>
  {@render children?.()}
</span>

<style>
  .inline-icon {
    display: inline-flex;
    align-items: center;
    justify-content: center;
  }
  .inline-icon :global(svg) {
    width: 100%;
    height: 100%;
  }
</style>
