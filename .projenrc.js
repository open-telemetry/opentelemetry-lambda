const { CdktfProject } = require('@medlypharmacy/medly-projen');
const project = new CdktfProject({
  defaultReleaseBranch: 'main',
  devDeps: ['@medlypharmacy/medly-projen@2.3.3'],
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
});
project.synth();
