<template>
	<div
		class="multiselect"
		:class="{'has-search-results': searchResultsVisible}"
		ref="multiselectRoot"
		tabindex="-1"
		@focus="focus"
	>
		<div class="control" :class="{'is-loading': loading || localLoading}">
			<div
				class="input-wrapper input"
				:class="{'has-multiple': hasMultiple}">
				<template v-if="Array.isArray(internalValue)">
					<template v-for="(item, key) in internalValue">
						<slot name="tag" :item="item">
						<span :key="`item${key}`" class="tag ml-2 mt-2">
							{{ label !== '' ? item[label] : item }}
							<a @click="() => remove(item)" class="delete is-small"></a>
						</span>
						</slot>
					</template>
				</template>

				<input
					type="text"
					class="input"
					v-model="query"
					@keyup="search"
					@keyup.enter.exact.prevent="() => createOrSelectOnEnter()"
					:placeholder="placeholder"
					@keydown.down.exact.prevent="() => preSelect(0)"
					ref="searchInput"
					@focus="handleFocus"
				/>
			</div>
		</div>

		<transition name="fade">
			<div class="search-results" :class="{'search-results-inline': inline}" v-if="searchResultsVisible">
				<button
					class="is-fullwidth"
					v-for="(data, key) in filteredSearchResults"
					:key="key"
					:ref="`result-${key}`"
					@keydown.up.prevent="() => preSelect(key - 1)"
					@keydown.down.prevent="() => preSelect(key + 1)"
					@click.prevent.stop="() => select(data)"
				>
					<span>
						<slot name="searchResult" :option="data">
							<span class="search-result">{{ label !== '' ? data[label] : data }}</span>
						</slot>
					</span>
					<span class="hint-text">
						{{ selectPlaceholder }}
					</span>
				</button>

				<button
					v-if="creatableAvailable"
					class="is-fullwidth"
					:ref="`result-${filteredSearchResults.length}`"
					@keydown.up.prevent="() => preSelect(filteredSearchResults.length - 1)"
					@keydown.down.prevent="() => preSelect(filteredSearchResults.length + 1)"
					@keyup.enter.prevent="create"
					@click.prevent.stop="create"
				>
					<span>
						<slot name="searchResult" :option="query">
							<span class="search-result">
								{{ query }}
							</span>
						</slot>
					</span>
					<span class="hint-text">
						{{ createPlaceholder }}
					</span>
				</button>
			</div>
		</transition>

	</div>
</template>

<script lang="ts">
import {defineComponent} from 'vue'
import {i18n} from '@/i18n'
import {closeWhenClickedOutside} from '@/helpers/closeWhenClickedOutside'

