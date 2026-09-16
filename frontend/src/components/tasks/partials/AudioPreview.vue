<template>
	<audio
		v-if="blobUrl"
		ref="playerRef"
		:src="blobUrl"
		:aria-label="attachment.file.name"
		class="audio-player"
		controls
		autoplay
		@play="stopOtherPlayers"
		@error="onAudioError"
	/>
	<XButton
		v-else
		:loading="loading"
		:aria-label="$t('task.attachment.playFile', {file: attachment.file.name})"
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
import {onBeforeUnmount, ref} from 'vue'
import {useI18n} from 'vue-i18n'
import AttachmentService from '@/services/attachment'
import type {IAttachment} from '@/modelTypes/IAttachment'
import {error} from '@/message'

const props = defineProps<{
	attachment: IAttachment
}>()

const {t} = useI18n({useScope: 'global'})

const attachmentService = new AttachmentService()
const blobUrl = ref<string | undefined>(undefined)
const playerRef = ref<HTMLAudioElement | null>(null)
const loading = ref(false)
let unmounted = false

// Fetched on demand: the download endpoint needs the auth header, so no plain-url streaming.
async function loadAudio() {
	if (loading.value || blobUrl.value) {
		return
	}

	loading.value = true
	try {
		const url = await attachmentService.getBlobUrl(props.attachment) as string
		if (unmounted) {
			window.URL.revokeObjectURL(url)
			return
		}
		blobUrl.value = url
	} catch (e) {
		error(e)
	} finally {
		loading.value = false
	}
}

// Undecodable files only fail once the element has the blob, so swap back to the play button.
function onAudioError() {
	if (unmounted || blobUrl.value === undefined) {
		return
	}

	if (playing === playerRef.value) {
		playing = null
	}

	window.URL.revokeObjectURL(blobUrl.value)
	blobUrl.value = undefined
	error({message: t('task.attachment.audioError')})
}

function stopOtherPlayers(e: Event) {
	const player = e.currentTarget as HTMLAudioElement

	if (playing !== null && playing !== player) {
		playing.pause()
	}
	playing = player
}

async function play() {
	if (blobUrl.value === undefined) {
		// The element autoplays as soon as it gets the blob.
		await loadAudio()
		return
	}

	try {
		await playerRef.value?.play()
	} catch {
		// AbortError when another player pauses this one mid-start, NotAllowedError without a user gesture - neither deserves a toast.
	}
}

onBeforeUnmount(() => {
	unmounted = true

	// A detached element keeps playing for as long as something references it.
	if (playing === playerRef.value) {
		playing?.pause()
		playing = null
	}

	if (blobUrl.value !== undefined) {
		window.URL.revokeObjectURL(blobUrl.value)
	}
})

defineExpose({play})
</script>

<style scoped lang="scss">
.audio-player {
	inline-size: 100%;
	max-inline-size: 30rem;
	block-size: 2.5rem;
	margin-block: 0 1em;
}

.audio-play {
	max-inline-size: 30rem;
	margin-block: 0 1em;
}
</style>
