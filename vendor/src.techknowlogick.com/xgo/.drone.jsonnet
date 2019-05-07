local BuildSwitchDryRun(version='go-latest', tags='latest', dry=false) = {
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
  [if !dry then 'when']: {
    branch: ['master'],
    event: {exclude: ['pull_request']}
  },
  [if dry then 'when']: {
    event: {include: ['pull_request']}
  },
};

local BuildWithDiffTags(version='go-latest', tags='latest') = BuildSwitchDryRun(version, tags, false);
local BuildWithDiffTagsDry(version='go-latest', tags='latest') = BuildSwitchDryRun(version, tags, true);
local BuildStep(version='go-latest') = BuildWithDiffTags(version, version);
local BuildStepDry(version='go-latest') = BuildSwitchDryRun(version, version, true);

{
kind: 'pipeline',
name: 'default',
steps: [
  BuildStepDry('base'),
  BuildStepDry('go-1.12.5'),
  BuildStepDry('go-1.11.10'),

  BuildStep('base'),
  BuildStep('go-1.12.5'),
  BuildStep('go-1.12.x'),
  BuildWithDiffTags(),
  BuildStep('go-1.11.10'),
  BuildStep('go-1.11.x'),
  BuildStep('go-1.12.4'),
  BuildStep('go-1.12.3'),
  BuildStep('go-1.12.2'),
  BuildStep('go-1.12.1'),
  BuildStep('go-1.12.0'),
  BuildStep('go-1.11.9'),
]
}
