<template>
	<div class="loader-container" v-bind:class="{ 'is-loading': loading}">
		<div class="card">
			<header class="card-header">
				<p class="card-header-title">
					Edit List
				</p>
			</header>
			<div class="card-content">
				<div class="content">
					<form  @submit.prevent="submit()">
						<div class="field">
							<label class="label" for="listtext">List Name</label>
							<div class="control">
								<input :class="{ 'disabled': loading}" :disabled="loading" class="input" type="text" id="listtext" placeholder="The list title goes here..." v-model="list.title">
							</div>
						</div>
						<div class="field">
							<label class="label" for="listdescription">Description</label>
							<div class="control">
								<textarea :class="{ 'disabled': loading}" :disabled="loading" class="textarea" placeholder="The lists description goes here..." id="listdescription" v-model="list.description"></textarea>
							</div>
						</div>
					</form>

					<div class="columns bigbuttons">
						<div class="column">
							<button @click="submit()" class="button is-success is-fullwidth" :class="{ 'is-loading': loading}">
								Save
							</button>
						</div>
						<div class="column is-1">
							<button @click="showDeleteModal = true" class="button is-danger is-fullwidth" :class="{ 'is-loading': loading}">
								<span class="icon is-small">
									<icon icon="trash-alt"/>
								</span>
							</button>
						</div>
					</div>
				</div>
			</div>
		</div>

		<manageusers :id="list.id" type="list" :userIsAdmin="userIsAdmin" />

		<manageteams :id="list.id" type="list" :userIsAdmin="userIsAdmin" />

		<modal
				v-if="showDeleteModal"
				@close="showDeleteModal = false"
				v-on:submit="deleteList()">
			<span slot="header">Delete the list</span>
			<p slot="text">Are you sure you want to delete this list and all of its contents?
				<br/>This includes all tasks and <b>CANNOT BE UNDONE!</b></p>
		</modal>
	</div>
</template>

<script>
    import auth from '../../auth'
    import router from '../../router'
    import {HTTP} from '../../http-common'
	import message from '../../message'
	import manageusers from '../sharing/user'
	import manageteams from '../sharing/team'

    export default {
        name: "EditList",
        data() {
            return {
                list: {title: '', description:''},
                error: '',
                loading: false,
                showDeleteModal: false,
				user: auth.user,
				userIsAdmin: false,
            }
        },
		components: {
			manageusers,
			manageteams,
		},
        beforeMount() {
            // Check if the user is already logged in, if so, redirect him to the homepage
            if (!auth.user.authenticated) {
                router.push({name: 'home'})
            }

            this.list.id = this.$route.params.id
        },
        created() {
            this.loadList()
        },
        watch: {
            // call again the method if the route changes
            '$route': 'loadList'
        },
        methods: {
            loadList() {
                const cancel = message.setLoading(this)

                HTTP.get(`lists/` + this.$route.params.id, {headers: {'Authorization': 'Bearer ' + localStorage.getItem('token')}})
                    .then(response => {
                        this.$set(this, 'list', response.data)
						if (response.data.owner.id === this.user.infos.id) {
							this.userIsAdmin = true
						}
                        cancel()
                    })
                    .catch(e => {
                        this.handleError(e)
                    })
            },
            submit() {
				const cancel = message.setLoading(this)

                HTTP.post(`lists/` + this.$route.params.id, this.list, {headers: {'Authorization': 'Bearer ' + localStorage.getItem('token')}})
                    .then(response => {
                        // Update the list in the parent
                        for (const n in this.$parent.namespaces) {
                            let lists = this.$parent.namespaces[n].lists
                            for (const l in lists) {
                                if (lists[l].id === response.data.id) {
                                    this.$set(this.$parent.namespaces[n].lists, l, response.data)
                                }
                            }
                        }
                        this.handleSuccess({message: 'The list was successfully updated.'})
						cancel()
                    })
                    .catch(e => {
                        cancel()
						this.handleError(e)
                    })
            },
            deleteList() {
				const cancel = message.setLoading(this)
                HTTP.delete(`lists/` + this.$route.params.id, {headers: {'Authorization': 'Bearer ' + localStorage.getItem('token')}})
                    .then(() => {
                        this.handleSuccess({message: 'The list was successfully deleted.'})
						cancel()
                        router.push({name: 'home'})
                    })
                    .catch(e => {
                        cancel()
						this.handleError(e)
                    })
            },
            handleError(e) {
                message.error(e, this)
            },
            handleSuccess(e) {
                message.success(e, this)
            }
        }
    }
</script>
