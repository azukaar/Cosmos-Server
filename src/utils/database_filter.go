package utils

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

// eventFilterColumns maps the field names users type in the events query box to columns.
var eventFilterColumns = map[string]string{
	"id":          "id",
	"_id":         "id",
	"node":        "node",
	"eventId":     "event_id",
	"label":       "label",
	"application": "application",
	"level":       "level",
	"date":        "date",
	"object":      "object",
	"_search":     "search",
}

var jsonPathSegment = regexp.MustCompile(`^[A-Za-z0-9_.\-]+$`)

// column kinds decide how a bound value is coerced.
const (
	kindText = "text"
	kindInt  = "int"
	kindDate = "date"
	kindJSON = "json"
)

type filterCtx struct {
	dialect string
	args    []interface{}
}

// TranslateEventFilter turns the Mongo-ish JSON filter the UI sends into a parameterized
// WHERE fragment. Values are never interpolated; unsupported operators are an error so
// the caller can answer 400 instead of silently dropping the filter.
func TranslateEventFilter(dialect string, filter map[string]interface{}) (string, []interface{}, error) {
	c := &filterCtx{dialect: dialect}
	sqlStr, err := c.conjunction(filter)
	if err != nil {
		return "", nil, err
	}
	return sqlStr, c.args, nil
}

func (c *filterCtx) conjunction(filter map[string]interface{}) (string, error) {
	parts := []string{}
	for _, key := range sortedKeys(filter) {
		value := filter[key]

		var (
			frag string
			err  error
		)
		switch key {
		case "$and":
			frag, err = c.group(value, " AND ")
		case "$or":
			frag, err = c.group(value, " OR ")
		default:
			if strings.HasPrefix(key, "$") {
				return "", fmt.Errorf("unsupported operator %q", key)
			}
			frag, err = c.field(key, value)
		}
		if err != nil {
			return "", err
		}
		parts = append(parts, frag)
	}

	if len(parts) == 0 {
		return "1=1", nil
	}
	return "(" + strings.Join(parts, " AND ") + ")", nil
}

func (c *filterCtx) group(value interface{}, join string) (string, error) {
	list, ok := value.([]interface{})
	if !ok || len(list) == 0 {
		return "", errors.New("$and/$or expects a non-empty array")
	}
	parts := []string{}
	for _, item := range list {
		m, ok := item.(map[string]interface{})
		if !ok {
			return "", errors.New("$and/$or expects objects")
		}
		frag, err := c.conjunction(m)
		if err != nil {
			return "", err
		}
		parts = append(parts, frag)
	}
	return "(" + strings.Join(parts, join) + ")", nil
}

func (c *filterCtx) field(path string, value interface{}) (string, error) {
	kind, err := fieldKind(path)
	if err != nil {
		return "", err
	}

	ops, isOps := operatorMap(value)
	if !isOps {
		expr, err := c.expr(path, kind)
		if err != nil {
			return "", err
		}
		bound, err := c.bind(kind, value)
		if err != nil {
			return "", err
		}
		return expr + " = " + bound, nil
	}

	parts := []string{}
	for _, op := range sortedKeys(ops) {
		frag, err := c.operator(path, kind, op, ops[op])
		if err != nil {
			return "", err
		}
		parts = append(parts, frag)
	}
	return "(" + strings.Join(parts, " AND ") + ")", nil
}

func (c *filterCtx) operator(path string, kind string, op string, value interface{}) (string, error) {
	expr, err := c.expr(path, kind)
	if err != nil {
		return "", err
	}

	switch op {
	case "$eq", "$ne", "$gt", "$gte", "$lt", "$lte":
		bound, err := c.bind(kind, value)
		if err != nil {
			return "", err
		}
		return expr + " " + comparisonOperators[op] + " " + bound, nil

	case "$in", "$nin":
		list, ok := value.([]interface{})
		if !ok || len(list) == 0 {
			return "", fmt.Errorf("%s expects a non-empty array", op)
		}
		bounds := []string{}
		for _, item := range list {
			bound, err := c.bind(kind, item)
			if err != nil {
				return "", err
			}
			bounds = append(bounds, bound)
		}
		negate := ""
		if op == "$nin" {
			negate = " NOT"
		}
		return expr + negate + " IN (" + strings.Join(bounds, ", ") + ")", nil

	case "$exists":
		want, ok := value.(bool)
		if !ok {
			return "", errors.New("$exists expects a boolean")
		}
		if want {
			return expr + " IS NOT NULL", nil
		}
		return expr + " IS NULL", nil

	case "$regex":
		pattern, ok := value.(string)
		if !ok {
			return "", errors.New("$regex expects a string")
		}
		like, err := regexToLike(pattern)
		if err != nil {
			return "", err
		}
		c.args = append(c.args, like)
		return expr + ` LIKE ? ESCAPE '\'`, nil
	}

	return "", fmt.Errorf("unsupported operator %q", op)
}

var comparisonOperators = map[string]string{
	"$eq": "=", "$ne": "<>", "$gt": ">", "$gte": ">=", "$lt": "<", "$lte": "<=",
}

// expr renders the column (or JSON path) reference, binding the path itself as a
// parameter so nothing user-supplied ever reaches the SQL text.
func (c *filterCtx) expr(path string, kind string) (string, error) {
	if kind != kindJSON {
		return eventFilterColumns[path], nil
	}

	segments := strings.TrimPrefix(path, "data.")
	if segments == "" || !jsonPathSegment.MatchString(segments) {
		return "", fmt.Errorf("unsupported field %q", path)
	}

	if c.dialect == DialectPostgres {
		c.args = append(c.args, "{"+strings.ReplaceAll(segments, ".", ",")+"}")
		return "(data::jsonb #>> ?::text[])", nil
	}
	c.args = append(c.args, "$."+segments)
	return "CAST(json_extract(data, ?) AS TEXT)", nil
}

