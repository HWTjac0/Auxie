<script lang="ts">
import TextInput from "../../../components/TextInput.svelte";
import Button from "../../../components/Button.svelte";
import CodeInput from "../../../components/CodeInput.svelte";

let username = $state("");
let joinCode = $state("");
</script>
<h2 class="sora-800">Join room</h2>
<form method="POST">
  <TextInput bind:value={username} name="username" placeholder="Username"/>
  <CodeInput bind:value={joinCode} name="joinCode" />
  <Button 
    bgColor="var(--auxie-warm-orange-500)" 
    shadowColor="var(--auxie-warm-orange-700)" 
    fontSize="14px"
  >
    Enter
  </Button>
</form>
<style>
  h2 {
    font-size: 32px;
    color: var(--auxie-cloud-white-50);
    text-align: center;
  }
</style>  
