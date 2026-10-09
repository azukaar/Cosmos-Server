// Community build stub of the Cosmos Pro feature set: types exist so the shared
// code and the SDK generator compile; handlers answer PRO001 and hooks do nothing.

package pro

import (
	"github.com/nats-io/nats.go"
	"time"
)

// RegisterCIResponder subscribes this node to its own op subject.
func RegisterCIResponder(nc *nats.Conn, self string) (*nats.Subscription, error) {
	// Pro feature stub.
	return nil, nil
}

// StartCIScheduler starts the leader loop, once per process.
func StartCIScheduler() {
	// Pro feature stub.
}

// SetCIClusterDomainProvider is the exported setter, same pattern as
// SetRegistryNodeProvider.
func SetCIClusterDomainProvider(f func() string) {
	// Pro feature stub.
}

// SetCIMetricPusher wires the metrics sink for CI counters.
func SetCIMetricPusher(f func(key string, value int, label, unit, object, setOperation, agglo string)) {
	// Pro feature stub.
}

// CIImagePullAuth is wired as docker.ImagePullAuthProvider.
func CIImagePullAuth(host string) (string, string) {
	// Pro feature stub.
	return "", ""
}

// CISource is where the code comes from.
// CIGitConnection is a set of git credentials shared by projects: one token
// per provider (or per self-hosted instance) instead of one per repository.
// A project names it in CISource.Connection; its token, username and API
// URL then stand in for the project's own.
type CIGitConnection struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	// Provider is one of github, gitlab, gitea, bitbucket or git.
	Provider string `json:"provider"`
	// Host is the git host the token is valid for (github.com,
	// git.example.com). The UI offers the connection for repositories on
	// that host, and self-hosted repository listings derive the API base
	// from it; empty matches the provider's public host. A pasted URL is
	// reduced to its hostname on save.
	Host string `json:"host,omitempty"`
	// APIURL overrides the provider API base for self-hosted instances;
	// empty derives it from the repository URL.
	APIURL string `json:"apiUrl,omitempty"`
	// Token authenticates clones and provider API calls. Write-only
	// through the API: reads answer "****" when set.
	Token string `json:"token,omitempty"`
	// Username pairs with Token for providers that need one (Bitbucket app
	// passwords, plain git basic auth). Defaults per provider.
	Username  string    `json:"username,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// CIRepository is one repository a connection can reach.
type CIRepository struct {
	// FullName is owner/repo (groups included on GitLab).
	FullName string `json:"fullName"`
	// URL is the https clone URL.
	URL           string `json:"url"`
	DefaultBranch string `json:"defaultBranch,omitempty"`
	Private       bool   `json:"private"`
	Description   string `json:"description,omitempty"`
}

type CISource struct {
	Provider string `json:"provider"`
	// RepoURL is the https clone URL (https://github.com/owner/repo[.git]).
	// An ssh clone URL (git@host:owner/repo.git, ssh://host/owner/repo) is
	// accepted on save and converted to its https form: clones and provider
	// calls authenticate with the token over https.
	RepoURL string `json:"repoUrl"`
	// Connection names a CIGitConnection whose credentials this project
	// uses. When set, Token and Username below are ignored (and cleared on
	// save); APIURL still applies when given, else the connection's does.
	Connection string `json:"connection,omitempty"`
	// APIURL overrides the provider API base for self-hosted GitLab / Gitea /
	// GitHub Enterprise; empty derives it from RepoURL.
	APIURL string `json:"apiUrl,omitempty"`
	// Token authenticates clones and provider API calls (webhook creation,
	// commit statuses, collaborator checks) for a project without a
	// connection. Write-only through the API.
	Token string `json:"token,omitempty"`
	// Username pairs with Token for providers that need one (Bitbucket app
	// passwords, plain git basic auth). Defaults per provider.
	Username string `json:"username,omitempty"`
	// RootDir is the sub-folder to build (monorepos).
	RootDir string `json:"rootDir,omitempty"`
	// Branches is the UI branch filter (globs). Empty = every branch. Rules
	// inside cosmos.json / .woodpecker.yml apply INSIDE this filter.
	Branches []string `json:"branches,omitempty"`
	// DefaultBranch is what a manual run builds when no branch is given.
	DefaultBranch string `json:"defaultBranch,omitempty"`
}

// CIBuildSettings are the UI overrides of the detected strategy. Everything
// here is also settable in cosmos.json `build`, which wins when present.
type CIBuildSettings struct {
	Strategy   string            `json:"strategy,omitempty"`
	Dockerfile string            `json:"dockerfile,omitempty"`
	PublishDir string            `json:"publishDir,omitempty"`
	Env        map[string]string `json:"env,omitempty"`
	// Platform is the image platform (default: the build node's).
	Platform       string `json:"platform,omitempty"`
	TimeoutMinutes int    `json:"timeoutMinutes,omitempty"`
}

// CIRegistryTarget is where artifacts go.
type CIRegistryTarget struct {
	// ArtifactQuotaBytes caps the project's stored file artifacts (0: the
	// default, ciDefaultArtifactQuota).
	ArtifactQuotaBytes int64 `json:"artifactQuotaBytes,omitempty"`
	// Registry: the docker registry to push to (its host prefixes the image); Image: repository name (default: project name).
	Registry string `json:"registry,omitempty"`
	Image    string `json:"image,omitempty"`
	// StaticRegistry receives static-site archives.
	StaticRegistry string `json:"staticRegistry,omitempty"`
	// PackageRegistry receives the package of an npm / pypi strategy. Empty
	// on save picks the only registry of that type.
	PackageRegistry string `json:"packageRegistry,omitempty"`
	// Tokens are Cosmos-minted deploy tokens keyed by registry. Never user-visible.
	Tokens map[string]string `json:"tokens,omitempty"`
}

// CIEnvironment maps a branch to a deployment target.
type CIEnvironment struct {
	// Name is the environment label ("production", "staging").
	Name string `json:"name"`
	// Deployment / Site is the target name; empty = the rule's default name.
	Deployment string `json:"deployment,omitempty"`
	// Host is the route hostname the target gets when CI creates it.
	Host       string `json:"host,omitempty"`
	AutoDeploy bool   `json:"autoDeploy"`
	// Env sets or overrides environment variables of the deployed
	// application for this environment only: the service the UI template
	// creates, or the cosmos.json services that take the artifact. Values
	// may use the build variables and the project secrets (${NAME}).
	Env map[string]string `json:"env,omitempty"`
}

// CIDeployTemplate is used when CI has to CREATE the deployment (first deploy).
// Afterwards the deployment is an ordinary user deployment and only its image
// is touched.
type CIDeployTemplate struct {
	Replicas int               `json:"replicas,omitempty"`
	Port     int               `json:"port,omitempty"`
	Host     string            `json:"host,omitempty"`
	Env      map[string]string `json:"env,omitempty"`
	Tags     []string          `json:"tags,omitempty"`
	Strategy string            `json:"strategy,omitempty"`
	// Healthcheck is a command run inside the container (["CMD", ...]).
	Healthcheck []string `json:"healthcheck,omitempty"`
	MemLimit    int64    `json:"memLimit,omitempty"`
	// SPA / Internal apply to static sites.
	SPA      bool `json:"spa,omitempty"`
	Internal bool `json:"internal,omitempty"`
}

// CIDeployRule is the UI deploy rule; cosmos.json `deploy` overrides it.
type CIDeployRule struct {
	Enabled bool   `json:"enabled"`
	Type    string `json:"type,omitempty"`
	// Name is the default target (deployment or site) name.
	Name string `json:"name,omitempty"`
	// Environments: branch glob -> environment. A branch matching no entry is
	// built but not deployed.
	Environments map[string]CIEnvironment `json:"environments,omitempty"`
	// PullRequests enables preview deployments, one per open PR, named
	// <name>pr<number>, removed when the PR closes or after PreviewTTL.
	PullRequests bool   `json:"pullRequests"`
	PreviewTTL   string `json:"previewTtl,omitempty"`
	// PreviewHost is the hostname template of previews; {{pr}} and {{name}}
	// are substituted ("pr-{{pr}}.preview.example.com").
	PreviewHost string           `json:"previewHost,omitempty"`
	Template    CIDeployTemplate `json:"template"`
}

// CISecret is a project secret. Value is write-only through the API.
type CISecret struct {
	Name           string    `json:"name"`
	Value          string    `json:"value,omitempty"`
	AvailableToPRs bool      `json:"availableToPRs"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

