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
import {computed, ref, shallowReactive, watchEffect} from 'vue'
import AttachmentService, {PREVIEW_SIZE} from '@/services/attachment'
import type {IAttachment} from '@/modelTypes/IAttachment'
import {canPreview, canPreviewPdf} from '@/models/attachment'

const props = defineProps<{
	modelValue?: IAttachment
}>()

const attachmentService = shallowReactive(new AttachmentService())
const blobUrl = ref<string | undefined>(undefined)
const isPdf = computed(() => props.modelValue && canPreviewPdf(props.modelValue))

watchEffect(async () => {
	if (props.modelValue && canPreview(props.modelValue)) {
		blobUrl.value = await attachmentService.getBlobUrl(props.modelValue, PREVIEW_SIZE.MD)
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
