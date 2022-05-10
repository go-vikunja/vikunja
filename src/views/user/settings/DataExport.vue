<template>
	<card :title="$t('user.export.title')">
		<p>
			{{ $t('user.export.description') }}
		</p>
		<template v-if="isLocalUser">
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
		</template>

		<x-button
			:loading="dataExportService.loading"
			@click="requestDataExport()"
			class="is-fullwidth mt-4">
			{{ $t('user.export.request') }}
		</x-button>
	</card>
</template>

<script lang="ts">
import {defineComponent} from 'vue'

export default defineComponent({
	name: 'user-settings-data-export',
})
</script>

<script setup lang="ts">
import {ref, computed, shallowReactive} from 'vue'
import {useStore} from 'vuex'
import {useI18n} from 'vue-i18n'

import DataExportService from '@/services/dataExport'
import { useTitle } from '@/composables/useTitle'
import {success} from '@/message'

const {t} = useI18n()
const store = useStore()

useTitle(() => `${t('user.export.title')} - ${t('user.settings.title')}`)

const dataExportService = shallowReactive(new DataExportService())
const password = ref('')
const errPasswordRequired = ref(false)
const isLocalUser = computed(() => store.state.auth.info?.isLocalUser)
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
