<script lang="ts">
  import { onNavigate } from "$app/navigation";
  import { page } from "$app/stores";
  import ArrowLeft from "../../components/icons/ArrowLeft.svelte";
  let username = $state("");
  let { children } = $props();

  onNavigate((navigation) => {
    if (!document.startViewTransition) return;

    return new Promise((resolve) => {
      document.startViewTransition(async () => {
        resolve();
        await navigation.complete;
      });
    });
  });

  let pageType = $derived(
    $page.url.pathname.includes("/join")
      ? "join"
      : $page.url.pathname.includes("/create")
        ? "create"
        : "welcome",
  );
</script>

<div class="wrapper {pageType}">
  <div class="background_gradient background_gradient_top"></div>
  <div class="background_gradient background_gradient_bottom"></div>
  <div class="overlay"></div>

  <header class="header sora-800">
    {#if pageType !== "welcome"}
      <nav class="top_nav">
        <a href="/" class="back-link">
          <ArrowLeft color="white" />
        </a>
      </nav>
    {/if}

    <img src="/assets/logo.png" alt="Auxie Logo" class="logo" />
    <h1>Auxie</h1>
  </header>

  <main class="setup_container">
    <div class="transition_wrapper">
      {@render children()}
    </div>
  </main>
</div>

<style>
  .top_nav {
    position: absolute;
    left: 20px;
    display: flex;
    align-items: center;
  }
  .back-link {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    padding: 10px;
    background: transparent;
    border-radius: 10px;
    transition: opacity 0.2s ease;
  }
  .back-link:hover {
    opacity: 0.7;
  }
  .setup_container :global(h2) {
    font-size: 32px;
    color: var(--auxie-cloud-white-50);
    text-align: center;
    margin-bottom: 20px;
  }

  .setup_container :global(form) {
    display: grid;
    row-gap: 10px;
  }
  .header {
    position: absolute;
    top: 0;
    left: 0;
    width: 100%;
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 11px;
    padding: 40px 0;
    z-index: 1000;
    color: var(--auxie-cloud-white-200);
  }

  .logo {
    height: 64px;
    width: auto;
    filter: drop-shadow(0 0 10px rgba(0, 0, 0, 0.3));
  }

  .overlay {
    background-color: var(--auxie-deep-navy-900);
    opacity: 0.56;
    z-index: 2;
    position: absolute;
    top: 0;
    left: 0;
    width: 100%;
    height: 100%;
  }

  .background_gradient {
    position: absolute;
    z-index: 0;
    width: 60%;
    height: 60%;
    filter: blur(150px);
    border-radius: 100%;
  }

  .background_gradient_top {
    top: -30%;
    left: -30%;
    background: var(--bg-gradient-top);
    animation: float-top 15s ease-in-out infinite alternate;
  }

  .background_gradient_bottom {
    bottom: -30%;
    right: -30%;
    background: var(--bg-gradient-bottom);
    animation: float-bottom 20s ease-in-out infinite alternate;
  }

  @keyframes float-top {
    0% {
      transform: translate(0, 0) scale(1) rotate(0deg);
      filter: blur(150px) hue-rotate(0deg);
    }
    50% {
      transform: translate(5%, 5%) scale(1.1) rotate(15deg);
      filter: blur(150px) hue-rotate(30deg);
    }
    100% {
      transform: translate(-5%, 5%) scale(0.9) rotate(-10deg);
      filter: blur(150px) hue-rotate(-15deg);
    }
  }

  @keyframes float-bottom {
    0% {
      transform: translate(0, 0) scale(1) rotate(0deg);
      filter: blur(150px) hue-rotate(0deg);
    }
    50% {
      transform: translate(-5%, -5%) scale(1.1) rotate(-20deg);
      filter: blur(150px) hue-rotate(-30deg);
    }
    100% {
      transform: translate(5%, -5%) scale(0.95) rotate(10deg);
      filter: blur(150px) hue-rotate(15deg);
    }
  }

  .wrapper {
    background-color: var(--auxie-deep-navy-900);
    height: 100vh;
    position: relative;
    overflow: hidden;
    display: flex;
    flex-direction: column;

    /* Default / Welcome Gradient Colors */
    --bg-gradient-top: linear-gradient(
      to top right,
      var(--auxie-electric-purple-600),
      var(--auxie-intense-mint-500)
    );
    --bg-gradient-bottom: linear-gradient(
      to top right,
      var(--auxie-razzmatazz-600),
      var(--auxie-electric-purple-300)
    );
  }

  .wrapper.join {
    --bg-gradient-top: linear-gradient(
      to top right,
      var(--auxie-vivid-blue-600),
      var(--auxie-intense-mint-400)
    );
    --bg-gradient-bottom: linear-gradient(
      to top right,
      var(--auxie-electric-purple-600),
      var(--auxie-vivid-blue-300)
    );
  }

  .wrapper.create {
    --bg-gradient-top: linear-gradient(
      to top right,
      var(--auxie-warm-orange-600),
      var(--auxie-vibrant-gold-400)
    );
    --bg-gradient-bottom: linear-gradient(
      to top right,
      var(--auxie-razzmatazz-600),
      var(--auxie-warm-orange-300)
    );
  }

  .setup_container {
    flex: 1;
    display: grid;
    justify-content: center;
    justify-items: center;
    align-content: center;
    z-index: 100;
    position: relative;
    width: 100%;
  }

  .transition_wrapper {
    view-transition-name: setup-content;
    width: 100%;
    max-width: 300px;
    display: flex;
    flex-direction: column;
    align-items: center;
  }

  /* View Transitions */
  ::view-transition-old(setup-content) {
    animation: 300ms cubic-bezier(0.4, 0, 0.2, 1) both slide-out;
  }

  ::view-transition-new(setup-content) {
    animation: 300ms cubic-bezier(0.4, 0, 0.2, 1) both slide-in;
  }

  @keyframes slide-out {
    from {
      opacity: 1;
      transform: translateY(0) scale(1);
    }
    to {
      opacity: 0;
      transform: translateY(-20px) scale(0.95);
    }
  }

  @keyframes slide-in {
    from {
      opacity: 0;
      transform: translateY(20px) scale(1.05);
    }
    to {
      opacity: 1;
      transform: translateY(0) scale(1);
    }
  }
</style>
