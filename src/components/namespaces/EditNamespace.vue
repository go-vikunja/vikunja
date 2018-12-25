<template>
	<div class="loader-container" v-bind:class="{ 'is-loading': loading}">
		<div class="card">
			<header class="card-header">
				<p class="card-header-title">
					Edit Namespace
				</p>
			</header>
			<div class="card-content">
				<div class="content">
					<form  @submit.prevent="submit()">
						<div class="field">
							<label class="label" for="namespacetext">Namespace Name</label>
							<div class="control">
								<input v-focus :class="{ 'disabled': loading}" :disabled="loading" class="input" type="text" id="namespacetext" placeholder="The namespace text is here..." v-model="namespace.name">
							</div>
						</div>
						<div class="field">
							<label class="label" for="namespacedescription">Description</label>
							<div class="control">
								<textarea :class="{ 'disabled': loading}" :disabled="loading" class="textarea" placeholder="The namespaces description goes here..." id="namespacedescription" v-model="namespace.description"></textarea>
							</div>
						</div>
					</form>

					<div class="columns bigbuttons">
						<div class="column">
							<button @click="submit()" class="button is-primary is-fullwidth" :class="{ 'is-loading': loading}">
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

		<manageusers :id="namespace.id" type="namespace" :userIsAdmin="userIsAdmin" />

		<manageteams :id="namespace.id" type="namespace" :userIsAdmin="userIsAdmin" />

		<modal
				v-if="showDeleteModal"
				@close="showDeleteModal = false"
				v-on:submit="deleteNamespace()">
			<span slot="header">Delete the namespace</span>
			<p slot="text">Are you sure you want to delete this namespace and all of its contents?
				<br/>This includes lists & tasks and <b>CANNOT BE UNDONE!</b></p>
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
        name: "EditNamespace",
        data() {
            return {
                namespace: {title: '', description:''},
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

            this.namespace.id = this.$route.params.id
        },
        created() {
            this.loadNamespace()
        },
        watch: {
            // call again the method if the route changes
            '$route': 'loadNamespace'
        },
		methods: {
            loadNamespace() {
				const cancel = message.setLoading(this)

                HTTP.get(`namespaces/` + this.$route.params.id, {headers: {'Authorization': 'Bearer ' + localStorage.getItem('token')}})
                    .then(response => {
                        this.$set(this, 'namespace', response.data)
						if (response.data.owner.id === this.user.infos.id) {
							this.userIsAdmin = true
						}
						cancel()
                    })
                    .catch(e => {
                        cancel()
						this.handleError(e)
                    })
            },
            submit() {
				const cancel = message.setLoading(this)

                HTTP.post(`namespaces/` + this.$route.params.id, this.namespace, {headers: {'Authorization': 'Bearer ' + localStorage.getItem('token')}})
                    .then(response => {
                        // Update the namespace in the parent
                        for (const n in this.$parent.namespaces) {
                            if (this.$parent.namespaces[n].id === response.data.id) {
                                response.data.lists = this.$parent.namespaces[n].lists
                                this.$set(this.$parent.namespaces, n, response.data)
                            }
                        }
                        this.handleSuccess({message: 'The namespace was successfully updated.'})
						cancel()
                    })
                    .catch(e => {
                        cancel()
						this.handleError(e)
                    })
            },
            deleteNamespace() {
				const cancel = message.setLoading(this)
                HTTP.delete(`namespaces/` + this.$route.params.id, {headers: {'Authorization': 'Bearer ' + localStorage.getItem('token')}})
                    .then(() => {
                        this.handleSuccess({message: 'The namespace was successfully deleted.'})
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

<style scoped>
	.bigbuttons{
		margin-top: 0.5rem;
	}
</style>