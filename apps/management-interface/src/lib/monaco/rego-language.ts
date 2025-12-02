import type { languages } from 'monaco-editor';

export const regoLanguageConfig: languages.LanguageConfiguration = {
  comments: {
    lineComment: '#',
  },
  brackets: [
    ['{', '}'],
    ['[', ']'],
    ['(', ')'],
  ],
  autoClosingPairs: [
    { open: '{', close: '}' },
    { open: '[', close: ']' },
    { open: '(', close: ')' },
    { open: '"', close: '"' },
    { open: '`', close: '`' },
  ],
  surroundingPairs: [
    { open: '{', close: '}' },
    { open: '[', close: ']' },
    { open: '(', close: ')' },
    { open: '"', close: '"' },
    { open: '`', close: '`' },
  ],
};

export const regoLanguageTokens: languages.IMonarchLanguage = {
  defaultToken: '',
  tokenPostfix: '.rego',

  keywords: [
    'package',
    'import',
    'as',
    'default',
    'not',
    'with',
    'some',
    'every',
    'if',
    'else',
    'true',
    'false',
    'null',
    'contains',
    'in',
  ],

  builtinFunctions: [
    // Aggregates
    'count',
    'sum',
    'product',
    'max',
    'min',
    'sort',
    'all',
    'any',

    // Arrays
    'array.concat',
    'array.slice',

    // Sets
    'set_diff',
    'intersection',
    'union',

    // Strings
    'concat',
    'contains',
    'endswith',
    'format_int',
    'indexof',
    'lower',
    'replace',
    'split',
    'sprintf',
    'startswith',
    'substring',
    'trim',
    'upper',

    // Regex
    'regex.match',
    'regex.find_all_string_submatch_n',
    'regex.find_n',
    'regex.is_valid',
    'regex.split',

    // Types
    'is_array',
    'is_boolean',
    'is_null',
    'is_number',
    'is_object',
    'is_set',
    'is_string',
    'type_name',
  ],

  operators: [
    '=',
    '==',
    '!=',
    '<',
    '<=',
    '>',
    '>=',
    '+',
    '-',
    '*',
    '/',
    '%',
    '&',
    '|',
    ':=',
  ],

  symbols: /[=><!~?:&|+\-*\/\^%]+/,

  escapes: /\\(?:[abfnrtv\\"']|x[0-9A-Fa-f]{1,4}|u[0-9A-Fa-f]{4}|U[0-9A-Fa-f]{8})/,

  tokenizer: {
    root: [
      // Identifiers and keywords
      [
        /[a-z_$][\w$]*/,
        {
          cases: {
            '@keywords': 'keyword',
            '@builtinFunctions': 'predefined',
            '@default': 'identifier',
          },
        },
      ],
      [/[A-Z][\w\$]*/, 'type.identifier'],

      // Whitespace
      { include: '@whitespace' },

      // Delimiters and operators
      [/[{}()\[\]]/, '@brackets'],
      [/[<>](?!@symbols)/, '@brackets'],
      [
        /@symbols/,
        {
          cases: {
            '@operators': 'operator',
            '@default': '',
          },
        },
      ],

      // Numbers
      [/\d*\.\d+([eE][\-+]?\d+)?/, 'number.float'],
      [/0[xX][0-9a-fA-F]+/, 'number.hex'],
      [/\d+/, 'number'],

      // Delimiter: after number because of .\d floats
      [/[;,.]/, 'delimiter'],

      // Strings
      [/"([^"\\]|\\.)*$/, 'string.invalid'], // non-terminated string
      [/"/, { token: 'string.quote', bracket: '@open', next: '@string' }],

      // Raw strings
      [/`/, { token: 'string.quote', bracket: '@open', next: '@rawstring' }],
    ],

    comment: [
      [/[^#]+/, 'comment'],
      [/#/, 'comment'],
    ],

    string: [
      [/[^\\"]+/, 'string'],
      [/@escapes/, 'string.escape'],
      [/\\./, 'string.escape.invalid'],
      [/"/, { token: 'string.quote', bracket: '@close', next: '@pop' }],
    ],

    rawstring: [
      [/[^`]/, 'string'],
      [/`/, { token: 'string.quote', bracket: '@close', next: '@pop' }],
    ],

    whitespace: [
      [/[ \t\r\n]+/, 'white'],
      [/#.*$/, 'comment'],
    ],
  },
};
