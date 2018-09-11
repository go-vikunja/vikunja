<template>
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
							<input :class="{ 'disabled': loading}" :disabled="loading" class="input" type="text" id="namespacetext" placeholder="The namespace text is here..." v-model="namespace.name">
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
						<button @click="submit()" class="button is-success is-fullwidth" :class="{ 'is-loading': loading}">
							Save
						</button>
					</div>
					<div class="column is-1">
						<button @click="deleteNamespace()" class="button is-danger is-fullwidth" :class="{ 'is-loading': loading}">
							<span class="icon is-small">
								<icon icon="trash-alt"/>
							</span>
						</button>
					</div>
				</div>
			</div>
		</div>
	</div>
</template>

<script>
    import auth from '../../auth'
    import router from '../../router'
    import {HTTP} from '../../http-common'
    import message from '../../message'
	
    export default {
        name: "EditNamespace",
        data() {
            return {
                namespace: {title: '', description:''},
                error: '',
                loading: false,
            }
        },
        beforeMount() {
            // Check if the user is already logged in, if so, redirect him to the homepage
            if (!auth.user.authenticated) {
                router.push({name: 'home'})
            }
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
                this.loading = true

                HTTP.get(`namespaces/` + this.$route.params.id, {headers: {'Authorization': 'Bearer ' + localStorage.getItem('token')}})
                    .then(response => {
                        this.$set(this, 'namespace', response.data)
                        this.loading = false
                    })
                    .catch(e => {
                        this.handleError(e)
                    })
            },
            submit() {
                this.loading = true

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
                    })
                    .catch(e => {
                        this.handleError(e)
                    })
            },
            deleteNamespace() {
				// TODO: add better looking modal to ask the user if he is sure
				if (!confirm('Are you sure you want to delete this namespace and all of its contents? This includes lists & tasks and CANNOT BE UNDONE!')) {
					return
				}

                HTTP.delete(`namespaces/` + this.$route.params.id, {headers: {'Authorization': 'Bearer ' + localStorage.getItem('token')}})
                    .then(() => {
                        this.handleSuccess({message: 'The namespace was successfully deleted.'})
                        router.push({name: 'home'})
                    })
                    .catch(e => {
                        this.handleError(e)
                    })
			},
            handleError(e) {
                this.loading = false
                message.error(e, this)
            },
            handleSuccess(e) {
                this.loading = false
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