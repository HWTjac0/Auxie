<script lang="ts">
const loginWithSpotify = () => {
   window.location.href = "http://127.0.0.1:8080/api/v1/auth/spotify/login";
};
</script>
<h1>Welcome to Auxie!</h1>
<button onclick={loginWithSpotify}>Login with Spotify</button>
