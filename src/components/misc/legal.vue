<template>
	<div class="legal-links">
		<BaseButton :href="imprintUrl" v-if="imprintUrl">{{ $t('navigation.imprint') }}</BaseButton>
		<span v-if="imprintUrl && privacyPolicyUrl"> | </span>
		<BaseButton :href="privacyPolicyUrl" v-if="privacyPolicyUrl">{{ $t('navigation.privacy') }}</BaseButton>
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
  margin-top: 1rem;
  text-align: right;
  color: var(--grey-300);
  font-size: 1rem;
}
</style>