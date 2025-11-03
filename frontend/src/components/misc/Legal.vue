<template>
	<div class="legal-links">
		<BaseButton
			v-if="imprintUrl"
			:href="imprintUrl"
		>
			{{ $t('navigation.imprint') }}
		</BaseButton>
		<span v-if="imprintUrl && privacyPolicyUrl"> | </span>
		<BaseButton
			v-if="privacyPolicyUrl"
			:href="privacyPolicyUrl"
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

const imprintUrl = computed(() => configStore.legal.imprintUrl)
const privacyPolicyUrl = computed(() => configStore.legal.privacyPolicyUrl)
</script>

<style lang="scss" scoped>
.legal-links {
  margin-block-start: 1rem;
  text-align: end;
  color: var(--grey-300);
  font-size: 1rem;
}
</style>
