<template>
	<!-- Preview image -->
	<img
		v-if="blobUrl"
		:src="blobUrl"
		alt="Attachment preview"
	>

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
import {type PropType, ref, shallowReactive, watchEffect} from 'vue'
import AttachmentService from '@/services/attachment'
import type {IAttachment} from '@/modelTypes/IAttachment'
import {SUPPORTED_IMAGE_SUFFIX} from '@/models/attachment'

const props = defineProps({
	modelValue: {
		type: Object as PropType<IAttachment>,
		default: undefined,
	},
})

const attachmentService = shallowReactive(new AttachmentService())
const blobUrl = ref<string | undefined>(undefined)

watchEffect(async () => {
	if (props.modelValue && canPreview(props.modelValue)) {
		blobUrl.value = await attachmentService.getBlobUrl(props.modelValue)
	}
})

function canPreview(attachment: IAttachment): boolean {
	return SUPPORTED_IMAGE_SUFFIX.some((suffix) => attachment.file.name.toLowerCase().endsWith(suffix))
}
</script>

<style scoped lang="scss">
img {
	width: 100%;
	border-radius: $radius;
	object-fit: cover;
}

.icon-wrapper {
	color: var(--grey-500);
}
</style>