<script lang="ts">
interface Props {
  value: string;
  placeholder: string;
  name?: string;
}
let { value = $bindable(), placeholder, name }: Props = $props();
</script>
<input type="text" bind:value={value} placeholder="{placeholder}" {name}>

<style>
input {
  width: 100%;
  font-family: "Onest", sans-serif;
  font-size: 16px;
  border-radius: 20px;
  corner-shape: squircle;
  border: 2px solid var(--auxie-deep-navy-400);
  padding: 11px 16px;
  background-color: var(--auxie-deep-navy-600);
  color: var(--auxie-cloud-white-100);
  box-shadow: 
    inset 0 -4px 4.3px 0 rgba(122, 131, 177, 0.15),
    inset 0 3px 6.8px 0 rgba(11, 13, 23, 0.45);
  outline: none;
  transition: border-color 0.2s ease, box-shadow 0.2s ease;
}

input::placeholder {
  color: var(--auxie-deep-navy-300);
}

input:hover {
  border-color: var(--auxie-deep-navy-400);
}

input:focus {
  border-color: var(--auxie-electric-purple-500);
  box-shadow: 
    0 0 10px rgba(112, 0, 255, 0.15),
    inset 0 -4px 4.3px 0 rgba(122, 131, 177, 0.15),
    inset 0 3px 6.8px 0 rgba(11, 13, 23, 0.45);
}
</style>
