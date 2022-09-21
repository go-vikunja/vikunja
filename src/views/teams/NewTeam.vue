<template>
	<create-edit
		:title="title"
		@create="newTeam()"
		:primary-disabled="team.name === ''"
	>
		<div class="field">
			<label class="label" for="teamName">{{ $t('team.attributes.name') }}</label>
			<div
				class="control is-expanded"
				:class="{ 'is-loading': teamService.loading }"
			>
				<input
					:class="{ 'disabled': teamService.loading }"
					class="input"
					id="teamName"
					:placeholder="$t('team.attributes.namePlaceholder')"
					type="text"
					v-focus
					v-model="team.name"
					@keyup.enter="newTeam"
				/>
			</div>
		</div>
		<p class="help is-danger" v-if="showError && team.name === ''">
			{{ $t('team.attributes.nameRequired') }}
		</p>
	</create-edit>
</template>

<script lang="ts">
export default { name: 'NewTeam' }
</script>

<script setup lang="ts">
import {reactive, ref, shallowReactive, computed} from 'vue'
import {useI18n} from 'vue-i18n'

import TeamModel from '@/models/team'
import TeamService from '@/services/team'

import CreateEdit from '@/components/misc/create-edit.vue'

import {useTitle} from '@/composables/useTitle'
import {useRouter} from 'vue-router'
import {success} from '@/message'

const {t} = useI18n()
const title = computed(() => t('team.create.title'))
useTitle(title)
const router = useRouter()

const teamService = shallowReactive(new TeamService())
const team = reactive(new TeamModel())
const showError = ref(false)

async function newTeam() {
	if (team.name === '') {
		showError.value = true
		return
	}
	showError.value = false

	const response = await teamService.create(team)
	router.push({
		name: 'teams.edit',
		params: { id: response.id },
	})
	success({message: t('team.create.success') })
}
</script>
