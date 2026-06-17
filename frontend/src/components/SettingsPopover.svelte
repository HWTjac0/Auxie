<script lang="ts">
let { onInvite, onLeave }: { onInvite: () => void, onLeave: () => void } = $props();

function handleInvite() {
  document.getElementById("actions_popover")?.hidePopover();
  onInvite();
}

function handleSettings() {
  document.getElementById("actions_popover")?.hidePopover();
  console.log("Settings");
}

function handleLeave() {
  document.getElementById("actions_popover")?.hidePopover();
  onLeave();
}
</script>

<div id="actions_popover" popover>
  <div class="popover_menu">
    <button class="popover_item onest-500" onclick={handleInvite}>
      Invite friends
    </button>
    <button class="popover_item onest-500" onclick={handleSettings}>
      Room settings
    </button>
    <div class="popover_divider"></div>
    <button class="popover_item popover_leave onest-500" onclick={handleLeave}>
      Leave room
    </button>
  </div>
</div>

<style>
  /* Zostaw istniejące style dla Popovera dokładnie takie jakie miałeś w pliku: */
  #actions_popover { position-anchor: --actions-anchor; inset: unset; top: anchor(--actions-anchor bottom); right: anchor(--actions-anchor right); margin-top: 10px; background-color: var(--auxie-deep-navy-700); border: 1px solid var(--auxie-deep-navy-600); border-radius: 12px; padding: 8px; color: var(--auxie-cloud-white-50); box-shadow: 0 10px 25px rgba(0, 0, 0, 0.5); }
  #actions_popover::backdrop { background: transparent; }
  .popover_menu { display: flex; flex-direction: column; gap: 4px; min-width: 160px; }
  .popover_item { display: flex; align-items: center; gap: 10px; padding: 10px 12px; background: transparent; border: none; border-radius: 8px; color: var(--auxie-cloud-white-600); font-size: 14px; cursor: pointer; text-align: left; text-decoration: none; transition: all 0.2s ease; }
  .popover_item:hover { background-color: var(--auxie-deep-navy-600); color: var(--auxie-cloud-white-50); }
  .popover_divider { height: 1px; background-color: var(--auxie-deep-navy-600); margin: 4px 0; }
  .popover_leave { color: var(--auxie-soft-crimson-400); }
  .popover_leave:hover { background-color: rgba(239, 68, 68, 0.1); color: var(--auxie-soft-crimson-500); }
</style>