<template>
	<div>
		<h3 class="has-text-weight-bold share-heading">
			{{ $t('project.share.links.title') }}
			<span
				v-tooltip="$t('project.share.links.explanation')"
				class="is-size-7 has-text-grey is-italic mis-3"
			>
				{{ $t('project.share.links.what') }}
			</span>
		</h3>

		<div class="sharables-project">
			<XButton
				v-if="!showNewForm"
				icon="plus"
				class="mbe-4"
				@click="showNewForm = true"
			>
				{{ $t('project.share.links.create') }}
			</XButton>

			<div
				v-if="showNewForm"
				class="p-4"
			>
				<FormField :label="$t('project.share.permission.title')">
					<template #default="{ id }">
						<div class="select">
							<select
								:id="id"
								v-model="draft.permission"
							>
								<option :value="PERMISSIONS.READ">
									{{ $t('project.share.permission.read') }}
								</option>
								<option :value="PERMISSIONS.READ_WRITE">
									{{ $t('project.share.permission.readWrite') }}
								</option>
								<option :value="PERMISSIONS.ADMIN">
									{{ $t('project.share.permission.admin') }}
								</option>
							</select>
						</div>
					</template>
				</FormField>
				<FormField
					id="linkShareName"
					v-model="draft.name"
					v-tooltip="$t('project.share.links.nameExplanation')"
					:label="$t('project.share.links.name')"
					:placeholder="$t('project.share.links.namePlaceholder')"
				/>
				<FormField
					id="linkSharePassword"
					v-model="draft.password"
					v-tooltip="$t('project.share.links.passwordExplanation')"
					:label="$t('project.share.links.password')"
					type="password"
					:placeholder="$t('user.auth.passwordPlaceholder')"
					autocomplete="new-password"
				/>
				<XButton
					icon="plus"
					:loading="isMutating"
					:disabled="isMutating"
					@click="add(projectId)"
				>
					{{ $t('project.share.share') }}
				</XButton>
			</div>

			<div
				v-if="linkShares.length > 0"
				class="has-horizontal-overflow"
			>
				<table class="table has-actions is-striped is-hoverable is-fullwidth">
					<thead>
						<tr>
							<th />
							<th v-if="availableViews.length > 0">
								{{ $t('project.share.links.view') }}
							</th>
							<th>{{ $t('project.share.attributes.delete') }}</th>
						</tr>
					</thead>
					<tbody>
						<tr
							v-for="s in linkShares"
							:key="s.id"
						>
							<td>
								<p
									v-if="s.name !== ''"
									class="mbe-2 is-italic"
								>
									{{ s.name }}
								</p>

								<p class="mbe-2">
									<i18n-t
										keypath="project.share.links.sharedBy"
										scope="global"
									>
										<strong>{{ getDisplayName(s.shared_by) }}</strong>
									</i18n-t>
								</p>

								<p class="mbe-2">
									<template v-if="s.permission === PERMISSIONS.ADMIN">
										<span class="icon is-small">
											<Icon icon="lock" />
										</span>&nbsp;
										{{ $t('project.share.permission.admin') }}
									</template>
									<template v-else-if="s.permission === PERMISSIONS.READ_WRITE">
										<span class="icon is-small">
											<Icon icon="pen" />
										</span>&nbsp;
										{{ $t('project.share.permission.readWrite') }}
									</template>
									<template v-else>
										<span class="icon is-small">
											<Icon icon="users" />
										</span>&nbsp;
										{{ $t('project.share.permission.read') }}
									</template>
								</p>
						
								<FormField
									:model-value="shareLinks[s.id]"
									readonly
									type="text"
								>
									<template #addon>
										<XButton
											v-tooltip="$t('misc.copy')"
											:shadow="false"
											@click="copy(shareLinks[s.id])"
										>
											<span class="icon">
												<Icon icon="paste" />
											</span>
										</XButton>
									</template>
								</FormField>
							</td>
							<td v-if="availableViews.length > 0">
								<div class="select">
									<select
										:value="selectedViews[s.id]"
										:aria-label="$t('project.share.links.view')"
										@change="pickedViews[s.id] = Number(($event.target as HTMLSelectElement).value)"
									>
										<option
											v-for="(view) in availableViews"
											:key="view.id"
											:value="view.id"
										>
											{{ view.title }}
										</option>
									</select>
								</div>
							</td>
							<td class="actions">
								<XButton
									danger
									icon="trash-alt"
									:aria-label="$t('project.share.links.remove')"
									@click="
										() => {
											linkIdToDelete = s.id
											showDeleteModal = true
										}
									"
								/>
							</td>
						</tr>
					</tbody>
				</table>
			</div>
		</div>

		<Modal
			:enabled="showDeleteModal"
			@close="showDeleteModal = false"
			@submit="remove(projectId)"
		>
			<template #header>
				<span>{{ $t('project.share.links.remove') }}</span>
			</template>

			<template #text>
				<p>{{ $t('project.share.links.removeText') }}</p>
			</template>
		</Modal>
	</div>
