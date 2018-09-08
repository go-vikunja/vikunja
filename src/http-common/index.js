import axios from 'axios'
let config = require('../../public/config.json')

export const HTTP = axios.create({
  baseURL: config.VIKUNJA_API_BASE_URL
})
