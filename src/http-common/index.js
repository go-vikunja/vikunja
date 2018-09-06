import axios from 'axios'
let config = require('../../siteconfig.json')

export const HTTP = axios.create({
  baseURL: config.API_URL
})
