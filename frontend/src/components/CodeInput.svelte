<script lang="ts">
interface Props {
  value: string;
  length?: number;
  name?: string;
}

let { value = $bindable(""), length = 6, name }: Props = $props();

let inputRef: HTMLInputElement | null = $state(null);

function forceCursorToEnd(e: Event) {
  if (inputRef) {
    inputRef.setSelectionRange(inputRef.value.length, inputRef.value.length);
  }
}

function handleInput(e: Event) {
  const target = e.target as HTMLInputElement;
  value = target.value.toUpperCase().slice(0, length);
  target.value = value;
  target.setSelectionRange(value.length, value.length);
}

function handleKeyDown(e: KeyboardEvent) {
  if (["ArrowLeft", "ArrowRight", "Home", "End"].includes(e.key)) {
    e.preventDefault();
  }
}
</script>

<div class="code-input-container">
  <input
    type="text"
    maxlength={length}
    bind:this={inputRef}
    value={value}
    oninput={handleInput}
    onselect={forceCursorToEnd}
    onclick={forceCursorToEnd}
    onkeydown={handleKeyDown}
    class="real-input"
    {name}
    autocomplete="off"
    autocorrect="off"
    autocapitalize="characters"
    spellcheck="false"
  />

  <div class="slots-overlay">
    {#each Array(length) as _, index}
      <div class="slot" class:filled={index < value.length} class:active={index === value.length}>
        {value[index] || "•"}
      </div>
    {/each}
  </div>
</div>

<style>
.code-input-container {
  position: relative;
  width: 100%;
  border-radius: 20px;
  corner-shape: squircle;
  border: 2px solid var(--auxie-deep-navy-500);
  background-color: var(--auxie-deep-navy-600);
  box-shadow: 
    inset 0 -4px 4.3px 0 rgba(122, 131, 177, 0.15),
    inset 0 3px 6.8px 0 rgba(11, 13, 23, 0.45);
  transition: border-color 0.2s ease, box-shadow 0.2s ease;
  display: flex;
  align-items: center;
  padding: 11px 16px;
}

.real-input {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  opacity: 0;
  cursor: text;
  z-index: 2;
  outline: none;
  background: transparent;
  border: none;
  font-size: 16px; 
}

.slots-overlay {
  display: flex;
  justify-content: space-between;
  width: 100%;
  align-items: center;
  z-index: 1;
  pointer-events: none;
}

.slot {
  flex: 1;
  text-align: center;
  font-family: "Onest", sans-serif;
  font-size: 18px;
  font-weight: bold;
  color: var(--auxie-cloud-white-100);
  transition: color 0.15s ease;
}

.slot:not(.filled) {
  color: var(--auxie-deep-navy-400);
}

.slot.active {
  color: var(--auxie-electric-purple-400);
}

.code-input-container:hover {
  border-color: var(--auxie-deep-navy-400);
}

.code-input-container:focus-within {
  border-color: var(--auxie-electric-purple-500);
  box-shadow: 
    0 0 10px rgba(112, 0, 255, 0.15),
    inset 0 -4px 4.3px 0 rgba(122, 131, 177, 0.15),
    inset 0 3px 6.8px 0 rgba(11, 13, 23, 0.45);
}
</style>
