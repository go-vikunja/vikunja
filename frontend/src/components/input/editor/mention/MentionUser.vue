<template>
	<NodeViewWrapper class="mention-user">
		<img
			:src="avatarUrl"
			alt=""
		>
		<span class="mention__label">
			{{ node.attrs.label ?? node.attrs.id }}
		</span>
	</NodeViewWrapper>
</template>

<script lang="ts" setup>
import { fetchAvatarBlobUrl } from '@/models/user'
import { nodeViewProps, NodeViewWrapper } from '@tiptap/vue-3'
import { watch, ref } from 'vue'
import type { IUser } from '@/modelTypes/IUser'

const props = defineProps(nodeViewProps)

const avatarUrl = ref<string>()

watch(
	() => props.node.attrs.id,
	async () => {
		const username = props.node.attrs.id as string
		avatarUrl.value = await fetchAvatarBlobUrl({username} as IUser, 32)
	},
	{immediate: true},
)
</script>

<style lang="scss">
.tiptap .mention-user {
    display: inline-flex;
    align-items: center;
    position: relative;
    inset-block-end: 0;
    padding-inline-start: 1.75rem;

    > img {
        border-radius: 100%;
        inline-size: 1.5rem;
        block-size: 1.5rem;
        position: absolute;
        inset-inline-start: 0;
    }
}
</style>