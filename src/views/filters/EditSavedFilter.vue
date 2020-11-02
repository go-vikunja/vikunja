<template>
	<div :class="{ 'is-loading': filterService.loading}" class="loader-container edit-list is-max-width-desktop">
		<div class="card">
			<header class="card-header">
				<p class="card-header-title">
					Edit Saved Filter
				</p>
			</header>
			<div class="card-content">
				<div class="content">
					<form @submit.prevent="save()">
						<div class="field">
							<label class="label" for="listtext">Filter Name</label>
							<div class="control">
								<input
									:class="{ 'disabled': filterService.loading}"
									:disabled="filterService.loading"
									@keyup.enter="save"
									class="input"
									id="listtext"
									placeholder="The list title goes here..."
									type="text"
									v-focus
									v-model="filter.title"/>
							</div>
						</div>
						<div class="field">
							<label class="label" for="listdescription">Description</label>
							<div class="control">
								<editor
									:class="{ 'disabled': filterService.loading}"
									:disabled="filterService.loading"
									:preview-is-default="false"
									id="listdescription"
									placeholder="The lists description goes here..."
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

					<div class="columns bigbuttons">
						<div class="column">
							<button :class="{ 'is-loading': filterService.loading}" @click="save()"
									class="button is-primary is-fullwidth">
								Save
							</button>
						</div>
						<div class="column is-1">
							<button :class="{ 'is-loading': filterService.loading}" @click="showDeleteModal = true"
									class="button is-danger is-fullwidth">
								<span class="icon is-small">
									<icon icon="trash-alt"/>
								</span>
							</button>
						</div>
					</div>
				</div>
			</div>
		</div>

		<modal
			@close="showDeleteModal = false"
			@submit="deleteList()"
			v-if="showDeleteModal">
			<span slot="header">Delete this saved filter</span>
			<p slot="text">
				Are you sure you want to delete this saved filter?
			</p>
		</modal>
	</div>
</template>

<script>
import ErrorComponent from '../../components/misc/error'
import LoadingComponent from '../../components/misc/loading'

import SavedFilterModel from '@/models/savedFilter'
import SavedFilterService from '@/services/savedFilter'
import ListModel from '@/models/list'
import Filters from '@/components/list/partials/filters'
import {objectToSnakeCase} from '@/helpers/case'

export default {
	name: 'EditFilter',
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
		Filters,
		editor: () => ({
			component: import(/* webpackChunkName: "editor" */ '../../components/input/editor'),
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
			const list = new ListModel({id: this.$route.params.id})

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
