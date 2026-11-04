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
// -----------------------------------------------------------------------------
// Comment preservation (cosmos.compose.<node> labels)
//
// HJSON comments are not representable in the JSON payload the backend
// stores, so we extract a node -> comment map client-side before saving and
// restore it into the editor on load. Keys mirror the JSON key path
// (e.g. "services.web.image"), matching the cosmos.compose.<path> labels the
// backend persists.
//
// The stored comment text is the FULL, verbatim comment block: original
// leading whitespace, the comment marker (//, # or /* */) and any suffix
// (*/), and every line of a multi-line comment — stored as a single label
// value. Restoration emits those exact lines again, only re-anchoring the
// indentation to the node's current depth (so the comment aligns even after
// the document is re-rendered by toHjson).
// -----------------------------------------------------------------------------

// Extract a map of key-path -> verbatim comment block for every node that has
// a comment immediately above it in the HJSON document. A comment block may be
// one or more consecutive // or # lines, or a /* ... */ block (single or
// multi-line). The whole block is captured as a single label value with its
// original indentation and markers preserved.
export const extractComments = (text) => {
  if (!text || typeof text !== 'string') return {};
  const comments = {};
  const lines = text.split('\n');
  // scope entries: { indent, key, kind } where kind is 'obj' or 'arr'.
  // Arrays keep an implicit element index; awaiting order is preserved by the
  // document order (we count non-comment lines inside the array).
  const scope = [];
  // Pending raw comment lines (verbatim, including indent + markers) before
  // the next key or array element line. Tracks an open block comment.
  let pending = [];
  let inBlockComment = false;
  let inMultiLineString = false;

  const isCommentLine = (line) => {
    const t = line.trim();
    return t.startsWith('//') || t.startsWith('#') || t.startsWith('/*');
  };
  const isBlockEnd = (line) => {
    const t = line.trim();
    return t.indexOf('*/') !== -1;
  };
  // pathFor builds the dotted key path for the current scope + an optional
  // extra key/index segment.
  const pathFor = (extra) => {
    const parts = [];
    for (const s of scope) {
      parts.push(s.key);
      if (s.kind === 'arr' && s.index !== undefined) parts.push(String(s.index));
    }
    if (extra !== undefined && extra !== null && extra !== '') parts.push(String(extra));
    return parts.join('.');
  };
  const flushPendingTo = (path) => {
    if (pending.length && !comments[path]) {
      comments[path] = pending.join('\n');
    }
    pending = [];
  };

  for (let i = 0; i < lines.length; i++) {
    const raw = lines[i];
    const trimmed = raw.trim();

    // HJSON ''' multiline strings are a single value spanned over multiple
    // lines; nothing inside them is a comment, an array element or a scope
    // change — skip until the closing '''.
    if (inMultiLineString) {
      if (trimmed.includes("'''")) inMultiLineString = false;
      pending = [];
      continue;
    }

    // Collect the raw comment block (verbatim lines).
    if (inBlockComment || (isCommentLine(raw) && !inBlockComment)) {
      pending.push(raw);
      if (!inBlockComment && trimmed.startsWith('/*')) {
        inBlockComment = true;
        if (isBlockEnd(raw)) inBlockComment = false;
      } else if (inBlockComment && isBlockEnd(raw)) {
        inBlockComment = false;
      }
      continue;
    }

    if (!trimmed) { pending = []; inBlockComment = false; continue; }

    const indent = raw.length - raw.trimStart().length;
    // Dedent: pop scopes strictly deeper than current indent.
    while (scope.length && scope[scope.length - 1].indent > indent) {
      scope.pop();
    }
    // Closing brace/bracket: pop that scope, cancel pending.
    if (trimmed === '}' || trimmed === ']' || /^}[,}]?$/.test(trimmed) || /^][,}]?$/.test(trimmed)) {
      if (scope.length) scope.pop();
      pending = [];
      continue;
    }
    // Account for the implicit index of any open array whose line is at this
    // indentation (i.e. we have an array element on this line).
    const arrScope = scope.length ? scope[scope.length - 1] : null;
    const insideArr = arrScope && arrScope.kind === 'arr' && indent > arrScope.indent;

    const keyMatch = /^("(?:[^"\\]|\\.)*"|[^:#\[\]{}"',\r\n\s/][^:#\[\]{}"',\r\n]*)?\s*:/.exec(trimmed);
    if (keyMatch) {
      let key = keyMatch[1].trim();
      if (key.startsWith('"') && key.endsWith('"')) {
        try { key = JSON.parse(key); } catch (e) { /* keep raw */ }
      }
      const path = pathFor(key);
      flushPendingTo(path);
      const afterColon = trimmed.slice(keyMatch[0].length).trim();
      const opensObj = afterColon.startsWith('{');
      const opensArr = afterColon.startsWith('[');
      if (opensObj || opensArr) {
        scope.push({ indent, key, kind: opensArr ? 'arr' : 'obj' });
      }
      // HJSON multiline string value: '''... possibly spanning further lines.
      if (afterColon.startsWith("'''") && !afterColon.includes("'''", 3)) {
        inMultiLineString = true;
      }
      continue;
    }

    // Array element line: attach pending comment at <arrPath>.<index>.
    if (insideArr && arrScope) {
      const idx = arrScope.index = (arrScope.index === undefined ? 0 : arrScope.index + 1);
      flushPendingTo(pathFor(null));
      // A bare ''' array element is a multiline string that continues below.
      if (trimmed.startsWith("'''") && !trimmed.includes("'''", 3)) {
        inMultiLineString = true;
      }
      continue;
    }

    // A top-level bare multiline string value (e.g. within an object value).
    if (trimmed.startsWith("'''") && !trimmed.includes("'''", 3)) {
      inMultiLineString = true;
      pending = [];
      continue;
    }

    pending = [];
  }

  return comments;
};

