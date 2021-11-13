<template>
    <button
        type="button"
        @click="$store.commit('toggleMenu')"
        class="menu-show-button"
        @shortkey="() => $store.commit('toggleMenu')"
        v-shortkey="['ctrl', 'e']"
        :aria-label="menuActive ? $t('misc.hideMenu') : $t('misc.showMenu')"
    />
</template>

<script setup>
import { computed} from 'vue'
import {store} from '@/store'

const menuActive = computed(() => store.menuActive)
</script>

<style lang="scss" scoped>
$lineWidth: 2rem;
$size: $lineWidth + 1rem;

.menu-show-button {
	// FIXME: create general button component
	appearance: none;
	background-color: transparent;
	border: 0;

	min-height: $size;
	width: $size;

	position: relative;

	$transformX: translateX(-50%);
	&::before,
	&::after {
		content: '';
		display: block;
		position: absolute;
		height: 3px;
		width: $lineWidth;
		left: 50%;
		transform: $transformX;
		background-color: $grey-400;
		border-radius: 2px;
		transition: all $transition;
	}

	&::before {
		top: 50%;
		transform: $transformX translateY(-0.4rem)
	}

	&::after {
		bottom: 50%;
		transform: $transformX translateY(0.4rem)
	}

	&:hover,
	&:focus {
		&::before,
		&::after {
			background-color: $grey-600;
		}

		&::before {
			transform: $transformX translateY(-0.5rem);
		}

		&::after {
			transform: $transformX translateY(0.5rem)
		}
	}
}
</style>