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
		<PhFile size="6em" />
	</div>
</template>

<script setup lang="ts">
import {ref, shallowReactive, watchEffect} from 'vue'
import {PhFile} from '@phosphor-icons/vue'
import AttachmentService, {PREVIEW_SIZE} from '@/services/attachment'
import type {IAttachment} from '@/modelTypes/IAttachment'
import {canPreview} from '@/models/attachment'

const props = defineProps<{
	modelValue?: IAttachment
}>()

const attachmentService = shallowReactive(new AttachmentService())
const blobUrl = ref<string | undefined>(undefined)

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
