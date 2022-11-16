<template>
	<div>
		<p class="has-text-weight-bold">
			{{ $t('list.share.links.title') }}
			<span
				class="is-size-7 has-text-grey is-italic ml-3"
				v-tooltip="$t('list.share.links.explanation')">
				{{ $t('list.share.links.what') }}
			</span>
		</p>

		<div class="sharables-list">
			<x-button
				v-if="!(linkShares.length === 0 || showNewForm)"
				@click="showNewForm = true"
				icon="plus"
				class="mb-4">
				{{ $t('list.share.links.create') }}
			</x-button>

			<div class="p-4" v-if="linkShares.length === 0 || showNewForm">
				<div class="field">
					<label class="label" for="linkShareRight">
						{{ $t('list.share.right.title') }}
					</label>
					<div class="control">
						<div class="select">
							<select v-model="selectedRight" id="linkShareRight">
								<option :value="RIGHTS.READ">
									{{ $t('list.share.right.read') }}
								</option>
								<option :value="RIGHTS.READ_WRITE">
									{{ $t('list.share.right.readWrite') }}
								</option>
								<option :value="RIGHTS.ADMIN">
									{{ $t('list.share.right.admin') }}
								</option>
							</select>
						</div>
					</div>
				</div>
				<div class="field">
					<label class="label" for="linkShareName">
						{{ $t('list.share.links.name') }}
					</label>
					<div class="control">
						<input
							id="linkShareName"
							class="input"
							:placeholder="$t('list.share.links.namePlaceholder')"
							v-tooltip="$t('list.share.links.nameExplanation')"
							v-model="name"
						/>
					</div>
				</div>
				<div class="field">
					<label class="label" for="linkSharePassword">
						{{ $t('list.share.links.password') }}
					</label>
					<div class="control">
						<input
							id="linkSharePassword"
							type="password"
							class="input"
							:placeholder="$t('user.auth.passwordPlaceholder')"
							v-tooltip="$t('list.share.links.passwordExplanation')"
							v-model="password"
						/>
					</div>
				</div>
				<x-button @click="add(listId)" icon="plus">
					{{ $t('list.share.share') }}
				</x-button>
			</div>

			<table
				class="table has-actions is-striped is-hoverable is-fullwidth link-share-list"
				v-if="linkShares.length > 0"
			>
				<thead>
				<tr>
					<th></th>
					<th>{{ $t('list.share.links.view') }}</th>
					<th>{{ $t('list.share.attributes.delete') }}</th>
				</tr>
				</thead>
				<tbody>
				<tr :key="s.id" v-for="s in linkShares">
					<td>
						<p class="mb-2 is-italic" v-if="s.name !== ''">
							{{ s.name }}
						</p>

						<p class="mb-2">
							<i18n-t keypath="list.share.links.sharedBy" scope="global">
								<strong>{{ getDisplayName(s.sharedBy) }}</strong>
							</i18n-t>
						</p>

						<p class="mb-2">
							<template v-if="s.right === RIGHTS.ADMIN">
								<span class="icon is-small">
									<icon icon="lock"/>
								</span>&nbsp;
								{{ $t('list.share.right.admin') }}
							</template>
							<template v-else-if="s.right === RIGHTS.READ_WRITE">
								<span class="icon is-small">
									<icon icon="pen"/>
								</span>&nbsp;
								{{ $t('list.share.right.readWrite') }}
							</template>
							<template v-else>
								<span class="icon is-small">
									<icon icon="users"/>
								</span>&nbsp;
								{{ $t('list.share.right.read') }}
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
			@submit="remove(listId)"
		>
			<template #header>
				<span>{{ $t('list.share.links.remove') }}</span>
			</template>

			<template #text>
				<p>{{ $t('list.share.links.removeText') }}</p>
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
import type {IList} from '@/modelTypes/IList'

import LinkShareService from '@/services/linkShare'

import {useCopyToClipboard} from '@/composables/useCopyToClipboard'
import {success} from '@/message'
import {getDisplayName} from '@/models/user'
import type {ListView} from '@/types/ListView'
import {LIST_VIEWS} from '@/types/ListView'
import {useConfigStore} from '@/stores/config'

const props = defineProps({
	listId: {
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

type SelectedViewMapper = Record<IList['id'], ListView>

const selectedView = ref<SelectedViewMapper>({})

const availableViews = computed<Record<ListView, string>>(() => ({
	list: t('list.list.title'),
	gantt: t('list.gantt.title'),
	table: t('list.table.title'),
	kanban: t('list.kanban.title'),
}))

const copy = useCopyToClipboard()
watch(
	() => props.listId,
	load,
	{immediate: true},
)

const configStore = useConfigStore()
const frontendUrl = computed(() => configStore.frontendUrl)

async function load(listId: IList['id']) {
	// If listId == 0 the list on the calling component wasn't already loaded, so we just bail out here
	if (listId === 0) {
		return
	}

	const links = await linkShareService.getAll({listId})
	links.forEach((l: ILinkShare) => {
		selectedView.value[l.id] = 'list'
	})
	linkShares.value = links
}

async function add(listId: IList['id']) {
	const newLinkShare = new LinkShareModel({
		right: selectedRight.value,
		listId,
		name: name.value,
		password: password.value,
	})
	await linkShareService.create(newLinkShare)
	selectedRight.value = RIGHTS.READ
	name.value = ''
	password.value = ''
	showNewForm.value = false
	success({message: t('list.share.links.createSuccess')})
	await load(listId)
}

async function remove(listId: IList['id']) {
	try {
		await linkShareService.delete(new LinkShareModel({
			id: linkIdToDelete.value,
			listId,
		}))
		success({message: t('list.share.links.deleteSuccess')})
		await load(listId)
	} finally {
		showDeleteModal.value = false
	}
}

function getShareLink(hash: string, view: ListView = LIST_VIEWS.LIST) {
	return frontendUrl.value + 'share/' + hash + '/auth?view=' + view
}
</script>

<style lang="scss" scoped>
// FIXME: I think this is not needed
.sharables-list:not(.card-content) {
	overflow-y: auto
}
</style>