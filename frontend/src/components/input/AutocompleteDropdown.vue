<script setup lang="ts">
import {type ComponentPublicInstance, computed, nextTick, ref, watch} from 'vue'

const TAB = 9,
	ENTER = 13,
	ESCAPE = 27,
	ARROW_UP = 38,
	ARROW_DOWN = 40

type state = 'unfocused' | 'focused'

const selectedIndex = ref(-1)
const state = ref<state>('unfocused')
const val = ref<string>('')
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

const emit = defineEmits(['blur'])

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
	options: any[],
	suggestion?: string,
	maxHeight?: number,
}>(), {
	maxHeight: 200,
})

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

function setState(stateName: state) {
	state.value = stateName
	if (stateName === 'unfocused') {
		emit('blur')
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
			select(-1)
			break
		case ARROW_DOWN:
			e.preventDefault()
			select(1)
			break
	}
}

const resultRefs = ref<(HTMLElement | null)[]>([])

function setResultRefs(el: Element | ComponentPublicInstance | null, index: number) {
	resultRefs.value[index] = el as (HTMLElement | null)
}

function select(offset: number) {

	let index = selectedIndex.value + offset

	if (!isFinite(index)) {
		index = 0
	}

	if (index >= props.options.length) {
		// At the last index, now moving back to the top
		index = 0
	}

	if (index < 0) {
		// Arrow up but we're already at the top
		index = props.options.length - 1
	}
	let elems = resultRefs.value[index]
	if (
		typeof elems === 'undefined'
	) {
		return
	}

	selectedIndex.value = index
	updateSuggestionScroll()

	if (Array.isArray(elems)) {
		elems[0].focus()
		return
	}
	elems?.focus()
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
			<slot
				name="input"
				:spacerText
				:placeholderText
				:onUpdateField
				:onFocusField
				:onKeydown
			>
				<div class="spacer">{{ spacerText }}</div>
				<div class="placeholder">{{ placeholderText }}</div>
				<textarea class="field"
						  @input="onUpdateField"
						  @focus="onFocusField"
						  @keydown="onKeydown"
						  :class="state"
						  :value="val"
						  ref="editorRef"></textarea>
			</slot>
		</div>
		<div class="suggestion-list" v-if="state === 'focused' && options.length">
			<div v-if="options && options.length" class="scroll-list">
				<div
					class="items"
					ref="suggestionScrollerRef"
					@keydown="onKeydown"
				>
					<button
						v-for="(item, index) in options"
						class="item"
						@click="onSelectValue(item)"
						:class="{ selected: index === selectedIndex }"
						:key="item"
						:ref="(el: Element | ComponentPublicInstance | null) => setResultRefs(el, index)"
					>
						<slot
							name="result"
							:item
							:selected="index === selectedIndex"
						>
							{{ item }}
						</slot>
					</button>
				</div>
			</div>
		</div>
	</div>
</template>

<style scoped lang="scss">
.autocomplete {
	position: relative;
	
	.suggestion-list {
		position: absolute;
		
		background: var(--white);
		border-radius: 0 0 var(--input-radius) var(--input-radius);
		border: 1px solid var(--primary);
		border-top: none;

		max-height: 50vh;
		overflow-x: auto;
		z-index: 100;
		max-width: 100%;
		min-width: 100%;
		margin-top: -2px;
		
		button {
			width: 100%;
			background: transparent;
			border: 0;

			font-size: .9rem;
			width: 100%;
			color: var(--grey-800);
			text-align: left;
			box-shadow: none;
			text-transform: none;
			font-family: $family-sans-serif;
			font-weight: normal;
			padding: .5rem .75rem;
			border: none;
			cursor: pointer;

			&:focus,
			&:hover {
				background: var(--grey-100);
				box-shadow: none !important;
			}

			&:active {
				background: var(--grey-100);
			}
		}
	}
}
</style>
