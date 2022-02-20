<template>
	<BaseButton
		class="menu-show-button"
		@click="$store.commit('toggleMenu')"
		@shortkey="() => $store.commit('toggleMenu')"
		v-shortcut="'Control+e'"
		:title="$t('keyboardShortcuts.toggleMenu')"
		:aria-label="menuActive ? $t('misc.hideMenu') : $t('misc.showMenu')"
	/>
</template>

<script setup lang="ts">
import {computed} from 'vue'
import {useStore} from 'vuex'

import BaseButton from '@/components/base/BaseButton.vue'

const store = useStore()
const menuActive = computed(() => store.state.menuActive)
</script>

<style lang="scss" scoped>
$lineWidth: 2rem;
$size: $lineWidth + 1rem;

.menu-show-button {
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
		background-color: var(--grey-400);
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
			background-color: var(--grey-600);
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