<template>
	<div
		ref="multiselectRoot"
		class="multiselect"
		:class="{'has-search-results': searchResultsVisible, 'is-disabled': disabled}"
		tabindex="-1"
		:aria-disabled="disabled"
		@focus="focus"
	>
		<div
			class="control"
			:class="{'is-loading': loading || localLoading}"
		>
			<div
				class="input-wrapper input"
				:class="{'has-multiple': hasMultiple, 'has-removal-button': removalAvailable && !disabled}"
			>
				<slot
					v-if="Array.isArray(internalValue)"
					name="items"
					:items="internalValue"
					:remove="remove"
				>
					<template v-for="(item, key) in internalValue">
						<slot
							name="tag"
							:item="item"
						>
							<span
								:key="`item${key}`"
								class="tag mis-2 mbs-2"
							>
								{{ label !== '' ? item[label] : item }}
								<BaseButton
									v-if="!disabled"
									class="delete is-small"
									@click="() => remove(item)"
								/>
							</span>
						</slot>
					</template>
				</slot>
				
				<input
					v-if="!disabled"
					:id="id"
					ref="searchInput"
					v-model="query"
					type="text"
					class="input"
					:name="name"
					:placeholder="placeholder"
					:autocomplete="autocompleteEnabled ? undefined : 'off'"
					:spellcheck="autocompleteEnabled ? undefined : 'false'"
					@keyup="search"
					@keyup.enter.exact.prevent="() => createOrSelectOnEnter()"
					@keydown.down.exact.prevent="() => preSelect(0)"
					@focus="handleFocus"
				>
				<BaseButton 
					v-if="removalAvailable && !disabled"
					class="removal-button"
					@click="resetSelectedValue"
				>
					<Icon icon="times" />
				</BaseButton>
			</div>
		</div>

		<CustomTransition name="fade">
			<div
				v-if="searchResultsVisible"
				class="search-results"
				:class="{'search-results-inline': inline}"
			>
				<BaseButton
					v-for="(data, index) in filteredSearchResults"
					:key="index"
					:ref="(el) => setResult(el, index)"
					class="search-result-button is-fullwidth"
					@keydown.up.prevent="() => preSelect(index - 1)"
					@keydown.down.prevent="() => preSelect(index + 1)"
					@click.prevent.stop="() => select(data)"
				>
					<span>
						<slot
							name="searchResult"
							:option="data"
						>
							<span class="search-result">{{ label !== '' ? data[label] : data }}</span>
						</slot>
					</span>
					<span class="hint-text">
						{{ selectPlaceholder }}
					</span>
				</BaseButton>

				<BaseButton
					v-if="creatableAvailable"
					:ref="(el) => setResult(el, filteredSearchResults.length)"
					class="search-result-button is-fullwidth"
					@keydown.up.prevent="() => preSelect(filteredSearchResults.length - 1)"
					@keydown.down.prevent="() => preSelect(filteredSearchResults.length + 1)"
					@keyup.enter.prevent="create"
					@click.prevent.stop="create"
				>
					<span>
						<slot
							name="searchResult"
							:option="query"
						>
							<span class="search-result">
								{{ query }}
							</span>
						</slot>
					</span>
					<span class="hint-text">
						{{ createPlaceholder }}
					</span>
				</BaseButton>
			</div>
		</CustomTransition>
	</div>
</template>

<script setup lang="ts" generic="T extends Record<string, unknown>">
import {computed, onBeforeUnmount, onMounted, ref, toRefs, watch, type ComponentPublicInstance} from 'vue'
import {useI18n} from 'vue-i18n'

import {closeWhenClickedOutside} from '@/helpers/closeWhenClickedOutside'

import BaseButton from '@/components/base/BaseButton.vue'
import CustomTransition from '@/components/misc/CustomTransition.vue'

const props = withDefaults(defineProps<{
	/** The object with the value, updated every time an entry is selected */
	modelValue: T | T[] | null,
	/** When true, shows a loading spinner */
	loading?: boolean
	/** The placeholder of the search input */
	placeholder?: string
	/** The search results where the @search listener needs to put the results into */
	searchResults?: T[]
	/** The name of the property of the searched object to show the user. If empty the component will show all raw data of an entry */
	label?: string
	/** The id attribute of the input element */
	id?: string
	/** The name attribute of the input element */
	name?: string
	/** If true, will provide an 'add this as a new value' entry which  fires an @create event when clicking on it. */
	creatable?: boolean
	/** The text shown next to the new value option. */
	createPlaceholder?: string
	/** The text shown next to an option. */
	selectPlaceholder?: string
	/** If true, allows for selecting multiple items. v-model will be an array with all selected values in that case. */
	multiple?: boolean
	/** If true, displays the search results inline instead of using a dropdown. */
	inline?: boolean
	/** If true, shows search results when no query is specified. */
	showEmpty?: boolean
	/** The delay in ms after which the search event will be fired. Used to avoid hitting the network on every keystroke. */
	searchDelay?: number
	/** If true, closes the dropdown after an entry is selected */
	closeAfterSelect?: boolean
	/** If false, the search input will get the autocomplete="off" attributes attached to it. */
	autocompleteEnabled?: boolean
	/** If true, disables the multiselect input */
	disabled?: boolean
}>(), {
	loading: false,
	placeholder: '',
	searchResults: () => [],
	label: '',
	creatable: false,
	createPlaceholder: () => useI18n().t('input.multiselect.createPlaceholder'),
	selectPlaceholder: () => useI18n().t('input.multiselect.selectPlaceholder'),
	multiple: false,
	inline: false,
	showEmpty: false,
	searchDelay: 200,
	closeAfterSelect: true,
	autocompleteEnabled: true,
	disabled: false,
	id: undefined,
	name: undefined,
})

