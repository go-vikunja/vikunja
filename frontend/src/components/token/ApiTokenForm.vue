<script setup lang="ts">
import {computed, onMounted, ref, watch} from 'vue'
import {useNow} from '@vueuse/core'
import XButton from '@/components/input/Button.vue'
import ApiTokenService from '@/services/apiToken'
import ApiTokenModel from '@/models/apiTokenModel'
import FancyCheckbox from '@/components/input/FancyCheckbox.vue'
import {MILLISECONDS_A_DAY} from '@/constants/date'
import Datepicker from '@/components/input/Datepicker.vue'
import FormField from '@/components/input/FormField.vue'
import type {IApiToken, IApiPermission} from '@/modelTypes/IApiToken'
import type {ApiTokenRoutes, ApiTokenPreset} from '@/modelTypes/IApiTokenSettings'

const props = withDefaults(defineProps<{
	ownerId?: number,
	loading?: boolean,
	initialTitle?: string,
	initialScopes?: string,
	routes?: ApiTokenRoutes,
	presets?: ApiTokenPreset[],
	lockedScopes?: IApiPermission,
}>(), {
	ownerId: 0,
	loading: false,
	initialTitle: '',
	initialScopes: '',
	routes: undefined,
	presets: undefined,
	lockedScopes: () => ({}),
})

const emit = defineEmits<{
	created: [token: IApiToken]
	cancel: []
}>()

const service = new ApiTokenService()
const now = useNow({interval: 60_000})

const DEFAULT_EXPIRY_DAYS = 30

function expiryDateIn(days: number) {
	return new Date(Date.now() + days * MILLISECONDS_A_DAY)
}

const availableRoutes = ref<ApiTokenRoutes>({})
const newToken = ref<IApiToken>(new ApiTokenModel())
const newTokenExpiry = ref<string | number>(DEFAULT_EXPIRY_DAYS)
const newTokenExpiryCustom = ref<Date | null>(expiryDateIn(DEFAULT_EXPIRY_DAYS))

watch(newTokenExpiry, (value, oldValue) => {
	if (value === 'custom' && !isNaN(Number(oldValue))) {
		newTokenExpiryCustom.value = expiryDateIn(Number(oldValue))
	}
})
const newTokenPermissions = ref<Record<string, Record<string, boolean>>>({})
const newTokenPermissionsGroup = ref<Record<string, boolean>>({})
const newTokenTitleValid = ref(true)
const newTokenExpiryValid = ref(true)
const newTokenPermissionValid = ref(true)
const apiTokenTitle = ref()

const defaultPresets: ApiTokenPreset[] = [
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

const presets = computed(() => props.presets ?? defaultPresets)

onMounted(async () => {
	const allRoutes: ApiTokenRoutes = props.routes ?? await service.getAvailableRoutes()

	const routesAvailable: ApiTokenRoutes = {}
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
			newTokenPermissions.value[group][r] = isLocked(group, r)
		})
		toggleGroupPermissionsFromChild(group, true)
	})
}

function isLocked(group: string, permission: string) {
	return props.lockedScopes[group]?.includes(permission) ?? false
}

function isGroupLocked(group: string) {
	return Object.keys(availableRoutes.value[group]).every(permission => isLocked(group, permission))
}

function applyPreset(preset: ApiTokenPreset) {
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
		newTokenPermissions.value[group][key] = checked || isLocked(group, key)
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
	return title.replace(/_/g, ' ')
}

async function createToken() {
	newTokenTitleValid.value = newToken.value.title.trim() !== ''
	if (!newTokenTitleValid.value) {
		apiTokenTitle.value?.focus()
		return
	}

	let hasPermissions = false

	newToken.value.permissions = {}
	Object.entries(newTokenPermissions.value).forEach(([key, ps]) => {
		const all = Object.entries(ps)
			.filter(([permission, selected]) => selected || isLocked(key, permission))
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
		newToken.value.expiresAt = expiryDateIn(expiry)
	} else {
		const customExpiry = newTokenExpiryCustom.value === null ? null : new Date(newTokenExpiryCustom.value)
		if (customExpiry === null || isNaN(customExpiry.getTime()) || customExpiry <= new Date()) {
			newTokenExpiryValid.value = false
			return
		}
		newTokenExpiryValid.value = true
		newToken.value.expiresAt = customExpiry
	}

	if (props.ownerId > 0) {
		(newToken.value as IApiToken & {ownerId: number}).ownerId = props.ownerId
	}

	const token = await service.create(newToken.value)

	// Reset before emitting: parents hide the form in their `created` handler, so
	// anything after the emit would write to a component that's already unmounting.
	newToken.value = new ApiTokenModel()
	newTokenExpiry.value = DEFAULT_EXPIRY_DAYS
	newTokenExpiryCustom.value = expiryDateIn(DEFAULT_EXPIRY_DAYS)
	newTokenExpiryValid.value = true
	resetPermissions()

	emit('created', token)
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
				<div
					v-if="newTokenExpiry === 'custom'"
					class="control mis-2"
				>
					<Datepicker
						v-model="newTokenExpiryCustom"
						:choose-date-label="$t('user.settings.apiTokens.attributes.expiresAt')"
						:show-shortcuts="false"
						:min-date="now"
						@update:modelValue="newTokenExpiryValid = true"
					/>
				</div>
			</div>
			<p
				v-if="!newTokenExpiryValid"
				class="help is-danger"
			>
				{{ $t('user.settings.apiTokens.expiryInvalid') }}
			</p>
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
						{{ preset.label ?? $t(`user.settings.apiTokens.presets.${preset.id}`) }}
					</XButton>
				</div>
			</div>

			<div
				v-for="(groupRoutes, group) in availableRoutes"
				:key="group"
				class="mbe-2"
			>
				<template
					v-if="Object.keys(groupRoutes).length >= 1"
				>
					<FancyCheckbox
						v-model="newTokenPermissionsGroup[group]"
						:disabled="isGroupLocked(group)"
						class="mie-2 is-capitalized has-text-weight-bold"
						@update:modelValue="checked => selectPermissionGroup(group, checked)"
					>
						{{ formatPermissionTitle(group) }}
					</FancyCheckbox>
					<br>
				</template>
				<template
					v-for="(paths, permission) in groupRoutes"
					:key="group+'-'+permission"
				>
					<FancyCheckbox
						v-model="newTokenPermissions[group][permission]"
						:disabled="isLocked(group, permission)"
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
