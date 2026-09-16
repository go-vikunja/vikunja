<template>
	<div>
		<h3 class="has-text-weight-bold share-heading">
			{{ $t('project.share.userTeam.shared', {type: typeNames}) }}
		</h3>
		<div v-if="userIsAdmin">
			<div class="field has-addons">
				<p
					class="control is-expanded"
					:class="{ 'is-loading': searchLoading }"
				>
					<Multiselect
						v-model="selected"
						:loading="searchLoading"
						:placeholder="$t('misc.searchPlaceholder')"
						:aria-label="$t('project.share.userTeam.search', {type: typeName})"
						:search-results="unsharedCandidates"
						:label="searchLabel"
						@search="search = $event"
					>
						<template #searchResult="{option}">
							<slot
								name="candidate"
								:candidate="option"
							/>
						</template>
					</Multiselect>
				</p>
				<p class="control">
					<XButton
						:class="{'is-loading': isMutating}"
						:aria-disabled="!selected || isMutating || undefined"
						@click="addSelected()"
					>
						{{ $t('project.share.share') }}
					</XButton>
				</p>
			</div>
		</div>

		<div
			v-if="identifiedShares.length > 0"
			class="has-horizontal-overflow mbe-4"
		>
			<table class="table has-actions is-striped is-hoverable is-fullwidth">
				<tbody>
					<tr
						v-for="share in identifiedShares"
						:key="share.id"
					>
						<slot
							name="share"
							:share="share"
						/>
						<td class="type">
							<template v-if="share.permission === PERMISSIONS.ADMIN">
								<span class="icon is-small">
									<Icon icon="lock" />
								</span>
								{{ $t('project.share.permission.admin') }}
							</template>
							<template v-else-if="share.permission === PERMISSIONS.READ_WRITE">
								<span class="icon is-small">
									<Icon icon="pen" />
								</span>
								{{ $t('project.share.permission.readWrite') }}
							</template>
							<template v-else>
								<span class="icon is-small">
									<Icon icon="users" />
								</span>
								{{ $t('project.share.permission.read') }}
							</template>
						</td>
						<td
							v-if="userIsAdmin"
							class="actions"
						>
							<div class="select">
								<select
									v-model="selectedPermission[share.id]"
									:aria-disabled="isMutating || undefined"
									class="mie-2"
									:aria-label="$t('project.share.userTeam.permissionFor', {sharable: shareName(share)})"
									@change="changePermission(share)"
								>
									<option
										:selected="share.permission === PERMISSIONS.READ"
										:value="PERMISSIONS.READ"
									>
										{{ $t('project.share.permission.read') }}
									</option>
									<option
										:selected="share.permission === PERMISSIONS.READ_WRITE"
										:value="PERMISSIONS.READ_WRITE"
									>
										{{ $t('project.share.permission.readWrite') }}
									</option>
									<option
										:selected="share.permission === PERMISSIONS.ADMIN"
										:value="PERMISSIONS.ADMIN"
									>
										{{ $t('project.share.permission.admin') }}
									</option>
								</select>
							</div>
							<XButton
								danger
								icon="trash-alt"
								:aria-label="$t('project.share.userTeam.remove', {type: typeName})"
								@click="shareToDelete = share"
							/>
						</td>
					</tr>
				</tbody>
			</table>
		</div>

		<Nothing v-else>
			{{ $t('project.share.userTeam.notShared', {type: typeNames}) }}
		</Nothing>

		<Modal
			:enabled="shareToDelete !== null"
			@close="shareToDelete = null"
			@submit="removeShare()"
		>
			<template #header>
				<span>{{
					$t('project.share.userTeam.removeHeader', {type: typeName, sharable: $t('project.list.title')})
				}}</span>
			</template>
			<template #text>
				<p>{{ $t('project.share.userTeam.removeText', {type: typeName, sharable: $t('project.list.title')}) }}</p>
			</template>
		</Modal>
	</div>
</template>

<script setup lang="ts" generic="Candidate extends Record<string, unknown> & {id?: number}, Share extends {id?: number, permission?: number}">
import {computed, ref, watch, type Ref} from 'vue'
import {normalizeSharePermission} from '@/client/queries/projectShares'
import {PERMISSIONS, type Permission} from '@/constants/permissions'
import Multiselect from '@/components/input/Multiselect.vue'
import Nothing from '@/components/misc/Nothing.vue'

type Identified<T extends {id?: number}> = T & {id: number}

const props = defineProps<{
	projectId: number
	userIsAdmin: boolean
	shares: Share[]
	candidates: Candidate[]
	searchLoading: boolean
	searchLabel: string
	isMutating: boolean
	typeName: string
	typeNames: string
	shareName: (share: Identified<Share>) => string
	add: (projectId: number, candidate: Identified<Candidate>) => Promise<unknown>
	updatePermission: (projectId: number, share: Identified<Share>, permission: Permission) => Promise<unknown>
	remove: (projectId: number, share: Identified<Share>) => Promise<unknown>
}>()

const search = defineModel<string>('search', {required: true})

defineSlots<{
	candidate(props: {candidate: Identified<Candidate>}): unknown
	share(props: {share: Identified<Share>}): unknown
}>()

function hasId<T extends {id?: number}>(item: T): item is Identified<T> {
	return typeof item.id === 'number'
}

const identifiedShares = computed(() => props.shares.filter(hasId))
const unsharedCandidates = computed(() => props.candidates.filter(hasId).filter(candidate => !identifiedShares.value.some(share => share.id === candidate.id)))
const selected = ref(null) as Ref<Identified<Candidate> | null>
const selectedPermission = ref<Record<number, Permission>>({})
const shareToDelete = ref(null) as Ref<Identified<Share> | null>

watch(identifiedShares, shares => {
	selectedPermission.value = Object.fromEntries(shares.map(share => [share.id, normalizeSharePermission(share.permission)]))
}, {immediate: true})
watch(() => props.projectId, () => {
	selected.value = null
	search.value = ''
	shareToDelete.value = null
})

async function addSelected() {
	const candidate = selected.value
	const projectId = props.projectId
	if (!candidate || props.isMutating) return
	try {
		await props.add(projectId, candidate)
	} catch {
		return
	}
	if (props.projectId === projectId && selected.value === candidate) {
		selected.value = null
		search.value = ''
	}
}

async function changePermission(share: Identified<Share>) {
	const projectId = props.projectId
	if (!props.isMutating) {
		await props.updatePermission(projectId, share, normalizeSharePermission(selectedPermission.value[share.id])).catch(() => {})
	}
	if (props.projectId === projectId) {
		selectedPermission.value[share.id] = normalizeSharePermission(identifiedShares.value.find(item => item.id === share.id)?.permission)
	}
}

async function removeShare() {
	const share = shareToDelete.value
	if (!share || props.isMutating) return
	await props.remove(props.projectId, share).catch(() => {})
	if (shareToDelete.value === share) shareToDelete.value = null
}
</script>
