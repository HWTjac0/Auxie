<script lang="ts">
const loginWithSpotify = () => {
   window.location.href = "http://127.0.0.1:8080/api/v1/auth/spotify/login";
};
</script>
<h1>Welcome to Auxie!</h1>
<div class="login_container">
  <a href="/create">Host part</a>
  <a href="/join">Enter as Guest </a>
  <button onclick={loginWithSpotify}>Login with Spotify</button>
</div>

<style>
.login_container {
  width: max-content;
  display: grid;
  row-gap: 10px;
}
</style>
