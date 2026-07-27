<template>
	<audio
		v-if="blobUrl"
		ref="playerRef"
		:src="blobUrl"
		class="audio-player"
		controls
		autoplay
		@play="stopOtherPlayers"
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

<script lang="ts">
// Module scope: starting a player pauses whichever one was playing before it.
let playing: HTMLAudioElement | null = null
</script>

<script setup lang="ts">
import {onBeforeUnmount, ref, shallowReactive} from 'vue'
import AttachmentService from '@/services/attachment'
import type {IAttachment} from '@/modelTypes/IAttachment'

const props = defineProps<{
	modelValue: IAttachment
}>()

const attachmentService = shallowReactive(new AttachmentService())
const blobUrl = ref<string | undefined>(undefined)
const playerRef = ref<HTMLAudioElement | null>(null)

// Fetched on demand: the download endpoint needs the auth header, so no plain-url streaming.
async function loadAudio() {
	blobUrl.value = await attachmentService.getBlobUrl(props.modelValue)
}

function stopOtherPlayers() {
	if (playing !== null && playing !== playerRef.value) {
		playing.pause()
	}
	playing = playerRef.value
}

onBeforeUnmount(() => {
	// A detached element keeps playing for as long as something references it.
	if (playing === playerRef.value) {
		playing?.pause()
		playing = null
	}

	if (blobUrl.value !== undefined) {
		window.URL.revokeObjectURL(blobUrl.value)
	}
})
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
