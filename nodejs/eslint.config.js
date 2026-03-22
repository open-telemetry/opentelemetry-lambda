'use strict';

const gtsConfig = require('gts');
const tseslint = require('typescript-eslint');
const globals = require('globals');

module.exports = [
  ...gtsConfig,

  // Global ignores (replaces .eslintignore files)
  {
    ignores: [
      '**/build/**',
      '**/node_modules/**',
      'sample-apps/aws-sdk/cdk.out/**',
      'sample-apps/aws-sdk/bin/**',
      'sample-apps/aws-sdk/lib/**',
    ],
  },

  // Node globals and rules for JS/MJS/CJS files
  {
    files: ['**/*.js', '**/*.mjs', '**/*.cjs'],
    languageOptions: {
      globals: {
        ...globals.node,
      },
    },
    rules: {
      'no-unused-vars': ['error', {varsIgnorePattern: '^_'}],
    },
  },

  // Custom rules for all .ts files
  {
    files: ['**/*.ts'],
    languageOptions: {
      parser: tseslint.parser,
      parserOptions: {
        project: true,
      },
      globals: {
        ...globals.node,
      },
    },
    rules: {
      '@typescript-eslint/no-this-alias': 'off',
      'eqeqeq': 'off',
      'prefer-rest-params': 'off',
      '@typescript-eslint/naming-convention': [
        'error',
        {
          selector: 'memberLike',
          modifiers: ['private', 'protected'],
          format: ['camelCase'],
          leadingUnderscore: 'require',
        },
      ],
      '@typescript-eslint/no-inferrable-types': [
        'error',
        {ignoreProperties: true},
      ],
      'arrow-parens': ['error', 'as-needed'],
      'prettier/prettier': [
        'error',
        {singleQuote: true, arrowParens: 'avoid'},
      ],
      '@typescript-eslint/no-require-imports': 'off',
    },
  },

  // Relaxed rules for test files
  {
    files: ['**/test/**/*.ts'],
    languageOptions: {
      globals: {
        ...globals.mocha,
      },
    },
    rules: {
      'no-empty': 'off',
      '@typescript-eslint/ban-ts-comment': 'off',
      '@typescript-eslint/no-empty-function': 'off',
      '@typescript-eslint/no-explicit-any': 'off',
      '@typescript-eslint/no-unused-vars': 'off',
      '@typescript-eslint/no-require-imports': 'off',
    },
  },

  // Relaxed rules for non-TS test files
  {
    files: ['**/test/**/*.js', '**/test/**/*.mjs', '**/test/**/*.cjs'],
    languageOptions: {
      globals: {
        ...globals.mocha,
      },
    },
    rules: {
      'no-empty': 'off',
      'no-unused-vars': 'off',
    },
  },
];
