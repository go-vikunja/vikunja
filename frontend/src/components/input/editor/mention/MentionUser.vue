<template>
	<NodeViewWrapper class="mention-user">
		<UserAvatar
			:user="user"
			:size="32"
			class="mention-user__avatar"
		/>
		<span class="mention__label">
			{{ node.attrs.label ?? node.attrs.id }}
		</span>
	</NodeViewWrapper>
</template>

<script lang="ts" setup>
import { nodeViewProps, NodeViewWrapper } from '@tiptap/vue-3'
import { computed } from 'vue'

import UserAvatar from '@/components/misc/UserAvatar.vue'

const props = defineProps(nodeViewProps)

const user = computed(() => ({username: props.node.attrs.id as string}))
</script>

<style lang="scss">
.tiptap .mention-user {
    display: inline-flex;
    align-items: center;
    position: relative;
    inset-block-end: 0;
    padding-inline-start: 1.75rem;

    .mention-user__avatar {
        border-radius: 100%;
        inline-size: 1.5rem;
        block-size: 1.5rem;
        position: absolute;
        inset-inline-start: 0;
    }
}
</style>