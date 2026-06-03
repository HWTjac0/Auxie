<script lang="ts">
let { children } = $props();
</script>

<div class="wrapper">
{@render children()}
</div>

<style>
.wrapper {
  font-family: Sora;
  background-color: var(--auxie-deep-navy-800);
  color: var(--auxie-cloud-white-200);
  height: 100vh;
  position: relative;
  display: grid;
}
</style>
