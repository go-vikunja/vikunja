<template>
	<audio
		v-if="blobUrl"
		:src="blobUrl"
		class="audio-player"
		controls
		autoplay
	/>
	<XButton
		v-else
		:disabled="attachmentService.loading"
		class="audio-play"
		icon="play"
		variant="secondary"
		:shadow="false"
		@click="loadAudio()"
	>
		{{ $t('task.attachment.play') }}
	</XButton>
</template>

<script setup lang="ts">
import {ref, shallowReactive} from 'vue'
import AttachmentService from '@/services/attachment'
import type {IAttachment} from '@/modelTypes/IAttachment'

const props = defineProps<{
	modelValue: IAttachment
}>()

const attachmentService = shallowReactive(new AttachmentService())
const blobUrl = ref<string | undefined>(undefined)

// The file is only fetched on demand: the download endpoint needs the auth header, so the
// player cannot stream from a plain url and would otherwise pull the whole file on page load.
async function loadAudio() {
	blobUrl.value = await attachmentService.getBlobUrl(props.modelValue)
}
</script>

<style scoped lang="scss">
.audio-player {
	inline-size: 100%;
	max-inline-size: 30rem;
	block-size: 2.5rem;
	margin-block: 0 1em;
}

.audio-play {
	margin-block: 0 1em;
}
</style>
