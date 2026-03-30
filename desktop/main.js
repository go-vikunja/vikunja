const {app, BrowserWindow, shell, ipcMain} = require('electron')
const path = require('path')
const express = require('express')
const eApp = express()
const portInUse = require('./portInUse.js')
const oauth = require('./oauth.js')

const frontendPath = 'frontend/'
const PROTOCOL = 'vikunja-desktop'

let mainWindow = null
let pendingDeepLinkUrl = null

// Ensure single instance so deep links reach the running app on Windows/Linux
const gotTheLock = app.requestSingleInstanceLock()
if (!gotTheLock) {
	app.quit()
	// Must exit the process immediately — app.quit() is async and the rest of this
	// file would still execute, potentially opening a blank window.
	process.exit(0)
}

// Register the custom protocol for deep links
if (process.defaultApp) {
	// During development, register with the path to the script
	if (process.argv.length >= 2) {
		app.setAsDefaultProtocolClient(PROTOCOL, process.execPath, [path.resolve(process.argv[1])])
	}
} else {
	app.setAsDefaultProtocolClient(PROTOCOL)
}

// Handle deep link on macOS (app already running or launched via URL)
app.on('open-url', (event, url) => {
	event.preventDefault()
	if (mainWindow) {
		handleDeepLink(url)
	} else {
		// Window not ready yet — buffer the URL for processing after createWindow()
		pendingDeepLinkUrl = url
	}
})

// Handle deep link on Windows/Linux when a second instance is launched
app.on('second-instance', (_event, argv) => {
	// Focus the main window
	if (mainWindow) {
		if (mainWindow.isMinimized()) mainWindow.restore()
		mainWindow.focus()
	}

	// Find the deep link URL in argv
	const deepLinkUrl = argv.find(arg => arg.startsWith(`${PROTOCOL}://`))
	if (deepLinkUrl) {
		handleDeepLink(deepLinkUrl)
	}
})

function handleDeepLink(url) {
	try {
		const parsed = new URL(url)
		if (parsed.hostname === 'callback') {
			const code = parsed.searchParams.get('code')
			if (code && mainWindow) {
				// Store the apiUrl that was used to start login so we can
				// exchange the code at the correct endpoint
				const apiUrl = pendingApiUrl
				if (!apiUrl) {
					mainWindow.webContents.send('oauth:error', 'No pending login session')
					return
				}

				oauth.exchangeCodeForTokens(apiUrl, code)
					.then(tokens => {
						mainWindow.webContents.send('oauth:tokens', tokens)
					})
					.catch(err => {
						mainWindow.webContents.send('oauth:error', err.message)
					})
			}
		}
	} catch {
		// Invalid URL, ignore
	}
}

let pendingApiUrl = null

// IPC: Start OAuth login flow
ipcMain.handle('oauth:start-login', async (_event, apiUrl) => {
	pendingApiUrl = apiUrl
	const authUrl = oauth.startLogin(apiUrl)
	await shell.openExternal(authUrl)
})

// IPC: Refresh access token
ipcMain.handle('oauth:refresh-token', async (_event, apiUrl, refreshToken) => {
	return oauth.refreshAccessToken(apiUrl, refreshToken)
})

function createWindow() {
	// Create the browser window.
	mainWindow = new BrowserWindow({
		width: 1680,
		height: 960,
		webPreferences: {
			preload: path.join(__dirname, 'preload.js'),
			nodeIntegration: false,
			contextIsolation: true,
			sandbox: true,
			webviewTag: false,
			navigateOnDragDrop: false,
		}
	})

	// Open external links in the browser, but only allow protocols
	// that the TipTap editor also allows (see frontend/src/components/input/editor/TipTap.vue).
	// TipTap allows: http, https (built-in) + ftp, git, obsidian, notion, message
	// We also allow mailto since it's a standard safe protocol for email links.
	mainWindow.webContents.setWindowOpenHandler(({ url }) => {
		try {
			const parsedUrl = new URL(url);
			const allowedProtocols = [
				'http:', 'https:', 'mailto:',
				'ftp:', 'git:', 'obsidian:', 'notion:', 'message:',
			];
			if (allowedProtocols.includes(parsedUrl.protocol)) {
				shell.openExternal(url);
			}
		} catch {
			// Invalid URL, ignore silently
		}
		return { action: 'deny' };
	});

	// Prevent same-window navigation to external origins.
	// Only allow navigation to the local express server.
	mainWindow.webContents.on('will-navigate', (event, navigationUrl) => {
		const parsedUrl = new URL(navigationUrl);
		// Allow navigations to the local express server
		if (parsedUrl.hostname === '127.0.0.1' || parsedUrl.hostname === 'localhost') {
			return;
		}
		event.preventDefault();
	});

	// Hide the toolbar
	mainWindow.setMenuBarVisibility(false)

	mainWindow.on('closed', () => {
		mainWindow = null
	})

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

			// Process any deep link that arrived before the page was ready,
			// either buffered from open-url or passed via process.argv on first launch
			mainWindow.webContents.once('did-finish-load', () => {
				if (!pendingDeepLinkUrl) {
					pendingDeepLinkUrl = process.argv.find(arg => arg.startsWith(`${PROTOCOL}://`)) || null
				}
				if (pendingDeepLinkUrl) {
					handleDeepLink(pendingDeepLinkUrl)
					pendingDeepLinkUrl = null
				}
			})
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
