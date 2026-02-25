<template>
	<div class="menu-bottom-block">
		<BaseButton
			class="menu-bottom-link"
			:href="computedUrl"
			target="_blank"
		>
			{{ $t('misc.poweredBy') }}
		</BaseButton>
		<span v-if="buildVersion && buildVersion !== 'dev'" class="build-version">
			{{ buildVersion }}
		</span>
	</div>
</template>

<script setup lang="ts">
import {computed} from 'vue'
import BaseButton from '@/components/base/BaseButton.vue'
import {POWERED_BY as poweredByUrl} from '@/urls'
import {VERSION} from '@/version.json'

const props = defineProps<{
	utmMedium: string;
}>()

const computedUrl = computed(() => `${poweredByUrl}&utm_medium=${props.utmMedium}`)
const buildVersion = computed(() => VERSION || '')
</script>

<style lang="scss">
.menu-bottom-block {
	text-align: center;
	padding-block-start: 1rem;
	padding-block-end: 1rem;
}

.menu-bottom-link {
	color: var(--grey-300);
	display: block;
	font-size: .8rem;
}

.build-version {
	display: block;
	color: var(--grey-500);
	font-size: .65rem;
	margin-top: 2px;
	opacity: 0.6;
	font-family: monospace;
	letter-spacing: 0.3px;
}
</style>
