<template>
	<create-edit
		:title="$t('team.create.title')"
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
import {defineComponent} from 'vue'
import TeamModel from '../../models/team'
import TeamService from '../../services/team'
import CreateEdit from '@/components/misc/create-edit.vue'

export default defineComponent({
	name: 'NewTeam',
	data() {
		return {
			teamService: new TeamService(),
			team: new TeamModel(),
			showError: false,
		}
	},
	components: {
		CreateEdit,
	},
	mounted() {
		this.setTitle(this.$t('team.create.title'))
	},
	methods: {
		async newTeam() {
			if (this.team.name === '') {
				this.showError = true
				return
			}
			this.showError = false

			const response = await this.teamService.create(this.team)
			this.$router.push({
				name: 'teams.edit',
				params: { id: response.id },
			})
			this.$message.success({message: this.$t('team.create.success') })
		},
	},
})
</script>
