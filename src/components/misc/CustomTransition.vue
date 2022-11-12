<template>
	<transition name="flash-background">
		<slot  />
	</transition>
</template>

<script setup lang="ts">
defineProps<{
	name: 'flash-background'
}>()
</script>

<style scoped lang="scss">
$flash-background-duration: 750ms;

.flash-background-enter-from,
.flash-background-enter-active {
  animation: flash-background $flash-background-duration ease 1;
}

@keyframes flash-background {
  0% {
    background: var(--primary-light);
  }
  100% {
    background: transparent;
  }
}

@media (prefers-reduced-motion: reduce) {
	@keyframes flash-background {
		0% {
			background: transparent;
		}
	}
}
</style>