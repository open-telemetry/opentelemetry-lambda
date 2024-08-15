module.exports = {
  plugins: [
    "@typescript-eslint",
    "header"
  ],
  extends: [
      "../../node_modules/gts",
  ],
  parser: "@typescript-eslint/parser",
  parserOptions: {
      "project": "./tsconfig.json"
  },
  rules: {
    "@typescript-eslint/no-this-alias": "off",
    "eqeqeq": "off",
    "prefer-rest-params": "off",
    "@typescript-eslint/naming-convention": [
        "error",
        {
          "selector": "memberLike",
          "modifiers": ["private", "protected"],
          "format": ["camelCase"],
          "leadingUnderscore": "require"
        }
    ],
    "@typescript-eslint/no-inferrable-types": ["error", { ignoreProperties: true }],
    "arrow-parens": ["error", "as-needed"],
    "prettier/prettier": ["error", { "singleQuote": true, "arrowParens": "avoid" }],
    "@typescript-eslint/no-require-imports": "off",
  },
  overrides: [
    {
      "files": ["test/**/*.ts"],
      "rules": {
        "no-empty": "off",
        "@typescript-eslint/ban-ts-ignore": "off",
        "@typescript-eslint/no-empty-function": "off",
        "@typescript-eslint/no-explicit-any": "off",
        "@typescript-eslint/no-unused-vars": "off",
        "@typescript-eslint/no-var-requires": "off"
      }
    }
  ]
};
