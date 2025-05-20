const {app, BrowserWindow, shell} = require('electron')
const path = require('path')
const express = require('express')
const eApp = express()
const portInUse = require('./portInUse.js')

const frontendPath = 'frontend/'

function createWindow() {
	// Create the browser window.
	const mainWindow = new BrowserWindow({
		width: 1680,
		height: 960,
		webPreferences: {
			nodeIntegration: true,
		}
	})

	// Open external links in the browser
	mainWindow.webContents.setWindowOpenHandler(({ url }) => {
  	shell.openExternal(url);
	  return { action: 'deny' };
	});

	// Hide the toolbar
	mainWindow.setMenuBarVisibility(false)

	// We try to use the same port every time and only use a different one if that does not succeed.
	let port = 45735
	portInUse(port, used => {
		if(used) {
			console.log(`Port ${port} already used, switching to a random one`)
			port = 0 // This lets express choose a random port
		}

		// Start a local express server to serve static files
		eApp.use(express.static(path.join(__dirname, frontendPath)))
		// Handle urls set by the frontend - use app.use as catch-all instead of route pattern
		eApp.use((request, response) => {
			response.sendFile(path.join(__dirname, frontendPath, 'index.html'))
		})
		const server = eApp.listen(port,  '127.0.0.1', () => {
			console.log(`Server started on port ${server.address().port}`)
			mainWindow.loadURL(`http://127.0.0.1:${server.address().port}`)
		})
	})
}

// This method will be called when Electron has finished
// initialization and is ready to create browser windows.
// Some APIs can only be used after this event occurs.
app.whenReady().then(() => {
	createWindow()

	app.on('activate', function () {
		// On macOS it's common to re-create a window in the app when the
		// dock icon is clicked and there are no other windows open.
		if (BrowserWindow.getAllWindows().length === 0) createWindow()
	})
})

// Quit when all windows are closed, except on macOS. There, it's common
// for applications and their menu bar to stay active until the user quits
// explicitly with Cmd + Q.
app.on('window-all-closed', () => {
	if (process.platform !== 'darwin') app.quit()
})

