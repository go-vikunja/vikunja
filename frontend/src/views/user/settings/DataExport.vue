<template>
	<Card :title="$t('user.export.title')">
		<Message
			v-if="exportInfo"
			class="mbe-4"
		>
			<div class="export-message">
				<p>
					<i18n-t
						keypath="user.export.ready"
						scope="global"
					>
						<time
							v-tooltip="formatDateLong(exportInfo.expires)"
							:datetime="formatISO(exportInfo.expires)"
						>
							{{ formattedExpiresDate }}
						</time>
					</i18n-t>
				</p>
				<XButton
					:to="{name:'user.export.download'}"
					class="button"
				>
					{{ $t('misc.download') }}
				</XButton>
			</div>
		</Message>
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
			class="is-fullwidth mbs-4"
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
import {formatISO, formatDateLong, formatDisplayDate} from '@/helpers/time/formatDate'

import Message from '@/components/misc/Message.vue'

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

const formattedExpiresDate = computed(() => exportInfo.value ? formatDisplayDate(new Date(exportInfo.value.expires)) : '')

onMounted(async () => {
	try {
		const data = await dataExportService.status()
		exportInfo.value = data
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

<style lang="scss" scoped>
.export-message {
	display: flex;
	justify-content: space-between;
	align-items: center;
	inline-size: 100%;
	gap: .5rem;
	
	> p {
		margin-block-end: 0;
	}
	
	@media (max-width: $mobile) {
		flex-direction: column;
		align-items: flex-start;
		
		> p {
			margin-block-end: 1rem;
		}
		
		> :deep(.button) {
			inline-size: 100%;
		}
	}
}
</style>
