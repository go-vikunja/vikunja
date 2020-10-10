const {app, BrowserWindow, protocol, shell} = require('electron')
const path = require('path')

function createWindow() {
	// Create the browser window.
	const win = new BrowserWindow({
		width: 800,
		height: 600,
		webPreferences: {
			nodeIntegration: true,
		},
	})

	// Remove external links in the browser
	win.webContents.on('new-window', function (e, url) {
		e.preventDefault()
		shell.openExternal(url)
	})

	// Hide the toolbar
	win.setMenuBarVisibility(false)

	// The starting point of the app
	win.loadFile('./frontend/index.html')
}

// This method will be called when Electron has finished
// initialization and is ready to create browser windows.
// Some APIs can only be used after this event occurs.
app.whenReady().then(() => {
	const root = path.normalize(`${__dirname}/..`)

	// file:// interceptor to serve all assets without running a web server in the background
	protocol.interceptFileProtocol('file', (request, callback) => {
		let url = request.url.substr(7)    /* all urls start with 'file://' */
		if (!url.startsWith(root)) {
			if (url.startsWith('css/fonts') || url.startsWith('css/img')) {
				url = url.substr(4)
			}
			url = path.normalize(`${root}/frontend/${url}`)
		}
		callback({path: url})
	})

	createWindow()
})
// Quit when all windows are closed, except on macOS. There, it's common
// for applications and their menu bar to stay active until the user quits
// explicitly with Cmd + Q.
app.on('window-all-closed', () => {
	if (process.platform !== 'darwin') {
		app.quit()
	}
})

app.on('activate', () => {
	// On macOS it's common to re-create a window in the app when the
	// dock icon is clicked and there are no other windows open.
	if (BrowserWindow.getAllWindows().length === 0) {
		createWindow()
	}
})

