<template>
	<img
		:src="blobUrl"
		alt="Attachment preview"
	>
</template>

<script setup lang="ts">
import {type PropType, ref, shallowReactive, watchEffect} from 'vue'
import AttachmentService from '@/services/attachment'
import type { IAttachment } from '@/modelTypes/IAttachment'

const props = defineProps({
	modelValue: {
		type: Object as PropType<IAttachment>,
		default: undefined,
	},
})

const attachmentService = shallowReactive(new AttachmentService())
const blobUrl = ref<string | undefined>(undefined)

watchEffect(async () => {
	if (props.modelValue) {
		blobUrl.value = await attachmentService.getBlobUrl(props.modelValue)
	}
})
</script>

<style scoped lang="scss">
img {
	width: 100%;
	border-radius: $radius;
	object-fit: cover;
}
</style>