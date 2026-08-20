<template>
	<audio
		v-if="blobUrl"
		ref="playerRef"
		:src="blobUrl"
		class="audio-player"
		controls
		autoplay
		@play="stopOtherPlayers"
		@error="onAudioError"
	/>
	<XButton
		v-else
		:loading="loading"
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
// Module scope, so that starting one attachment stops whichever was playing before it -
// every player autoplays as soon as its file arrives.
let playing: HTMLAudioElement | null = null
</script>

<script setup lang="ts">
import {onBeforeUnmount, ref} from 'vue'
import {useI18n} from 'vue-i18n'
import AttachmentService from '@/services/attachment'
import type {IAttachment} from '@/modelTypes/IAttachment'
import {error} from '@/message'

const props = defineProps<{
	modelValue: IAttachment
}>()

const {t} = useI18n({useScope: 'global'})

const attachmentService = new AttachmentService()
const blobUrl = ref<string | undefined>(undefined)
const playerRef = ref<HTMLAudioElement | null>(null)
const loading = ref(false)
let unmounted = false

// The file is only fetched on demand: the download endpoint needs the auth header, so the
// player cannot stream from a plain url and would otherwise pull the whole file on page load.
async function loadAudio() {
	if (loading.value || blobUrl.value) {
		return
	}

	loading.value = true
	try {
		const url = await attachmentService.getBlobUrl(props.modelValue)
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
	if (blobUrl.value === undefined) {
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
