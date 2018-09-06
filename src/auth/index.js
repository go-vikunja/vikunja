import {HTTP} from '../http-common'
import router from '../router'
// const API_URL = 'http://localhost:8082/api/v1/'
// const LOGIN_URL = 'http://localhost:8082/login'

export default {

  user: {
    authenticated: false,
    infos: {}
  },

  login (context, creds, redirect) {
    HTTP.post('login', {
      username: creds.username,
      password: creds.password
    })
      .then(response => {
        // Save the token to local storage for later use
        localStorage.removeItem('token') // Delete an eventually preexisting old token
        localStorage.setItem('token', response.data.token)

        // Tell others the user is autheticated
        this.user.authenticated = true
        this.getUserInfos()

        // Hide the loader
        context.loading = false

        // Redirect if nessecary
        if (redirect) {
          router.push({ name: redirect })
        }
      })
      .catch(e => {
        // Hide the loader
        context.loading = false
        if (e.response) {
          context.error = e.response.data.message
          if (e.response.status === 401) {
            context.error = 'Wrong username or password.'
          }
        }
      })
  },

  logout () {
    localStorage.removeItem('token')
    router.push({ name: 'login' })
    this.user.authenticated = false
  },

  checkAuth () {
    let jwt = localStorage.getItem('token')
    this.getUserInfos()
    this.user.authenticated = false
    if (jwt) {
      let infos = this.user.infos
      let ts = Math.round((new Date()).getTime() / 1000)
      if (infos.exp >= ts) {
        this.user.authenticated = true
      }
    }
  },

  getUserInfos () {
    let jwt = localStorage.getItem('token')
    if (jwt) {
      this.user.infos = this.parseJwt(localStorage.getItem('token'))
      return this.parseJwt(localStorage.getItem('token'))
    } else {
      return {}
    }
  },

  parseJwt (token) {
    let base64Url = token.split('.')[1]
    let base64 = base64Url.replace('-', '+').replace('_', '/')
    return JSON.parse(window.atob(base64))
  },

  getAuthHeader () {
    return {
      'Authorization': 'Bearer ' + localStorage.getItem('token')
    }
  },

  getToken () {
    return localStorage.getItem('token')
  }
}
