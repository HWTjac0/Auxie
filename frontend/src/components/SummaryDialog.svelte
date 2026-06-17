<script lang="ts">
  import { onMount } from "svelte";
  import { goto } from "$app/navigation";

  interface TrackHistory {
    title: string;
    //duration: string; // np. "03:45"
    startTime: string | null; // np. "22:15" lub null jeśli pominięto
    skipped: boolean;
  }

  interface Props {
    history: TrackHistory[];
  }

  let { history }: Props = $props();
  let dialog: HTMLDialogElement;
  let copySuccess = $state(false);
  let randomTitle = $state("");

  const titles = [
    "Party Hits: Best Songs of the Night",
    "Top Tracks We Loved at the Party",
    "Favorite Bangers from Tonight",
    "Best & Most Played Songs of the Party",
    "Vibe Check: Our Top Party Anthems",
    "Epic Night – Best Songs Recap"
  ];

  export function show() {
    randomTitle = titles[Math.floor(Math.random() * titles.length)];
    dialog?.showModal();
  }

  function handleClose() {
    dialog?.close();
    window.location.href = "/";
  }

  async function copyToClipboard() {
    const textToCopy = history.map(track => {
      const timeInfo = track.skipped ? "utwór pominięty" : track.startTime;
      return `- ${track.title} - ${timeInfo}`;
    }).join("\n");

    try {
      await navigator.clipboard.writeText(`${randomTitle}\n\n${textToCopy}`);
      copySuccess = true;
      setTimeout(() => copySuccess = false, 2000);
    } catch (err) {
      console.error("Failed to copy: ", err);
    }
  }
</script>

<dialog bind:this={dialog}>
  <h3 class="dialog-title">{randomTitle}</h3>
  
  <div class="history-list">
    {#each history as track}
      <div class="track-item">
        <div class="track-info">
          <span class="track-title">{track.title}</span>
        </div>
        <div class="track-status {track.skipped ? 'skipped' : ''}">
          {track.skipped ? "pominięty" : track.startTime}
        </div>
      </div>
    {/each}
    {#if history.length === 0}
      <p class="empty-msg">Brak utworów do wyświetlenia.</p>
    {/if}
  </div>

  <div class="actions">
    <button class="btn-copy" onclick={copyToClipboard}>
      {copySuccess ? "Skopiowano!" : "Skopiuj listę"}
    </button>
    <button class="btn-close" onclick={handleClose}>Zamknij</button>
  </div>
</dialog>

<style>
  dialog {
    border: 2px solid var(--auxie-deep-navy-600);
    border-radius: 20px;
    padding: 24px;
    background: var(--auxie-deep-navy-700);
    color: var(--auxie-cloud-white-100);
    max-width: 90vw;
    width: 400px;
    box-sizing: border-box;
    position: fixed;
    top: 50%;
    left: 50%;
    transform: translate(-50%, -50%);
    margin: 0;
    font-family: "Onest", sans-serif;
    box-shadow: 0 10px 40px rgba(0, 0, 0, 0.5);
  }

  dialog[open] {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 20px;
  }

  dialog::backdrop {
    background: rgba(8, 10, 15, 0.8);
    backdrop-filter: blur(8px);
  }

  .dialog-title {
    margin: 0;
    font-size: 18px;
    font-family: "Sora", sans-serif;
    font-weight: 800;
    text-align: center;
    color: var(--auxie-intense-mint-500);
  }

  .history-list {
    width: 100%;
    max-height: 250px;
    overflow-y: auto;
    display: flex;
    flex-direction: column;
    gap: 8px;
    background: var(--auxie-deep-navy-800);
    padding: 12px;
    border-radius: 12px;
    border: 1px solid var(--auxie-deep-navy-600);
  }

  .track-item {
    display: flex;
    justify-content: space-between;
    font-size: 13px;
    padding-bottom: 6px;
    border-bottom: 1px solid var(--auxie-deep-navy-600);
  }

  .track-item:last-child {
    border-bottom: none;
    padding-bottom: 0;
  }

  .track-info {
    display: flex;
    gap: 5px;
    color: var(--auxie-cloud-white-100);
  }

  .track-duration {
    color: var(--auxie-cloud-white-600);
  }

  .track-status {
    color: var(--auxie-vivid-blue-400);
    font-family: monospace;
  }

  .track-status.skipped {
    color: var(--auxie-soft-crimson-400);
  }

  .empty-msg {
    text-align: center;
    font-size: 13px;
    color: var(--auxie-cloud-white-600);
  }

  .actions {
    display: flex;
    width: 100%;
    gap: 10px;
  }

  .actions button {
    flex: 1;
    border-radius: 12px;
    padding: 10px;
    font-size: 14px;
    font-weight: 600;
    cursor: pointer;
    transition: all 0.2s;
  }

  .btn-copy {
    background: var(--auxie-deep-navy-600);
    border: 1px solid var(--auxie-deep-navy-500);
    color: var(--auxie-cloud-white-200);
  }

  .btn-copy:hover {
    background: var(--auxie-deep-navy-500);
  }

  .btn-close {
    background: var(--auxie-intense-mint-500);
    border: none;
    color: var(--auxie-deep-navy-900);
  }

  .btn-close:hover {
    background: var(--auxie-intense-mint-600);
  }
</style>