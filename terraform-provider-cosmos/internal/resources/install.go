package resources

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	cosmossdk "github.com/azukaar/cosmos-server/go-sdk"
	"github.com/azukaar/terraform-provider-cosmos/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource = &installResource{}
)

func NewInstallResource() resource.Resource {
	return &installResource{}
}

type installResource struct {
	client *client.CosmosClient
}

type installModel struct {
	// Database
	MongoDBMode types.String `tfsdk:"mongodb_mode"`
	MongoDB     types.String `tfsdk:"mongodb"`

	// HTTPS
	Hostname               types.String `tfsdk:"hostname"`
	HTTPSCertificateMode   types.String `tfsdk:"https_certificate_mode"`
	SSLEmail               types.String `tfsdk:"ssl_email"`
	UseWildcardCertificate types.Bool   `tfsdk:"use_wildcard_certificate"`
	DNSChallengeProvider   types.String `tfsdk:"dns_challenge_provider"`
	DNSChallengeConfig     types.Map    `tfsdk:"dns_challenge_config"`
	TLSCert                types.String `tfsdk:"tls_cert"`
	TLSKey                 types.String `tfsdk:"tls_key"`
	AllowHTTPLocalIPAccess types.Bool   `tfsdk:"allow_http_local_ip_access"`

	// Admin
	Nickname types.String `tfsdk:"nickname"`
	Password types.String `tfsdk:"password"`
	Email    types.String `tfsdk:"email"`

	// Optional
	ClearConfig         types.Bool   `tfsdk:"clear_config"`
	ConstellationConfig types.String `tfsdk:"constellation_config"`
	Licence             types.String `tfsdk:"licence"`

	// Computed
	AdminToken     types.String `tfsdk:"admin_token"`
	AdminTokenName types.String `tfsdk:"admin_token_name"`
}

func (r *installResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_install"
}

