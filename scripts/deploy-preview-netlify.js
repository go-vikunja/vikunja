const slugify = require('slugify')
const {exec} = require('child_process')
const axios = require('axios')

const BOT_USER_ID = 513
const giteaToken = process.env.GITEA_TOKEN
const siteId = process.env.NETLIFY_SITE_ID
const branchSlug = slugify(process.env.DRONE_SOURCE_BRANCH)
const prNumber = process.env.DRONE_PULL_REQUEST

const prIssueCommentsUrl = `https://kolaente.dev/api/v1/repos/vikunja/frontend/issues/${prNumber}/comments`
const alias = `${prNumber}-${branchSlug}`
const fullPreviewUrl = `https://${alias}--vikunja-frontend-preview.netlify.app`

const promiseExec = cmd => {
	return new Promise((resolve, reject) => {
		exec(cmd, (error, stdout, stderr) => {
			if (error) {
				reject(error)
				return
			}

			resolve(stdout)
		})
	})
}

(async function () {
	let stdout = await promiseExec(`./node_modules/.bin/netlify link --id ${siteId}`)
	console.log(stdout)
	stdout = await promiseExec(`./node_modules/.bin/netlify deploy --alias ${alias}`)
	console.log(stdout)

	const {data} = await axios.get(prIssueCommentsUrl)
	const hasComment = data.some(c => c.user.id === BOT_USER_ID)

	if (hasComment) {
		console.log(`PR #${prNumber} already has a comment with a link, not sending another comment.`)
		return
	}

	await axios.post(prIssueCommentsUrl, {
		body: `
Hi ${process.env.DRONE_COMMIT_AUTHOR}!

Thank you for creating a PR!

I've deployed the changes of this PR on a preview environment under this URL: ${fullPreviewUrl}

You can use this url to view the changes live and test them out.
You will need to manually connect this to an api running somehwere. The easiest to use is https://try.vikunja.io/.

Have a nice day!

> Beep boop, I'm a bot.
`,
	}, {
		headers: {
			'Content-Type': 'application/json',
			'accept': 'application/json',
			'Authorization': `token ${giteaToken}`,
		},
	})

	console.log(`Preview comment sent successfully to PR #${prNumber}!`)
})()