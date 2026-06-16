<script lang="ts">
  let { size = 24, color = "currentColor" }: { size?: number, color?: string } = $props();
</script>

<svg
  width={size}
  height={size}
  viewBox="0 0 24 24"
  fill="none"
  xmlns="http://www.w3.org/2000/svg"
>
  <path
    d="M5.5 5.27V18.73C5.5 19.3 6.13 19.64 6.61 19.32L16.2 12.92C16.62 12.64 16.62 12.02 16.2 11.74L6.61 5.34C6.13 5.02 5.5 5.36 5.5 5.93V5.27Z"
    fill={color}
  />
  <rect x="16" y="5" width="2.5" height="14" rx="1" fill={color} />
</svg>
