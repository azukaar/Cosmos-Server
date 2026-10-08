package docker

import (
	"encoding/json"
	"strings"

	conttype "github.com/docker/docker/api/types/container"
)

// HJSON comment preservation.
//
// The compose editor renders HJSON (a superset of JSON with comments). The
// JSON payload sent to the API cannot carry comments, so the editor extracts a
// node -> comment map (client-side, keyed by JSON key path such as
// "services.web.image") and sends it as "$$comments". The backend stores each
// entry as a cosmos.compose.<path> label on the container config, and returns
// them on export so the editor can restore the comments before the nodes.

const composeCommentsLabelPrefix = "cosmos.compose."

// extractCommentsConfig pulls the optional "$$comments" map (node path ->
// comment text, HJSON comments from the compose editor) out of a
// CreateService request body.
func extractCommentsConfig(rawBody []byte) map[string]string {
	var bodyMap map[string]interface{}
	if err := json.Unmarshal(rawBody, &bodyMap); err != nil {
		return nil
	}
	raw, ok := bodyMap["$$comments"].(map[string]interface{})
	if !ok {
		return nil
	}
	out := map[string]string{}
	for k, v := range raw {
		if s, ok := v.(string); ok && s != "" {
			out[k] = s
		}
	}
	return out
}

// setComposeCommentsLabels writes the node -> comment map as
// cosmos.compose.<path> labels on the container config (paths are sanitized to
// be docker-label-safe: dots are path separators we keep, everything else stays
// alphanumeric/._-).
func setComposeCommentsLabels(conf *conttype.Config, comments map[string]string) {
	if conf == nil || len(comments) == 0 {
		return
	}
	if conf.Labels == nil {
		conf.Labels = make(map[string]string)
	}
	for path, comment := range comments {
		key := composeCommentsLabelPrefix + sanitizeLabelPath(path)
		conf.Labels[key] = comment
	}
}

// getComposeComments collects the cosmos.compose.* labels from a container
// config into a node path -> comment map (strips the prefix).
func getComposeComments(conf *conttype.Config) map[string]string {
	if conf == nil || len(conf.Labels) == 0 {
		return nil
	}
	out := map[string]string{}
	for k, v := range conf.Labels {
		if strings.HasPrefix(k, composeCommentsLabelPrefix) {
			path := strings.TrimPrefix(k, composeCommentsLabelPrefix)
			out[path] = v
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// sanitizeLabelPath ensures a node path is safe to use as (part of) a docker
// label key: allowed chars are letters, digits, '_', '-' and '.' (the path
// separator). Everything else is dropped.
func sanitizeLabelPath(path string) string {
	var b strings.Builder
	for _, r := range path {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' || r == '-' || r == '.' {
			b.WriteRune(r)
		}
	}
	return b.String()
}