</template>

<script setup lang="ts">
import {ref, watch, computed} from 'vue'
import {useQuery} from '@tanstack/vue-query'
import type {LinkSharing} from '@/client/generated'
import {linkSharesQuery, createLinkShareDraft, useCreateLinkShareMutation, useDeleteLinkShareMutation} from '@/client/queries/linkShares'
import {PERMISSIONS} from '@/constants/permissions'
import FormField from '@/components/input/FormField.vue'
import {useCopyToClipboard} from '@/composables/useCopyToClipboard'
import {getDisplayName} from '@/helpers/user'
import {useConfigStore} from '@/stores/config'
import {useProjectViews} from '@/composables/useProjectViews'

const props = withDefaults(defineProps<{projectId?: number}>(), {projectId: 0})
const sharesQuery = useQuery(computed(() => ({...linkSharesQuery(props.projectId), enabled: props.projectId > 0})))
type VisibleShare = LinkSharing & Required<Pick<LinkSharing, 'id' | 'hash'>>
const linkShares = computed(() => (sharesQuery.data.value ?? []).filter((share): share is VisibleShare => typeof share.id === 'number' && typeof share.hash === 'string'))
const createMutation = useCreateLinkShareMutation()
const deleteMutation = useDeleteLinkShareMutation()
const isMutating = computed(() => createMutation.isPending.value || deleteMutation.isPending.value)
const draft = ref(createLinkShareDraft())
const showDeleteModal = ref(false)
const linkIdToDelete = ref(0)
const showNewForm = ref(false)
const {views: availableViews} = useProjectViews(() => props.projectId)
const copy = useCopyToClipboard()
const pickedViews = ref<Record<number, number>>({})
const configStore = useConfigStore()

watch(() => props.projectId, () => {
	draft.value = createLinkShareDraft()
	showNewForm.value = false
	showDeleteModal.value = false
}, {flush: 'sync'})

async function add(projectId: number) {
	if (projectId <= 0 || isMutating.value) return
	try {
		await createMutation.mutateAsync({projectId, share: draft.value})
	} catch {
		return
	} finally {
		// Evicts the plaintext password from the mutation cache.
		createMutation.reset()
	}
	if (props.projectId === projectId) {
		draft.value = createLinkShareDraft()
		showNewForm.value = false
	}
}

function remove(projectId: number) {
	if (isMutating.value) return
	deleteMutation.mutate({projectId, id: linkIdToDelete.value}, {
		onSettled: () => {
			if (props.projectId === projectId) showDeleteModal.value = false
		},
	})
}

const selectedViews = computed(() => {
	const viewIds = availableViews.value.flatMap(view => view.id === undefined ? [] : [view.id])
	return Object.fromEntries(linkShares.value.map(share => {
		const picked = pickedViews.value[share.id]
		return [share.id, picked !== undefined && viewIds.includes(picked) ? picked : viewIds[0] ?? null]
	}))
})

const shareLinks = computed(() => Object.fromEntries(linkShares.value.map(share => {
	const viewId = selectedViews.value[share.id]
	return [share.id, `${configStore.frontend_url}share/${share.hash}/auth${viewId ? `?view=${viewId}` : ''}`]
})))
</script>

<style lang="scss" scoped>
// FIXME: I think this is not needed
.sharables-project:not(.card-content) {
	overflow-y: auto
}
</style>
