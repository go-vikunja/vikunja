local BuildSwitchDryRun(version='go-latest', tags='latest', dry=false, depends='') = {
  name: if dry then 'dry-run-' + version else 'build-' + version,
  pull: 'always',
  image: 'plugins/docker',
  settings: {
    dockerfile: 'docker/' + version + '/Dockerfile',
    context: 'docker/' + version,
    password: {
      from_secret: 'docker_password'
    },
    username: {
      from_secret: 'docker_username'
    },
    repo: 'techknowlogick/xgo',
    tags: tags,
    dry_run: dry
  },
  [if depends != '' then 'depends_on']: [
    depends
  ],
  [if !dry then 'when']: {
    branch: ['master'],
    event: {exclude: ['pull_request']}
  },
  [if dry then 'when']: {
    event: ['pull_request']
  },
};

local BuildWithDiffTags(version='go-latest', tags='latest', depends='') = BuildSwitchDryRun(version, tags, false, depends);
local BuildWithDiffTagsDry(version='go-latest', tags='latest', depends='') = BuildSwitchDryRun(version, tags, true, depends);
local BuildStep(version='go-latest', depends='') = BuildWithDiffTags(version, version, depends);
local BuildStepDry(version='go-latest', depends='') = BuildSwitchDryRun(version, version, true, depends);

{
kind: 'pipeline',
name: 'default',
steps: [
  BuildStepDry('base'),
  BuildStepDry('go-1.14.1', 'dry-run-base'),
  BuildStepDry('go-1.13.9', 'dry-run-base'),

  BuildStep('base'),
  BuildStep('go-1.14.1', 'build-base'),
  BuildStep('go-1.14.x', 'build-go-1.14.1'),
  BuildStep('go-1.13.9', 'build-base'),
  BuildStep('go-1.13.x', 'build-go-1.13.9'),
  BuildWithDiffTags('go-latest', 'latest', 'build-go-1.14.x'),
]
}
