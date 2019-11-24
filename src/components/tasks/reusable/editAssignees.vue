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
		<template slot="tag" slot-scope="{ option, remove }">
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
	import message from '../../../message'
	import multiselect from 'vue-multiselect'

	import UserModel from '../../../models/user'
	import ListUserService from '../../../services/listUsers'
	import TaskAssigneeService from '../../../services/taskAssignee'
	import TaskAssigneeModel from '../../../models/taskAssignee'
	import User from '../../global/user'

	export default {
		name: 'editAssignees',
		components: {
			User,
			multiselect,
		},
		props: {
			taskID: {
				type: Number,
				required: true,
			},
			listID: {
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
				const taskAssignee = new TaskAssigneeModel({user_id: user.id, task_id: this.taskID})
				this.taskAssigneeService.create(taskAssignee)
					.then(() => {
						message.success({message: 'The user was successfully assigned.'}, this)
					})
					.catch(e => {
						message.error(e, this)
					})
			},
			removeAssignee(user) {
				const taskAssignee = new TaskAssigneeModel({user_id: user.id, task_id: this.taskID})
				this.taskAssigneeService.delete(taskAssignee)
					.then(() => {
						// Remove the assignee from the list
						for (const a in this.assignees) {
							if (this.assignees[a].id === user.id) {
								this.assignees.splice(a, 1)
							}
						}
						message.success({message: 'The user was successfully unassigned.'}, this)
					})
					.catch(e => {
						message.error(e, this)
					})
			},
			findUser(query) {
				if (query === '') {
					this.clearAllFoundUsers()
					return
				}

				this.listUserService.getAll({listID: this.listID}, {s: query})
					.then(response => {
						// Filter the results to not include users who are already assigned
						this.$set(this, 'foundUsers', differenceWith(response, this.assignees, (first, second) => {
							return first.id === second.id
						}))
					})
					.catch(e => {
						message.error(e, this)
					})
			},
			clearAllFoundUsers() {
				this.$set(this, 'foundUsers', [])
			},
		},
	}
</script>