func fieldKind(path string) (string, error) {
	if col, ok := eventFilterColumns[path]; ok {
		switch col {
		case "date":
			return kindDate, nil
		case "id":
			return kindInt, nil
		default:
			return kindText, nil
		}
	}
	if strings.HasPrefix(path, "data.") {
		return kindJSON, nil
	}
	return "", fmt.Errorf("unsupported field %q", path)
}

func (c *filterCtx) bind(kind string, value interface{}) (string, error) {
	v, err := coerce(kind, value)
	if err != nil {
		return "", err
	}
	c.args = append(c.args, v)
	return "?", nil
}

func coerce(kind string, value interface{}) (interface{}, error) {
	// {"$date": ...} wrappers survive from the extended-JSON the UI used to send
	if m, ok := value.(map[string]interface{}); ok {
		if raw, ok := m["$date"]; ok && len(m) == 1 {
			return coerce(kindDate, raw)
		}
		return nil, errors.New("unsupported nested object as a value")
	}

	switch kind {
	case kindDate:
		switch v := value.(type) {
		case string:
			t, err := time.Parse(time.RFC3339, v)
			if err != nil {
				return nil, fmt.Errorf("invalid date %q", v)
			}
			return TimeToMillis(t), nil
		case float64:
			return int64(v), nil
		}
		return nil, errors.New("invalid date value")

	case kindInt:
		switch v := value.(type) {
		case float64:
			return int64(v), nil
		case string:
			n, err := strconv.ParseInt(v, 10, 64)
			if err != nil {
				return nil, fmt.Errorf("invalid integer %q", v)
			}
			return n, nil
		}
		return nil, errors.New("invalid integer value")

	case kindJSON:
		// JSON paths are compared as text in both dialects
		switch v := value.(type) {
		case string:
			return v, nil
		case bool:
			return strconv.FormatBool(v), nil
		case float64:
			return strconv.FormatFloat(v, 'f', -1, 64), nil
		case nil:
			return nil, nil
		}
		return nil, errors.New("unsupported value type")

	default:
		switch v := value.(type) {
		case string:
			return v, nil
		case bool:
			return v, nil
		case float64:
			return v, nil
		case nil:
			return nil, nil
		}
		return nil, errors.New("unsupported value type")
	}
}

// operatorMap reports whether a value is an operator object ({"$gt": 1}) rather than
// a literal. Mixing operator and plain keys is rejected as ambiguous.
func operatorMap(value interface{}) (map[string]interface{}, bool) {
	m, ok := value.(map[string]interface{})
	if !ok || len(m) == 0 {
		return nil, false
	}
	for k := range m {
		if !strings.HasPrefix(k, "$") {
			return nil, false
		}
	}
	if _, isDate := m["$date"]; isDate && len(m) == 1 {
		return nil, false
	}
	return m, true
}

// regexMeta are the characters a regex gives special meaning to; escaped they are the literal character.
const regexMeta = `.*+?()[]{}|^$\`

// regexEscapable is what may follow a backslash: the metacharacters plus "/", which JS-style patterns escape.
const regexEscapable = regexMeta + `/`

// regexToLike converts the small regex subset the events UI produces into a LIKE
// pattern. Anything else is rejected rather than silently mistranslated.
func regexToLike(pattern string) (string, error) {
	body := pattern
	anchoredStart := strings.HasPrefix(body, "^")
	if anchoredStart {
		body = body[1:]
	}
	anchoredEnd := endsWithAnchor(body)
	if anchoredEnd {
		body = body[:len(body)-1]
	}

	// ".*" is a wildcard, a lone "." one character, "\x" a literal x; nothing else is supported
	var b strings.Builder
	for i := 0; i < len(body); {
		r, size := utf8.DecodeRuneInString(body[i:])
		switch {
		case r == '\\':
			i += size
			if i >= len(body) {
				return "", fmt.Errorf("unsupported regex %q", pattern)
			}
			esc, escSize := utf8.DecodeRuneInString(body[i:])
			// only escaped metacharacters are literals; \d, \w, \s ... are classes
			if !strings.ContainsRune(regexEscapable, esc) {
				return "", fmt.Errorf("unsupported regex %q", pattern)
			}
			writeLikeLiteral(&b, esc)
			i += escSize

		case r == '.':
			if strings.HasPrefix(body[i+size:], "*") {
				b.WriteByte('%')
				i += size + 1
			} else {
				b.WriteByte('_')
				i += size
			}

		case strings.ContainsRune(regexMeta, r):
			return "", fmt.Errorf("unsupported regex %q", pattern)

		default:
			writeLikeLiteral(&b, r)
			i += size
		}
	}

	out := b.String()
	if !anchoredStart {
		out = "%" + out
	}
	if !anchoredEnd {
		out = out + "%"
	}
	return out, nil
}

// endsWithAnchor reports whether a trailing "$" is the end anchor (even count of preceding backslashes) or an escaped literal.
func endsWithAnchor(s string) bool {
	if !strings.HasSuffix(s, "$") {
		return false
	}
	slashes := 0
	for i := len(s) - 2; i >= 0 && s[i] == '\\'; i-- {
		slashes++
	}
	return slashes%2 == 0
}

func writeLikeLiteral(b *strings.Builder, r rune) {
	if r == '%' || r == '_' || r == '\\' {
		b.WriteByte('\\')
	}
	b.WriteRune(r)
}
