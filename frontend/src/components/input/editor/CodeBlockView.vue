<template>
	<NodeViewWrapper
		as="pre"
		class="code-block"
	>
		<BaseButton
			v-tooltip="t('misc.copy')"
			class="code-block__copy d-print-none"
			contenteditable="false"
			:aria-label="t('misc.copy')"
			@mousedown.stop
			@click.stop="copy(node.textContent)"
			@dblclick.stop
			@keydown.stop
		>
			<Icon icon="copy" />
		</BaseButton>
		<NodeViewContent
			as="code"
			:class="node.attrs.language ? extension.options.languageClassPrefix + node.attrs.language : undefined"
		/>
	</NodeViewWrapper>
</template>

<script setup lang="ts">
import {nodeViewProps, NodeViewContent, NodeViewWrapper} from '@tiptap/vue-3'
import {useI18n} from 'vue-i18n'

import BaseButton from '@/components/base/BaseButton.vue'
import {useCopyToClipboard} from '@/composables/useCopyToClipboard'

defineProps(nodeViewProps)

const {t} = useI18n({useScope: 'global'})
const copy = useCopyToClipboard()
</script>

<style lang="scss">
.tiptap pre.code-block {
	position: relative;
	padding-inline-end: 3rem;

	.code-block__copy {
		position: absolute;
		inset-block-start: .5rem;
		inset-inline-end: .5rem;
		display: inline-flex;
		align-items: center;
		justify-content: center;
		inline-size: 1.5rem;
		block-size: 1.5rem;
		border-radius: $radius;
		background: var(--grey-200);
		color: var(--grey-700);
		font-size: .875rem;

		&:hover {
			background: var(--grey-300);
		}

		&:focus-visible {
			outline: 2px solid var(--primary);
			outline-offset: 2px;
		}
	}
}
</style>
