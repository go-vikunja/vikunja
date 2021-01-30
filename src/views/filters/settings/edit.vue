<template>
	<create-edit
		title="Edit This Saved Filter"
		primary-icon=""
		primary-label="Save"
		@primary="save"
		tertary="Delete"
		@tertary="$router.push({ name: 'filter.list.settings.delete', params: { id: $route.params.listId } })"
	>
		<form @submit.prevent="save()">
			<div class="field">
				<label class="label" for="title">Filter Title</label>
				<div class="control">
					<input
						:class="{ 'disabled': filterService.loading}"
						:disabled="filterService.loading"
						@keyup.enter="save"
						class="input"
						id="title"
						placeholder="The title goes here..."
						type="text"
						v-focus
						v-model="filter.title"/>
				</div>
			</div>
			<div class="field">
				<label class="label" for="description">Description</label>
				<div class="control">
					<editor
						:class="{ 'disabled': filterService.loading}"
						:disabled="filterService.loading"
						:preview-is-default="false"
						id="description"
						placeholder="The description goes here..."
						v-model="filter.description"
					/>
				</div>
			</div>
			<div class="field">
				<label class="label" for="filters">Filters</label>
				<div class="control">
					<filters
						:class="{ 'disabled': filterService.loading}"
						:disabled="filterService.loading"
						class="has-no-shadow has-no-border"
						v-model="filters"
					/>
				</div>
			</div>
		</form>
	</create-edit>
</template>

<script>
import ErrorComponent from '@/components/misc/error'
import LoadingComponent from '@/components/misc/loading'
import CreateEdit from '@/components/misc/create-edit'

import SavedFilterModel from '@/models/savedFilter'
import SavedFilterService from '@/services/savedFilter'
import ListModel from '@/models/list'
import Filters from '@/components/list/partials/filters'
import {objectToSnakeCase} from '@/helpers/case'

export default {
	name: 'filter-settings-edit',
	data() {
		return {
			filter: SavedFilterModel,
			filterService: SavedFilterService,
			filters: {
				sort_by: ['done', 'id'],
				order_by: ['asc', 'desc'],
				filter_by: ['done'],
				filter_value: ['false'],
				filter_comparator: ['equals'],
				filter_concat: 'and',
				filter_include_nulls: true,
			},

			showDeleteModal: false,
		}
	},
	components: {
		CreateEdit,
		Filters,
		editor: () => ({
			component: import(/* webpackChunkName: "editor" */ '@/components/input/editor'),
			loading: LoadingComponent,
			error: ErrorComponent,
			timeout: 60000,
		}),
	},
	created() {
		this.filterService = new SavedFilterService()
		this.loadSavedFilter()
	},
	watch: {
		// call again the method if the route changes
		'$route': 'loadSavedFilter',
	},
	methods: {
		loadSavedFilter() {
			// We assume the listId in the route is the pseudolist
			const list = new ListModel({id: this.$route.params.listId})

			this.filter = new SavedFilterModel({id: list.getSavedFilterId()})
			this.filterService.get(this.filter)
				.then(r => {
					this.filter = r
					this.filters = objectToSnakeCase(this.filter.filters)
				})
				.catch(e => this.error(e, this))
		},
		save() {
			this.filter.filters = this.filters
			this.filterService.update(this.filter)
				.then(r => {
					this.$store.dispatch('namespaces/loadNamespaces')
					this.success({message: 'The filter was saved successfully.'}, this)
					this.filter = r
					this.filters = objectToSnakeCase(this.filter.filters)
				})
				.catch(e => this.error(e, this))
		},
	},
}
</script>
