<template>
	<p class="upgrade-hint has-text-grey is-size-7">
		<slot />
		<BaseButton
			v-if="upgradeUrl"
			:href="upgradeUrl"
			class="upgrade-hint__link"
		>
			{{ $t('entitlement.upgrade') }}
		</BaseButton>
	</p>
</template>

<script setup lang="ts">
import {computed} from 'vue'

import BaseButton from '@/components/base/BaseButton.vue'
import {useConfigStore} from '@/stores/config'

const configStore = useConfigStore()
const upgradeUrl = computed(() => configStore.upgradeUrl)
</script>

<style lang="scss" scoped>
.upgrade-hint__link {
	margin-inline-start: .25rem;
	text-decoration: underline;
}
</style>
