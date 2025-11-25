<template>
	<NodeViewWrapper class="mention-user">
		<img :src="avatarUrl">
		<span class="mention__label">
			{{ node.attrs.label ?? node.attrs.id }}
		</span>
	</NodeViewWrapper>
</template>

<script lang="ts" setup>
import { fetchAvatarBlobUrl } from '@/models/user'
import { nodeViewProps, NodeViewWrapper } from '@tiptap/vue-3'
import { watch, ref } from 'vue'

const props = defineProps(nodeViewProps)

const avatarUrl = ref('')

watch(
	() => props.node.attrs.id,
	async () => {
		avatarUrl.value = await fetchAvatarBlobUrl({username: props.node.attrs.id}, 32)
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