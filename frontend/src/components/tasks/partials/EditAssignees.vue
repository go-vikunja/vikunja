<template>
	<Multiselect
		v-model="assignees"
		class="edit-assignees"
		:class="{'has-assignees': assignees.length > 0}"
		:loading="projectUserService.loading"
		:placeholder="$t('task.assignee.placeholder')"
		:multiple="true"
		:search-results="foundUsers"
		label="name"
		:select-placeholder="$t('task.assignee.selectPlaceholder')"
		:autocomplete-enabled="false"
		@search="findUser"
		@select="addAssignee"
	>
		<template #items="{items}">
			<AssigneeList
				:assignees="items"
				:disabled="disabled"
				can-remove
				@remove="removeAssignee"
			/>
		</template>
		<template #searchResult="{option: user}">
			<User
				:avatar-size="24"
				:show-username="true"
				:user="user"
			/>
		</template>
	</Multiselect>
</template>

<script setup lang="ts">
import {ref, shallowReactive, watch, nextTick} from 'vue'
import {useI18n} from 'vue-i18n'

import User from '@/components/misc/User.vue'
import Multiselect from '@/components/input/Multiselect.vue'

import {includesById} from '@/helpers/utils'
import ProjectUserService from '@/services/projectUsers'
import {success} from '@/message'
import {useTaskStore} from '@/stores/tasks'

import type {IUser} from '@/modelTypes/IUser'
import {getDisplayName} from '@/models/user'
import AssigneeList from '@/components/tasks/partials/AssigneeList.vue'

const props = withDefaults(defineProps<{
	modelValue: IUser[] | undefined,
	// Pass 0 to use this picker in pre-creation contexts (e.g. inline
	// quick-add). Assignees are then emitted via v-model only — the caller
	// is responsible for persisting them once the task exists.
	taskId: number,
	projectId: number,
	disabled?: boolean,
}>(), {
	disabled: false,
})

const emit = defineEmits<{
	'update:modelValue': [value: IUser[] | undefined],
}>()

const taskStore = useTaskStore()
const {t} = useI18n({useScope: 'global'})

const projectUserService = shallowReactive(new ProjectUserService())
const foundUsers = ref<IUser[]>([])
const assignees = ref<IUser[]>([])
let isAdding = false

watch(
	() => props.modelValue,
	(value) => {
		assignees.value = value
	},
	{
		immediate: true,
		deep: true,
	},
)

async function addAssignee(user: IUser) {
	if (isAdding) {
		return
	}

	if (props.taskId === 0) {
		// taskId 0 means the task doesn't exist yet; skip the API call and let the caller persist on creation.
		emit('update:modelValue', assignees.value)
		return
	}

	try {
		nextTick(() => isAdding = true)

		await taskStore.addAssignee({user: user, taskId: props.taskId})
		emit('update:modelValue', assignees.value)
		success({message: t('task.assignee.assignSuccess')})
	} finally {
		nextTick(() => isAdding = false)
	}
}

async function removeAssignee(user: IUser) {
	if (props.taskId !== 0) {
		await taskStore.removeAssignee({user: user, taskId: props.taskId})
	}

	// Remove the assignee from the project
	const idx = assignees.value.findIndex(a => a.id === user.id)
	if (idx !== -1) {
		assignees.value.splice(idx, 1)
	}
	emit('update:modelValue', assignees.value)
	if (props.taskId !== 0) {
		success({message: t('task.assignee.unassignSuccess')})
	}
}

async function findUser(query: string) {
	const response = await projectUserService.getAll({projectId: props.projectId}, {s: query}) as IUser[]

	// Filter the results to not include users who are already assigned
	foundUsers.value = response
		.filter(({id}) => !includesById(assignees.value, id))
		.map(u => {
			// Users may not have a display name set, so we fall back on the username in that case
			u.name = getDisplayName(u)
			return u
		})
}
</script>

<style lang="scss">
.edit-assignees.has-assignees.multiselect .input {
	padding-inline-start: 0;
}
</style>
