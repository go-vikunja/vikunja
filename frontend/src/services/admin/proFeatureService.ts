import {AuthenticatedHTTPFactory, getApiV2Url} from '@/helpers/fetcher'

import type {ProFeature} from '@/constants/proFeatures'

export interface ProFeatureState {
	feature: ProFeature
	licensed: boolean
	perUserToggleable: boolean
	defaultEnabled: boolean
	defaultSource: 'code' | 'instance'
}

export interface UserProFeatureState {
	feature: ProFeature
	override: boolean | null
	effective: boolean
}

function parseProFeatureState(raw: Record<string, unknown>): ProFeatureState {
	return {
		feature: raw.feature as ProFeature,
		licensed: !!raw.licensed,
		perUserToggleable: !!raw.per_user_toggleable,
		defaultEnabled: !!raw.default_enabled,
		defaultSource: raw.default_source as 'code' | 'instance',
	}
}

function parseUserProFeatureState(raw: Record<string, unknown>): UserProFeatureState {
	return {
		feature: raw.feature as ProFeature,
		override: (raw.override ?? null) as boolean | null,
		effective: !!raw.effective,
	}
}

export function useAdminProFeatureService() {
	const http = AuthenticatedHTTPFactory()

	async function getAll(): Promise<ProFeatureState[]> {
		const {data} = await http.get(getApiV2Url('admin/pro-features'))
		return (data ?? []).map(parseProFeatureState)
	}

	async function setInstanceDefault(feature: ProFeature, defaultEnabled: boolean): Promise<ProFeatureState[]> {
		const {data} = await http.put(getApiV2Url(`admin/pro-features/${feature}`), {default_enabled: defaultEnabled})
		return (data ?? []).map(parseProFeatureState)
	}

	async function resetInstanceDefault(feature: ProFeature): Promise<void> {
		await http.delete(getApiV2Url(`admin/pro-features/${feature}`))
	}

	async function getForUser(userId: number): Promise<UserProFeatureState[]> {
		const {data} = await http.get(getApiV2Url(`admin/users/${userId}/pro-features`))
		return (data ?? []).map(parseUserProFeatureState)
	}

	async function setUserOverride(userId: number, feature: ProFeature, enabled: boolean): Promise<UserProFeatureState[]> {
		const {data} = await http.put(getApiV2Url(`admin/users/${userId}/pro-features/${feature}`), {enabled})
		return (data ?? []).map(parseUserProFeatureState)
	}

	async function clearUserOverride(userId: number, feature: ProFeature): Promise<UserProFeatureState[]> {
		const {data} = await http.delete(getApiV2Url(`admin/users/${userId}/pro-features/${feature}`))
		return (data ?? []).map(parseUserProFeatureState)
	}

	return {getAll, setInstanceDefault, resetInstanceDefault, getForUser, setUserOverride, clearUserOverride}
}
