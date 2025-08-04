<script setup lang="ts">
import ApiTokenService from '@/services/apiToken'
import {computed, onMounted, ref} from 'vue'
import {useFlatpickrLanguage} from '@/helpers/useFlatpickrLanguage'
import {formatDateSince, formatDisplayDate} from '@/helpers/time/formatDate'
import XButton from '@/components/input/Button.vue'
import BaseButton from '@/components/base/BaseButton.vue'
import ApiTokenModel from '@/models/apiTokenModel'
import FancyCheckbox from '@/components/input/FancyCheckbox.vue'
import {MILLISECONDS_A_DAY} from '@/constants/date'
import flatPickr from 'vue-flatpickr-component'
import 'flatpickr/dist/flatpickr.css'
import {useI18n} from 'vue-i18n'
import Message from '@/components/misc/Message.vue'
import type {IApiToken} from '@/modelTypes/IApiToken'

const service = new ApiTokenService()
const tokens = ref<IApiToken[]>([])
const apiDocsUrl = window.API_URL + '/docs'
const showCreateForm = ref(false)
const availableRoutes = ref(null)
const newToken = ref<IApiToken>(new ApiTokenModel())
const newTokenExpiry = ref<string | number>(30)
const newTokenExpiryCustom = ref(new Date())
const newTokenPermissions = ref({})
const newTokenPermissionsGroup = ref({})
const newTokenTitleValid = ref(true)
const newTokenPermissionValid = ref(true)
const apiTokenTitle = ref()
const tokenCreatedSuccessMessage = ref('')

const showDeleteModal = ref<boolean>(false)
const tokenToDelete = ref<IApiToken>()

const {t} = useI18n()

const now = new Date()

const flatPickerConfig = computed(() => ({
	altFormat: t('date.altFormatLong'),
	altInput: true,
	dateFormat: 'Y-m-d H:i',
	enableTime: true,
	time_24hr: true,
	locale: useFlatpickrLanguage().value,
	minDate: now,
}))

onMounted(async () => {
	tokens.value = await service.getAll()
	const allRoutes = await service.getAvailableRoutes()

	const routesAvailable = {}
	const keys = Object.keys(allRoutes)
	keys.sort((a, b) => (a === 'other' ? 1 : b === 'other' ? -1 : 0))
	keys.forEach(key => {
		routesAvailable[key] = allRoutes[key]
	})
	
	availableRoutes.value = routesAvailable
	
	resetPermissions()
})

function resetPermissions() {
	newTokenPermissions.value = {}
	Object.entries(availableRoutes.value).forEach(entry => {
		const [group, routes] = entry
		newTokenPermissions.value[group] = {}
		Object.keys(routes).forEach(r => {
			newTokenPermissions.value[group][r] = false
		})
	})
}

async function deleteToken() {
	await service.delete(tokenToDelete.value)
	showDeleteModal.value = false
	const index = tokens.value.findIndex(el => el.id === tokenToDelete.value.id)
	tokenToDelete.value = null
	if (index === -1) {
		return
	}
	tokens.value.splice(index, 1)
}

async function createToken() {
	if (!newTokenTitleValid.value) {
		apiTokenTitle.value.focus()
		return
	}
	
	let hasPermissions = false

	newToken.value.permissions = {}
	Object.entries(newTokenPermissions.value).forEach(([key, ps]) => {
		const all = Object.entries(ps)
			// eslint-disable-next-line @typescript-eslint/no-unused-vars
			.filter(([_, v]) => v)
			.map(p => p[0])
		if (all.length > 0) {
			newToken.value.permissions[key] = all
			hasPermissions = true
		}
	})
	
	if(!hasPermissions) {
		newTokenPermissionValid.value = false
		return
	}
	
	const expiry = Number(newTokenExpiry.value)
	if (!isNaN(expiry)) {
		// if it's a number, we assume it's the number of days in the future
		newToken.value.expiresAt = new Date((+new Date()) + expiry * MILLISECONDS_A_DAY)
	} else {
		newToken.value.expiresAt = new Date(newTokenExpiryCustom.value)
	}

	const token = await service.create(newToken.value)
	tokenCreatedSuccessMessage.value = t('user.settings.apiTokens.tokenCreatedSuccess', {token: token.token})
	newToken.value = new ApiTokenModel()
	newTokenExpiry.value = 30
	newTokenExpiryCustom.value = new Date()
	resetPermissions()
	tokens.value.push(token)
	showCreateForm.value = false
}

