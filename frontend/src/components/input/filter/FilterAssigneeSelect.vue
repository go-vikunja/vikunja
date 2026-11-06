<script setup lang="ts">
import {computed, ref} from 'vue'
import {useI18n} from 'vue-i18n'
import Multiselect from '@/components/input/Multiselect.vue'
import {useProjectUserSearch, useUserSearch} from '@/composables/useUserSearch'
import {getDisplayName} from '@/helpers/user'

const props = defineProps<{
	projectId?: number
}>()
const usernames = defineModel<string[]>({required: true})
const {t} = useI18n({useScope: 'global'})
const query = ref('')
const enabled = ref(false)
const globalSearch = useUserSearch(() => props.projectId ? '' : query.value)
const projectSearch = useProjectUserSearch(() => props.projectId ?? 0, query, enabled)
const users = computed(() => props.projectId ? projectSearch.users.value : globalSearch.users.value)
const loading = computed(() => props.projectId ? projectSearch.isFetching.value : globalSearch.isFetching.value)
const selected = computed(() => usernames.value.map(username => ({
	username,
	title: username,
})))
const results = computed(() => users.value.flatMap(user => user.username && !usernames.value.includes(user.username)
	? [{
		username: user.username,
		title: `${getDisplayName(user)} (${user.username})`,
	}] : []))
</script>

<template>
	<Multiselect
		:model-value="selected"
		:search-results="results"
		:loading="loading"
		:multiple="true"
		:show-empty="true"
		:close-after-select="false"
		:aria-label="t('task.attributes.assignees')"
		:placeholder="t('task.assignee.placeholder')"
		:autocomplete-enabled="false"
		label="title"
		@focusin="enabled = true"
		@search="query = $event; enabled = true"
		@update:modelValue="usernames = Array.isArray($event) ? $event.map(user => user.username) : []"
	/>
</template>
