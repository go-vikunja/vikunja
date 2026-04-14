<template>
	<div class="emoji-items">
		<template v-if="items.length">
			<button
				v-for="(item, index) in items"
				:key="item.shortcode"
				:ref="el => setItemRef(el, index)"
				type="button"
				class="emoji-item"
				:class="{ 'is-selected': index === selectedIndex }"
				@click="selectItem(index)"
			>
				<span class="emoji-glyph">{{ item.emoji }}</span>
				<div class="emoji-info">
					<p class="emoji-shortcode">:{{ item.shortcode }}:</p>
					<p class="emoji-annotation">{{ item.annotation }}</p>
				</div>
			</button>
		</template>
		<div
			v-else
			class="emoji-item no-results"
		>
			{{ $t('input.editor.emoji.empty') }}
		</div>
	</div>
</template>

<script lang="ts" setup>
import {ref, watch, nextTick} from 'vue'
import type {EmojiEntry} from './emojiData'

const props = defineProps<{
	items: EmojiEntry[]
	command: (item: EmojiEntry) => void
}>()

const selectedIndex = ref(0)
const itemEls = ref<HTMLElement[]>([])

function setItemRef(el: Element | null, index: number) {
	if (el instanceof HTMLElement) {
		itemEls.value[index] = el
	}
}

watch(() => props.items, () => {
	selectedIndex.value = 0
	itemEls.value = []
})

watch(selectedIndex, async idx => {
	await nextTick()
	itemEls.value[idx]?.scrollIntoView({block: 'nearest'})
})

function selectItem(index: number) {
	const item = props.items[index]
	if (item) props.command(item)
}

function onKeyDown({event}: {event: KeyboardEvent}): boolean {
	if (props.items.length === 0) return false

	if (event.key === 'ArrowUp') {
		selectedIndex.value = ((selectedIndex.value + props.items.length) - 1) % props.items.length
		return true
	}
	if (event.key === 'ArrowDown') {
		selectedIndex.value = (selectedIndex.value + 1) % props.items.length
		return true
	}
	if (event.key === 'Enter' || event.key === 'Tab') {
		if (event.isComposing) return false
		selectItem(selectedIndex.value)
		return true
	}
	return false
}

defineExpose({onKeyDown})
</script>

<style lang="scss" scoped>
.emoji-items {
	padding: 0.2rem;
	position: relative;
	border-radius: 0.5rem;
	background: var(--white);
	color: var(--grey-900);
	overflow: hidden;
	font-size: 0.9rem;
	box-shadow: var(--shadow-md);
	min-inline-size: 240px;
	max-block-size: 300px;
	overflow-y: auto;
}

.emoji-item {
	display: flex;
	align-items: center;
	margin: 0;
	inline-size: 100%;
	text-align: start;
	background: transparent;
	border-radius: $radius;
	border: 0;
	padding: 0.4rem 0.6rem;
	transition: background-color $transition;

	&.is-selected, &:hover {
		background: var(--grey-100);
		cursor: pointer;
	}

	&.no-results {
		color: var(--grey-500);
		cursor: default;
	}
}

.emoji-glyph {
	font-size: 1.4rem;
	margin-inline-end: 0.75rem;
	flex-shrink: 0;
}

.emoji-info {
	display: flex;
	flex-direction: column;
	min-inline-size: 0;
	flex: 1;

	p {
		margin: 0;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}
}

.emoji-shortcode {
	font-family: monospace;
	font-weight: 500;
	color: var(--grey-800);
}

.emoji-annotation {
	font-size: 0.75rem;
	color: var(--grey-500);
}
</style>
