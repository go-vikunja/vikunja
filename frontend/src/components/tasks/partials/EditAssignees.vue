<template>
	<Multiselect
		v-model="assignees"
		class="edit-assignees"
		:class="{'has-assignees': assignees.length > 0}"
		:loading="usersLoading"
		:placeholder="$t('task.assignee.placeholder')"
		:multiple="true"
		:search-results="foundUsers"
		:show-empty="true"
		label="name"
		:select-placeholder="$t('task.assignee.selectPlaceholder')"
		:autocomplete-enabled="false"
		@search="findUser"
		@select="addAssignee"
		@focus="preloadUsers"
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
import {ref, computed, watch, nextTick} from 'vue'
import {useI18n} from 'vue-i18n'

import User from '@/components/misc/User.vue'
import Multiselect from '@/components/input/Multiselect.vue'

import {useProjectUserSearch} from '@/composables/useUserSearch'
import {success} from '@/message'
import {useAuthStore} from '@/stores/auth'
import {useAddTaskAssigneeMutation, useRemoveTaskAssigneeMutation} from '@/client/queries/taskMutations'

import type {User as IUser} from '@/client/generated'
import {getDisplayName, type UserWithId} from '@/models/user'
import AssigneeList from '@/components/tasks/partials/AssigneeList.vue'

const props = withDefaults(defineProps<{
	modelValue: IUser[] | undefined,
	taskId: number,
	projectId: number,
	disabled?: boolean,
}>(), {
	disabled: false,
})

const emit = defineEmits<{
	'update:modelValue': [value: IUser[] | undefined],
}>()

const authStore = useAuthStore()
const addAssigneeMutation = useAddTaskAssigneeMutation()
const removeAssigneeMutation = useRemoveTaskAssigneeMutation()
const {t} = useI18n({useScope: 'global'})

const userSearch = ref('')
const searchEnabled = ref(false)
const {users: userResults, isFetching: usersLoading} = useProjectUserSearch(() => props.projectId, userSearch, searchEnabled)
const assignees = ref<UserWithId[]>([])
let isAdding = false

function preloadUsers() {
	searchEnabled.value = true
}

watch(
	() => props.modelValue,
	(value) => {
		assignees.value = withIds(value ?? [])
	},
	{
		immediate: true,
		deep: true,
	},
)

async function addAssignee(user: UserWithId) {
	if (isAdding) {
		return
	}

	try {
		nextTick(() => isAdding = true)

		await addAssigneeMutation.mutateAsync({user: user, taskId: props.taskId})
		emit('update:modelValue', assignees.value)
		success({message: t('task.assignee.assignSuccess')})
	} finally {
		nextTick(() => isAdding = false)
	}
}

async function removeAssignee(user: UserWithId) {
	await removeAssigneeMutation.mutateAsync({user: user, taskId: props.taskId})

	assignees.value = assignees.value.filter(a => a.id !== user.id)
	emit('update:modelValue', assignees.value)
	success({message: t('task.assignee.unassignSuccess')})
}

function findUser(query = '') {
	userSearch.value = query
	searchEnabled.value = true
}

function withIds(users: IUser[]): UserWithId[] {
	return users.flatMap(user => user.id === undefined ? [] : [{...user, id: user.id}])
}

const foundUsers = computed(() => withIds(userResults.value)
	.filter(({id}) => !assignees.value.some(assignee => assignee.id === id))
	.map(user => ({...user, name: getDisplayName(user)}))
	.sort((a, b) => {
		if (a.id === authStore.info?.id) return -1
		if (b.id === authStore.info?.id) return 1
		return a.name.localeCompare(b.name)
	}),
)
</script>

<style lang="scss">
.edit-assignees.has-assignees.multiselect .input {
	padding-inline-start: 0;
}
</style>
