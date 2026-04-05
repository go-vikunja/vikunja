<script setup lang="ts">
import {computed, onMounted, ref} from 'vue'
import {useFlatpickrLanguage} from '@/helpers/useFlatpickrLanguage'
import XButton from '@/components/input/Button.vue'
import ApiTokenService from '@/services/apiToken'
import ApiTokenModel from '@/models/apiTokenModel'
import FancyCheckbox from '@/components/input/FancyCheckbox.vue'
import {MILLISECONDS_A_DAY} from '@/constants/date'
import flatPickr from 'vue-flatpickr-component'
import 'flatpickr/dist/flatpickr.css'
import {useI18n} from 'vue-i18n'
import FormField from '@/components/input/FormField.vue'
import type {IApiToken} from '@/modelTypes/IApiToken'

const props = withDefaults(defineProps<{
	ownerId?: number,
	loading?: boolean,
	initialTitle?: string,
	initialScopes?: string,
}>(), {
	ownerId: 0,
	loading: false,
	initialTitle: '',
	initialScopes: '',
})

const emit = defineEmits<{
	created: [token: IApiToken]
	cancel: []
}>()

const service = new ApiTokenService()
const {t} = useI18n()
const now = new Date()

const availableRoutes = ref(null)
const newToken = ref<IApiToken>(new ApiTokenModel())
const newTokenExpiry = ref<string | number>(30)
const newTokenExpiryCustom = ref(new Date())
const newTokenPermissions = ref({})
const newTokenPermissionsGroup = ref({})
const newTokenTitleValid = ref(true)
const newTokenPermissionValid = ref(true)
const apiTokenTitle = ref()

interface TokenPreset {
	id: string
	groups: Record<string, string[] | '*'>
}

const presets: TokenPreset[] = [
	{
		id: 'readOnly',
		groups: {
			'*': ['read_one', 'read_all'],
		},
	},
	{
		id: 'tasks',
		groups: {
			'tasks': '*',
			'tasks_attachments': '*',
			'tasks_assignees': '*',
			'tasks_labels': '*',
			'tasks_comments': '*',
			'tasks_relations': '*',
			'labels': ['read_one', 'read_all', 'create'],
			'projects': ['read_one', 'read_all', 'views_buckets_tasks'],
			'projects_views': ['read_one', 'read_all'],
			'projects_views_tasks': ['read_one', 'read_all'],
		},
	},
	{
		id: 'projects',
		groups: {
			'projects': '*',
			'projects_views': '*',
			'projects_teams': '*',
			'projects_users': '*',
			'projects_shares': '*',
			'projects_webhooks': '*',
			'projects_buckets': '*',
			'projects_views_tasks': '*',
			'tasks': ['read_one', 'read_all'],
			'teams': ['read_one', 'read_all'],
		},
	},
	{
		id: 'fullAccess',
		groups: {
			'*': '*',
		},
	},
]

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
	const allRoutes = await service.getAvailableRoutes()

	const routesAvailable = {}
	const keys = Object.keys(allRoutes)
	keys.sort((a, b) => (a === 'other' ? 1 : b === 'other' ? -1 : 0))
	keys.forEach(key => {
		routesAvailable[key] = allRoutes[key]
	})

	availableRoutes.value = routesAvailable
	resetPermissions()

	// Apply initial values from props (e.g. from query parameters)
	if (props.initialTitle) {
		newToken.value.title = props.initialTitle
		newTokenTitleValid.value = true
	}

	if (props.initialScopes) {
		const requestedScopes: Record<string, string[]> = {}
		for (const scope of props.initialScopes.split(',')) {
			const [group, permission] = scope.split(':')
			if (group && permission) {
				if (!requestedScopes[group]) {
					requestedScopes[group] = []
				}
				requestedScopes[group].push(permission)
			}
		}
		for (const [group, permissions] of Object.entries(requestedScopes)) {
			if (newTokenPermissions.value[group]) {
				for (const permission of permissions) {
					if (newTokenPermissions.value[group][permission] !== undefined) {
						newTokenPermissions.value[group][permission] = true
					}
				}
				toggleGroupPermissionsFromChild(group, true)
			}
		}
	}
})

