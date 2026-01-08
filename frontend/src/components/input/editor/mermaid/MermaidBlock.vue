<template>
	<NodeViewWrapper class="mermaid-block">
		<div
			v-if="isEditing"
			class="mermaid-editor"
		>
			<textarea
				ref="textareaRef"
				v-model="code"
				class="mermaid-textarea"
				@blur="renderDiagram"
			/>
		</div>
		<div
			v-else
			class="mermaid-preview"
			@dblclick="startEditing"
		>
			<div
				ref="mermaidRef"
				class="mermaid-diagram"
			/>
			<div
				v-if="error"
				class="mermaid-error"
			>
				{{ error }}
			</div>
		</div>
	</NodeViewWrapper>
</template>

<script setup lang="ts">
import {ref, onMounted, watch} from 'vue'
import {NodeViewWrapper, nodeViewProps} from '@tiptap/vue-3'
import mermaid from 'mermaid'

const props = defineProps(nodeViewProps)

const code = ref('')
const isEditing = ref(false)
const error = ref('')
const textareaRef = ref<HTMLTextAreaElement>()
const mermaidRef = ref<HTMLDivElement>()

// Initialize mermaid
mermaid.initialize({
	startOnLoad: false,
	theme: 'default',
	securityLevel: 'strict',
})

onMounted(() => {
	// Get initial code from node
	code.value = props.node.textContent || ''
	if (code.value) {
		renderDiagram()
	}
})

watch(() => props.node.textContent, (newContent) => {
	if (!isEditing.value && newContent !== code.value) {
		code.value = newContent || ''
		renderDiagram()
	}
})

function startEditing() {
	isEditing.value = true
	setTimeout(() => {
		textareaRef.value?.focus()
	}, 0)
}

async function renderDiagram() {
	isEditing.value = false
	error.value = ''

	if (!code.value.trim()) {
		return
	}

	try {
		const id = `mermaid-${Math.random().toString(36).substr(2, 9)}`
		const {svg} = await mermaid.render(id, code.value)
		
		if (mermaidRef.value) {
			mermaidRef.value.innerHTML = svg
		}

		// Update node content
		const from = props.getPos()
		if (from === undefined) return
		
		const to = from + props.node.nodeSize
		props.editor.commands.setTextSelection({from, to})
		props.editor.commands.insertContent({
			type: 'mermaid',
			content: [{type: 'text', text: code.value}],
		})
	} catch (err) {
		error.value = err instanceof Error ? err.message : 'Failed to render diagram'
		console.error('Mermaid render error:', err)
	}
}
</script>

<style lang="scss" scoped>
.mermaid-block {
	margin: 1rem 0;
	border: 1px solid var(--grey-200);
	border-radius: 4px;
	overflow: hidden;
}

.mermaid-editor {
	.mermaid-textarea {
		inline-size: 100%;
		min-block-size: 200px;
		padding: 1rem;
		font-family: Monaco, Menlo, 'Ubuntu Mono', Consolas, source-code-pro, monospace;
		font-size: 14px;
		border: none;
		outline: none;
		resize: vertical;
		background: var(--grey-50);
	}
}

.mermaid-preview {
	padding: 1rem;
	cursor: pointer;
	background: var(--white);
	
	&:hover {
		background: var(--grey-50);
	}
}

.mermaid-diagram {
	display: flex;
	justify-content: center;
	align-items: center;
	
	:deep(svg) {
		max-inline-size: 100%;
		block-size: auto;
	}
}

.mermaid-error {
	color: var(--danger);
	padding: 1rem;
	background: var(--danger-light);
	border-radius: 4px;
	margin-block-start: 1rem;
}
</style>
