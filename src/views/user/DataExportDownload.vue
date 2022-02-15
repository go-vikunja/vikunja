<template>
	<div class="content">
		<h1>{{ $t('user.export.downloadTitle') }}</h1>
		<template v-if="isLocalUser">
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
		</template>

		<x-button
			v-focus
			:loading="dataExportService.loading"
			@click="download()"
			class="mt-4">
			{{ $t('misc.download') }}
		</x-button>
	</div>
</template>

<script setup lang="ts">
import {ref, computed, reactive} from 'vue'
import DataExportService from '@/services/dataExport'
import {store} from '@/store'

const dataExportService = reactive(new DataExportService())
const password = ref('')
const errPasswordRequired = ref(false)
const passwordInput = ref(null)

const isLocalUser = computed(() => store.state.auth.info?.isLocalUser)

function download() {
	if (password.value === '' && isLocalUser.value) {
		errPasswordRequired.value = true
		passwordInput.value.focus()
		return
	}

	dataExportService.download(password.value)
}
</script>
