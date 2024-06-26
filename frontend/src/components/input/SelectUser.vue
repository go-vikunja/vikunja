<template>
	<Multiselect
		v-model="selectedUsers"
		:search-results="foundUsers"
		:loading="userService.loading"
		:multiple="true"
		:placeholder="$t('team.edit.search')"
		label="username"
		@search="findUsers"
	/>
</template>

<script setup lang="ts">
import {computed, ref, shallowReactive, watchEffect} from 'vue'

import Multiselect from '@/components/input/Multiselect.vue'

import type {IUser} from '@/modelTypes/IUser'

import UserService from '@/services/user'
import {includesById} from '@/helpers/utils'

const props = withDefaults(defineProps<{
	modelValue: IUser[] | undefined
}>(), {
	modelValue: () => [],
})

const emit = defineEmits<{
	'update:modelValue': [value: IUser[]]
}>()

const users = ref<IUser[]>([])

watchEffect(() => {
	users.value = props.modelValue
})

const selectedUsers = computed({
	get() {
		return users.value
	},
	set: (value) => {
		users.value = value
		emit('update:modelValue', value)
	},
})

const userService = shallowReactive(new UserService())
const foundUsers = ref<IUser[]>([])

async function findUsers(query: string) {
	if (query === '') {
		foundUsers.value = []
		return
	}

	const response = await userService.getAll({}, {s: query}) as IUser[]

	// Filter selected items from the results
	foundUsers.value = response.filter(({id}) => !includesById(users.value, id))
}
</script>