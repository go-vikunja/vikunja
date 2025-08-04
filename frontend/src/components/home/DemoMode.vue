<script setup lang="ts">
import {computed, ref} from 'vue'
import {useConfigStore} from '@/stores/config'
import BaseButton from '@/components/base/BaseButton.vue'

const configStore = useConfigStore()
const hide = ref(false)
const enabled = computed(() => configStore.demoModeEnabled && !hide.value)
</script>

<template>
	<div
		v-if="enabled"
		class="demo-mode-banner"
	>
		<p>
			{{ $t('demo.title') }}
			<strong class="is-uppercase">{{ $t('demo.everythingWillBeDeleted') }}</strong>
		</p>
		<BaseButton
			class="hide-button"
			@click="() => hide = true"
		>
			<Icon icon="times" />
		</BaseButton>
	</div>
</template>

<style scoped lang="scss">
.demo-mode-banner {
	position: fixed;
	inset-block-end: 0;
	inset-inline: 0;
	background: var(--danger);
	z-index: 100;
	padding: .5rem;
	text-align: center;
	
	&, strong {
		color: hsl(220, 13%, 91%) !important; // --grey-200 in light mode, hardcoded because the color should not change
	}
}

.hide-button {
	padding: .25rem .5rem;
	cursor: pointer;
	position: absolute;
	inset-inline-end: .5rem;
	inset-block-start: .25rem;
}
</style>