function formatPermissionTitle(title: string): string {
	return title.replaceAll('_', ' ')
}

function selectPermissionGroup(group: string, checked: boolean) {
	Object.entries(availableRoutes.value[group]).forEach(entry => {
		const [key] = entry
		newTokenPermissions.value[group][key] = checked
	})
}

function toggleGroupPermissionsFromChild(group: string, checked: boolean) {
	if (checked) {
		// Check if all permissions of that group are checked and check the "select all" checkbox in that case
		let allChecked = true
		Object.entries(availableRoutes.value[group]).forEach(entry => {
			const [key] = entry
			if (!newTokenPermissions.value[group][key]) {
				allChecked = false
			}
		})

		if (allChecked) {
			newTokenPermissionsGroup.value[group] = true
		}
	} else {
		newTokenPermissionsGroup.value[group] = false
	}
}
</script>

<template>
	<Card :title="$t('user.settings.apiTokens.title')">
		<Message
			v-if="tokenCreatedSuccessMessage !== ''"
			class="has-text-centered mbe-4"
		>
			{{ tokenCreatedSuccessMessage }}<br>
			{{ $t('user.settings.apiTokens.tokenCreatedNotSeeAgain') }}
		</Message>

		<p>
			{{ $t('user.settings.apiTokens.general') }}
			<BaseButton :href="apiDocsUrl">
				{{ $t('user.settings.apiTokens.apiDocs') }}
			</BaseButton>
			.
		</p>

		<table
			v-if="tokens.length > 0"
			class="table"
		>
			<tr>
				<th>{{ $t('misc.id') }}</th>
				<th>{{ $t('user.settings.apiTokens.attributes.title') }}</th>
				<th>{{ $t('user.settings.apiTokens.attributes.permissions') }}</th>
				<th>{{ $t('user.settings.apiTokens.attributes.expiresAt') }}</th>
				<th>{{ $t('misc.created') }}</th>
				<th class="has-text-end">
					{{ $t('misc.actions') }}
				</th>
			</tr>
			<tr
				v-for="tk in tokens"
				:key="tk.id"
			>
				<td>{{ tk.id }}</td>
				<td>{{ tk.title }}</td>
				<td class="is-capitalized">
					<template
						v-for="(v, p) in tk.permissions"
						:key="'permission-' + p"
					>
						<strong>{{ formatPermissionTitle(p) }}:</strong>
						{{ v.map(formatPermissionTitle).join(', ') }}
						<br>
					</template>
				</td>
				<td>
					{{ formatDisplayDate(tk.expiresAt) }}
					<p
						v-if="tk.expiresAt < new Date()"
						class="has-text-danger"
					>
						{{ $t('user.settings.apiTokens.expired', {ago: formatDateSince(tk.expiresAt)}) }}
					</p>
				</td>
				<td>{{ formatDisplayDate(tk.created) }}</td>
				<td class="has-text-end">
					<XButton
						variant="secondary"
						@click="() => {tokenToDelete = tk; showDeleteModal = true}"
					>
						{{ $t('misc.delete') }}
					</XButton>
				</td>
			</tr>
		</table>

		<form
			v-if="showCreateForm"
			@submit.prevent="createToken"
		>
			<!-- Title -->
			<div class="field">
				<label
					class="label"
					for="apiTokenTitle"
				>{{ $t('user.settings.apiTokens.attributes.title') }}</label>
				<div class="control">
					<input
						id="apiTokenTitle"
						ref="apiTokenTitle"
						v-model="newToken.title"
						v-focus
						class="input"
						type="text"
						:placeholder="$t('user.settings.apiTokens.attributes.titlePlaceholder')"
						@keyup="() => newTokenTitleValid = newToken.title !== ''"
						@focusout="() => newTokenTitleValid = newToken.title !== ''"
					>
				</div>
				<p
					v-if="!newTokenTitleValid"
					class="help is-danger"
				>
					{{ $t('user.settings.apiTokens.titleRequired') }}
				</p>
			</div>

			<!-- Expiry -->
			<div class="field">
				<label
					class="label"
					for="apiTokenExpiry"
				>
					{{ $t('user.settings.apiTokens.attributes.expiresAt') }}
				</label>
				<div class="is-flex">
					<div class="control select">
						<select
							id="apiTokenExpiry"
							v-model="newTokenExpiry"
							class="select"
						>
							<option value="30">
								{{ $t('user.settings.apiTokens.30d') }}
							</option>
							<option value="60">
								{{ $t('user.settings.apiTokens.60d') }}
							</option>
							<option value="90">
								{{ $t('user.settings.apiTokens.90d') }}
							</option>
							<option value="custom">
								{{ $t('misc.custom') }}
							</option>
						</select>
					</div>
					<flat-pickr
						v-if="newTokenExpiry === 'custom'"
						v-model="newTokenExpiryCustom"
						class="mis-2"
						:config="flatPickerConfig"
					/>
				</div>
			</div>

			<!-- Permissions -->
			<div class="field">
				<label class="label">{{ $t('user.settings.apiTokens.attributes.permissions') }}</label>
				<p>{{ $t('user.settings.apiTokens.permissionExplanation') }}</p>
				<div
					v-for="(routes, group) in availableRoutes"
					:key="group"
					class="mbe-2"
				>
					<template
						v-if="Object.keys(routes).length >= 1"
					>
						<FancyCheckbox
							v-model="newTokenPermissionsGroup[group]"
							class="mie-2 is-capitalized has-text-weight-bold"
							@update:modelValue="checked => selectPermissionGroup(group, checked)"
						>
							{{ formatPermissionTitle(group) }}
						</FancyCheckbox>
						<br>
					</template>
					<template
						v-for="(paths, route) in routes"
						:key="group+'-'+route"
					>
						<FancyCheckbox
							v-model="newTokenPermissions[group][route]"
							class="mis-4 mie-2 is-capitalized"
							@update:modelValue="checked => toggleGroupPermissionsFromChild(group, checked)"
						>
							{{ formatPermissionTitle(route) }}
						</FancyCheckbox>
						<br>
					</template>
				</div>
			</div>

			<p
				v-if="!newTokenPermissionValid"
				class="help is-danger"
			>
				{{ $t('user.settings.apiTokens.permissionRequired') }}
			</p>
			<XButton
				:loading="service.loading"
				@click="createToken"
			>
				{{ $t('user.settings.apiTokens.createToken') }}
			</XButton>
		</form>

		<XButton
			v-else
			icon="plus"
			class="mbe-4"
			:loading="service.loading"
			@click="() => showCreateForm = true"
		>
			{{ $t('user.settings.apiTokens.createAToken') }}
		</XButton>

		<Modal
			:enabled="showDeleteModal"
			@close="showDeleteModal = false"
			@submit="deleteToken()"
		>
			<template #header>
				{{ $t('user.settings.apiTokens.delete.header') }}
			</template>

			<template #text>
				<p>
					{{ $t('user.settings.apiTokens.delete.text1', {token: tokenToDelete.title}) }}<br>
					{{ $t('user.settings.apiTokens.delete.text2') }}
				</p>
			</template>
		</Modal>
	</Card>
</template>
