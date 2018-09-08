import axios from 'axios'
//let config = require('../../siteconfig.json')
//import config from '../config/dev.env'
//import Vue from 'vue'

import config from '../config'

config.initConfig()
let conf = config.get()
/*
conf.then(function (r) {
    // eslint-disable-next-line
    console.log(r)
})*/
config.configReady()
// eslint-disable-next-line
console.log(conf)

export const HTTP = axios.create({
  baseURL: conf.VIKUNJA_API_BASE_URL
})
