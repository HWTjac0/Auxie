<script lang="ts">
  import TextInput from "./TextInput.svelte";

  let { searchQuery = $bindable("") }: { searchQuery?: string } = $props();

  let dialogElement: HTMLDialogElement;

  export function show() {
    dialogElement.showModal();
  }

  export function close() {
    dialogElement.close();
  }
</script>

<dialog bind:this={dialogElement} class="search-dialog">
  <div class="dialog-content">
    <div class="dialog-header">
      <h2 class="onest-500">Search for a song</h2>
      <button class="close-btn" onclick={close}>&times;</button>
    </div>
    
    <div class="search-section">
      <TextInput bind:value={searchQuery} placeholder="Type song name..." />
      <div class="search-results">
        <p class="empty-state">No matching songs found</p>
      </div>
    </div>
  </div>
</dialog>

<style>
  .search-dialog {
    padding: 0;
    border: none;
    border-radius: 20px;
    background-color: var(--auxie-deep-navy-700);
    color: var(--auxie-cloud-white-50);
    max-width: 500px;
    width: 90%;
    margin: auto;
    box-shadow: 0 10px 40px rgba(0, 0, 0, 0.5);
  }

  .search-dialog::backdrop {
    background: rgba(0, 0, 0, 0.6);
    backdrop-filter: blur(4px);
  }

  .dialog-content {
    display: flex;
    flex-direction: column;
    padding: 20px;
    gap: 20px;
  }

  .dialog-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }

  .dialog-header h2 {
    margin: 0;
    font-size: 20px;
  }

  .close-btn {
    background: transparent;
    border: none;
    color: var(--auxie-cloud-white-400);
    font-size: 28px;
    cursor: pointer;
    line-height: 1;
    padding: 0;
    display: flex;
    align-items: center;
    justify-content: center;
    transition: color 0.2s;
  }

  .close-btn:hover {
    color: var(--auxie-soft-crimson-400);
  }

  .search-section {
    display: flex;
    flex-direction: column;
    gap: 15px;
  }

  .search-results {
    background-color: var(--auxie-deep-navy-600);
    border-radius: 12px;
    padding: 15px;
    min-height: 250px;
    display: flex;
    align-items: center;
    justify-content: center;
    border: 1px solid var(--auxie-deep-navy-500);
  }

  .empty-state {
    color: var(--auxie-cloud-white-600);
    font-family: "Onest", sans-serif;
  }
</style>
