const path = require('path')
const express = require('express')
const app = express()

const p = path.join(__dirname, '..', 'dist-dev')
const port = 4173

app.use(express.static(p))
// Handle urls set by the frontend
app.get('*', (request, response, next) => {
	response.sendFile(`${p}/index.html`)
})
app.listen(port, '127.0.0.1', () => {
	console.log(`Serving files from ${p}`)
	console.log(`Server started on port ${port}`)
})
