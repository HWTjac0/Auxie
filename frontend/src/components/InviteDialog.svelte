<script lang="ts">
  import { onMount } from "svelte";

  interface Props {
    slug: string;
  }

  let { slug }: Props = $props();

  let dialog: HTMLDialogElement;
  let QRCode: any = $state(null);
  let qrcodeContainer: HTMLDivElement;

  let inviteUrl = $derived(slug && typeof window !== "undefined" ? `${window.location.origin}/room/${slug}` : "");
  let joinCode = $derived(slug ? slug.split("-").at(-1)?.toUpperCase() : "");

  onMount(async () => {
    const module = await import("$lib/vendor/qrcode.min.js");
    QRCode = module.default;
  });

  $effect(() => {
    if (QRCode && qrcodeContainer && inviteUrl) {
      qrcodeContainer.innerHTML = "";
      new QRCode(qrcodeContainer, inviteUrl);
    }
  });

  export function show() {
    dialog?.showModal();
  }

  export function close() {
    dialog?.close();
  }

  let copySuccess = $state(false);
  async function copyToClipboard() {
    try {
      await navigator.clipboard.writeText(inviteUrl);
      copySuccess = true;
      setTimeout(() => {
        copySuccess = false;
      }, 2000);
    } catch (err) {
      console.error("Failed to copy text: ", err);
    }
  }
</script>

<dialog bind:this={dialog}>
  <h3 class="dialog-title">Invite friends</h3>
  
  <div class="qrcode-wrapper">
    <div bind:this={qrcodeContainer} class="qrcode-container"></div>
  </div>
  
  <div class="code-section">
    <span class="code-label">Join Code</span>
    <span class="code-value">{joinCode}</span>
  </div>

  <div class="url-section">
    <input type="text" readonly value={inviteUrl} class="url-input" />
    <button onclick={copyToClipboard} class="copy-btn">
      {copySuccess ? "Copied!" : "Copy"}
    </button>
  </div>

  <button class="close-btn" onclick={close}>Close</button>
</dialog>

<style>
  dialog {
    border: 2px solid var(--auxie-deep-navy-600);
    border-radius: 20px;
    padding: 24px;
    background: var(--auxie-deep-navy-700);
    color: var(--auxie-cloud-white-100);
    max-width: 90vw;
    width: 320px;
    box-sizing: border-box;
    position: fixed;
    top: 50%;
    left: 50%;
    transform: translate(-50%, -50%);
    margin: 0;
    font-family: "Onest", sans-serif;
    box-shadow: 0 10px 40px rgba(0, 0, 0, 0.5);
    corner-shape: squircle;
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

  .qrcode-wrapper {
    background: #ffffff;
    padding: 12px;
    border-radius: 12px;
    display: flex;
    justify-content: center;
    align-items: center;
  }

  .qrcode-container {
    width: 160px;
    height: 160px;
  }

  .qrcode-container :global(img), .qrcode-container :global(canvas) {
    width: 100%;
    height: 100%;
    display: block;
  }

  .code-section {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 4px;
  }

  .code-label {
    font-size: 12px;
    color: var(--auxie-cloud-white-600);
    text-transform: uppercase;
    letter-spacing: 0.05em;
  }

  .code-value {
    font-size: 24px;
    font-weight: 800;
    color: var(--auxie-intense-mint-500);
    font-family: monospace;
    letter-spacing: 0.1em;
  }

  .url-section {
    display: flex;
    width: 100%;
    gap: 8px;
    background: var(--auxie-deep-navy-800);
    border: 1px solid var(--auxie-deep-navy-500);
    border-radius: 12px;
    padding: 4px;
    box-sizing: border-box;
  }

  .url-input {
    flex: 1;
    background: none;
    border: none;
    color: var(--auxie-cloud-white-200);
    font-size: 14px;
    padding: 8px;
    outline: none;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    min-width: 0;
  }

  .copy-btn {
    background: var(--auxie-intense-mint-500);
    color: var(--auxie-deep-navy-900);
    border: none;
    border-radius: 8px;
    padding: 6px 12px;
    font-size: 12px;
    font-weight: bold;
    cursor: pointer;
    transition: background-color 0.2s;
  }

  .copy-btn:hover {
    background: var(--auxie-intense-mint-600);
  }

  .close-btn {
    background: var(--auxie-deep-navy-600);
    border: 1px solid var(--auxie-deep-navy-500);
    border-radius: 12px;
    padding: 10px;
    font-size: 14px;
    font-weight: 600;
    color: var(--auxie-cloud-white-200);
    cursor: pointer;
    transition: background 0.2s, color 0.2s;
    width: 100%;
  }

  .close-btn:hover {
    background: var(--auxie-deep-navy-500);
    color: var(--auxie-cloud-white-100);
  }
</style>
