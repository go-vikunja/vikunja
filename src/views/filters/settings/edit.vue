<template>
	<create-edit
		:title="$t('filters.edit.title')"
		primary-icon=""
		:primary-label="$t('misc.save')"
		@primary="save"
		:tertary="$t('misc.delete')"
		@tertary="$router.push({ name: 'filter.list.settings.delete', params: { id: $route.params.listId } })"
	>
		<form @submit.prevent="save()">
			<div class="field">
				<label class="label" for="title">{{ $t('filters.attributes.title') }}</label>
				<div class="control">
					<input
						:class="{ 'disabled': filterService.loading}"
						:disabled="filterService.loading || null"
						@keyup.enter="save"
						class="input"
						id="title"
						:placeholder="$t('filters.attributes.titlePlaceholder')"
						type="text"
						v-focus
						v-model="filter.title"/>
				</div>
			</div>
			<div class="field">
				<label class="label" for="description">{{ $t('filters.attributes.description') }}</label>
				<div class="control">
					<editor
						:class="{ 'disabled': filterService.loading}"
						:disabled="filterService.loading"
						:preview-is-default="false"
						id="description"
						:placeholder="$t('filters.attributes.descriptionPlaceholder')"
						v-model="filter.description"
					/>
				</div>
			</div>
			<div class="field">
				<label class="label" for="filters">{{ $t('filters.title') }}</label>
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
import ErrorComponent from '@/components/misc/error.vue'
import LoadingComponent from '@/components/misc/loading.vue'
import CreateEdit from '@/components/misc/create-edit.vue'

import SavedFilterModel from '@/models/savedFilter'
import SavedFilterService from '@/services/savedFilter'
import ListModel from '@/models/list'
import Filters from '@/components/list/partials/filters.vue'
import {objectToSnakeCase} from '@/helpers/case'

export default {
	name: 'filter-settings-edit',
	data() {
		return {
			filter: SavedFilterModel,
			filterService: new SavedFilterService(),
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
			component: import('@/components/input/editor.vue'),
			loading: LoadingComponent,
			error: ErrorComponent,
			timeout: 60000,
		}),
	},
	watch: {
		// call again the method if the route changes
		'$route': {
			handler: 'loadSavedFilter',
			deep: true,
			immediate: true,
		},
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
				.catch(e => this.$message.error(e))
		},
		save() {
			this.filter.filters = this.filters
			this.filterService.update(this.filter)
				.then(r => {
					this.$store.dispatch('namespaces/loadNamespaces')
					this.$message.success({message: this.$t('filters.attributes.edit.success')})
					this.filter = r
					this.filters = objectToSnakeCase(this.filter.filters)
					this.$router.back()
				})
				.catch(e => this.$message.error(e))
		},
	},
}
</script>
