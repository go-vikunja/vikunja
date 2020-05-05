import axios from 'axios'

export const HTTP = axios.create({
  baseURL: window.API_URL
})
