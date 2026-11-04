<script setup lang="ts">
import {computed, nextTick, ref} from 'vue'
import {useI18n} from 'vue-i18n'
import {useQuery} from '@tanstack/vue-query'

import {useTitle} from '@/composables/useTitle'
import {useConfigStore} from '@/stores/config'
import {identitiesQuery, useUnlinkIdentityMutation} from '@/client/queries/identities'
import {totpQuery} from '@/client/queries/totp'
import type {LinkedIdentity} from '@/client/generated'
import {formatDisplayDate} from '@/helpers/time/formatDate'
import {redirectToProvider} from '@/helpers/redirectToProvider'
import ErrorMessage from '@/components/misc/Error.vue'
import ReauthFields from '@/components/user/ReauthFields.vue'
import type {IProvider} from '@/types/IProvider'

defineOptions({name: 'UserSettingsLinkedAccounts'})

const {t} = useI18n({useScope: 'global'})
useTitle(() => `${t('user.settings.linkedAccounts.title')} - ${t('user.settings.title')}`)

const configStore = useConfigStore()
const identityQuery = useQuery(identitiesQuery())
const identities = computed(() => identityQuery.data.value ?? [])
const totp = useQuery(computed(() => ({
	...totpQuery(),
	enabled: configStore.totp_enabled,
})))
const unlinkNeedsTotp = computed(() => totp.data.value?.enabled === true || totp.isError.value)
const unlinkTotpPending = computed(() => configStore.totp_enabled && totp.isPending.value)

const unlinkedProviders = computed(() => {
	const linked = new Set(identities.value.map(i => i.provider_key))
	return configStore.auth.openid_connect.providers.filter((p: IProvider) => !linked.has(p.key))
})

function identityName(identity: LinkedIdentity): string {
	return identity.provider_name || identity.issuer || ''
}

function unlinkLabel(identity: LinkedIdentity): string {
	return t('user.settings.linkedAccounts.unlinkHeader', {provider: identityName(identity)})
}

const unlinkMutation = useUnlinkIdentityMutation()
const identityToUnlink = ref<LinkedIdentity | null>(null)
const unlinkTitle = computed(() => identityToUnlink.value
	? t('user.settings.linkedAccounts.unlinkHeader', {provider: identityName(identityToUnlink.value)})
	: '')
const description = ref<HTMLElement | null>(null)
const linkButtons = ref<HTMLElement | null>(null)
const password = ref('')
const totpPasscode = ref('')

function openUnlink(identity: LinkedIdentity) {
	identityToUnlink.value = identity
	password.value = ''
	totpPasscode.value = ''
}

function closeUnlink() {
	identityToUnlink.value = null
	password.value = ''
	totpPasscode.value = ''
}

async function unlink() {
	const identity = identityToUnlink.value
	if (unlinkMutation.isPending.value || unlinkTotpPending.value || !identity?.id || !password.value) return
	try {
		await unlinkMutation.mutateAsync({
			id: identity.id,
			password: password.value,
			totp_passcode: unlinkNeedsTotp.value ? totpPasscode.value : undefined,
		})
	} catch { return }
	if (identityToUnlink.value === identity) {
		closeUnlink()
		await nextTick()
		const target = linkButtons.value?.querySelector<HTMLElement>('button, a') ?? description.value
		target?.focus()
	}
}
</script>

<template>
	<Card
		:title="$t('user.settings.linkedAccounts.title')"
		:loading="identityQuery.isPending.value"
	>
		<p
			ref="description"
			tabindex="-1"
			class="mbe-4"
		>
			{{ $t('user.settings.linkedAccounts.description') }}
		</p>

		<ErrorMessage v-if="identityQuery.isError.value" />

		<div
			v-if="identities.length > 0"
			class="has-horizontal-overflow"
		>
			<table class="table">
				<thead>
					<tr>
						<th>{{ $t('user.settings.linkedAccounts.provider') }}</th>
						<th>{{ $t('user.settings.linkedAccounts.linkedAt') }}</th>
						<th class="has-text-end">
							{{ $t('misc.actions') }}
						</th>
					</tr>
				</thead>
				<tbody>
					<tr
						v-for="identity in identities"
						:key="identity.id"
					>
						<td>{{ identityName(identity) }}</td>
						<td>{{ formatDisplayDate(identity.created) }}</td>
						<td class="has-text-end">
							<XButton
								variant="secondary"
								:aria-label="unlinkLabel(identity)"
								@click="openUnlink(identity)"
							>
								{{ $t('user.settings.linkedAccounts.unlink') }}
							</XButton>
						</td>
					</tr>
				</tbody>
			</table>
		</div>

		<p
			v-else-if="identityQuery.isSuccess.value"
			class="mbe-4"
		>
			{{ $t('user.settings.linkedAccounts.none') }}
		</p>

		<div
			v-if="identityQuery.isSuccess.value && unlinkedProviders.length > 0"
			ref="linkButtons"
			class="link-buttons"
		>
			<XButton
				v-for="provider in unlinkedProviders"
				:key="provider.key"
				icon="link"
				@click="redirectToProvider(provider, {linking: true})"
			>
				{{ $t('user.settings.linkedAccounts.link', {provider: provider.name}) }}
			</XButton>
		</div>

		<Modal
			v-if="identityToUnlink"
			variant="hint-modal"
			:aria-label="unlinkTitle"
			@close="closeUnlink()"
		>
			<Card
				class="has-no-shadow"
				:title="unlinkTitle"
			>
				<p class="mbe-4">
					{{ $t('user.settings.linkedAccounts.unlinkText') }}
				</p>
				<ReauthFields
					v-model:password="password"
					v-model:passcode="totpPasscode"
					password-id="unlinkIdentityPassword"
					passcode-id="unlinkIdentityPasscode"
					:show-passcode="unlinkNeedsTotp"
					autofocus
					@submit="unlink()"
				/>

				<template #footer>
					<XButton
						variant="tertiary"
						@click="closeUnlink()"
					>
						{{ $t('misc.cancel') }}
					</XButton>
					<XButton
						danger
						:loading="unlinkMutation.isPending.value || unlinkTotpPending"
						@click="unlink()"
					>
						{{ $t('user.settings.linkedAccounts.unlink') }}
					</XButton>
				</template>
			</Card>
		</Modal>
	</Card>
</template>

<style lang="scss" scoped>
.link-buttons {
	display: flex;
	flex-wrap: wrap;
	gap: .5rem;
}
</style>
