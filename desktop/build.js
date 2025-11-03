// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-present Vikunja and contributors. All rights reserved.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

const fs = require('fs')
const path = require('path')
const {execSync} = require('child_process')

// Helper function to copy directory recursively
async function copyDir(src, dest) {
	// Create destination directory if it doesn't exist
	if (!fs.existsSync(dest)) {
		await fs.promises.mkdir(dest, { recursive: true })
	}

	// Get all files in source directory
	const entries = await fs.promises.readdir(src, { withFileTypes: true })

	for (const entry of entries) {
		const srcPath = path.join(src, entry.name)
		const destPath = path.join(dest, entry.name)

		if (entry.isDirectory()) {
			// Recursively copy subdirectories
			await copyDir(srcPath, destPath)
		} else {
			// Copy files
			await fs.promises.copyFile(srcPath, destPath)
		}
	}
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
	const renameDistFiles = args[1] === 'true' || false
	const frontendSourceDir = path.resolve(__dirname, '../frontend/dist')
	const frontendDir = path.resolve(__dirname, 'frontend')
	const indexFilePath = path.join(frontendDir, 'index.html')
	const packageJsonPath = path.join(__dirname, 'package.json')

	console.log(`Building version ${versionPlaceholder}`)

	try {
		console.log('Step 1: Copying frontend files...')
		if (fs.existsSync(frontendDir)) {
			console.log('Removing existing frontend directory...')
			await fs.promises.rm(frontendDir, { recursive: true, force: true })
		}
		await fs.promises.mkdir(frontendDir, { recursive: true })
		
		await copyDir(frontendSourceDir, frontendDir)

		console.log('Step 2: Modifying index.html...')
		await replaceTextInFile(indexFilePath, /\/api\/v1/g, '')

		console.log('Step 3: Updating version in package.json...')
		await replaceTextInFile(packageJsonPath, /\${version}/g, versionPlaceholder)
		await replaceTextInFile(
			packageJsonPath,
			/"version": ".*"/,
			`"version": "${versionPlaceholder}"`,
		)

		console.log('Step 4: Installing dependencies and building...')
		execSync('pnpm dist', {stdio: 'inherit'})

		if (renameDistFiles) {
			console.log('Step 5: Renaming release files...')
			await renameDistFilesToUnstable(versionPlaceholder)
		}

		console.log('All steps completed successfully!')
	} catch (err) {
		console.error('An error occurred:', err.message)
		process.exit(1)
	}
}

main()
