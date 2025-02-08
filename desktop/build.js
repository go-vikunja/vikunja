// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-present Vikunja and contributors. All rights reserved.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public Licensee as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public Licensee for more details.
//
// You should have received a copy of the GNU Affero General Public Licensee
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

const fs = require('fs')
const path = require('path')
const https = require('https')
const {execSync} = require('child_process')
const unzipper = require('unzipper')

// Helper function to download a file
async function downloadFile(url, dest) {
	return new Promise((resolve, reject) => {
		const file = fs.createWriteStream(dest)
		https.get(url, (response) => {
			if (response.statusCode !== 200) {
				return reject(new Error(`Failed to download file: ${response.statusCode}`))
			}
			response.pipe(file)
			file.on('finish', () => {
				file.close(resolve)
			})
		}).on('error', (err) => {
			fs.unlink(dest, () => reject(err))
		})
	})
}

// Helper function to unzip a file to a directory
async function unzipFile(zipPath, destDir) {
	return fs.createReadStream(zipPath)
		.pipe(unzipper.Extract({path: destDir}))
		.promise()
}

// Helper function to replace text in a file
async function replaceTextInFile(filePath, searchValue, replaceValue) {
	const data = await fs.promises.readFile(filePath, 'utf8')
	const result = data.replace(searchValue, replaceValue)
	await fs.promises.writeFile(filePath, result, 'utf8')
}

async function renameDistFilesToUnstable(currentVersion) {
	const directory = 'dist'
	const files = await fs.promises.readdir(directory)
	for (const file of files) {
		if (file.includes(currentVersion)) {
			const newName = file.replace(currentVersion, 'unstable')
			await fs.promises.rename(
				path.join(directory, file),
				path.join(directory, newName),
			)
			console.log(`Renamed: ${file} -> ${newName}`)
		}
	}
}

// Main function to execute the script steps
async function main() {
	const args = process.argv.slice(2)
	if (args.length === 0) {
		console.error('Error: Version placeholder argument is required.')
		console.error('Usage: node build-script.js <version-placeholder> [rename-version]')
		process.exit(1)
	}

	const versionPlaceholder = args[0]
	const renameDistFiles = args[1] || false
	const frontendZipUrl = 'https://dl.vikunja.io/frontend/vikunja-frontend-unstable.zip'
	const zipFilePath = path.resolve(__dirname, 'vikunja-frontend-unstable.zip')
	const frontendDir = path.resolve(__dirname, 'frontend')
	const indexFilePath = path.join(frontendDir, 'index.html')
	const packageJsonPath = path.join(__dirname, 'package.json')

	console.log(`Building version ${versionPlaceholder}`)

	try {
		console.log('Step 1: Downloading frontend zip...')
		await downloadFile(frontendZipUrl, zipFilePath)

		console.log('Step 2: Unzipping frontend package...')
		await unzipFile(zipFilePath, frontendDir)

		console.log('Step 3: Modifying index.html...')
		await replaceTextInFile(indexFilePath, /\/api\/v1/g, '')

		console.log('Step 4: Updating version in package.json...')
		await replaceTextInFile(packageJsonPath, /\${version}/g, versionPlaceholder)
		await replaceTextInFile(
			packageJsonPath,
			/"version": ".*"/,
			`"version": "${versionPlaceholder}"`,
		)

		console.log('Step 5: Installing dependencies and building...')
		execSync('pnpm dist', {stdio: 'inherit'})

		if (renameDistFiles) {
			console.log('Step 6: Renaming release files...')
			await renameDistFilesToUnstable(versionPlaceholder)
		}

		console.log('All steps completed successfully!')
	} catch (err) {
		console.error('An error occurred:', err.message)
	} finally {
		// Cleanup the zip file
		if (fs.existsSync(zipFilePath)) {
			fs.unlinkSync(zipFilePath)
		}
	}
}

main()
