<template>
	<div class="card is-fullwidth">
		<header class="card-header">
			<p class="card-header-title">
				Share links
			</p>
		</header>
		<div class="card-content content sharables-list">
			<form @submit.prevent="add()" class="add-form">
				<p>
					Share with a link:
				</p>
				<div class="field has-addons">
					<div class="control">
						<div class="select">
							<select v-model="selectedRight">
								<option :value="rights.READ">Read only</option>
								<option :value="rights.READ_WRITE">Read & write</option>
								<option :value="rights.ADMIN">Admin</option>
							</select>
						</div>
					</div>
					<div class="control">
						<button class="button is-success" type="submit">
							Share
						</button>
					</div>
				</div>
			</form>
			<table class="table is-striped is-hoverable is-fullwidth link-share-list">
				<tbody>
				<tr>
					<th>Link</th>
					<th>Shared by</th>
					<th>Right</th>
					<th>Delete</th>
				</tr>
				<template v-if="linkShares.length > 0">
					<tr :key="s.id" v-for="s in linkShares">
						<td>
							<div class="field has-addons no-input-mobile">
								<div class="control">
									<input :value="getShareLink(s.hash)" class="input" readonly type="text"/>
								</div>
								<div class="control">
									<a @click="copy(getShareLink(s.hash))" class="button is-success noshadow" v-tooltip="'Copy to clipboard'">
									<span class="icon">
										<icon icon="paste"/>
									</span>
									</a>
								</div>
							</div>
						</td>
						<td>
							{{ s.sharedBy.getDisplayName() }}
						</td>
						<td class="type">
							<template v-if="s.right === rights.ADMIN">
						<span class="icon is-small">
							<icon icon="lock"/>
						</span>
								Admin
							</template>
							<template v-else-if="s.right === rights.READ_WRITE">
						<span class="icon is-small">
							<icon icon="pen"/>
						</span>
								Write
							</template>
							<template v-else>
						<span class="icon is-small">
							<icon icon="users"/>
						</span>
								Read-only
							</template>
						</td>
						<td class="actions">
							<button @click="() => {linkIdToDelete = s.id; showDeleteModal = true}"
									class="button is-danger icon-only">
							<span class="icon">
								<icon icon="trash-alt"/>
							</span>
							</button>
						</td>
					</tr>
				</template>
				</tbody>
			</table>
		</div>

		<modal
			@close="showDeleteModal = false"
			@submit="remove()"
			v-if="showDeleteModal">
			<span slot="header">Remove a link share</span>
			<p slot="text">Are you sure you want to remove this link share?<br/>
				It will no longer be possible to access this list with this link share.<br/>
				<b>This CANNOT BE UNDONE!</b></p>
		</modal>
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
		listId() { // watch it
			this.load()
		},
	},
	computed: mapState({
		frontendUrl: state => state.config.frontendUrl,
	}),
	methods: {
		load() {
			// If listId == 0 the list on the calling component wasn't already loaded, so we just bail out here
			if (this.listId === 0) {
				return
			}

			this.linkShareService.getAll({listId: this.listId})
				.then(r => {
					this.linkShares = r
				})
				.catch(e => {
					this.error(e, this)
				})
		},
		add() {
			let newLinkShare = new LinkShareModel({right: this.selectedRight, listId: this.listId})
			this.linkShareService.create(newLinkShare)
				.then(() => {
					this.selectedRight = rights.READ
					this.success({message: 'The link share was successfully created'}, this)
					this.load()
				})
				.catch(e => {
					this.error(e, this)
				})
		},
		remove() {
			let linkshare = new LinkShareModel({id: this.linkIdToDelete, listId: this.listId})
			this.linkShareService.delete(linkshare)
				.then(() => {
					this.success({message: 'The link share was successfully deleted'}, this)
					this.load()
				})
				.catch(e => {
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
