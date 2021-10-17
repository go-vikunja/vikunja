<template>
	<div class="content">
		<h1>{{ $t('user.export.downloadTitle') }}</h1>
		<p>{{ $t('user.export.descriptionPasswordRequired') }}</p>
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
			v-focus
			:loading="dataExportService.loading"
			@click="download()"
			class="mt-4">
			{{ $t('misc.download') }}
		</x-button>
	</div>
</template>

<script>
import DataExportService from '../../services/dataExport'

export default {
	name: 'data-export-download',
	data() {
		return {
			dataExportService: DataExportService,
			password: '',
			errPasswordRequired: false,
		}
	},
	created() {
		this.dataExportService = new DataExportService()
	},
	methods: {
		download() {
			if (this.password === '') {
				this.errPasswordRequired = true
				this.$refs.passwordInput.focus()
				return
			}

			this.dataExportService.download(this.password)
		},
	},
}
</script>
