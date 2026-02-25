<template>
	<div class="menu-bottom-block">
		<BaseButton
			class="menu-bottom-link"
			:href="computedUrl"
			target="_blank"
		>
			{{ $t('misc.poweredBy') }}
		</BaseButton>
		<a
			v-if="buildVersion && buildVersion !== 'dev'"
			class="build-version"
			:href="commitUrl"
			target="_blank"
			rel="noopener"
		>
			{{ buildVersion }}
		</a>
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
const commitHash = computed(() => {
	// Extract hash from format like "custom-cc068123c"
	const match = buildVersion.value.match(/([0-9a-f]{7,})$/i)
	return match ? match[1] : ''
})
const commitUrl = computed(() =>
	commitHash.value
		? `https://github.com/trbom5c/vikunja/commit/${commitHash.value}`
		: '#',
)
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
	text-decoration: none;
	cursor: pointer;
	transition: opacity 0.15s, color 0.15s;

	&:hover {
		opacity: 1;
		color: var(--primary);
		text-decoration: underline;
	}
}
</style>
