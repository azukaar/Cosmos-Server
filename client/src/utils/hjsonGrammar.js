import Prism from 'prismjs/components/prism-core';

// ==============================|| HJSON GRAMMAR ||============================== //
//
// Cosmos displays compose/JSON payloads as HJSON (see utils/hjson.jsx) with
// quotes: 'always' (all strings quoted) and separator: true (trailing commas).
// Prism's stock `json` grammar only colors *quoted* keys, so HJSON's unquoted
// keys (`services:`) would render white. This grammar extends JSON with
// HJSON's unquoted keys and keeps the familiar color scheme:
//   - keys (quoted or unquoted)            -> property  (red/Okaidia)
//   - string values                        -> string    (green)
//   - numbers / booleans / null            -> number/boolean/null
//   - comments (# // /* */)                 -> comment   (grey)
//   - multiline strings (''' ... ''')       -> string    (green)
//
// Register it as `hjson` and override the `json` grammar so the compose
// editor (which resolves the 'json' language) picks it up transparently.

const hjsonProperty = {
  // Quoted key - 'key': value, handles escaped quotes inside.
  pattern: /"(?:\\.|[^"\\])*"(?=\s*:)/,
  greedy: true,
};

const hjsonUnquotedProperty = {
  // Unquoted HJSON key (services:, my-app:): starts after a line start,
  // brace/bracket or comma (plus indentation), extends to the ':'.
  // '/' is excluded as the *first* char so comment lines (//, /*) can never
  // be misread as keys; '/' inside a key (my/app:) is still fine.
  pattern: /(?<=^|[\r\n{[,])\s*[^:#\[\]{}"',\r\n\s/][^:#\[\]{}"',\r\n]*(?=\s*:)/,
  greedy: true,
  alias: 'property',
};

const hjsonString = {
  // Quoted string value - handles escaped quotes (\" or \\), plus any of
  // #, //, : inside. The (?!\s*:) excludes quoted keys (handled above).
  pattern: /"(?:\\.|[^"\\])*"(?!\s*:)/,
  greedy: true,
};

const hjsonMlString = {
  // HJSON multiline string ''' ... ''' (may span lines).
  pattern: /'''[\s\S]*?'''/,
  greedy: true,
  alias: 'string',
};

const hjsonGrammar = {
  'unquoted-property': hjsonUnquotedProperty,
  property: hjsonProperty,
  'ml-string': hjsonMlString,
  string: hjsonString,

  comment: [
    // /* ... */ block comments (multiline). Tried first so they span lines.
    {
      pattern: /\/\*[\s\S]*?(?:\*\/|$)/,
      greedy: true,
    },
    // # and // line comments (to end of line). No /m or $: with Prism's
    // greedy engine, /m + $ made # comments not match.
    {
      pattern: /(?:\/\/.*|#[^\r\n]*)/,
      greedy: true,
    },
  ],

  number: /-?\b\d+(?:\.\d+)?(?:e[+-]?\d+)?\b/i,

  punctuation: /[{}[\],]/,

  operator: /:/,

  boolean: /\b(?:false|true)\b/,

  null: {
    pattern: /\bnull\b/,
    alias: 'keyword',
  },
};

Prism.languages.hjson = hjsonGrammar;

// Override the stock json language so existing `.json` highlight callers
// (e.g. the compose editor) also render HJSON colors.
Prism.languages.json = hjsonGrammar;

// Register hjson as an alias of json for any contextual language (like
// react-simple-code-editor resolves languages by name). Safe to call
// multiple times.
export const registerHjson = (name = 'json') => {
  Prism.languages[name] = hjsonGrammar;
};

export default hjsonGrammar;
