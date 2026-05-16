<script lang="ts">
  import { onNavigate } from '$app/navigation';

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
</script>

<div class="wrapper"> 
  <div class="background_gradient background_gradient_top"></div>
  <div class="background_gradient background_gradient_bottom"></div>
  <div class="overlay"> </div>
  
  <header class="header sora-800">
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
  background: linear-gradient(to top right,  var(--auxie-electric-purple-600), var(--auxie-intense-mint-500));
  animation: float-top 15s ease-in-out infinite alternate;
}

.background_gradient_bottom {
  bottom: -30%;
  right: -30%;
  background: linear-gradient(to top right, var(--auxie-razzmatazz-600), var(--auxie-electric-purple-300));
  animation: float-bottom 20s ease-in-out infinite alternate;
}

@keyframes float-top {
  0% { transform: translate(0, 0) scale(1) rotate(0deg); filter: blur(150px) hue-rotate(0deg); }
  50% { transform: translate(5%, 5%) scale(1.1) rotate(15deg); filter: blur(150px) hue-rotate(30deg); }
  100% { transform: translate(-5%, 5%) scale(0.9) rotate(-10deg); filter: blur(150px) hue-rotate(-15deg); }
}

@keyframes float-bottom {
  0% { transform: translate(0, 0) scale(1) rotate(0deg); filter: blur(150px) hue-rotate(0deg); }
  50% { transform: translate(-5%, -5%) scale(1.1) rotate(-20deg); filter: blur(150px) hue-rotate(-30deg); }
  100% { transform: translate(5%, -5%) scale(0.95) rotate(10deg); filter: blur(150px) hue-rotate(15deg); }
}

.wrapper {
  background-color: var(--auxie-deep-navy-900);
  height: 100vh;
  position: relative;
  overflow: hidden;
  display: flex;
  flex-direction: column;
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
  from { opacity: 1; transform: translateY(0) scale(1); }
  to { opacity: 0; transform: translateY(-20px) scale(0.95); }
}

@keyframes slide-in {
  from { opacity: 0; transform: translateY(20px) scale(1.05); }
  to { opacity: 1; transform: translateY(0) scale(1); }
}
</style>
