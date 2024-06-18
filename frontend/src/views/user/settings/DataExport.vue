<template>
	<Card :title="$t('user.export.title')">
		<p>
			{{ $t('user.export.description') }}
		</p>
		<template v-if="isLocalUser">
			<p>
				{{ $t('user.export.descriptionPasswordRequired') }}
			</p>
			<div class="field">
				<label
					class="label"
					for="currentPasswordDataExport"
				>
					{{ $t('user.settings.currentPassword') }}
				</label>
				<div class="control">
					<input
						id="currentPasswordDataExport"
						ref="passwordInput"
						v-model="password"
						class="input"
						:class="{'is-danger': errPasswordRequired}"
						:placeholder="$t('user.settings.currentPasswordPlaceholder')"
						type="password"
						@keyup="() => errPasswordRequired = password === ''"
					>
				</div>
				<p
					v-if="errPasswordRequired"
					class="help is-danger"
				>
					{{ $t('user.deletion.passwordRequired') }}
				</p>
			</div>
		</template>

		<x-button
			:loading="dataExportService.loading"
			class="is-fullwidth mt-4"
			@click="requestDataExport()"
		>
			{{ $t('user.export.request') }}
		</x-button>
	</Card>
</template>

<script lang="ts">
export default {name: 'UserSettingsDataExport'}
</script>

<script setup lang="ts">
import {ref, computed, shallowReactive} from 'vue'
import {useI18n} from 'vue-i18n'

import DataExportService from '@/services/dataExport'
import {useTitle} from '@/composables/useTitle'
import {success} from '@/message'
import {useAuthStore} from '@/stores/auth'

const {t} = useI18n({useScope: 'global'})
const authStore = useAuthStore()

useTitle(() => `${t('user.export.title')} - ${t('user.settings.title')}`)

const dataExportService = shallowReactive(new DataExportService())
const password = ref('')
const errPasswordRequired = ref(false)
const isLocalUser = computed(() => authStore.info?.isLocalUser)
const passwordInput = ref()

async function requestDataExport() {
	if (password.value === '' && isLocalUser.value) {
		errPasswordRequired.value = true
		passwordInput.value.focus()
		return
	}

	await dataExportService.request(password.value)
	success({message: t('user.export.success')})
	password.value = ''
}
</script>
