<script lang="ts">
import Button from '../../../components/Button.svelte';

const loginWithSpotify = () => {
   window.location.href = "http://127.0.0.1:8080/api/v1/auth/spotify/login";
};
</script>

<h1>Welcome to Auxie!</h1>

<div class="login_container">
  <div class="button_group">

    <Button href="/create" class="onest-800">
      Host party
    </Button>
    
    <Button href="/join" bgColor="var(--auxie-electric-purple-500)" shadowColor="var(--auxie-electric-purple-700)">
      Enter as Guest
    </Button>
  </div>
  <div class="separator_row">
    <div class="separator"></div>
    <p>Or login with</p>
    <div class="separator"></div>
  </div>
  <div class="button_group">
    <Button onclick={loginWithSpotify} bgColor="var(--auxie-intense-mint-500)" shadowColor="var(--auxie-intense-mint-700)">
      Login with Spotify
    </Button>
  </div>
</div>

<style>
.button_group {
  width: 100%;
  display: grid;
  row-gap: 18px;
}
.separator_row {
  font-family: "Onest";
  font-weight:bold;
  font-size: 14px;
  color: var(--auxie-cloud-white-200);
  display: grid;
  grid-template-columns: 30% 1fr 30%;
  justify-items: center;
  align-items: center;
}
.separator {
  width: 100%;
  height: 3px;
  border-radius: 4px;
  background-color: var(--auxie-deep-navy-500);
}
.login_container {
  width: 100%;
  display: grid;
  row-gap: 40px;
}
</style>
