<template>
	<div>
		<p class="has-text-weight-bold">
			Share Links
			<span

				class="is-size-7"
				v-tooltip="'Share Links allow you to easily share a list with other users who don\'t have an account on Vikunja.'">
				What is a share link?
			</span>
		</p>

		<div class="sharables-list">

			<x-button
				v-if="!(linkShares.length === 0 || showNewForm)"
				@click="showNewForm = true"
				icon="plus"
				class="mb-4">
				Create a new link share
			</x-button>

			<div class="p-4" v-if="linkShares.length === 0 || showNewForm">
				<div class="field">
					<label class="label" for="linkShareRight">
						Right
					</label>
					<div class="control">
						<div class="select">
							<select v-model="selectedRight" id="linkShareRight">
								<option :value="rights.READ">Read only</option>
								<option :value="rights.READ_WRITE">
									Read & write
								</option>
								<option :value="rights.ADMIN">Admin</option>
							</select>
						</div>
					</div>
				</div>
				<div class="field">
					<label class="label" for="linkShareName">
						Name (optional)
					</label>
					<div class="control">
						<input
							id="linkShareName"
							class="input"
							placeholder="e.g. Lorem Ipsum"
							v-tooltip="'All actions done by this link share will show up with the name.'"
							v-model="name"
						/>
					</div>
				</div>
				<div class="field">
					<label class="label" for="linkSharePassword">
						Password (optional)
					</label>
					<div class="control">
						<input
							id="linkSharePassword"
							type="password"
							class="input"
							placeholder="e.g. ••••••••••••"
							v-tooltip="'When authenticating, the user will be required to enter this password.'"
							v-model="password"
						/>
					</div>
				</div>
				<x-button @click="add" icon="plus">Share</x-button>
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
					this.error(e, this)
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
			const linkshare = new LinkShareModel({
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
