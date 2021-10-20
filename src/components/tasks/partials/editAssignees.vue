<template>
	<div
		tabindex="-1"
		@focus="focus"
	>
		<multiselect
			:loading="listUserService.loading"
			:placeholder="$t('task.assignee.placeholder')"
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
import {includesById} from '@/helpers/utils'
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
		modelValue: {
			type: Array,
		},
	},
	emits: ['update:modelValue'],
	data() {
		return {
			newAssignee: new UserModel(),
			listUserService: new ListUserService(),
			foundUsers: [],
			assignees: [],
			taskAssigneeService: new TaskAssigneeService(),
		}
	},
	watch: {
		modelValue: {
			handler(value) {
				this.assignees = value
			},
			immediate: true,
			deep: true,
		},
	},
	methods: {
		async addAssignee(user) {
			await this.$store.dispatch('tasks/addAssignee', {user: user, taskId: this.taskId})
			this.$emit('update:modelValue', this.assignees)
			this.$message.success({message: this.$t('task.assignee.assignSuccess')})
		},

		async removeAssignee(user) {
			await this.$store.dispatch('tasks/removeAssignee', {user: user, taskId: this.taskId})

			// Remove the assignee from the list
			for (const a in this.assignees) {
				if (this.assignees[a].id === user.id) {
					this.assignees.splice(a, 1)
				}
			}
			this.$message.success({message: this.$t('task.assignee.unassignSuccess')})
		},

		async findUser(query) {
			if (query === '') {
				this.clearAllFoundUsers()
				return
			}

			const response = await this.listUserService.getAll({listId: this.listId}, {s: query})

			// Filter the results to not include users who are already assigned
			this.foundUsers = response.filter(({id}) => !includesById(this.assignees, id))
		},

		clearAllFoundUsers() {
			this.foundUsers = []
		},

		focus() {
			this.$refs.multiselect.focus()
		},
	},
}
</script>

<style lang="scss" scoped>
.assignee {
	position: relative;

	&:not(:first-child) {
		margin-left: -1.5rem;
	}

	:deep(.user img) {
		border: 2px solid $white;
		margin-right: 0;
	}

	.remove-assignee {
		position: absolute;
		top: 4px;
		left: 2px;
		color: $red;
		background: $white;
		padding: 0 4px;
		display: block;
		border-radius: 100%;
		font-size: .75rem;
		width: 18px;
		height: 18px;
		z-index: 100;
	}
}
</style>