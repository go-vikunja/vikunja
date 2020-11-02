<template>
	<multiselect
		:clear-on-select="true"
		:close-on-select="false"
		:disabled="disabled"
		:hide-selected="true"
		:internal-search="true"
		:loading="listUserService.loading"
		:multiple="true"
		:options="foundUsers"
		:options-limit="300"
		:searchable="true"
		:showNoOptions="false"
		@search-change="findUser"
		@select="addAssignee"
		label="username"
		placeholder="Type to assign a user..."
		select-label="Assign this user"
		track-by="id"
		v-model="assignees"
	>
		<template slot="tag" slot-scope="{ option }">
			<user :avatar-size="30" :show-username="false" :user="option"/>
			<a @click="removeAssignee(option)" class="remove-assignee" v-if="!disabled">
				<icon icon="times"/>
			</a>
		</template>
		<template slot="clear" slot-scope="props">
			<div
				@mousedown.prevent.stop="clearAllFoundUsers(props.search)"
				class="multiselect__clear"
				v-if="newAssignee !== null && newAssignee.id !== 0"></div>
		</template>
		<span slot="noResult">No user found. Consider changing the search query.</span>
	</multiselect>
</template>

<script>
import differenceWith from 'lodash/differenceWith'

import UserModel from '../../../models/user'
import ListUserService from '../../../services/listUsers'
import TaskAssigneeService from '../../../services/taskAssignee'
import User from '../../misc/user'
import LoadingComponent from '../../misc/loading'
import ErrorComponent from '../../misc/error'

export default {
	name: 'editAssignees',
	components: {
		User,
		multiselect: () => ({
			component: import(/* webpackChunkName: "multiselect" */ 'vue-multiselect'),
			loading: LoadingComponent,
			error: ErrorComponent,
			timeout: 60000,
		}),
	},
	props: {
		taskId: {
			type: Number,
			required: true,
		},
		listId: {
			type: Number,
			required: true,
		},
		disabled: {
			default: false,
		},
		value: {
			type: Array,
		},
	},
	data() {
		return {
			newAssignee: UserModel,
			listUserService: ListUserService,
			foundUsers: [],
			assignees: [],
			taskAssigneeService: TaskAssigneeService,
		}
	},
	created() {
		this.assignees = this.value
		this.listUserService = new ListUserService()
		this.newAssignee = new UserModel()
		this.taskAssigneeService = new TaskAssigneeService()
	},
	watch: {
		value(newVal) {
			this.assignees = newVal
		},
	},
	methods: {
		addAssignee(user) {
			this.$store.dispatch('tasks/addAssignee', {user: user, taskId: this.taskId})
				.then(() => {
					this.$emit('input', this.assignees)
				})
				.catch(e => {
					this.error(e, this)
				})
		},
		removeAssignee(user) {
			this.$store.dispatch('tasks/removeAssignee', {user: user, taskId: this.taskId})
				.then(() => {
					// Remove the assignee from the list
					for (const a in this.assignees) {
						if (this.assignees[a].id === user.id) {
							this.assignees.splice(a, 1)
						}
					}
				})
				.catch(e => {
					this.error(e, this)
				})
		},
		findUser(query) {
			if (query === '') {
				this.clearAllFoundUsers()
				return
			}

			this.listUserService.getAll({listId: this.listId}, {s: query})
				.then(response => {
					// Filter the results to not include users who are already assigned
					this.$set(this, 'foundUsers', differenceWith(response, this.assignees, (first, second) => {
						return first.id === second.id
					}))
				})
				.catch(e => {
					this.error(e, this)
				})
		},
		clearAllFoundUsers() {
			this.$set(this, 'foundUsers', [])
		},
	},
}
</script>
