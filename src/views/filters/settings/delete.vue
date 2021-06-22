<template>
	<modal
		@close="$router.back()"
		@submit="deleteSavedFilter()"
	>
		<span slot="header">Delete this saved filter</span>
		<p slot="text">
			Are you sure you want to delete this saved filter?
		</p>
	</modal>
</template>

<script>
import SavedFilterModel from '@/models/savedFilter'
import SavedFilterService from '@/services/savedFilter'
import ListModel from '@/models/list'

export default {
	name: 'filter-settings-delete',
	data() {
		return {
			filterService: SavedFilterService,
		}
	},
	created() {
		this.filterService = new SavedFilterService()
	},
	methods: {
		deleteSavedFilter() {
			// We assume the listId in the route is the pseudolist
			const list = new ListModel({id: this.$route.params.listId})
			const filter = new SavedFilterModel({id: list.getSavedFilterId()})

			this.filterService.delete(filter)
				.then(() => {
					this.$store.dispatch('namespaces/loadNamespaces')
					this.success({message: 'The filter was deleted successfully.'})
					this.$router.push({name: 'namespaces.index'})
				})
				.catch(e => this.error(e))
		},
	},
}
</script>
