<template>
	<!-- Preview image -->
	<img
		v-if="blobUrl"
		:src="blobUrl"
		alt="Attachment preview"
	>

	<!-- PDF icon -->
	<div
		v-else-if="isPdf"
		class="icon-wrapper"
	>
		<Icon
			size="6x"
			icon="file-pdf"
		/>
	</div>

	<!-- Audio icon -->
	<div
		v-else-if="isAudio"
		class="icon-wrapper"
	>
		<Icon
			size="6x"
			icon="volume-high"
		/>
	</div>

	<!-- Video icon -->
	<div
		v-else-if="isVideo"
		class="icon-wrapper"
	>
		<Icon
			size="6x"
			icon="play"
		/>
	</div>

	<!-- Fallback -->
	<div
		v-else
		class="icon-wrapper"
	>
		<Icon
			size="6x"
			icon="file"
		/>
	</div>
</template>

<script setup lang="ts">
import {computed, ref, watchEffect} from 'vue'
import {PREVIEW_SIZE} from '@/helpers/attachments'
import {attachmentBlobUrl} from '@/helpers/attachments'
import type {TaskAttachment as IAttachment} from '@/client/generated'
import {canPreviewAudio, canPreviewImage, canPreviewPdf, canPreviewVideo} from '@/helpers/attachmentPreview'

const props = defineProps<{
	modelValue?: IAttachment
}>()

const blobUrl = ref<string | undefined>(undefined)
const isPdf = computed(() => props.modelValue && canPreviewPdf(props.modelValue))
const isAudio = computed(() => props.modelValue && canPreviewAudio(props.modelValue))
const isVideo = computed(() => props.modelValue && canPreviewVideo(props.modelValue))

watchEffect(async onCleanup => {
	let active = true
	let ownedUrl: string | undefined
	blobUrl.value = undefined
	onCleanup(() => {
		active = false
		if (ownedUrl) URL.revokeObjectURL(ownedUrl)
	})
	const attachment = props.modelValue
	if (!attachment || !canPreviewImage(attachment)) {
		return
	}

	try {
		const url = await attachmentBlobUrl({id: attachment.id!, task_id: attachment.task_id!}, PREVIEW_SIZE.MD)
		// a newer attachment may have won the race while this one was in flight
		if (active) {
			ownedUrl = url
			blobUrl.value = url
		} else {
			URL.revokeObjectURL(url)
		}
	} catch {
		// fall back to the generic file icon
	}
})
</script>

<style scoped lang="scss">
img {
	inline-size: 100%;
	border-radius: $radius;
	object-fit: cover;
}

.icon-wrapper {
	color: var(--grey-500);
}
</style>
