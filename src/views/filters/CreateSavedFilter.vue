<template>
	<div class="modal-mask keyboard-shortcuts-modal">
		<div @click.self="$router.back()" class="modal-container">
			<div class="modal-content">
				<div class="card has-background-white has-no-shadow">
					<header class="card-header">
						<p class="card-header-title">Create A Saved Filter</p>
					</header>
					<div class="card-content content">
						<p>
							A saved filter is a virtual list which is computed from a set of filters each time it is
							accessed. Once created, it will appear in a special namespace.
						</p>
						<div class="field">
							<label class="label" for="title">Title</label>
							<div class="control">
								<input
									v-model="savedFilter.title"
									:class="{ 'disabled': savedFilterService.loading}"
									:disabled="savedFilterService.loading"
									class="input"
									id="Title"
									placeholder="The saved filter title goes here..."
									type="text"
									v-focus
								/>
							</div>
						</div>
						<div class="field">
							<label class="label" for="description">Description</label>
							<div class="control">
								<editor
									v-model="savedFilter.description"
									:class="{ 'disabled': savedFilterService.loading}"
									:disabled="savedFilterService.loading"
									:preview-is-default="false"
									id="description"
									placeholder="The description goes here..."
									v-if="editorActive"
								/>
							</div>
						</div>
						<div class="field">
							<label class="label" for="filters">Filters</label>
							<div class="control">
								<filters
									:class="{ 'disabled': savedFilterService.loading}"
									:disabled="savedFilterService.loading"
									class="has-no-shadow has-no-border"
									v-model="filters"
								/>
							</div>
						</div>
						<button
							:class="{ 'disabled': savedFilterService.loading}"
							:disabled="savedFilterService.loading"
							@click="create()"
							class="button is-primary is-fullwidth">
							Create new saved filter
						</button>
					</div>
				</div>
			</div>
		</div>
	</div>
</template>

<script>
import LoadingComponent from '@/components/misc/loading'
import ErrorComponent from '@/components/misc/error'
import Filters from '@/components/list/partials/filters'
import SavedFilterService from '@/services/savedFilter'
import SavedFilterModel from '@/models/savedFilter'

export default {
	name: 'CreateSavedFilter',
	data() {
		return {
			editorActive: false,
			filters: {
				sort_by: ['done', 'id'],
				order_by: ['asc', 'desc'],
				filter_by: ['done'],
				filter_value: ['false'],
				filter_comparator: ['equals'],
				filter_concat: 'and',
				filter_include_nulls: true,
			},
			savedFilterService: SavedFilterService,
			savedFilter: SavedFilterModel,
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
		this.editorActive = false
		this.$nextTick(() => this.editorActive = true)

		this.savedFilterService = new SavedFilterService()
		this.savedFilter = new SavedFilterModel()
	},
	methods: {
		create() {
			this.savedFilter.filters = this.filters
			this.savedFilterService.create(this.savedFilter)
				.then(r => {
					this.$store.dispatch('namespaces/loadNamespaces')
					this.$router.push({name: 'list.index', params: {listId: r.getListId()}})
				})
				.catch(e => this.error(e, this))
		},
	},
}
</script>
