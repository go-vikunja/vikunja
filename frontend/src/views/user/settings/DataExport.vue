<template>
	<Card :title="$t('user.export.title')">
		<div
			v-if="exportInfo"
			class="message is-info"
		>
			<p>{{ $t('user.export.ready') }}</p>
			<p>{{ $t('misc.created') }}: {{ new Date(exportInfo.created).toLocaleString() }}</p>
			<p>{{ $t('user.export.expires') }}: {{ new Date(exportInfo.expires).toLocaleString() }}</p>
			<RouterLink
				:to="{name:'user.export.download'}"
				class="button is-link mt-3"
			>
				{{ $t('misc.download') }}
			</RouterLink>
		</div>
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

		<XButton
			:loading="dataExportService.loading"
			class="is-fullwidth mt-4"
			@click="requestDataExport()"
		>
			{{ $t('user.export.request') }}
		</XButton>
	</Card>
</template>


<script setup lang="ts">
import {ref, computed, shallowReactive, onMounted} from 'vue'
import {useI18n} from 'vue-i18n'

import DataExportService from '@/services/dataExport'
import {useTitle} from '@/composables/useTitle'
import {success} from '@/message'
import {useAuthStore} from '@/stores/auth'

defineOptions({name: 'UserSettingsDataExport'})

const {t} = useI18n({useScope: 'global'})
const authStore = useAuthStore()

useTitle(() => `${t('user.export.title')} - ${t('user.settings.title')}`)

const dataExportService = shallowReactive(new DataExportService())
interface ExportInfo {
       id: number
       size: number
       created: string
       expires: string
}
const exportInfo = ref<ExportInfo | null>(null)
const password = ref('')
const errPasswordRequired = ref(false)
const isLocalUser = computed(() => authStore.info?.isLocalUser)
const passwordInput = ref()

onMounted(async () => {
	try {
		const data = await dataExportService.status()
		if (Object.keys(data).length > 0) {
			exportInfo.value = data
		}
	} catch {
		exportInfo.value = null
	}
})

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
