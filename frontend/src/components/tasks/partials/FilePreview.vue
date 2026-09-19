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
import {computed, ref, watch} from 'vue'
import {fetchAttachmentBlobUrl} from '@/helpers/attachments'
import type {TaskAttachment as IAttachment} from '@/client/generated'
import type {AttachmentIdentity} from '@/client/queries/attachments'
import {canPreviewAudio, canPreviewImage, canPreviewPdf, canPreviewVideo} from '@/helpers/attachmentPreview'

const props = defineProps<{
	modelValue?: IAttachment
}>()

const blobUrl = ref<string | undefined>(undefined)
const isPdf = computed(() => props.modelValue && canPreviewPdf(props.modelValue))
const isAudio = computed(() => props.modelValue && canPreviewAudio(props.modelValue))
const isVideo = computed(() => props.modelValue && canPreviewVideo(props.modelValue))

const previewIdentity = computed<AttachmentIdentity | null>(() => {
	const attachment = props.modelValue
	if (!attachment?.id || !attachment.task_id || !canPreviewImage(attachment)) {
		return null
	}
	return {
		id: attachment.id,
		task_id: attachment.task_id,
	}
})

// Keyed on the ids, not the prop object: a list refetch hands over an equal attachment as a new object.
watch(
	() => previewIdentity.value && `${previewIdentity.value.task_id}-${previewIdentity.value.id}`,
	async (_key, _previous, onCleanup) => {
		const identity = previewIdentity.value
		let active = true
		onCleanup(() => {
			active = false
		})
		blobUrl.value = undefined
		if (identity === null) {
			return
		}

		try {
			const url = await fetchAttachmentBlobUrl(identity, 'md')
			if (active) blobUrl.value = url
		} catch {
			// noop
		}
	},
	{immediate: true},
)
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
