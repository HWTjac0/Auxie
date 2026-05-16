<script lang="ts">
let { children } = $props();
</script>

<div class="wrapper"> 
  <div class="background_gradient background_gradient_top"></div>
  <div class="background_gradient background_gradient_bottom"></div>
  <div class="overlay"> </div>
  <div class="setup_container">
    {@render children()}
  </div>
</div>

<style>
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
}
.setup_container {
  height: 100%;
  display: grid;
  justify-content: center;
  justify-items: center;
  align-content: center;
  z-index: 100;
  position: relative;
}

</style>
