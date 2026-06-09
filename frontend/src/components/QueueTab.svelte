<script lang="ts">
import MusicalNote from "./icons/MusicalNote.svelte";

// Tutaj w przyszłości można przekazać listę utworów w kolejce
</script>

<div class="queue-tab">
  <div class="queue-header">
    <h2 class="onest-500">Queue</h2>
  </div>
  
  <div class="empty-state">
    <div class="empty-icon">
    <MusicalNote size={40}/>
    </div>
    <h3 class="onest-500">Queue is empty</h3>
    <p class="onest-300">Search for desired track to be played below and add it to the queue!</p>
  </div>
</div>

<style>
  .queue-tab {
    display: flex;
    flex-direction: column;
    gap: 15px;
    padding: 10px 0;
  }

  .queue-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 0 5px;
  }

  h2 {
    font-size: 18px;
    color: var(--auxie-cloud-white-50);
    margin: 0;
  }

  .empty-state {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 10px;
    text-align: center;
    padding: 40px 20px;
    background-color: var(--auxie-deep-navy-700);
    border: 1px dashed var(--auxie-deep-navy-500);
    border-radius: 16px;
  }

  .empty-icon {
    margin-bottom: 10px;
    opacity: 0.5;
    filter: drop-shadow(0 0 10px rgba(255, 255, 255, 0.2));
  }

  .empty-state h3 {
    color: var(--auxie-cloud-white-200);
    margin: 0;
    font-size: 18px;
  }

  .empty-state p {
    color: var(--auxie-cloud-white-600);
    margin: 0;
    font-size: 14px;
    max-width: 250px;
    line-height: 1.4;
  }
</style>