// CITrust is the per-project trust model.
type CITrust struct {
	// PullRequests: off | collaborators | nosecrets | approval.
	PullRequests string `json:"pullRequests"`
}

// CIWebhook is the inbound hook state.
type CIWebhook struct {
	// Secret signs inbound deliveries (HMAC / shared token, per provider).
	Secret string `json:"secret"`
	// ProviderHookID is the id of the hook Cosmos created through the provider
	// API, when it could.
	ProviderHookID string    `json:"providerHookId,omitempty"`
	Registered     bool      `json:"registered"`
	Error          string    `json:"error,omitempty"`
	LastDeliveryAt time.Time `json:"lastDeliveryAt,omitempty"`
}

// CIPreview is one live pull-request preview deployment.
type CIPreview struct {
	PR         int       `json:"pr"`
	Branch     string    `json:"branch"`
	Deployment string    `json:"deployment"`
	Host       string    `json:"host,omitempty"`
	Build      int       `json:"build"`
	CreatedAt  time.Time `json:"createdAt"`
	ExpiresAt  time.Time `json:"expiresAt"`
}

// CIProjectStats is the running tally shown on cards.
type CIProjectStats struct {
	Succeeded      int       `json:"succeeded"`
	Failed         int       `json:"failed"`
	LastStatus     string    `json:"lastStatus,omitempty"`
	LastBuild      int       `json:"lastBuild,omitempty"`
	LastBuildAt    time.Time `json:"lastBuildAt,omitempty"`
	LastDurationMs int64     `json:"lastDurationMs,omitempty"`
}

// CIProject is the persistent record of one connected repository.
type CIProject struct {
	Name        string           `json:"name"`
	Description string           `json:"description,omitempty"`
	Source      CISource         `json:"source"`
	Build       CIBuildSettings  `json:"build"`
	Deploy      CIDeployRule     `json:"deploy"`
	Registry    CIRegistryTarget `json:"registry"`
	Secrets     []CISecret       `json:"secrets,omitempty"`
	Trust       CITrust          `json:"trust"`
	// Tags select the nodes builds may run on (AND semantics, like
	// deployments). Empty = any node.
	Tags    []string  `json:"tags,omitempty"`
	Enabled bool      `json:"enabled"`
	Webhook CIWebhook `json:"webhook"`
	// LastBuildNumber is the per-project build counter (CAS-bumped).
	LastBuildNumber int            `json:"lastBuildNumber"`
	Previews        []CIPreview    `json:"previews,omitempty"`
	Stats           CIProjectStats `json:"stats"`
	CreatedAt       time.Time      `json:"createdAt"`
	UpdatedAt       time.Time      `json:"updatedAt"`
}
