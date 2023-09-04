<template>
	<Multiselect
		:loading="projectUserService.loading"
		:placeholder="$t('task.assignee.placeholder')"
		:multiple="true"
		@search="findUser"
		:search-results="foundUsers"
		@select="addAssignee"
		label="name"
		:select-placeholder="$t('task.assignee.selectPlaceholder')"
		v-model="assignees"
		:autocomplete-enabled="false"
	>
		<template #items="{items}">
			<assignee-list :assignees="items" :remove="removeAssignee"/>
		</template>
		<template #searchResult="{option: user}">
			<user :avatar-size="24" :show-username="true" :user="user"/>
		</template>
	</Multiselect>
</template>

<script setup lang="ts">
import {ref, shallowReactive, watch, nextTick, type PropType} from 'vue'
import {useI18n} from 'vue-i18n'

import User from '@/components/misc/user.vue'
import Multiselect from '@/components/input/multiselect.vue'

import {includesById} from '@/helpers/utils'
import ProjectUserService from '@/services/projectUsers'
import {success} from '@/message'
import {useTaskStore} from '@/stores/tasks'

import type {IUser} from '@/modelTypes/IUser'
import {getDisplayName} from '@/models/user'
import AssigneeList from '@/components/tasks/partials/assigneeList.vue'

const props = defineProps({
	taskId: {
		type: Number,
		required: true,
	},
	projectId: {
		type: Number,
		required: true,
	},
	disabled: {
		default: false,
	},
	modelValue: {
		type: Array as PropType<IUser[]>,
		default: () => [],
	},
})
const emit = defineEmits(['update:modelValue'])

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
	await taskStore.removeAssignee({user: user, taskId: props.taskId})

	// Remove the assignee from the project
	for (const a in assignees.value) {
		if (assignees.value[a].id === user.id) {
			assignees.value.splice(a, 1)
		}
	}
	success({message: t('task.assignee.unassignSuccess')})
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
