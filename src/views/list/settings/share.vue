<template>
	<create-edit
		:title="$t('list.share.header')"
		primary-label=""
	>
		<component
			:id="list.id"
			:is="manageUsersComponent"
			:userIsAdmin="userIsAdmin"
			shareType="user"
			type="list"/>
		<component
			:id="list.id"
			:is="manageTeamsComponent"
			:userIsAdmin="userIsAdmin"
			shareType="team"
			type="list"/>

		<link-sharing :list-id="$route.params.listId" v-if="linkSharingEnabled" class="mt-4"/>
	</create-edit>
</template>

<script>
import ListService from '@/services/list'
import ListModel from '@/models/list'
import {CURRENT_LIST} from '@/store/mutation-types'

import CreateEdit from '@/components/misc/create-edit.vue'
import LinkSharing from '@/components/sharing/linkSharing.vue'
import userTeam from '@/components/sharing/userTeam.vue'

export default {
	name: 'list-setting-share',
	data() {
		return {
			list: ListModel,
			listService: ListService,
			manageUsersComponent: '',
			manageTeamsComponent: '',
		}
	},
	components: {
		CreateEdit,
		LinkSharing,
		userTeam,
	},
	computed: {
		linkSharingEnabled() {
			return this.$store.state.config.linkSharingEnabled
		},
		userIsAdmin() {
			return this.list.owner && this.list.owner.id === this.$store.state.auth.info.id
		},
	},
	created() {
		this.listService = new ListService()
		this.loadList()
	},
	methods: {
		loadList() {
			const list = new ListModel({id: this.$route.params.listId})

			this.listService.get(list)
				.then(r => {
					this.$set(this, 'list', r)
					this.$store.commit(CURRENT_LIST, r)
					// This will trigger the dynamic loading of components once we actually have all the data to pass to them
					this.manageTeamsComponent = 'userTeam'
					this.manageUsersComponent = 'userTeam'
					this.setTitle(this.$t('list.share.title', {list: this.list.title}))
				})
				.catch(e => {
					this.error(e)
				})
		},
	},
}
</script>
