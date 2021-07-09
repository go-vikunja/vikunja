<template>
	<div>
		<p class="has-text-weight-bold">
			{{ $t('list.share.links.title') }}
			<span
				class="is-size-7"
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
								<option :value="rights.READ">
									{{ $t('list.share.right.read') }}
								</option>
								<option :value="rights.READ_WRITE">
									{{ $t('list.share.right.readWrite') }}
								</option>
								<option :value="rights.ADMIN">
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
				<x-button @click="add" icon="plus">
					{{ $t('list.share.share') }}
				</x-button>
			</div>

			<table
				class="table has-actions is-striped is-hoverable is-fullwidth link-share-list"
				v-if="linkShares.length > 0"
			>
				<thead>
				<tr>
					<th>{{ $t('list.share.attributes.link') }}</th>
					<th>{{ $t('list.share.attributes.name') }}</th>
					<th>{{ $t('list.share.attributes.sharedBy') }}</th>
					<th>{{ $t('list.share.attributes.right') }}</th>
					<th>{{ $t('list.share.attributes.delete') }}</th>
				</tr>
				</thead>
				<tbody>
				<tr :key="s.id" v-for="s in linkShares">
					<td>
						<div class="field has-addons no-input-mobile">
							<div class="control">
								<input
									:value="getShareLink(s.hash)"
									class="input"
									readonly
									type="text"
								/>
							</div>
							<div class="control">
								<x-button
									@click="copy(getShareLink(s.hash))"
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
						<template v-if="s.name !== ''">
							{{ s.name }}
						</template>
						<i v-else>{{ $t('list.share.links.noName') }}</i>
					</td>
					<td>
						{{ s.sharedBy.getDisplayName() }}
					</td>
					<td class="type">
						<template v-if="s.right === rights.ADMIN">
							<span class="icon is-small">
								<icon icon="lock"/>
							</span>&nbsp;
							{{ $t('list.share.right.admin') }}
						</template>
						<template v-else-if="s.right === rights.READ_WRITE">
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

		<transition name="modal">
			<modal
				@close="showDeleteModal = false"
				@submit="remove()"
				v-if="showDeleteModal"
			>
				<span slot="header">{{ $t('list.share.links.remove') }}</span>
				<p slot="text">
					{{ $t('list.share.links.removeText') }}
				</p>
			</modal>
		</transition>
	</div>
</template>

<script>
import rights from '../../models/rights'

import LinkShareService from '../../services/linkShare'
import LinkShareModel from '../../models/linkShare'

import copy from 'copy-to-clipboard'
import {mapState} from 'vuex'

export default {
	name: 'linkSharing',
	props: {
		listId: {
			default: 0,
			required: true,
		},
	},
	data() {
		return {
			linkShares: [],
			linkShareService: LinkShareService,
			newLinkShare: LinkShareModel,
			rights: rights,
			selectedRight: rights.READ,
			name: '',
			password: '',
			showDeleteModal: false,
			linkIdToDelete: 0,
			showNewForm: false,
		}
	},
	beforeMount() {
		this.linkShareService = new LinkShareService()
	},
	created() {
		this.linkShareService = new LinkShareService()
		this.load()
	},
	watch: {
		listId() {
			// watch it
			this.load()
		},
	},
	computed: mapState({
		frontendUrl: (state) => state.config.frontendUrl,
	}),
	methods: {
		load() {
			// If listId == 0 the list on the calling component wasn't already loaded, so we just bail out here
			if (this.listId === 0) {
				return
			}

			this.linkShareService
				.getAll({listId: this.listId})
				.then((r) => {
					this.linkShares = r
				})
				.catch((e) => {
					this.error(e)
				})
		},
		add() {
			const newLinkShare = new LinkShareModel({
				right: this.selectedRight,
				listId: this.listId,
				name: this.name,
				password: this.password,
			})
			this.linkShareService
				.create(newLinkShare)
				.then(() => {
					this.selectedRight = rights.READ
					this.name = ''
					this.password = ''
					this.showNewForm = false
					this.success({message: this.$t('list.share.links.createSuccess')})
					this.load()
				})
				.catch((e) => {
					this.error(e)
				})
		},
		remove() {
			const linkshare = new LinkShareModel({
				id: this.linkIdToDelete,
				listId: this.listId,
			})
			this.linkShareService
				.delete(linkshare)
				.then(() => {
					this.success({message: this.$t('list.share.links.deleteSuccess')})
					this.load()
				})
				.catch((e) => {
					this.error(e)
				})
				.finally(() => {
					this.showDeleteModal = false
				})
		},
		copy(text) {
			copy(text)
		},
		getShareLink(hash) {
			return this.frontendUrl + 'share/' + hash + '/auth'
		},
	},
}
</script>
