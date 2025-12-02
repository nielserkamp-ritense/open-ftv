import type { languages } from 'monaco-editor';

export const cedarLanguageConfig: languages.LanguageConfiguration = {
  comments: {
    lineComment: '//',
    blockComment: ['/*', '*/'],
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
  ],
  surroundingPairs: [
    { open: '{', close: '}' },
    { open: '[', close: ']' },
    { open: '(', close: ')' },
    { open: '"', close: '"' },
  ],
};

export const cedarLanguageTokens: languages.IMonarchLanguage = {
  defaultToken: '',
  tokenPostfix: '.cedar',

  keywords: [
    'permit',
    'forbid',
    'when',
    'unless',
    'if',
    'then',
    'else',
    'in',
    'has',
    'like',
    'is',
    'principal',
    'action',
    'resource',
    'context',
    'true',
    'false',
  ],

  builtinFunctions: [
    // IP address functions
    'ip',
    'isIpv4',
    'isIpv6',
    'isLoopback',
    'isMulticast',
    'isInRange',

    // Decimal functions
    'decimal',
    'lessThan',
    'lessThanOrEqual',
    'greaterThan',
    'greaterThanOrEqual',

    // String functions
    'contains',
    'containsAll',
    'containsAny',

    // Set operations
    'set',
  ],

  operators: [
    '==',
    '!=',
    '<',
    '<=',
    '>',
    '>=',
    '&&',
    '||',
    '!',
    '+',
    '-',
    '*',
    '.',
  ],

  symbols: /[=><!~?:&|+\-*\/\^%\.]+/,

  escapes: /\\(?:[abfnrtv\\"']|x[0-9A-Fa-f]{1,4}|u[0-9A-Fa-f]{4}|U[0-9A-Fa-f]{8})/,

  tokenizer: {
    root: [
      // Identifiers and keywords
      [
        /[a-zA-Z_][\w]*/,
        {
          cases: {
            '@keywords': 'keyword',
            '@builtinFunctions': 'predefined',
            '@default': 'identifier',
          },
        },
      ],

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
      [/\d+/, 'number'],

      // Delimiter
      [/[;,.]/, 'delimiter'],

      // Strings
      [/"([^"\\]|\\.)*$/, 'string.invalid'], // non-terminated string
      [/"/, { token: 'string.quote', bracket: '@open', next: '@string' }],

      // Entity UIDs (e.g., User::"alice")
      [/[A-Z][a-zA-Z0-9]*::/, 'type.identifier'],
    ],

    comment: [
      [/[^\/*]+/, 'comment'],
      [/\/\*/, 'comment', '@push'],
      [/\*\//, 'comment', '@pop'],
      [/[\/*]/, 'comment'],
    ],

    string: [
      [/[^\\"]+/, 'string'],
      [/@escapes/, 'string.escape'],
      [/\\./, 'string.escape.invalid'],
      [/"/, { token: 'string.quote', bracket: '@close', next: '@pop' }],
    ],

    whitespace: [
      [/[ \t\r\n]+/, 'white'],
      [/\/\*/, 'comment', '@comment'],
      [/\/\/.*$/, 'comment'],
    ],
  },
};
