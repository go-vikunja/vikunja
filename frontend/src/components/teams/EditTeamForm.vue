<template>
	<Card
		class="is-fullwidth"
		:title="$t('team.edit.title', {team: team.name})"
	>
		<form @submit.prevent="save()">
			<FormField
				id="teamtext"
				v-model="teamDraft.name"
				v-focus
				:label="$t('team.attributes.name')"
				:disabled="membersLoading"
				:loading="membersLoading"
				:placeholder="$t('team.attributes.namePlaceholder')"
				type="text"
				:error="showErrorTeamnameRequired && teamDraft.name === '' ? $t('team.attributes.nameRequired') : null"
			/>
			<FormField
				v-if="configStore.public_teams_enabled"
				:label="$t('team.attributes.isPublic')"
			>
				<FancyCheckbox
					v-model="teamDraft.is_public"
					:disabled="membersLoading || undefined"
					:class="{ 'disabled': loading }"
				>
					{{ $t('team.attributes.isPublicDescription') }}
				</FancyCheckbox>
			</FormField>
			<FormField :label="$t('team.attributes.description')">
				<Editor
					id="teamdescription"
					v-model="teamDraft.description"
					:class="{ disabled: loading }"
					:disabled="loading"
					:placeholder="$t('team.attributes.descriptionPlaceholder')"
				/>
			</FormField>

			<div class="field has-addons mbs-4">
				<div class="control is-fullwidth">
					<XButton
						:loading="loading"
						class="is-fullwidth"
						type="submit"
					>
						{{ $t('misc.save') }}
					</XButton>
				</div>
				<div class="control">
					<XButton
						:loading="loading"
						danger
						icon="trash-alt"
						:aria-label="$t('team.edit.delete.header')"
						@click="emit('delete')"
					/>
				</div>
			</div>
		</form>
	</Card>
</template>

<script lang="ts" setup>
import {ref} from 'vue'
import Editor from '@/components/input/AsyncEditor'
import FancyCheckbox from '@/components/input/FancyCheckbox.vue'
import FormField from '@/components/input/FormField.vue'
import {createTeamDraft} from '@/client/queries/teams'
import {useConfigStore} from '@/stores/config'
import type {Team, TeamWritable} from '@/client/generated'

const props = defineProps<{
	team: Team
	loading: boolean
	membersLoading: boolean
}>()

const emit = defineEmits<{
	save: [team: Required<TeamWritable>]
	delete: []
}>()

const configStore = useConfigStore()
// eslint-disable-next-line vue/no-setup-props-reactivity-loss -- seeded once so refetches keep the draft; parent keys on team id
const teamDraft = ref(createTeamDraft(props.team))
const showErrorTeamnameRequired = ref(false)

function save() {
	showErrorTeamnameRequired.value = teamDraft.value.name === ''
	if (showErrorTeamnameRequired.value) return
	emit('save', teamDraft.value)
}
</script>

