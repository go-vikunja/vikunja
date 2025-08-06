<template>
	<BaseButton
		v-shortcut="'Mod+e'"
		class="menu-show-button"
		:title="$t('keyboardShortcuts.toggleMenu')"
		:aria-label="menuActive ? $t('misc.hideMenu') : $t('misc.showMenu')"
		@click="baseStore.toggleMenu()"
		@shortkey="() => baseStore.toggleMenu()"
	/>
</template>

<script setup lang="ts">
import {computed} from 'vue'
import {useBaseStore} from '@/stores/base'

import BaseButton from '@/components/base/BaseButton.vue'

const baseStore = useBaseStore()
const menuActive = computed(() => baseStore.menuActive)
</script>

<style lang="scss" scoped>
$line-width: 2rem;
$size: $line-width + 1rem;

.menu-show-button {
	min-block-size: $size;
	inline-size: $size;

	position: relative;

	$transform-x: translateX(-50%);

	&::before,
	&::after {
		content: '';
		display: block;
		position: absolute;
		block-size: 3px;
		inline-size: $line-width;
		inset-inline-start: 50%;
		transform: $transform-x;
		background-color: var(--grey-400);
		border-radius: 2px;
		transition: all $transition;
	}

	&::before {
		inset-block-start: 50%;
		transform: $transform-x translateY(-0.4rem)
	}

	&::after {
		inset-block-end: 50%;
		transform: $transform-x translateY(0.4rem)
	}

	&:hover,
	&:focus {
		&::before,
		&::after {
			background-color: var(--grey-600);
		}

		&::before {
			transform: $transform-x translateY(-0.5rem);
		}

		&::after {
			transform: $transform-x translateY(0.5rem)
		}
	}
}
</style>
