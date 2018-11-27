<template>
	<div class="content">
		<h3>Create a new namespace</h3>
		<form @submit.prevent="newNamespace">
			<div class="field is-grouped">
				<p class="control has-icons-left is-expanded" v-bind:class="{ 'is-loading': loading}">
					<input class="input" v-bind:class="{ 'disabled': loading}" v-model="namespace.name" type="text" placeholder="The namespace's name goes here...">
					<span class="icon is-small is-left">
										<icon icon="layer-group"/>
					</span>
				</p>
				<p class="control">
					<button type="submit" class="button is-success">
						<span class="icon is-small">
							<icon icon="plus"/>
						</span>
						Add
					</button>
				</p>
			</div>
		</form>
	</div>
</template>

<script>
    import auth from '../../auth'
    import router from '../../router'
    import {HTTP} from '../../http-common'
    import message from '../../message'

    export default {
        name: "NewNamespace",
        data() {
            return {
                namespace: {title: ''},
                error: '',
                loading: false
            }
        },
        beforeMount() {
            // Check if the user is already logged in, if so, redirect him to the homepage
            if (!auth.user.authenticated) {
                router.push({name: 'home'})
            }
        },
        methods: {
            newNamespace() {
				const cancel = message.setLoading(this)

                HTTP.put(`namespaces`, this.namespace, {headers: {'Authorization': 'Bearer ' + localStorage.getItem('token')}})
                    .then(() => {
                        this.$parent.loadNamespaces()
                        this.handleSuccess({message: 'The namespace was successfully created.'})
						cancel()
                    })
                    .catch(e => {
                        this.handleError(e)
						cancel()
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

</style>