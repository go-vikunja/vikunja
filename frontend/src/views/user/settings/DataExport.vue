<template>
	<Card :title="$t('user.export.title')">
		<Message
			v-if="exportInfo && exportExpires"
			class="mbe-4"
		>
			<div class="export-message">
				<p>
					<i18n-t
						keypath="user.export.ready"
						scope="global"
					>
						<time
							v-tooltip="formatDateLong(exportExpires)"
							:datetime="formatISO(exportExpires)"
						>
							{{ formattedExpiresDate }}
						</time>
					</i18n-t>
				</p>
				<XButton
					:to="{name:'user.export.download'}"
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
			<FormField
				id="currentPasswordDataExport"
				ref="passwordInput"
				v-model="password"
				:label="$t('user.settings.currentPassword')"
				:class="{'is-danger': errPasswordRequired}"
				:placeholder="$t('user.settings.currentPasswordPlaceholder')"
				type="password"
				:error="errPasswordRequired ? $t('user.deletion.passwordRequired') : null"
				@keyup="() => errPasswordRequired = password === ''"
			/>
		</template>

		<XButton
			:loading="requestMutation.isPending.value"
			class="is-fullwidth mbs-4"
			@click="requestDataExport()"
		>
			{{ $t('user.export.request') }}
		</XButton>
	</Card>
</template>

<script setup lang="ts">
import {ref, computed} from 'vue'
import {useI18n} from 'vue-i18n'

import {useQuery} from '@tanstack/vue-query'
import {dataExportQuery, useRequestExportMutation} from '@/client/queries/dataExport'
import {parseDateOrNull} from '@/helpers/parseDateOrNull'
import {useTitle} from '@/composables/useTitle'
import {useAuthStore} from '@/stores/auth'
import {formatISO, formatDateLong, formatDisplayDate} from '@/helpers/time/formatDate'

import Message from '@/components/misc/Message.vue'
import FormField from '@/components/input/FormField.vue'

defineOptions({name: 'UserSettingsDataExport'})

const {t} = useI18n({useScope: 'global'})
const authStore = useAuthStore()

useTitle(() => `${t('user.export.title')} - ${t('user.settings.title')}`)

const status = useQuery(dataExportQuery())
const exportInfo = computed(() => status.data.value?.id ? status.data.value : null)
const exportExpires = computed(() => parseDateOrNull(exportInfo.value?.expires))
const requestMutation = useRequestExportMutation()
const password = ref('')
const errPasswordRequired = ref(false)
const isLocalUser = computed(() => authStore.info?.is_local_user)
const passwordInput = ref()

const formattedExpiresDate = computed(() => exportExpires.value ? formatDisplayDate(exportExpires.value) : '')

async function requestDataExport() {
	if (password.value === '' && isLocalUser.value) {
		errPasswordRequired.value = true
		passwordInput.value?.focus()
		return
	}

	try {
		await requestMutation.mutateAsync(password.value)
	} catch { return }
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
