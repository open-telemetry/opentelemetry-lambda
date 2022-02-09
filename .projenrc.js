const { CdktfProject } = require('@medlypharmacy/medly-projen');
const project = new CdktfProject({
  defaultReleaseBranch: 'main',
  devDeps: ['@medlypharmacy/medly-projen@2.3.9'],
  name: 'opentelemetry-lambda',
  enabledEnvs: [
    'dev'
  ],
  terraformProviders: [
    'aws@~> 3.53.0',
  ],
  terraformModules: [],
  gitignore: [
    'bin/*',
    'extension.zip',
    'build',
    '.gradle',
    '.idea',
    '.terraform*',
    'terraform.*',
    '.vscode/',
  ],
  preBuildSteps: [
    {
      'uses': 'actions/setup-go@v2',
      'with': {
        'go-version': '^1.17',
      },
    },
    {    
      'uses': 'actions/cache@v2',
      'with': {
        'path': '~/go/pkg/mod',
        'key': "${{ runner.os }}-go-${{ hashFiles('**/go.sum') }}",
        'restore-keys': '${{ runner.os }}-go-',
      },
    },
    {
      'name': 'Build collector',
      'working-directory': 'collector',
      'run': 'make package'
    },
    {
      'name': 'Build Node.js wrapper',
      'working-directory': 'nodejs',
      'run': 'npm install'
    },
  ]
});
project.synth();
