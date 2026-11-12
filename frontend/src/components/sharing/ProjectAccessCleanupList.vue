<template>
	<p>{{ intro }}</p>
	<!-- Keep in sync with cleanupAfterProjectAccessLoss. -->
	<ul class="cleanup-list">
		<li>{{ $t(`project.share.accessCleanup.${perspective}.assignments`) }}</li>
		<li>{{ $t(`project.share.accessCleanup.${perspective}.subscriptions`) }}</li>
		<li>{{ $t(`project.share.accessCleanup.${perspective}.webhooks`) }}</li>
	</ul>
</template>

<script setup lang="ts">
const {perspective = 'others'} = defineProps<{
	intro: string
	perspective?: 'self' | 'others'
}>()
</script>

<style lang="scss" scoped>
.cleanup-list {
	display: inline-block;
	text-align: start;
	margin-inline-start: 0;
	margin-block-start: 0;
	padding-inline-start: 1.25rem;
}
</style>
