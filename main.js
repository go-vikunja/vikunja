const {app, BrowserWindow, shell} = require('electron')
const express = require('express')
const eApp = express()

const frontendPath = 'frontend/'

function createWindow() {
	// Create the browser window.
	const mainWindow = new BrowserWindow({
		width: 800,
		height: 600,
		webPreferences: {
			nodeIntegration: true,
		}
	})

	// Open external links in the browser
	mainWindow.webContents.on('new-window', function (e, url) {
		e.preventDefault()
		shell.openExternal(url)
	})

	// Hide the toolbar
	mainWindow.setMenuBarVisibility(false)

	// Start a local express server to serve static files
	eApp.use(express.static(frontendPath))
	// Handle urls set by the frontend
	eApp.get('*', (request, response, next) => {
		response.sendFile(`${__dirname}/${frontendPath}index.html`);
	})
	const server = eApp.listen(0,  '127.0.0.1', () => {
		console.log(`Server started on port ${server.address().port}`)
		mainWindow.loadURL(`http://127.0.0.1:${server.address().port}`)
	})

	// Open the DevTools.
	// mainWindow.webContents.openDevTools()
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

