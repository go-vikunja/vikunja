<template>
	<Message
		variant="warning"
		class="upgrade-hint"
	>
		<slot />
		<BaseButton
			v-if="upgradeUrl"
			:href="upgradeUrl"
			target="_blank"
			rel="noopener noreferrer"
			class="upgrade-hint__link"
		>
			{{ $t('entitlement.upgrade') }}
		</BaseButton>
	</Message>
</template>

<script setup lang="ts">
import {computed} from 'vue'

import BaseButton from '@/components/base/BaseButton.vue'
import Message from '@/components/misc/Message.vue'
import {useConfigStore} from '@/stores/config'

const configStore = useConfigStore()
const upgradeUrl = computed(() => configStore.upgradeUrl)
</script>

<style lang="scss" scoped>
.upgrade-hint__link {
	margin-inline-start: .5rem;
	font-weight: bold;
	text-decoration: underline;
}
</style>
