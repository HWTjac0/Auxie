<script lang="ts">
import type { Snippet } from "svelte";

interface Props {
  size?: number | string;
  strokeWidth?: number | string;
  color?: string;
  class?: string;
  children?: Snippet;
  [key: string]: any;
}

let {
  size = 24,
  strokeWidth = 1.5,
  color = "currentColor",
  class: className = "",
  children,
  ...restProps
}: Props = $props();
</script>

<svg
  xmlns="http://www.w3.org/2000/svg"
  fill="none"
  viewBox="0 0 24 24"
  stroke={color}
  stroke-width={strokeWidth}
  width={size}
  height={size}
  class={className}
  {...restProps}
>
  {@render children?.()}
</svg>
