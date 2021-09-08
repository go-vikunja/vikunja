<template>
	<modal
		@close="$router.back()"
		@submit="deleteSavedFilter()"
	>
		<template #header><span>{{ $t('filters.delete.header') }}</span></template>
		
		<template #text>
			<p>{{ $t('filters.delete.text') }}</p>
		</template>
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
			filterService: new SavedFilterService(),
		}
	},
	methods: {
		deleteSavedFilter() {
			// We assume the listId in the route is the pseudolist
			const list = new ListModel({id: this.$route.params.listId})
			const filter = new SavedFilterModel({id: list.getSavedFilterId()})

			this.filterService.delete(filter)
				.then(() => {
					this.$store.dispatch('namespaces/loadNamespaces')
					this.$message.success({message: this.$t('filters.delete.success')})
					this.$router.push({name: 'namespaces.index'})
				})
				.catch(e => this.$message.error(e))
		},
	},
}
</script>
