<template>
	<div>
		<p class="has-text-weight-bold">
			{{ $t('project.share.links.title') }}
			<span
				v-tooltip="$t('project.share.links.explanation')"
				class="is-size-7 has-text-grey is-italic mis-3"
			>
				{{ $t('project.share.links.what') }}
			</span>
		</p>

		<div class="sharables-project">
			<XButton
				v-if="!(linkShares.length === 0 || showNewForm)"
				icon="plus"
				class="mbe-4"
				@click="showNewForm = true"
			>
				{{ $t('project.share.links.create') }}
			</XButton>

			<div
				v-if="linkShares.length === 0 || showNewForm"
				class="p-4"
			>
				<div class="field">
					<label
						class="label"
						for="linkShareRight"
					>
						{{ $t('project.share.permission.title') }}
					</label>
					<div class="control">
						<div class="select">
							<select
								id="linkShareRight"
								v-model="selectedPermission"
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
					</div>
				</div>
				<div class="field">
					<label
						class="label"
						for="linkShareName"
					>
						{{ $t('project.share.links.name') }}
					</label>
					<div class="control">
						<input
							id="linkShareName"
							v-model="name"
							v-tooltip="$t('project.share.links.nameExplanation')"
							class="input"
							:placeholder="$t('project.share.links.namePlaceholder')"
						>
					</div>
				</div>
				<div class="field">
					<label
						class="label"
						for="linkSharePassword"
					>
						{{ $t('project.share.links.password') }}
					</label>
					<div class="control">
						<input
							id="linkSharePassword"
							v-model="password"
							v-tooltip="$t('project.share.links.passwordExplanation')"
							type="password"
							class="input"
							:placeholder="$t('user.auth.passwordPlaceholder')"
						>
					</div>
				</div>
				<XButton
					icon="plus"
					@click="add(projectId)"
				>
					{{ $t('project.share.share') }}
				</XButton>
			</div>

			<table
				v-if="linkShares.length > 0"
				class="table has-actions is-striped is-hoverable is-fullwidth"
			>
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
									<strong>{{ getDisplayName(s.sharedBy) }}</strong>
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
						
							<div class="field has-addons no-input-mobile">
								<div class="control">
									<input
										:value="shareLinks[s.id]"
										class="input"
										readonly
										type="text"
									>
								</div>
								<div class="control">
									<XButton
										v-tooltip="$t('misc.copy')"
										:shadow="false"
										@click="copy(shareLinks[s.id])"
									>
										<span class="icon">
											<Icon icon="paste" />
										</span>
									</XButton>
								</div>
							</div>
						</td>
						<td v-if="availableViews.length > 0">
							<div class="select">
								<select v-model="selectedViews[s.id]">
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
								class="is-danger"
								icon="trash-alt"
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
import {ref, watch, computed, shallowReactive} from 'vue'
import {useI18n} from 'vue-i18n'

import {PERMISSIONS} from '@/constants/permissions'
import LinkShareModel from '@/models/linkShare'

import type {ILinkShare} from '@/modelTypes/ILinkShare'
import type {IProject} from '@/modelTypes/IProject'

import LinkShareService from '@/services/linkShare'

import {useCopyToClipboard} from '@/composables/useCopyToClipboard'
import {success} from '@/message'
import {getDisplayName} from '@/models/user'
import {useConfigStore} from '@/stores/config'
import {useProjectStore} from '@/stores/projects'
import type {IProjectView} from '@/modelTypes/IProjectView'

const props = withDefaults(defineProps<{
	projectId?: IProject['id'],
}>(), {
	projectId: 0,
})

const {t} = useI18n({useScope: 'global'})

const linkShares = ref<ILinkShare[]>([])
const linkShareService = shallowReactive(new LinkShareService())
const selectedPermission = ref(PERMISSIONS.READ)
const name = ref('')
const password = ref('')
const showDeleteModal = ref(false)
const linkIdToDelete = ref(0)
const showNewForm = ref(false)

const projectStore = useProjectStore()

const availableViews = computed<IProjectView[]>(() => projectStore.projects[props.projectId]?.views || [])
const copy = useCopyToClipboard()
watch(
	() => props.projectId,
	load,
	{immediate: true},
)

const configStore = useConfigStore()
const frontendUrl = computed(() => configStore.frontendUrl)

async function load(projectId: IProject['id']) {
	// If projectId == 0 the project on the calling component wasn't already loaded, so we just bail out here
	if (projectId === 0) {
		return
	}

	linkShares.value = await linkShareService.getAll({projectId})
}

type SelectedViewMapper = Record<IProject['id'], IProjectView['id']>

const selectedViews = ref<SelectedViewMapper>({})

watch(() => ([linkShares.value, availableViews.value]), ([newLinkShares, newProjectViews]) => {
	if (!newLinkShares?.length || !newProjectViews?.length) {
		selectedViews.value = {}
		return
	}

	newLinkShares.forEach((linkShare) => {
		selectedViews.value[linkShare.id] = newProjectViews.length > 0 ? newProjectViews[0].id : null
	})
}, {
	immediate:true,
	deep: true,
})


async function add(projectId: IProject['id']) {
	const newLinkShare = new LinkShareModel({
		permission: selectedPermission.value,
		projectId,
		name: name.value,
		password: password.value,
	})
	await linkShareService.create(newLinkShare)
	selectedPermission.value = PERMISSIONS.READ
	name.value = ''
	password.value = ''
	showNewForm.value = false
	success({message: t('project.share.links.createSuccess')})
	await load(projectId)
}

async function remove(projectId: IProject['id']) {
	try {
		await linkShareService.delete(new LinkShareModel({
			id: linkIdToDelete.value,
			projectId,
		}))
		success({message: t('project.share.links.deleteSuccess')})
		await load(projectId)
	} finally {
		showDeleteModal.value = false
	}
}

function getShareLink(hash: string, viewId: IProjectView['id']|null) {
	return frontendUrl.value + 'share/' + hash + '/auth' + (viewId ? '?view=' + viewId : '')
}

const shareLinks = computed(() => {
	return linkShares.value.reduce((links, linkShare) => {
		links[linkShare.id] = getShareLink(linkShare.hash, selectedViews.value[linkShare.id] ?? null)
		return links
	}, {} as {[id: string]: string },
	)
})
</script>

<style lang="scss" scoped>
// FIXME: I think this is not needed
.sharables-project:not(.card-content) {
	overflow-y: auto
}
</style>
