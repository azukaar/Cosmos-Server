import Hjson from 'hjson';

// ==============================|| HJSON HELPERS ||============================== //
//
// Cosmos stores service/compose payloads as JSON. For display and editing we
// prefer HJSON (https://hjson.github.io/) — a human-friendly superset of JSON
// that allows comments, unquoted keys and relaxed strings. Since HJSON is a
// superset of JSON, every valid JSON document is already valid HJSON, so the
// backend never notices the difference: we only convert at the edges of the UI.
//
// NOTE on quotes:
//   - quotes: 'min'        -> strings that would be misread as numbers or
//                             booleans are kept quoted ("1", "true");
//                             everything else is bare where safe.
//   - separator: false     -> no trailing commas.
//   - multiline handled in postProcessMultiline below: strings containing
//     \n (but NO \r) are displayed as ''' block strings. HJSON's own
//     multiline would strip \r (corrupting CRLF data), so we only convert
//     the safe subset ourselves.

// Rewrite HJSON string values that contain real newlines (but no carriage
// returns) into HJSON multiline ''' ... ''' blocks, preserving indentation.
// Strings that contain \r are left as escaped single-line strings (HJSON
// would lose the \r when parsing ''' blocks, so we never emit them).
// indent = indent of the opening ''' token. contentIndent = indent of the
// body lines. HJSON strips the body's common indent, so contentIndent must
// make each body line exactly contentIndent wide for a clean round-trip.
const toMultilineBlock = (indent, value, contentIndent) => {
  const lines = value.split('\n');
  const body = lines.map((l) => (contentIndent || indent) + (l || '')).join('\n');
  return `${indent}'''\n${body}\n${indent}'''`;
};

// Rewrite values containing real newlines (but no carriage returns) into
// HJSON multiline ''' ... ''' blocks. Handles both:
//   key: "a\nb"            (object value)
//   "a\nb"                 (array element, on its own line)
// CRLF strings (\r) are left escaped, since HJSON strips \r in ''' blocks.
const postProcessMultiline = (hjson) => {
  let out = hjson.replace(
    /^(\s*)([^:\n]+):\s*("(?:(?:\\.)|[^"\\])*")(,?)$/gm,
    (match, indent, key, quoted, comma) => {
      let value;
      try { value = JSON.parse(quoted); } catch (e) { return match; }
      if (typeof value !== 'string' || !value.includes('\n') || value.includes('\r')) {
        return match;
      }
      const block = toMultilineBlock(indent, value, indent + '  ');
      return `${indent}${key}:${block}${comma}`;
    }
  );
  // array element: a double-quoted string alone on its line.
  out = out.replace(
    /^(\s*)("(?:(?:\\.)|[^"\\])*")(,?)$/gm,
    (match, indent, quoted, comma) => {
      let value;
      try { value = JSON.parse(quoted); } catch (e) { return match; }
      if (typeof value !== 'string' || !value.includes('\n') || value.includes('\r')) {
        return match;
      }
      return `${toMultilineBlock(indent, value, indent)}${comma}`;
    }
  );
  return out;
};

// Bare array elements containing ':' (e.g. BACKEND_HOSTNAME=https://piped...,
// or 8080:80) are emitted unquoted by quotes:'min'. Our Prism grammar would
// then misread them as key:value (the : splits) or a comment (//). This pass
// tracks [ ] scope by indentation and quotes those elements so they stay single
// green strings. Object values are left alone (the grammar colors them; a bare
// object value may also legitimately contain ':').
const quoteArrayColons = (hjson) => {
  const lines = hjson.split('\n');
  const scope = []; // { indent, kind: 'obj' | 'arr' } - current open bracket scopes
  const out = [];

  for (let i = 0; i < lines.length; i++) {
    const raw = lines[i];
    const trimmed = raw.trim();
    const indent = raw.length - raw.trimStart().length;

    // pop scopes strictly deeper than this indent (dedent)
    while (scope.length && scope[scope.length - 1].indent > indent) {
      scope.pop();
    }
    // closing brace/bracket pops the matching scope
    if (trimmed === '}' || trimmed === ']' || /^}[,}]?$/.test(trimmed) || /^][,}]?$/.test(trimmed)) {
      if (scope.length) scope.pop();
      out.push(raw);
      continue;
    }
    if (!trimmed) { out.push(raw); continue; }

    const arrScope = scope.length ? scope[scope.length - 1] : null;
    const insideArr = arrScope && arrScope.kind === 'arr' && indent > arrScope.indent;

    // Inside an array there are no object key:value lines (nested objects are
    // separate { } scopes tracked above), so any BARE element containing ':' is
    // an ambiguous string (BACKEND_HOSTNAME=..., 8080:80). Quote it so the
    // grammar colors it green instead of splitting at ':' / treating // as a
    // comment. Always valid HJSON and round-trips exactly.
    if (insideArr && !trimmed.startsWith('"') && trimmed.includes(':')) {
      const indentStr = raw.slice(0, raw.length - raw.trimStart().length);
      out.push(indentStr + JSON.stringify(trimmed));
      continue;
    }

    out.push(raw);

    // track nested scopes opened by this line's value
    const afterColon = trimmed.replace(/^.*?:\s*/, '');
    if (afterColon.startsWith('[')) {
      scope.push({ indent, kind: 'arr' });
    } else if (afterColon.startsWith('{')) {
      scope.push({ indent, kind: 'obj' });
    }
  }
  return out.join('\n');
};

// Render a JS object as pretty HJSON.
export const toHjson = (obj) => {
  try {
    return postProcessMultiline(
      quoteArrayColons(
        Hjson.stringify(obj, {
          space: 2,
          quotes: 'min',
          separator: false,
          bracesSameLine: true,
          multiline: 'off',
        })
      )
    );
  } catch (e) {
    // Never break the UI on a malformed payload — fall back to JSON.
    return JSON.stringify(obj, null, 2);
  }
};

// Parse text that may be either JSON or HJSON (HJSON accepts JSON, but being
// explicit costs nothing and keeps intent clear). Returns the parsed object,
// or throws with a normalized, user-facing message on invalid input.
export const parseJsonOrHjson = (text) => {
  const trimmed = text == null ? '' : String(text).trim();
  if (!trimmed) {
    throw new Error('Empty input');
  }

  // Fast path: plain JSON is the common case for machine-written payloads.
  let jsonErr;
  try {
    return JSON.parse(trimmed);
  } catch (err) {
    jsonErr = err;
    // Fall through to HJSON, which tolerates comments / unquoted keys.
  }

  try {
    return Hjson.parse(trimmed);
  } catch (hjsonErr) {
    // Prefer the JSON error message when the document actually looked like
    // JSON; otherwise surface the HJSON one.
    const looksLikeJson = /^[\s]*[[{]/.test(trimmed);
    const err = looksLikeJson && jsonErr ? jsonErr : hjsonErr;
    throw new Error(err && err.message ? err.message : 'Invalid JSON/HJSON input');
  }
};