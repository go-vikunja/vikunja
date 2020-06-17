<template>
	<multiselect
			:multiple="true"
			:close-on-select="false"
			:clear-on-select="true"
			:options-limit="300"
			:hide-selected="true"
			v-model="assignees"
			:options="foundUsers"
			:searchable="true"
			:loading="listUserService.loading"
			:internal-search="true"
			@search-change="findUser"
			@select="addAssignee"
			placeholder="Type to assign a user..."
			label="username"
			track-by="id"
			select-label="Assign this user"
			:showNoOptions="false"
		>
		<template slot="tag" slot-scope="{ option }">
			<user :user="option" :show-username="false" :avatar-size="30"/>
			<a @click="removeAssignee(option)" class="remove-assignee">
				<icon icon="times"/>
			</a>
		</template>
		<template slot="clear" slot-scope="props">
			<div class="multiselect__clear" v-if="newAssignee !== null && newAssignee.id !== 0"
				@mousedown.prevent.stop="clearAllFoundUsers(props.search)"></div>
		</template>
		<span slot="noResult">No user found. Consider changing the search query.</span>
	</multiselect>
</template>

<script>
	import {differenceWith} from 'lodash'
	import multiselect from 'vue-multiselect'

	import UserModel from '../../../models/user'
	import ListUserService from '../../../services/listUsers'
	import TaskAssigneeService from '../../../services/taskAssignee'
	import User from '../../misc/user'

	export default {
		name: 'editAssignees',
		components: {
			User,
			multiselect,
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
			initialAssignees: {
				type: Array,
				default: () => [],
			}
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
			this.assignees = this.initialAssignees
			this.listUserService = new ListUserService()
			this.newAssignee = new UserModel()
			this.taskAssigneeService = new TaskAssigneeService()
		},
		watch: {
			initialAssignees(newVal) {
				this.assignees = newVal
			}
		},
		methods: {
			addAssignee(user) {
				this.$store.dispatch('tasks/addAssignee', {user: user, taskId: this.taskId})
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
