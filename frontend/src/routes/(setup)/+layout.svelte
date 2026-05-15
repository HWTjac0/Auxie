<script lang="ts">
let { children } = $props();
</script>

<div class="wrapper">
  <div class="setup_container">
    {@render children()}
  </div>
</div>

<style>
.wrapper {
  background-color: var(--auxie-deep-navy-900);
  height: 100vh;
}
.setup_container {
  height: 100%;
  display: grid;
  justify-content: center;
  justify-items: center;
  align-content: center;
}

</style>
