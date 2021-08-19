<template>
	<div
		tabindex="-1"
		@focus="focus"
	>
		<multiselect
			:loading="listUserService.loading"
			:placeholder="$t('task.assignee.placeholder')"
			:disabled="disabled"
			:multiple="true"
			@search="findUser"
			:search-results="foundUsers"
			@select="addAssignee"
			label="username"
			:select-placeholder="$t('task.assignee.selectPlaceholder')"
			v-model="assignees"
			ref="multiselect"
		>
			<template #tag="props">
				<span class="assignee">
					<user :avatar-size="32" :show-username="false" :user="props.item"/>
					<a @click="removeAssignee(props.item)" class="remove-assignee" v-if="!disabled">
						<icon icon="times"/>
					</a>
				</span>
			</template>
		</multiselect>
	</div>
</template>

<script>
import differenceWith from 'lodash/differenceWith'

import UserModel from '../../../models/user'
import ListUserService from '../../../services/listUsers'
import TaskAssigneeService from '../../../services/taskAssignee'
import User from '../../misc/user'
import Multiselect from '@/components/input/multiselect.vue'

export default {
	name: 'editAssignees',
	components: {
		User,
		Multiselect,
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
					this.success({message: this.$t('task.assignee.assignSuccess')})
				})
				.catch(e => {
					this.error(e)
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
					this.success({message: this.$t('task.assignee.assignSuccess')})
				})
				.catch(e => {
					this.error(e)
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
					this.error(e)
				})
		},
		clearAllFoundUsers() {
			this.$set(this, 'foundUsers', [])
		},
		focus() {
			this.$refs.multiselect.focus()
		},
	},
}
</script>