export default defineComponent({
	name: 'multiselect',
	data() {
		return {
			query: '',
			searchTimeout: null,
			localLoading: false,
			showSearchResults: false,
			internalValue: null,
		}
	},
	props: {
		// When true, shows a loading spinner
		loading: {
			type: Boolean,
			default() {
				return false
			},
		},
		// The placeholder of the search input
		placeholder: {
			type: String,
			default() {
				return ''
			},
		},
		// The search results where the @search listener needs to put the results into
		searchResults: {
			type: Array,
			default() {
				return []
			},
		},
		// The name of the property of the searched object to show the user.
		// If empty the component will show all raw data of an entry.
		label: {
			type: String,
			default() {
				return ''
			},
		},
		// The object with the value, updated every time an entry is selected.
		modelValue: {
			default() {
				return null
			},
		},
		// If true, will provide an "add this as a new value" entry which fires an @create event when clicking on it.
		creatable: {
			type: Boolean,
			default: false,
		},
		// The text shown next to the new value option.
		createPlaceholder: {
			type: String,
			default() {
				return i18n.global.t('input.multiselect.createPlaceholder')
			},
		},
		// The text shown next to an option.
		selectPlaceholder: {
			type: String,
			default() {
				return i18n.global.t('input.multiselect.selectPlaceholder')
			},
		},
		// If true, allows for selecting multiple items. v-model will be an array with all selected values in that case.
		multiple: {
			type: Boolean,
			default: false,
		},
		// If true, displays the search results inline instead of using a dropdown.
		inline: {
			type: Boolean,
			default: false,
		},
		// If true, shows search results when no query is specified.
		showEmpty: {
			type: Boolean,
			default: true,
		},
		// The delay in ms after which the search event will be fired. Used to avoid hitting the network on every keystroke.
		searchDelay: {
			type: Number,
			default() {
				return 200
			},
		},
		closeAfterSelect: {
			type: Boolean,
			default: true,
		},
	},

	/**
	 * Available events:
	 *   @search: Triggered every time the search query input changes
	 *   @select: Triggered every time an option from the search results is selected. Also triggers a change in v-model.
	 *   @create: If nothing or no exact match was found and `creatable` is true, this event is triggered with the current value of the search query.
	 *   @remove: If `multiple` is enabled, this will be fired every time an item is removed from the array of selected items.
	 */
	emits: ['update:modelValue', 'search', 'select', 'create', 'remove'],

	mounted() {
		document.addEventListener('click', this.hideSearchResultsHandler)
	},
	beforeUnmount() {
		document.removeEventListener('click', this.hideSearchResultsHandler)
	},
	watch: {
		modelValue: {
			handler(value) {
				this.setSelectedObject(value)
			},
			immediate: true,
			deep: true,
		},
	},
	computed: {
		searchResultsVisible() {
			if (this.query === '' && !this.showEmpty) {
				return false
			}

			return this.showSearchResults && (
				(this.filteredSearchResults.length > 0) ||
				(this.creatable && this.query !== '')
			)
		},
		creatableAvailable() {
			return this.creatable && this.query !== '' && !this.filteredSearchResults.some(elem => {
				// Don't make create available if we have an exact match in our search results.
				if (this.label !== '') {
					return elem[this.label] === this.query
				}

				return elem === this.query
			})
		},
		filteredSearchResults() {
			if (this.multiple && this.internalValue !== null && Array.isArray(this.internalValue)) {
				return this.searchResults.filter(item => !this.internalValue.some(e => e === item))
			}

			return this.searchResults
		},
		hasMultiple() {
			return this.multiple && Array.isArray(this.internalValue) && this.internalValue.length > 0
		},
	},
	methods: {
		// Searching will be triggered with a 200ms delay to avoid searching on every keyup event.
		search() {

			// Updating the query with a binding does not work on mobile for some reason,
			// getting the value manual does.
			this.query = this.$refs.searchInput.value

			if (this.searchTimeout !== null) {
				clearTimeout(this.searchTimeout)
				this.searchTimeout = null
			}

			this.localLoading = true

			this.searchTimeout = setTimeout(() => {
				this.$emit('search', this.query)
				setTimeout(() => {
					this.localLoading = false
				}, 100) // The duration of the loading timeout of the services
				this.showSearchResults = true
			}, this.searchDelay)
		},
		hideSearchResultsHandler(e) {
			closeWhenClickedOutside(e, this.$refs.multiselectRoot, this.closeSearchResults)
		},
		closeSearchResults() {
			this.showSearchResults = false
		},
		handleFocus() {
			// We need the timeout to avoid the hideSearchResultsHandler hiding the search results right after the input
			// is focused. That would lead to flickering pre-loaded search results and hiding them right after showing.
			setTimeout(() => {
				this.showSearchResults = true
			}, 10)
		},
		select(object) {
			if (this.multiple) {
				if (this.internalValue === null) {
					this.internalValue = []
				}

				this.internalValue.push(object)
			} else {
				this.internalValue = object
			}

			this.$emit('update:modelValue', this.internalValue)
			this.$emit('select', object)
			this.setSelectedObject(object)
			if (this.closeAfterSelect && this.filteredSearchResults.length > 0 && !this.creatableAvailable) {
				this.closeSearchResults()
			}
		},
		setSelectedObject(object, resetOnly = false) {
			this.internalValue = object

			// We assume we're getting an array when multiple is enabled and can therefore leave the query
			// value etc as it is
			if (this.multiple) {
				this.query = ''
				return
			}

			if (object === null) {
				this.query = ''
				return
			}

			if (resetOnly) {
				return
			}

			this.query = this.label !== '' ? object[this.label] : object
		},
		preSelect(index) {
			if (index < 0) {
				this.$refs.searchInput.focus()
				return
			}

			const elems = this.$refs[`result-${index}`]
			if (typeof elems === 'undefined' || elems.length === 0) {
				return
			}

			if (Array.isArray(elems)) {
				elems[0].focus()
				return
			}

			elems.focus()
		},
		create() {
			if (this.query === '') {
				return
			}

			this.$emit('create', this.query)
			this.setSelectedObject(this.query, true)
			this.closeSearchResults()
		},
		createOrSelectOnEnter() {

			if (!this.creatableAvailable && this.searchResults.length === 1) {
				this.select(this.searchResults[0])
				return
			}

			if (!this.creatableAvailable) {
				return
			}

			this.create()
		},
		remove(item) {
			for (const ind in this.internalValue) {
				if (this.internalValue[ind] === item) {
					this.internalValue.splice(ind, 1)
					break
				}
			}

			this.$emit('update:modelValue', this.internalValue)
			this.$emit('remove', item)
		},
		focus() {
			this.$refs.searchInput.focus()
		},
	},
})
</script>

<style lang="scss" scoped>
.multiselect {
	width: 100%;
	position: relative;

	.control.is-loading::after {
		top: .75rem;
	}

	&.has-search-results .input-wrapper {
		border-radius: $radius $radius 0 0;
		border-color: var(--primary) !important;
		background: var(--white) !important;

		&, &:focus-within {
			border-bottom-color: var(--grey-200) !important;
		}
	}

	.input-wrapper {
		padding: 0;
		background: var(--white) !important;
		border-color: var(--grey-200) !important;
		flex-wrap: wrap;
		height: auto;

		&:hover {
			border-color: var(--grey-300) !important;
		}

		.input {
			display: flex;
			max-width: 100%;
			width: 100%;
			align-items: center;
			border: none !important;
			background: transparent;
			height: auto;

			&::placeholder {
				font-style: normal !important;
			}
		}

		&.has-multiple .input {
			max-width: 250px;

			input {
				padding-left: 0;
			}
		}

		&:focus-within {
			border-color: var(--primary) !important;
			background: var(--white) !important;
		}

		.loader {
			margin: 0 .5rem;
		}
	}

	.search-results {
		background: var(--white);
		border-radius: 0 0 $radius $radius;
		border: 1px solid var(--primary);
		border-top: none;

		max-height: 50vh;
		overflow-x: auto;
		position: absolute;
		z-index: 100;
		max-width: 100%;
		min-width: 100%;

		&-inline {
			position: static;
		}

		button {
			background: transparent;
			text-align: left;
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
				padding-left: .5rem;
			}

			&:focus, &:hover {
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
	}
}
</style>