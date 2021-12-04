<template>
	<div class="legal-links">
		<a :href="imprintUrl" rel="noreferrer noopener nofollow" target="_blank" v-if="imprintUrl">{{ $t('navigation.imprint') }}</a>
		<span v-if="imprintUrl && privacyPolicyUrl"> | </span>
		<a :href="privacyPolicyUrl" rel="noreferrer noopener nofollow" target="_blank" v-if="privacyPolicyUrl">{{ $t('navigation.privacy') }}</a>
	</div>
</template>

<script lang="ts" setup>
import {computed} from 'vue'
import {useStore} from 'vuex'

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