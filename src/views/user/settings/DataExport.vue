<template>
	<card :title="$t('user.export.title')">
		<p>
			{{ $t('user.export.description') }}
		</p>
		<p>
			{{ $t('user.export.descriptionPasswordRequired') }}
		</p>
		<div class="field">
			<label class="label" for="currentPasswordDataExport">
				{{ $t('user.settings.currentPassword') }}
			</label>
			<div class="control">
				<input
					class="input"
					:class="{'is-danger': errPasswordRequired}"
					id="currentPasswordDataExport"
					:placeholder="$t('user.settings.currentPasswordPlaceholder')"
					type="password"
					v-model="password"
					@keyup="() => errPasswordRequired = password === ''"
					ref="passwordInput"
				/>
			</div>
			<p class="help is-danger" v-if="errPasswordRequired">
				{{ $t('user.deletion.passwordRequired') }}
			</p>
		</div>

		<x-button
			:loading="dataExportService.loading"
			@click="requestDataExport()"
			class="is-fullwidth mt-4">
			{{ $t('user.export.request') }}
		</x-button>
	</card>
</template>

<script>
import DataExportService from '@/services/dataExport'

export default {
	name: 'user-settings-data-export',
	data() {
		return {
			dataExportService: new DataExportService(),
			password: '',
			errPasswordRequired: false,
		}
	},
	mounted() {
		this.setTitle(`${this.$t('user.export.title')} - ${this.$t('user.settings.title')}`)
	},
	methods: {
		async requestDataExport() {
			if (this.password === '') {
				this.errPasswordRequired = true
				this.$refs.passwordInput.focus()
				return
			}

			await this.dataExportService.request(this.password)
			this.$message.success({message: this.$t('user.export.success')})
			this.password = ''
		},
	},
}
</script>
