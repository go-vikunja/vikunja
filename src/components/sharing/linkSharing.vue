<template>
	<div>
		<p class="has-text-weight-bold">
			{{ $t('project.share.links.title') }}
			<span
				class="is-size-7 has-text-grey is-italic ml-3"
				v-tooltip="$t('project.share.links.explanation')">
				{{ $t('project.share.links.what') }}
			</span>
		</p>

		<div class="sharables-project">
			<x-button
				v-if="!(linkShares.length === 0 || showNewForm)"
				@click="showNewForm = true"
				icon="plus"
				class="mb-4">
				{{ $t('project.share.links.create') }}
			</x-button>

			<div class="p-4" v-if="linkShares.length === 0 || showNewForm">
				<div class="field">
					<label class="label" for="linkShareRight">
						{{ $t('project.share.right.title') }}
					</label>
					<div class="control">
						<div class="select">
							<select v-model="selectedRight" id="linkShareRight">
								<option :value="RIGHTS.READ">
									{{ $t('project.share.right.read') }}
								</option>
								<option :value="RIGHTS.READ_WRITE">
									{{ $t('project.share.right.readWrite') }}
								</option>
								<option :value="RIGHTS.ADMIN">
									{{ $t('project.share.right.admin') }}
								</option>
							</select>
						</div>
					</div>
				</div>
				<div class="field">
					<label class="label" for="linkShareName">
						{{ $t('project.share.links.name') }}
					</label>
					<div class="control">
						<input
							id="linkShareName"
							class="input"
							:placeholder="$t('project.share.links.namePlaceholder')"
							v-tooltip="$t('project.share.links.nameExplanation')"
							v-model="name"
						/>
					</div>
				</div>
				<div class="field">
					<label class="label" for="linkSharePassword">
						{{ $t('project.share.links.password') }}
					</label>
					<div class="control">
						<input
							id="linkSharePassword"
							type="password"
							class="input"
							:placeholder="$t('user.auth.passwordPlaceholder')"
							v-tooltip="$t('project.share.links.passwordExplanation')"
							v-model="password"
						/>
					</div>
				</div>
				<x-button @click="add(projectId)" icon="plus">
					{{ $t('project.share.share') }}
				</x-button>
			</div>

			<table
				class="table has-actions is-striped is-hoverable is-fullwidth"
				v-if="linkShares.length > 0"
			>
				<thead>
				<tr>
					<th></th>
					<th>{{ $t('project.share.links.view') }}</th>
					<th>{{ $t('project.share.attributes.delete') }}</th>
				</tr>
				</thead>
				<tbody>
				<tr :key="s.id" v-for="s in linkShares">
					<td>
						<p class="mb-2 is-italic" v-if="s.name !== ''">
							{{ s.name }}
						</p>

						<p class="mb-2">
							<i18n-t keypath="project.share.links.sharedBy" scope="global">
								<strong>{{ getDisplayName(s.sharedBy) }}</strong>
							</i18n-t>
						</p>

						<p class="mb-2">
							<template v-if="s.right === RIGHTS.ADMIN">
								<span class="icon is-small">
									<icon icon="lock"/>
								</span>&nbsp;
								{{ $t('project.share.right.admin') }}
							</template>
							<template v-else-if="s.right === RIGHTS.READ_WRITE">
								<span class="icon is-small">
									<icon icon="pen"/>
								</span>&nbsp;
								{{ $t('project.share.right.readWrite') }}
							</template>
							<template v-else>
								<span class="icon is-small">
									<icon icon="users"/>
								</span>&nbsp;
								{{ $t('project.share.right.read') }}
							</template>
						</p>
						
						<div class="field has-addons no-input-mobile">
							<div class="control">
								<input
										:value="getShareLink(s.hash, selectedView[s.id])"
										class="input"
										readonly
										type="text"
								/>
							</div>
							<div class="control">
								<x-button
										@click="copy(getShareLink(s.hash, selectedView[s.id]))"
										:shadow="false"
										v-tooltip="$t('misc.copy')"
								>
									<span class="icon">
										<icon icon="paste"/>
									</span>
								</x-button>
							</div>
						</div>
					</td>
					<td>
						<div class="select">
							<select v-model="selectedView[s.id]">
								<option
									v-for="(title, key) in availableViews"
									:value="key"
									:key="key">
									{{ title }}
								</option>
							</select>
						</div>
					</td>
					<td class="actions">
						<x-button
							@click="
									() => {
										linkIdToDelete = s.id
										showDeleteModal = true
									}
								"
							class="is-danger"
							icon="trash-alt"
						/>
					</td>
				</tr>
				</tbody>
			</table>
		</div>

		<modal
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
		</modal>
	</div>
</template>

<script setup lang="ts">
import {ref, watch, computed, shallowReactive} from 'vue'
import {useI18n} from 'vue-i18n'

import {RIGHTS} from '@/constants/rights'
import LinkShareModel from '@/models/linkShare'

import type {ILinkShare} from '@/modelTypes/ILinkShare'
import type {IProject} from '@/modelTypes/IProject'

import LinkShareService from '@/services/linkShare'

import {useCopyToClipboard} from '@/composables/useCopyToClipboard'
import {success} from '@/message'
import {getDisplayName} from '@/models/user'
import type {ProjectView} from '@/types/ProjectView'
import {PROJECT_VIEWS} from '@/types/ProjectView'
import {useConfigStore} from '@/stores/config'

const props = defineProps({
	projectId: {
		default: 0,
		required: true,
	},
})

const {t} = useI18n({useScope: 'global'})

const linkShares = ref<ILinkShare[]>([])
const linkShareService = shallowReactive(new LinkShareService())
const selectedRight = ref(RIGHTS.READ)
const name = ref('')
const password = ref('')
const showDeleteModal = ref(false)
const linkIdToDelete = ref(0)
const showNewForm = ref(false)

type SelectedViewMapper = Record<IProject['id'], ProjectView>

const selectedView = ref<SelectedViewMapper>({})

const availableViews = computed<Record<ProjectView, string>>(() => ({
	list: t('project.list.title'),
	gantt: t('project.gantt.title'),
	table: t('project.table.title'),
	kanban: t('project.kanban.title'),
}))

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

	const links = await linkShareService.getAll({projectId})
	links.forEach((l: ILinkShare) => {
		selectedView.value[l.id] = 'list'
	})
	linkShares.value = links
}

async function add(projectId: IProject['id']) {
	const newLinkShare = new LinkShareModel({
		right: selectedRight.value,
		projectId,
		name: name.value,
		password: password.value,
	})
	await linkShareService.create(newLinkShare)
	selectedRight.value = RIGHTS.READ
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

function getShareLink(hash: string, view: ProjectView = PROJECT_VIEWS.LIST) {
	return frontendUrl.value + 'share/' + hash + '/auth?view=' + view
}
</script>

<style lang="scss" scoped>
// FIXME: I think this is not needed
.sharables-project:not(.card-content) {
	overflow-y: auto
}
</style>