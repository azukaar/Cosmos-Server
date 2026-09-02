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

// Render a JS object as pretty HJSON.
export const toHjson = (obj) => {
  try {
    return postProcessMultiline(
      Hjson.stringify(obj, {
        space: 2,
        quotes: 'min',
        separator: false,
        bracesSameLine: true,
        multiline: 'off',
      })
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