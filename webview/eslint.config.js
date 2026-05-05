import js from '@eslint/js';
import tseslint from 'typescript-eslint';
import pluginVue from 'eslint-plugin-vue';
import globals from 'globals';

export default tseslint.config(
  // 忽略目录
  {
    ignores: ['dist/**', 'node_modules/**'],
  },

  // JS 推荐规则
  js.configs.recommended,

  // TypeScript 推荐规则
  ...tseslint.configs.recommended,

  // Vue 推荐规则（vue3-essential：组件结构/语法正确性；格式由 Prettier 负责）
  ...pluginVue.configs['flat/essential'],

  // 项目自定义规则
  {
    files: ['src/**/*.ts', 'src/**/*.vue'],
    languageOptions: {
      globals: {
        ...globals.browser,
        ...globals.es2020,
      },
      parserOptions: {
        parser: tseslint.parser,
        project: './tsconfig.json',
        extraFileExtensions: ['.vue'],
      },
    },
    rules: {
      // TS：与 tsconfig strict 保持一致，关闭重复检查
      '@typescript-eslint/no-unused-vars': 'warn',
      '@typescript-eslint/no-explicit-any': 'warn',
      '@typescript-eslint/no-non-null-assertion': 'warn',
      '@typescript-eslint/no-unused-expressions': ['error', {
        allowShortCircuit: true, allowTernary: true
      }],

      // Vue：组件命名使用 PascalCase
      'vue/component-name-in-template-casing': ['error', 'PascalCase'],
      // Vue：多词组件名（内置组件豁免）
      'vue/multi-word-component-names': ['error', {
        ignores: ['index', 'Index'],
      }],
      // Vue：禁止 v-html（XSS 风险）
      'vue/no-v-html': 'off',
      // 关闭与 Prettier 冲突的 HTML 格式规则
      'vue/html-self-closing': 'off',
      'vue/html-indent': 'off',
      'vue/html-closing-bracket-newline': 'off',
      'vue/multiline-html-element-content-newline': 'off',
      'vue/singleline-html-element-content-newline': 'off',
      'vue/max-attributes-per-line': 'off',
      'vue/attributes-order': 'off',

      // 通用
      'no-empty': 'off',
      'no-console': 'off',
      'eqeqeq': ['error', 'always'],
    },
  },
);
