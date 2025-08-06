<template>
	<transition :name="name">
		<!-- eslint-disable-next-line -->
		<slot/>
	</transition>
</template>

<script setup lang="ts">
withDefaults(defineProps<{
	name?: 'fade' | 'flash-background' | 'width' | 'modal'
}>(), {
	name: 'fade',
})
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

.fade-enter-active,
.fade-leave-active {
  transition: opacity $transition-duration;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}

.width-enter-active,
.width-leave-active {
  transition: width $transition-duration;
}

.width-enter-from,
.width-leave-to {
  inline-size: 0;
}

.modal-enter,
.modal-leave-active {
	opacity: 0;
}

.modal-enter .modal-container,
.modal-leave-active .modal-container {
	transform: scale(0.9);
}
</style>
