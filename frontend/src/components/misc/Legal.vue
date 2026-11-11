<template>
	<div class="legal-links">
		<template
			v-for="(link, index) in legalLinks"
			:key="index"
		>
			<span v-if="index > 0"> | </span>
			<BaseButton :href="link.url">
				{{ link.text }}
			</BaseButton>
		</template>
	</div>
</template>

<script lang="ts" setup>
import BaseButton from '@/components/base/BaseButton.vue'
import {useLegalLinks} from '@/composables/useLegalLinks'

const legalLinks = useLegalLinks()
</script>

<style lang="scss" scoped>
.legal-links {
  margin-block-start: 1rem;
  text-align: end;
  color: var(--grey-300);
  font-size: 1rem;
}
</style>
