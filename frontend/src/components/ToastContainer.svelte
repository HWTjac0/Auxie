<script lang="ts">
import { toasts } from "../lib/toasts.svelte";
import { flip } from "svelte/animate";
import { fade, fly } from "svelte/transition";
</script>

<div class="toast-container">
  {#each toasts.list as toast (toast.id)}
    <div
      class="toast-item {toast.type}"
      animate:flip={{ duration: 300 }}
      in:fly={{ y: 30, duration: 350, opacity: 0 }}
      out:fade={{ duration: 150 }}
    >
      <div class="toast-indicator"></div>
      <div class="toast-content onest-500">
        {toast.message}
      </div>
      <button class="toast-close" onclick={() => toasts.remove(toast.id)}>&times;</button>
    </div>
  {/each}
</div>

<style>
  .toast-container {
    position: fixed;
    bottom: 24px;
    right: 24px;
    display: flex;
    flex-direction: column;
    gap: 12px;
    z-index: 9999;
    max-width: 380px;
    width: calc(100% - 48px);
    pointer-events: none;
  }

  .toast-item {
    pointer-events: auto;
    display: flex;
    align-items: center;
    background: rgba(15, 18, 36, 0.9);
    backdrop-filter: blur(12px);
    border: 1px solid rgba(255, 255, 255, 0.08);
    box-shadow: 
      0 12px 32px rgba(0, 0, 0, 0.3),
      inset 0 1px 0 rgba(255, 255, 255, 0.1);
    border-radius: 16px;
    padding: 14px 18px;
    gap: 12px;
    position: relative;
    overflow: hidden;
    transition: transform 0.2s ease;
  }

  .toast-item:hover {
    transform: translateY(-2px);
  }

  .toast-indicator {
    width: 4px;
    height: 100%;
    position: absolute;
    left: 0;
    top: 0;
  }

  .toast-content {
    color: var(--auxie-cloud-white-100);
    font-size: 14px;
    line-height: 1.4;
    flex: 1;
    font-family: "Onest", sans-serif;
  }

  .toast-close {
    background: transparent;
    border: none;
    color: var(--auxie-cloud-white-600);
    font-size: 24px;
    cursor: pointer;
    line-height: 1;
    padding: 0;
    margin-left: auto;
    display: flex;
    align-items: center;
    justify-content: center;
    transition: color 0.2s;
  }

  .toast-close:hover {
    color: var(--auxie-cloud-white-100);
  }

  /* Colors based on toast type */
  .toast-item.info .toast-indicator {
    background-color: #3b82f6; /* Blue */
  }
  .toast-item.info {
    border-left: 2px solid #3b82f6;
  }

  .toast-item.success .toast-indicator {
    background-color: var(--auxie-intense-mint-500, #10b981); /* Green */
  }
  .toast-item.success {
    border-left: 2px solid var(--auxie-intense-mint-500, #10b981);
  }

  .toast-item.warning .toast-indicator {
    background-color: var(--auxie-warm-orange-500, #f59e0b); /* Orange */
  }
  .toast-item.warning {
    border-left: 2px solid var(--auxie-warm-orange-500, #f59e0b);
  }

  .toast-item.error .toast-indicator {
    background-color: var(--auxie-razzmatazz-500);
    box-shadow: 0 0 10px rgba(239, 35, 107, 0.4);
  }

  .toast-item.joined .toast-indicator {
    background-color: var(--auxie-intense-mint-500);
    box-shadow: 0 0 10px rgba(0, 255, 135, 0.4);
  }

  .toast-item.left .toast-indicator {
    background-color: var(--auxie-razzmatazz-600);
    box-shadow: 0 0 10px rgba(239, 35, 107, 0.4);
  }

  .toast-item.track .toast-indicator {
    background-color: var(--auxie-electric-purple-500);
    box-shadow: 0 0 10px rgba(138, 43, 226, 0.4);
  }
  .toast-item.error {
    border-left: 2px solid var(--auxie-soft-crimson-500, #ef4444);
  }
</style>
