/* eslint-disable */
import axios from "axios";

function sleep(ms) {
    return new Promise(resolve => setTimeout(resolve, ms));
}


export default {

    initConfig() {
        //this.config = {VIKUNJA_API_BASE_URL: '/api/v1/'}

        axios.get('config.json')
            .then(function (response) {
                /*console.log('response', response.data);
                console.log('self', self.config);
                self.config = response.data*/
                // eslint-disable-next-line
                //console.log(response.data);

                localStorage.removeItem('config')
                localStorage.setItem('config', JSON.stringify(response.data))
            })
            .catch(function (error) {
                // eslint-disable-next-line
                console.log(error);
            })

        /*console.log('final', conf.data);
        return conf.data*/
    },

    async configReady() {
        while(!localStorage.getItem('config')){
            await sleep(100);
        }
        return true
    },

    get() {
        this.configReady()
        return JSON.parse(localStorage.getItem('config'))
    },

    VIKUNJA_API_BASE_URL: '/api/v1/'
}