// Given an HJSON document and a node -> comment map (verbatim comment blocks),
// return a new document with each comment re-inserted on lines immediately
// before its node. Stored indentation is re-anchored to the node's current
// depth; the comment markers and any block-comment suffixes are preserved.
export const injectComments = (text, comments) => {
  if (!comments || !Object.keys(comments).length) return text;
  if (!text || typeof text !== 'string') return text;

  const lines = text.split('\n');
  // scope entries: { indent, key, kind } (kind 'obj' | 'arr')
  const scope = [];
  const out = [];
  const inserted = {};
  let inMultiLineString = false;

  const pathFor = (extra) => {
    const parts = [];
    for (const s of scope) {
      parts.push(s.key);
      if (s.kind === 'arr' && s.index !== undefined) parts.push(String(s.index));
    }
    if (extra !== undefined && extra !== null && extra !== '') parts.push(String(extra));
    return parts.join('.');
  };
  const emitComment = (path) => {
    if (comments[path] && !inserted[path]) {
      inserted[path] = true;
      const nodeIndent = '  '.repeat(scope.length + 1);
      const block = String(comments[path]);
      const blockLines = block.split('\n');
      const baseIndent = (blockLines[0].match(/^\s*/) || [''])[0].length;
      for (const cl of blockLines) {
        const lineIndent = (cl.match(/^\s*/) || [''])[0].length;
        const rel = Math.max(0, lineIndent - baseIndent);
        const content = cl.replace(/^\s+/, '');
        out.push(content ? nodeIndent + ' '.repeat(rel) + content : nodeIndent);
      }
    }
  };

  for (let i = 0; i < lines.length; i++) {
    const raw = lines[i];
    const trimmed = raw.trim();

    // Inside a ''' multiline string: single value, skip (no comments there).
    if (inMultiLineString) {
      if (trimmed.includes("'''")) inMultiLineString = false;
      out.push(raw);
      continue;
    }

    const indent = raw.length - raw.trimStart().length;

    while (scope.length && scope[scope.length - 1].indent > indent) {
      scope.pop();
    }
    if (trimmed === '}' || trimmed === ']' || /^}[,}]?$/.test(trimmed) || /^][,}]?$/.test(trimmed)) {
      if (scope.length) scope.pop();
      out.push(raw);
      continue;
    }

    const arrScope = scope.length ? scope[scope.length - 1] : null;
    const insideArr = arrScope && arrScope.kind === 'arr' && indent > arrScope.indent;

    const keyMatch = /^("(?:[^"\\]|\\.)*"|[^:#\[\]{}"',\r\n\s/][^:#\[\]{}"',\r\n]*)?\s*:/.exec(trimmed);
    if (keyMatch) {
      let key = keyMatch[1].trim();
      if (key.startsWith('"') && key.endsWith('"')) {
        try { key = JSON.parse(key); } catch (e) { /* keep */ }
      }
      const path = pathFor(key);
      emitComment(path);
      out.push(raw);
      const afterColon = trimmed.slice(keyMatch[0].length).trim();
      const opensArr = afterColon.startsWith('[');
      const opensObj = afterColon.startsWith('{');
      if (opensObj || opensArr) {
        scope.push({ indent, key, kind: opensArr ? 'arr' : 'obj' });
      }
      if (afterColon.startsWith("'''") && !afterColon.includes("'''", 3)) {
        inMultiLineString = true;
      }
      continue;
    }

    // Array element: emit comment at <arrPath>.<index> before the element.
    if (insideArr && arrScope) {
      const idx = arrScope.index = (arrScope.index === undefined ? 0 : arrScope.index + 1);
      emitComment(pathFor(null));
      out.push(raw);
      if (trimmed.startsWith("'''") && !trimmed.includes("'''", 3)) {
        inMultiLineString = true;
      }
      continue;
    }

    // Bare ''' multiline string value.
    if (trimmed.startsWith("'''") && !trimmed.includes("'''", 3)) {
      inMultiLineString = true;
      out.push(raw);
      continue;
    }

    out.push(raw);
  }

  return out.join('\n');
};