const emit = defineEmits<{
	'update:modelValue': [value: T | T[] | null],
	/**
	 * Triggered every time the search query input changes
	 */
	'search': [query: string],
	/**
	 * Triggered every time an option from the search results is selected. Also triggers a change in v-model.
	 */
	'select': [value: T],
	/**
	 * If nothing or no exact match was found and `creatable` is true, this event is triggered with the current value of the search query.
	 */
	'create': [query: string],
	/**
	 * If `multiple` is enabled, this will be fired every time an item is removed from the array of selected items.
	 */
	'remove': [value: T],
}>()

function elementInResults(elem: string | T, label: string, query: string): boolean {
	// Don't make create available if we have an exact match in our search results.
	if (label !== '') {
		return (elem as Record<string, unknown>)[label] === query
	}

	return elem === query
}

const query = ref<string | T>('')
const searchTimeout = ref<ReturnType<typeof setTimeout> | null>(null)
const localLoading = ref(false)
const showSearchResults = ref(false)

const internalValue = ref<string | T | T[] | null>(null)

onMounted(() => document.addEventListener('click', hideSearchResultsHandler))
onBeforeUnmount(() => document.removeEventListener('click', hideSearchResultsHandler))

const {modelValue, searchResults} = toRefs(props)

watch(
	modelValue,
	(value) => setSelectedObject(value),
	{
		immediate: true,
		deep: true,
	},
)

const searchResultsVisible = computed(() => {
	if (query.value === '' && !props.showEmpty) {
		return false
	}

	return showSearchResults.value && (
		(filteredSearchResults.value.length > 0) ||
		(props.creatable && query.value !== '')
	)
})

const creatableAvailable = computed(() => {
	const hasResult = filteredSearchResults.value.some((elem) => elementInResults(elem, props.label, query.value as string))
	const hasQueryAlreadyAdded = Array.isArray(internalValue.value) && internalValue.value.some(elem => elementInResults(elem, props.label, query.value))

	return props.creatable
		&& query.value !== ''
		&& !(hasResult || hasQueryAlreadyAdded)
})

const filteredSearchResults = computed(() => {
	const currentInternal = internalValue.value
	if (props.multiple && currentInternal !== null && Array.isArray(currentInternal)) {
		return searchResults.value.filter((item) => !currentInternal.some(e => e === item))
	}

	return searchResults.value
})

const hasMultiple = computed(() => {
	return props.multiple && Array.isArray(internalValue.value) && internalValue.value.length > 0
})

const removalAvailable = computed(() => !props.multiple && internalValue.value !== null && query.value !== '' && !(props.loading || localLoading.value))
function resetSelectedValue() {
	select(null)
}

const searchInput = ref<HTMLInputElement | null>(null)

// Searching will be triggered with a 200ms delay to avoid searching on every keyup event.
function search() {

	// Updating the query with a binding does not work on mobile for some reason,
	// getting the value manual does.
	query.value = searchInput.value?.value || ''

	if (searchTimeout.value !== null) {
		clearTimeout(searchTimeout.value)
		searchTimeout.value = null
	}

	localLoading.value = true

	searchTimeout.value = setTimeout(() => {
		emit('search', query.value)
		setTimeout(() => {
			localLoading.value = false
		}, 100) // The duration of the loading timeout of the services
		showSearchResults.value = true
	}, props.searchDelay)
}

const multiselectRoot = ref<HTMLElement | null>(null)

function hideSearchResultsHandler(e: MouseEvent) {
	closeWhenClickedOutside(e, multiselectRoot.value, closeSearchResults)
}

function closeSearchResults() {
	showSearchResults.value = false
}

function handleFocus() {
	// We need the timeout to avoid the hideSearchResultsHandler hiding the search results right after the input
	// is focused. That would lead to flickering pre-loaded search results and hiding them right after showing.
	setTimeout(() => {
		showSearchResults.value = true
	}, 10)
}

function select(object: T | null) {
	if (props.multiple) {
		if (internalValue.value === null) {
			internalValue.value = []
		}

		internalValue.value.push(object)
	} else {
		internalValue.value = object
	}

	emit('update:modelValue', internalValue.value)
	emit('select', object)
	setSelectedObject(object)
	if (props.closeAfterSelect && filteredSearchResults.value.length > 0 && !creatableAvailable.value) {
		closeSearchResults()
	}
}

