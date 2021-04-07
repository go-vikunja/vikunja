<template>
	<div>
		<p class="has-text-weight-bold">Share Links</p>
		<div class="sharables-list">
			<div class="p-4">
				<p>Share with a link:</p>
				<div class="field has-addons">
					<div class="control">
						<input
							class="input"
							placeholder="Name"
							v-tooltip="'All actions done by this link share will show up with the name.'"
							v-model="name"
						/>
					</div>
					<div class="control">
						<div class="select">
							<select v-model="selectedRight">
								<option :value="rights.READ">Read only</option>
								<option :value="rights.READ_WRITE">
									Read & write
								</option>
								<option :value="rights.ADMIN">Admin</option>
							</select>
						</div>
					</div>
					<div class="control">
						<x-button @click="add"> Share</x-button>
					</div>
				</div>
			</div>
			<table
				class="table has-actions is-striped is-hoverable is-fullwidth link-share-list"
				v-if="linkShares.length > 0"
			>
				<thead>
				<tr>
					<th>Link</th>
					<th>Name</th>
					<th>Shared&nbsp;by</th>
					<th>Right</th>
					<th>Delete</th>
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
									v-tooltip="'Copy to clipboard'"
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
						<i v-else>No name set</i>
					</td>
					<td>
						{{ s.sharedBy.getDisplayName() }}
					</td>
					<td class="type">
						<template v-if="s.right === rights.ADMIN">
							<span class="icon is-small">
								<icon icon="lock"/>
							</span>&nbsp;
							Admin
						</template>
						<template v-else-if="s.right === rights.READ_WRITE">
							<span class="icon is-small">
								<icon icon="pen"/>
							</span>&nbsp;
							Write
						</template>
						<template v-else>
							<span class="icon is-small">
								<icon icon="users"/>
							</span>&nbsp;
							Read-only
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
				<span slot="header">Remove a link share</span>
				<p slot="text">
					Are you sure you want to remove this link share?<br/>
					It will no longer be possible to access this list with this link
					share.<br/>
					<b>This CANNOT BE UNDONE!</b>
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
			showDeleteModal: false,
			linkIdToDelete: 0,
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
					this.error(e, this)
				})
		},
		add() {
			let newLinkShare = new LinkShareModel({
				right: this.selectedRight,
				listId: this.listId,
				name: this.name,
			})
			this.linkShareService
				.create(newLinkShare)
				.then(() => {
					this.selectedRight = rights.READ
					this.success(
						{message: 'The link share was successfully created'},
						this
					)
					this.load()
				})
				.catch((e) => {
					this.error(e, this)
				})
		},
		remove() {
			let linkshare = new LinkShareModel({
				id: this.linkIdToDelete,
				listId: this.listId,
			})
			this.linkShareService
				.delete(linkshare)
				.then(() => {
					this.success(
						{message: 'The link share was successfully deleted'},
						this
					)
					this.load()
				})
				.catch((e) => {
					this.error(e, this)
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
