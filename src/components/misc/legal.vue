<template>
	<div class="legal-links">
		<BaseButton :href="imprintUrl" v-if="imprintUrl">{{ $t('navigation.imprint') }}</BaseButton>
		<span v-if="imprintUrl && privacyPolicyUrl"> | </span>
		<BaseButton :href="privacyPolicyUrl" v-if="privacyPolicyUrl">{{ $t('navigation.privacy') }}</BaseButton>
	</div>
</template>

<script lang="ts" setup>
import {computed} from 'vue'
import {useStore} from 'vuex'

import BaseButton from '@/components/base/BaseButton.vue'

const store = useStore()

const imprintUrl = computed(() => store.state.config.legal.imprintUrl)
const privacyPolicyUrl = computed(() => store.state.config.legal.privacyPolicyUrl)
</script>

<style lang="scss" scoped>
.legal-links {
  margin-top: 1rem;
  text-align: right;
  color: var(--grey-300);
  font-size: 1rem;
}
</style>