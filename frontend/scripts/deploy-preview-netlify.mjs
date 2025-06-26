import { exec } from 'node:child_process'

function createSlug(string) {
	return String(string)
		.trim()
		.normalize('NFKD')
		.toLowerCase()
		.replace(/[.\s/]/g, '-')
		.replace(/[^A-Za-z\d-]/g, '')
}

const BOT_USER_ID = 513
const giteaToken = process.env.GITEA_TOKEN
const siteId = process.env.NETLIFY_SITE_ID
const branchSlug = createSlug(process.env.DRONE_SOURCE_BRANCH)
const prNumber = process.env.DRONE_PULL_REQUEST

const prIssueCommentsUrl = `https://kolaente.dev/api/v1/repos/vikunja/vikunja/issues/${prNumber}/comments`
const alias = `${prNumber}-${branchSlug}`.substring(0,37)
const fullPreviewUrl = `https://${alias}--vikunja-frontend-preview.netlify.app`

const promiseExec = cmd => {
	return new Promise((resolve, reject) => {
		exec(cmd, (error, stdout) => {
			if (error) {
				reject(error)
				return
			}

			resolve(stdout)
		})
	})
}

(async function () {
	let stdout = await promiseExec(`/home/node/docker-netlify-cli/node_modules/.bin/netlify link --id ${siteId}`)
	console.log(stdout)
	stdout = await promiseExec(`/home/node/docker-netlify-cli/node_modules/.bin/netlify deploy --alias ${alias}`)
	console.log(stdout)

	const data = await fetch(prIssueCommentsUrl).then(response => response.json())
	const hasComment = data.some(c => c.user.id === BOT_USER_ID)

	if (hasComment) {
		console.log(`PR #${prNumber} already has a comment with a link, not sending another comment.`)
		return
	}

	const message = `
Hi ${process.env.DRONE_COMMIT_AUTHOR}!

Thank you for creating a PR!

I've deployed the frontend changes of this PR on a preview environment under this URL: ${fullPreviewUrl}

You can use this url to view the changes live and test them out.
You will need to manually connect this to an api running somewhere. The easiest to use is https://try.vikunja.io/.

This preview does not contain any changes made to the api, only the frontend.

Have a nice day!

> Beep boop, I'm a bot.
`

	try {
		const response = await fetch(prIssueCommentsUrl, {
			method: 'POST',
			body: JSON.stringify({
				body: message,
			}),
			headers: {
				'Content-Type': 'application/json',
				'accept': 'application/json',
				'Authorization': `token ${giteaToken}`,
			},
		})
		if (!response.ok) {
			throw new Error(`HTTP error, status = ${response.status}`)
		}
		console.log(`Preview comment sent successfully to PR #${prNumber}!`)
	} catch (e) {
		console.log(`Could not send preview comment to PR #${prNumber}! ${e.message}`)
	}
})()
