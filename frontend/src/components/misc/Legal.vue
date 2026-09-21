<template>
	<div class="legal-links">
		<BaseButton
			v-if="imprint_url"
			:href="imprint_url"
		>
			{{ $t('navigation.imprint') }}
		</BaseButton>
		<span v-if="imprint_url && privacy_policy_url"> | </span>
		<BaseButton
			v-if="privacy_policy_url"
			:href="privacy_policy_url"
		>
			{{ $t('navigation.privacy') }}
		</BaseButton>
	</div>
</template>

<script lang="ts" setup>
import {computed} from 'vue'
import BaseButton from '@/components/base/BaseButton.vue'
import {useConfigStore} from '@/stores/config'

const configStore = useConfigStore()

const imprint_url = computed(() => configStore.legal.imprint_url)
const privacy_policy_url = computed(() => configStore.legal.privacy_policy_url)
</script>

<style lang="scss" scoped>
.legal-links {
  margin-block-start: 1rem;
  text-align: end;
  color: var(--grey-300);
  font-size: 1rem;
}
</style>
