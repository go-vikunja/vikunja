<script setup lang="ts">
import {computed, nextTick, ref, watch} from 'vue'

const TAB = 9,
	ENTER = 13,
	ESCAPE = 27,
	ARROW_UP = 38,
	ARROW_DOWN = 40

type state = 'unfocused' | 'focused'

const selectedIndex = ref(0)
const state = ref<state>('unfocused')
const val = ref<string>('')
const isResizing = ref(false)
const model = defineModel<string>()

const suggestionScrollerRef = ref<HTMLInputElement | null>(null)
const containerRef = ref<HTMLInputElement | null>(null)
const editorRef = ref<HTMLInputElement | null>(null)

watch(
	() => model.value,
	newValue => {
		val.value = newValue
	},
)

const placeholderText = computed(() => {
	const value = (model.value || '').replace(/[\n\r\t]/gi, ' ')

	if (state.value === 'unfocused') {
		return value ? '' : props.suggestion
	}

	if (!value || !value.trim()) {
		return props.suggestion
	}

	return lookahead()
})

const spacerText = computed(() => {
	const value = (model.value || '').replace(/[\n\r\t]/gi, ' ')

	if (!value || !value.trim()) {
		return props.suggestion
	}

	return value
})

const props = withDefaults(defineProps<{
	options: string[],
	suggestion?: string,
	maxHeight?: number,
}>(), {
	maxHeight: 200,
})
function addSelectedIndex(offset: number) {
	let nextIndex = Math.max(
		0,
		Math.min(selectedIndex.value + offset, props.options.length - 1),
	)
	if (!isFinite(nextIndex)) {
		nextIndex = 0
	}
	selectedIndex.value = nextIndex
	updateSuggestionScroll()
}

function highlight(words: string, query: string) {
	return (words || '').replace(new RegExp(query, 'i'), '<mark class="scroll-term">' + query + '</mark>')
}

function lookahead() {
	if (!props.options.length) {
		return model.value
	}
	const index = Math.max(0, Math.min(selectedIndex.value, props.options.length - 1))
	const match = props.options[index]
	return model.value + (match ? match.substring(model.value?.length) : '')
}

function updateSuggestionScroll() {
	nextTick(() => {
		const scroller = suggestionScrollerRef.value
		const selectedItem = scroller?.querySelector('.selected')
		scroller.scrollTop = selectedItem ? selectedItem.offsetTop : 0
	})
}

function updateScrollWindowSize() {
	if (isResizing.value) {
		return
	}

	isResizing.value = true

	nextTick(() => {
		isResizing.value = false

		const scroller = suggestionScrollerRef.value
		const parent = containerRef.value
		if (scroller) {
			const rect = parent.getBoundingClientRect()
			const pxTop = rect.top
			const pxBottom = window.innerHeight - rect.bottom
			const maxHeight = Math.max(pxTop, pxBottom, props.maxHeight)
			const isReversed = pxBottom < props.maxHeight && pxTop > pxBottom
			scroller.style.maxHeight = Math.min(isReversed ? pxTop : pxBottom, props.maxHeight) + 'px'
			scroller.parentNode.style.transform =
				isReversed ? 'translateY(-100%) translateY(-1.4rem)'
					: 'translateY(.4rem)'
		}
	})
}

function setState(stateName: state) {
	state.value = stateName
	if (stateName === 'unfocused') {
		editorRef.value.blur()
	} else {
		updateScrollWindowSize()
	}
}

function onFocusField(e) {
	setState('focused')
}

function onKeydown(e) {
	switch (e.keyCode || e.which) {
		case ESCAPE:
			e.preventDefault()
			setState('unfocused')
			break
		case ARROW_UP:
			e.preventDefault()
			addSelectedIndex(-1)
			break
		case ARROW_DOWN:
			e.preventDefault()
			addSelectedIndex(1)
			break
		case ENTER:
		case TAB:
			e.preventDefault()
			onSelectValue(lookahead() || model.value)
			break
	}
}

function onSelectValue(value) {
	model.value = value
	selectedIndex.value = 0
	setState('unfocused')
}

function onUpdateField(e) {
	setState('focused')
	model.value = e.currentTarget.value
}


</script>

<template>
	<div class="autocomplete" ref="containerRef">
		<div class="entry-box">
			<div class="spacer">{{ spacerText }}</div>
			<div class="placeholder">{{ placeholderText }}</div>
			<textarea class="field"
					  @input="onUpdateField"
					  @focus="onFocusField"
					  @keydown="onKeydown"
					  :class="state"
					  :value="val"
					  ref="editorRef"></textarea>
		</div>
		<div class="suggestion-list" v-if="state === 'focused' && options.length">
			<div v-if="options && options.length" class="scroll-list">
				<div class="items" ref="suggestionScrollerRef" @keydown="onKeydown">
					<button 
							v-for="(item, index) in options"
							class="item"
							@click="onSelectValue(item)"
							:class="{ selected: index === selectedIndex }"
							:key="item"
							v-html="highlight(item, val)"></button>
				</div>
			</div>
		</div>
	</div>
</template>

<style scoped lang="scss">

.autocomplete {
	position: relative;

	* {
		font-size: 1rem;
		font-family: Consolas, Lucida Console, Courier New, monospace;
	}

	.entry-box {
		position: relative;
		width: 180px;
	}

	.spacer,
	.placeholder,
	.field {
		border: none;
		height: 100%;
		padding: .1rem .2rem;
		width: 100%;
	}

	.spacer {
		min-height: 1rem;
		visibility: hidden;
	}

	.placeholder {
		user-select: none;
		pointer-events: none;
		opacity: 0.4;
		z-index: 2;
	}

	.field {
		z-index: 1;

		&.focused {
			color: blue;
		}
	}

	.placeholder,
	.field {
		left: 0;
		outline: none;
		overflow: hidden;
		position: absolute;
		resize: none;
		top: 0;
	}

	.suggestion-list {
		position: absolute;
		width: 100%;
	}

	.scroll-list {
		position: absolute;
		width: 100%;
		border: solid 1px lightgray;
		box-shadow: 0 1px 3px rgba(0, 0, 0, 0.12), 0 1px 2px rgba(0, 0, 0, 0.24);
	}

	.items {
		margin: 0;
		max-height: 150px;
		overflow-y: auto;
		overflow-x: hidden;
		padding: 0;

		&::-webkit-scrollbar {
			width: 10px;
		}

		&::-webkit-scrollbar-thumb {
			background: #045068;
			border-radius: 20px;
		}

		&::-webkit-scrollbar-track {
			background: #dfe1e5;
		}
	}

	.item {
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
		border: solid 1px transparent;
		background-color: white;
		display: block;
		width: 100%;
		text-align: left;

		&:hover {
			cursor: pointer;
		}

		&:not(.selected):hover {
			background-color: #c1dae2;
			color: black;
		}

		&.selected {
			background-color: #00aee6;
			color: white;
		}
	}

	.scroll-term {
		font-weight: bold;
		background-color: unset;
		color: unset;
	}
}

</style>