function resetPermissions() {
	newTokenPermissions.value = {}
	newTokenPermissionsGroup.value = {}
	newTokenPermissionValid.value = true
	Object.entries(availableRoutes.value).forEach(entry => {
		const [group, routes] = entry
		newTokenPermissions.value[group] = {}
		newTokenPermissionsGroup.value[group] = false
		Object.keys(routes).forEach(r => {
			newTokenPermissions.value[group][r] = false
		})
	})
}

function applyPreset(preset: TokenPreset) {
	resetPermissions()

	for (const [groupKey, permissions] of Object.entries(preset.groups)) {
		if (groupKey === '*') {
			for (const group of Object.keys(availableRoutes.value)) {
				applyPermissionsToGroup(group, permissions)
			}
		} else if (availableRoutes.value[groupKey]) {
			applyPermissionsToGroup(groupKey, permissions)
		}
	}
}

function applyPermissionsToGroup(group: string, permissions: string[] | '*') {
	if (permissions === '*') {
		selectPermissionGroup(group, true)
		newTokenPermissionsGroup.value[group] = true
	} else {
		for (const perm of permissions) {
			if (newTokenPermissions.value[group]?.[perm] !== undefined) {
				newTokenPermissions.value[group][perm] = true
			}
		}
		toggleGroupPermissionsFromChild(group, true)
	}
}

function selectPermissionGroup(group: string, checked: boolean) {
	Object.entries(availableRoutes.value[group]).forEach(entry => {
		const [key] = entry
		newTokenPermissions.value[group][key] = checked
	})
	if (checked) {
		newTokenPermissionValid.value = true
	}
}

function toggleGroupPermissionsFromChild(group: string, checked: boolean) {
	if (checked) {
		newTokenPermissionValid.value = true
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

function formatPermissionTitle(title: string): string {
	return title.replaceAll('_', ' ')
}

async function createToken() {
	newTokenTitleValid.value = newToken.value.title.trim() !== ''
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

	if (!hasPermissions) {
		newTokenPermissionValid.value = false
		return
	}

	const expiry = Number(newTokenExpiry.value)
	if (!isNaN(expiry)) {
		newToken.value.expiresAt = new Date((+new Date()) + expiry * MILLISECONDS_A_DAY)
	} else {
		newToken.value.expiresAt = new Date(newTokenExpiryCustom.value)
	}

	if (props.ownerId > 0) {
		(newToken.value as IApiToken & {ownerId: number}).ownerId = props.ownerId
	}

	const token = await service.create(newToken.value)
	emit('created', token)

	newToken.value = new ApiTokenModel()
	newTokenExpiry.value = 30
	newTokenExpiryCustom.value = new Date()
	resetPermissions()
}
</script>

<template>
	<form @submit.prevent="createToken">
		<!-- Title -->
		<FormField
			id="apiTokenTitle"
			ref="apiTokenTitle"
			v-model="newToken.title"
			v-focus
			:label="$t('user.settings.apiTokens.attributes.title')"
			type="text"
			:placeholder="$t('user.settings.apiTokens.attributes.titlePlaceholder')"
			:error="newTokenTitleValid ? null : $t('user.settings.apiTokens.titleRequired')"
			@keyup="() => newTokenTitleValid = newToken.title !== ''"
			@focusout="() => newTokenTitleValid = newToken.title !== ''"
		/>

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

			<!-- Presets -->
			<div class="preset-buttons mbe-4">
				<label class="label">{{ $t('user.settings.apiTokens.presets.title') }}</label>
				<div
					class="is-flex"
					style="gap: .5rem; flex-wrap: wrap;"
				>
					<XButton
						v-for="preset in presets"
						:key="preset.id"
						variant="secondary"
						type="button"
						@click="applyPreset(preset)"
					>
						{{ $t(`user.settings.apiTokens.presets.${preset.id}`) }}
					</XButton>
				</div>
			</div>

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
					v-for="(paths, permission) in routes"
					:key="group+'-'+permission"
				>
					<FancyCheckbox
						v-model="newTokenPermissions[group][permission]"
						class="mis-4 mie-2 is-capitalized"
						@update:modelValue="checked => toggleGroupPermissionsFromChild(group, checked)"
					>
						{{ formatPermissionTitle(permission) }}
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
			:loading="loading"
			type="submit"
		>
			{{ $t('user.settings.apiTokens.createToken') }}
		</XButton>
		<XButton
			variant="tertiary"
			type="button"
			@click="emit('cancel')"
		>
			{{ $t('misc.cancel') }}
		</XButton>
	</form>
</template>