func (r *installResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	allReplace := []planmodifier.String{stringplanmodifier.RequiresReplace()}

	resp.Schema = schema.Schema{
		Description: "Bootstraps a fresh Cosmos server via /api/setup. Calls the unauthenticated install endpoint with database, HTTPS, and admin credentials, then captures a freshly-issued admin API token (admin_token) for use by downstream resources via a chained provider alias. All input fields require resource replacement; the resource is one-shot.",
		Attributes: map[string]schema.Attribute{
			"mongodb_mode": schema.StringAttribute{
				Description:   "Database mode: \"Create\" (provision MongoDB container), \"Provided\" (use external mongodb URI), or \"DisableUserManagement\" (no DB).",
				Required:      true,
				PlanModifiers: allReplace,
			},
			"mongodb": schema.StringAttribute{
				Description:   "MongoDB connection string. Required when mongodb_mode = \"Provided\".",
				Optional:      true,
				Sensitive:     true,
				PlanModifiers: allReplace,
			},
			"hostname": schema.StringAttribute{
				Description:   "Public hostname for the Cosmos server.",
				Required:      true,
				PlanModifiers: allReplace,
			},
			"https_certificate_mode": schema.StringAttribute{
				Description:   "HTTPS mode (e.g. LETSENCRYPT, SELFSIGNED, PROVIDED, DISABLED).",
				Optional:      true,
				PlanModifiers: allReplace,
			},
			"ssl_email": schema.StringAttribute{
				Description:   "Email for Let's Encrypt registration.",
				Optional:      true,
				PlanModifiers: allReplace,
			},
			"use_wildcard_certificate": schema.BoolAttribute{
				Description: "Whether to issue a wildcard certificate via DNS challenge.",
				Optional:    true,
			},
			"dns_challenge_provider": schema.StringAttribute{
				Description:   "DNS provider for the DNS-01 challenge (e.g. \"cloudflare\").",
				Optional:      true,
				PlanModifiers: allReplace,
			},
			"dns_challenge_config": schema.MapAttribute{
				Description: "Provider-specific DNS challenge environment variables.",
				Optional:    true,
				Sensitive:   true,
				ElementType: types.StringType,
			},
			"tls_cert": schema.StringAttribute{
				Description:   "PEM-encoded TLS certificate when https_certificate_mode = \"PROVIDED\".",
				Optional:      true,
				Sensitive:     true,
				PlanModifiers: allReplace,
			},
			"tls_key": schema.StringAttribute{
				Description:   "PEM-encoded TLS key when https_certificate_mode = \"PROVIDED\".",
				Optional:      true,
				Sensitive:     true,
				PlanModifiers: allReplace,
			},
			"allow_http_local_ip_access": schema.BoolAttribute{
				Description: "Whether to allow plain HTTP access on local IPs.",
				Optional:    true,
			},
			"nickname": schema.StringAttribute{
				Description:   "Admin user nickname.",
				Required:      true,
				PlanModifiers: allReplace,
			},
			"password": schema.StringAttribute{
				Description:   "Admin user password. Must satisfy Cosmos password requirements (min 9 chars, mixed case, digit, symbol).",
				Required:      true,
				Sensitive:     true,
				PlanModifiers: allReplace,
			},
			"email": schema.StringAttribute{
				Description:   "Admin user email.",
				Optional:      true,
				PlanModifiers: allReplace,
			},
			"clear_config": schema.BoolAttribute{
				Description: "If true, wipe existing /config before setup. Use with care.",
				Optional:    true,
			},
			"constellation_config": schema.StringAttribute{
				Description:   "Optional raw constellation config (joining an existing cluster).",
				Optional:      true,
				Sensitive:     true,
				PlanModifiers: allReplace,
			},
			"licence": schema.StringAttribute{
				Description:   "Optional Cosmos Pro licence key. Validated via ProcessLicence at the end of setup.",
				Optional:      true,
				Sensitive:     true,
				PlanModifiers: allReplace,
			},
			"admin_token": schema.StringAttribute{
				Description: "Admin API token issued during setup. Use this in a chained provider alias for downstream resources.",
				Computed:    true,
				Sensitive:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"admin_token_name": schema.StringAttribute{
				Description: "Name of the admin API token issued during setup.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *installResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.CosmosClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.CosmosClient, got %T", req.ProviderData),
		)
		return
	}
	r.client = c
}

func (r *installResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan installModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createAdminToken := true
	body := cosmossdk.MainSetupJSON{
		MongodbMode:      stringPtrIfSet(plan.MongoDBMode),
		Mongodb:          stringPtrIfSet(plan.MongoDB),
		Hostname:         stringPtrIfSet(plan.Hostname),
		Nickname:         stringPtrIfSet(plan.Nickname),
		Password:         stringPtrIfSet(plan.Password),
		Email:            stringPtrIfSet(plan.Email),
		CreateAdminToken: &createAdminToken,
	}

	if !plan.HTTPSCertificateMode.IsNull() && !plan.HTTPSCertificateMode.IsUnknown() {
		body.HttpsCertificateMode = stringPtrIfSet(plan.HTTPSCertificateMode)
	}
	if !plan.SSLEmail.IsNull() && !plan.SSLEmail.IsUnknown() {
		body.SslEmail = stringPtrIfSet(plan.SSLEmail)
	}
	if !plan.UseWildcardCertificate.IsNull() && !plan.UseWildcardCertificate.IsUnknown() {
		v := plan.UseWildcardCertificate.ValueBool()
		body.UseWildcardCertificate = &v
	}
	if !plan.DNSChallengeProvider.IsNull() && !plan.DNSChallengeProvider.IsUnknown() {
		body.DnsChallengeProvider = stringPtrIfSet(plan.DNSChallengeProvider)
	}
	if !plan.DNSChallengeConfig.IsNull() && !plan.DNSChallengeConfig.IsUnknown() {
		var dns map[string]string
		if d := plan.DNSChallengeConfig.ElementsAs(ctx, &dns, false); d.HasError() {
			resp.Diagnostics.Append(d...)
			return
		}
		if len(dns) > 0 {
			body.DNSChallengeConfig = &dns
		}
	}
	if !plan.TLSCert.IsNull() && !plan.TLSCert.IsUnknown() {
		body.TlsCert = stringPtrIfSet(plan.TLSCert)
	}
	if !plan.TLSKey.IsNull() && !plan.TLSKey.IsUnknown() {
		body.TlsKey = stringPtrIfSet(plan.TLSKey)
	}
	if !plan.AllowHTTPLocalIPAccess.IsNull() && !plan.AllowHTTPLocalIPAccess.IsUnknown() {
		v := plan.AllowHTTPLocalIPAccess.ValueBool()
		body.AllowHTTPLocalIPAccess = &v
	}
	if !plan.ClearConfig.IsNull() && !plan.ClearConfig.IsUnknown() {
		v := plan.ClearConfig.ValueBool()
		body.ClearConfig = &v
	}
	if !plan.ConstellationConfig.IsNull() && !plan.ConstellationConfig.IsUnknown() {
		body.ConstellationConfig = stringPtrIfSet(plan.ConstellationConfig)
	}
	if !plan.Licence.IsNull() && !plan.Licence.IsUnknown() {
		body.Licence = stringPtrIfSet(plan.Licence)
	}

	httpResp, err := r.client.Raw.PostApiSetup(ctx, body)
	if err != nil {
		resp.Diagnostics.AddError("Error calling /api/setup", err.Error())
		return
	}

	rawData, err := client.ParseRawResponse(httpResp)
	if err != nil {
		resp.Diagnostics.AddError("Error parsing setup response", err.Error())
		return
	}

	var setupResult struct {
		AdminToken     string `json:"adminToken"`
		AdminTokenName string `json:"adminTokenName"`
	}
	if len(rawData) > 0 && string(rawData) != "null" {
		if err := json.Unmarshal(rawData, &setupResult); err != nil {
			resp.Diagnostics.AddError("Error parsing setup response data", err.Error())
			return
		}
	}

	plan.AdminToken = types.StringValue(setupResult.AdminToken)
	plan.AdminTokenName = types.StringValue(setupResult.AdminTokenName)

	// /api/setup triggers a server restart (~1s after responding). Poll the
	// new endpoint until it comes back up so dependents don't race. The
	// scheme depends on the user's HTTPS choice — DISABLED (or unset) means
	// the server stays on plain HTTP after restart.
	hostname := plan.Hostname.ValueString()
	if hostname != "" {
		scheme := "https"
		mode := strings.ToUpper(plan.HTTPSCertificateMode.ValueString())
		if mode == "" || mode == "DISABLED" {
			scheme = "http"
		}
		// 3-minute budget: LE issuance + redirect + restart can each add latency.
		if err := waitForServer(ctx, scheme, hostname, 3*time.Minute, 5*time.Second); err != nil {
			resp.Diagnostics.AddError("Server did not come back up after /api/setup", err.Error())
			return
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// waitForServer polls <scheme>://<hostname>/cosmos/api/status until any HTTP
// response is received or `timeout` elapses. Any status code (e.g. 200 in
// NewInstall mode, 401 once configured) proves the server is back online.
// TLS verification is skipped because the cert may be self-signed or a
// freshly-issued LE cert that hasn't propagated yet.
func waitForServer(ctx context.Context, scheme, hostname string, timeout, interval time.Duration) error {
	url := scheme + "://" + hostname + "/cosmos/api/status"
	httpClient := &http.Client{
		Timeout: 5 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	deadline := time.Now().Add(timeout)
	var lastErr error
	for {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		httpResp, err := httpClient.Get(url)
		if err == nil {
			httpResp.Body.Close()
			return nil
		}
		lastErr = err
		if time.Now().After(deadline) {
			return fmt.Errorf("timed out after %s waiting for %s: %w", timeout, url, lastErr)
		}
		time.Sleep(interval)
	}
}

func (r *installResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// cosmos_install is one-shot: drift detection is intentionally a no-op.
	// The pre-bootstrap base_url (used to call /api/setup) only serves
	// ACME challenges + redirect after install, so probing it for status
	// is unreliable. Just preserve state.
	var state installModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *installResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Update not supported",
		"All cosmos_install fields require replacement. This is a provider bug if reached.",
	)
}

func (r *installResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state installModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Best-effort: revoke the issued admin token. The server itself stays up;
	// cosmos_install only models the bootstrap, not the server lifecycle.
	tokenName := state.AdminTokenName.ValueString()
	token := state.AdminToken.ValueString()
	if tokenName == "" || token == "" {
		return
	}

	tokenClient, err := client.NewCosmosClient(r.client.BaseURL, token, false)
	if err != nil {
		resp.Diagnostics.AddWarning("Could not build token-revocation client", err.Error())
		return
	}

	httpResp, err := tokenClient.Raw.DeleteApiApiTokens(ctx, cosmossdk.ConfigapiDeleteAPITokenRequest{Name: tokenName})
	if err != nil {
		resp.Diagnostics.AddWarning("Could not revoke admin token", err.Error())
		return
	}
	if err := client.CheckResponse(httpResp); err != nil {
		if !client.IsNotFound(err) {
			resp.Diagnostics.AddWarning("Admin token revocation returned an error", err.Error())
		}
	}
}

func stringPtrIfSet(v types.String) *string {
	if v.IsNull() || v.IsUnknown() {
		return nil
	}
	s := v.ValueString()
	if s == "" {
		return nil
	}
	return &s
}

