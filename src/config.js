import {HTTP} from './http-common'

export default {
    config: null,

    getConfig() {
        return this.config
    },

    initConfig() {
        return HTTP.get('info')
            .then(r => {
                this.config = r.data
            })
    }
}