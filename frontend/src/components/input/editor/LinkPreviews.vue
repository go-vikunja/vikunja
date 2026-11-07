<template>
	<LinkPreviewCard
		v-for="url in urls"
		:key="url"
		:href="url"
	/>
</template>

<script setup lang="ts">
import {computed} from 'vue'
import {useConfigStore} from '@/stores/config'
import {getFullBaseUrl} from '@/helpers/getFullBaseUrl'
import {standaloneLinkUrls} from '@/helpers/standaloneLinkUrls'
import LinkPreviewCard from './LinkPreviewCard.vue'

const props = defineProps<{content: string}>()
const config = useConfigStore()
const urls = computed(() => standaloneLinkUrls(props.content, [
	new URL(getFullBaseUrl(), window.location.origin).href,
	config.frontend_url,
]))
</script>
