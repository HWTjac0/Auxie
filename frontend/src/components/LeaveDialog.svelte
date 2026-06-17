<script lang="ts">
  interface Props {
    isHost: boolean;
    onConfirm: () => void;
  }

  let { isHost, onConfirm }: Props = $props();
  let dialog: HTMLDialogElement;

  export function show() {
    dialog?.showModal();
  }

  export function close() {
    dialog?.close();
  }

  function handleConfirm() {
    close();
    onConfirm();
  }
</script>

<dialog bind:this={dialog}>
  <h3 class="dialog-title">Wyjście z pokoju</h3>
  
  <div class="message-section">
    {#if isHost}
      <p class="warning-text">Jesteś gospodarzem. Wyjście spowoduje <strong>zakończenie sesji</strong> i wyrzucenie wszystkich osób z pokoju!</p>
    {:else}
      <p>Czy na pewno chcesz opuścić ten pokój?</p>
    {/if}
  </div>

  <div class="actions">
    <button class="btn-cancel" onclick={close}>Anuluj</button>
    <button class="btn-confirm" onclick={handleConfirm}>Wyjdź</button>
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
    width: 340px;
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
    font-size: 20px;
    font-family: "Sora", sans-serif;
    font-weight: 800;
  }

  .message-section {
    text-align: center;
    font-size: 15px;
    color: var(--auxie-cloud-white-200);
    line-height: 1.5;
  }

  .warning-text {
    color: var(--auxie-soft-crimson-400);
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

  .btn-cancel {
    background: var(--auxie-deep-navy-600);
    border: 1px solid var(--auxie-deep-navy-500);
    color: var(--auxie-cloud-white-200);
  }

  .btn-cancel:hover {
    background: var(--auxie-deep-navy-500);
  }

  .btn-confirm {
    background: var(--auxie-soft-crimson-500);
    border: none;
    color: white;
  }

  .btn-confirm:hover {
    background: var(--auxie-soft-crimson-400);
  }
</style>