function setSelectedObject(object: string | T | null, resetOnly = false) {
	internalValue.value = object

	// We assume we're getting an array when multiple is enabled and can therefore leave the query
	// value etc as it is
	if (props.multiple) {
		query.value = ''
		return
	}

	if (object === null) {
		query.value = ''
		return
	}

	if (resetOnly) {
		return
	}

	query.value = props.label !== '' ? object[props.label] : object
}

const results = ref<(Element | ComponentPublicInstance)[]>([])

function setResult(el: Element | ComponentPublicInstance | null, index: number) {
	if (el === null) {
		delete results.value[index]
	} else {
		results.value[index] = el
	}
}

function preSelect(index: number) {
	if (index < 0) {
		searchInput.value?.focus()
		return
	}

	const elems = results.value[index]
	if (typeof elems === 'undefined' || elems.length === 0) {
		return
	}

	if (Array.isArray(elems)) {
		elems[0].focus()
		return
	}

	elems.focus()
}

function create() {
	if (query.value === '') {
		return
	}

	emit('create', query.value)
	setSelectedObject(query.value, true)
	closeSearchResults()
}

function createOrSelectOnEnter() {
	if (!creatableAvailable.value && searchResults.value.length === 1) {
		select(searchResults.value[0])
		return
	}

	if (!creatableAvailable.value) {
		// Check if there's an exact match for our search term
		const exactMatch = filteredSearchResults.value.find((elem) => elementInResults(elem, props.label, query.value as string))
		if (exactMatch) {
			select(exactMatch)
		}

		return
	}

	create()
}

function remove(item: T) {
	for (const ind in internalValue.value) {
		if (internalValue.value[ind] === item) {
			internalValue.value.splice(ind, 1)
			break
		}
	}

	emit('update:modelValue', internalValue.value)
	emit('remove', item)
}

function focus() {
	searchInput.value?.focus()
}
</script>

<style lang="scss" scoped>
.multiselect {
	inline-size: 100%;
	position: relative;

	.control.is-loading::after {
		inset-block-start: .75rem;
	}

	&.is-disabled {
		.input-wrapper, .input, & {
			cursor: default !important;

			&:focus-within,&:focus-visible, &:hover, &:focus {
				border-color: transparent !important;
				background: transparent !important;
				box-shadow: none;
			}
		}
	}
}

.input-wrapper {
	padding: 0;
	background: var(--white);
	border-color: var(--grey-200);
	flex-wrap: wrap;
	block-size: auto;

	&:hover {
		border-color: var(--grey-300) !important;
	}

	.input {
		display: flex;
		max-inline-size: 100%;
		inline-size: 100%;
		align-items: center;
		border: none !important;
		background: transparent;
		block-size: auto;

		&::placeholder {
			font-style: normal !important;
		}
	}

	&.has-multiple .input {
		max-inline-size: 250px;

		input {
			padding-inline-start: 0;
		}
	}

	&:focus-within {
		border-color: var(--primary) !important;
		background: var(--white) !important;
	}

	// doesn't seem to be used. maybe inside the slot?
	.loader {
		margin: 0 .5rem;
	}
}

.has-search-results .input-wrapper {
	border-radius: $radius $radius 0 0;
	border-color: var(--primary) !important;
	background: var(--white) !important;

	&, &:focus-within {
		border-block-end-color: var(--grey-200) !important;
	}
}

.search-results {
	background: var(--white);
	border-radius: 0 0 $radius $radius;
	border: 1px solid var(--primary);
	border-block-start: none;

	max-block-size: 50vh;
	overflow-x: auto;
	position: absolute;
	z-index: 100;
	max-inline-size: 100%;
	min-inline-size: 100%;
}

.search-results-inline {
	position: static;
}

.search-result-button {
	background: transparent;
	text-align: start;
	box-shadow: none;
	border-radius: 0;
	text-transform: none;
	font-family: $family-sans-serif;
	font-weight: normal;
	padding: .5rem;
	border: none;
	cursor: pointer;
	color: var(--grey-800);

	display: flex;
	justify-content: space-between;
	align-items: center;
	overflow: hidden;

	&:focus,
	&:hover {
		background: var(--grey-100);
		box-shadow: none !important;

		.hint-text {
			color: var(--text);
		}
	}

	&:active {
		background: var(--grey-200);
	}
}

.search-result {
	white-space: nowrap;
	text-overflow: ellipsis;
	overflow: hidden;
	padding: .5rem .75rem;
}


.hint-text {
	font-size: .75rem;
	color: transparent;
	transition: color $transition;
	padding-inline-start: .5rem;
}

.has-removal-button {
	position: relative;
}

.removal-button {
	position: absolute;
	inset-inline-end: .5rem;
	color: var(--danger);
}
</style>
