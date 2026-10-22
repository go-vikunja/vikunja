<template>
	<audio
		v-if="blobUrl"
		ref="playerRef"
		:src="blobUrl"
		:aria-label="attachment.file?.name"
		class="audio-player"
		controls
		autoplay
		@play="stopOtherPlayers"
		@error="onAudioError"
	/>
	<XButton
		v-else
		:loading="loading"
		:aria-label="$t('task.attachment.playFile', {file: attachment.file?.name})"
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
import {onBeforeUnmount, ref, watch} from 'vue'
import {useI18n} from 'vue-i18n'
import {fetchAttachmentUrl, releaseAttachmentUrl} from '@/helpers/attachments'
import {isRequestContextAbort} from '@/client/requestContext'
import type {TaskAttachment as IAttachment} from '@/client/generated'
import {error} from '@/message'

const props = defineProps<{
	attachment: IAttachment
}>()

const {t} = useI18n({useScope: 'global'})

const blobUrl = ref<string | undefined>(undefined)
const playerRef = ref<HTMLAudioElement | null>(null)
const loading = ref(false)
let unmounted = false
let previewEpoch = 0

// Fetched on demand: the download endpoint needs the auth header, so no plain-url streaming.
async function loadAudio() {
	if (loading.value || blobUrl.value) {
		return
	}

	loading.value = true
	const epoch = previewEpoch
	try {
		const url = await fetchAttachmentUrl({id: props.attachment.id!, task_id: props.attachment.task_id!})
		if (unmounted || epoch !== previewEpoch) {
			releaseAttachmentUrl(url)
			return
		}
		blobUrl.value = url
	} catch (e) {
		if (!unmounted && epoch === previewEpoch && !isRequestContextAbort(e)) error(e)
	} finally {
		if (epoch === previewEpoch) loading.value = false
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

	releaseAttachmentUrl(blobUrl.value)
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

function releaseAudio() {
	previewEpoch++
	loading.value = false
	if (playing === playerRef.value) {
		playing?.pause()
		playing = null
	}
	if (blobUrl.value !== undefined) {
		releaseAttachmentUrl(blobUrl.value)
		blobUrl.value = undefined
	}
}

// Scalar key: an array getter always compares changed, so every list refetch would kill playback.
watch(() => `${props.attachment.task_id}-${props.attachment.id}`, releaseAudio, {flush: 'sync'})
onBeforeUnmount(() => {
	unmounted = true
	releaseAudio()
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
