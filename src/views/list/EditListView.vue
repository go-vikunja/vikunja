<template>
	<div>
		<edit-filter v-if="isSavedFilter"/>
		<edit-list v-else/>
	</div>
</template>

<script>
import EditList from '@/views/list/EditList'
import EditFilter from '@/views/filters/EditSavedFilter'
import {mapState} from 'vuex'
import {getSavedFilterIdFromListId} from '@/helpers/savedFilter'

export default {
	name: 'EditListView',
	components: {
		EditFilter,
		EditList,
	},
	computed: mapState({
		isSavedFilter: state => getSavedFilterIdFromListId(state.currentList.id) > 0
	})
}
</script>

