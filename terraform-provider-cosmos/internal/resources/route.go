package resources

import (
	"context"
	"fmt"

	cosmossdk "github.com/azukaar/cosmos-server/go-sdk"
	"github.com/azukaar/terraform-provider-cosmos/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

var (
	_ resource.Resource                = &routeResource{}
	_ resource.ResourceWithImportState = &routeResource{}
)

func NewRouteResource() resource.Resource {
	return &routeResource{}
}

type routeResource struct {
	client *client.CosmosClient
}

// ---------- Terraform model ----------

type routeSmartShieldModel struct {
	Enabled               types.Bool    `tfsdk:"enabled"`
	MaxGlobalSimultaneous types.Int64   `tfsdk:"max_global_simultaneous"`
	PerUserByteLimit      types.Int64   `tfsdk:"per_user_byte_limit"`
	PerUserRequestLimit   types.Int64   `tfsdk:"per_user_request_limit"`
	PerUserSimultaneous   types.Int64   `tfsdk:"per_user_simultaneous"`
	PerUserTimeBudget     types.Float32 `tfsdk:"per_user_time_budget"`
	PolicyStrictness      types.Int64   `tfsdk:"policy_strictness"`
	PrivilegedGroups      types.Int64   `tfsdk:"privileged_groups"`
}

var smartShieldAttrTypes = map[string]attr.Type{
	"enabled":                  types.BoolType,
	"max_global_simultaneous":  types.Int64Type,
	"per_user_byte_limit":      types.Int64Type,
	"per_user_request_limit":   types.Int64Type,
	"per_user_simultaneous":    types.Int64Type,
	"per_user_time_budget":     types.Float32Type,
	"policy_strictness":        types.Int64Type,
	"privileged_groups":        types.Int64Type,
}

type routeAdditionalFilterModel struct {
	Name  types.String `tfsdk:"name"`
	Type  types.String `tfsdk:"type"`
	Value types.String `tfsdk:"value"`
}

var additionalFilterAttrTypes = map[string]attr.Type{
	"name":  types.StringType,
	"type":  types.StringType,
	"value": types.StringType,
}

type routeModel struct {
	Name                      types.String `tfsdk:"name"`
	Target                    types.String `tfsdk:"target"`
	Host                      types.String `tfsdk:"host"`
	Mode                      types.String `tfsdk:"mode"`
	Description               types.String `tfsdk:"description"`
	AuthEnabled               types.Bool   `tfsdk:"auth_enabled"`
	AdminOnly                 types.Bool   `tfsdk:"admin_only"`
	Disabled                  types.Bool   `tfsdk:"disabled"`
	HideFromDashboard         types.Bool   `tfsdk:"hide_from_dashboard"`
	RestrictToConstellation   types.Bool   `tfsdk:"restrict_to_constellation"`
	BlockCommonBots           types.Bool   `tfsdk:"block_common_bots"`
	BlockAPIAbuse             types.Bool   `tfsdk:"block_api_abuse"`
	DisableHeaderHardening    types.Bool   `tfsdk:"disable_header_hardening"`
	SpoofHostname             types.Bool   `tfsdk:"spoof_hostname"`
	AcceptInsecureHTTPSTarget types.Bool   `tfsdk:"accept_insecure_https_target"`
	SkipURLClean              types.Bool   `tfsdk:"skip_url_clean"`
	StripPathPrefix           types.Bool   `tfsdk:"strip_path_prefix"`
	Corsorigin                types.String `tfsdk:"cors_origin"`
	Icon                      types.String `tfsdk:"icon"`
	MaxBandwith               types.Int64  `tfsdk:"max_bandwidth"`
	OverwriteHostHeader       types.String `tfsdk:"overwrite_host_header"`
	PathPrefix                types.String `tfsdk:"path_prefix"`
	ThrottlePerMinute         types.Int64  `tfsdk:"throttle_per_minute"`
	Timeout                   types.Int64  `tfsdk:"timeout"`
	Tunnel                    types.String `tfsdk:"tunnel"`
	TunneledHost              types.String `tfsdk:"tunneled_host"`
	LBMode                    types.String `tfsdk:"lb_mode"`
	LBStickyMode              types.Bool   `tfsdk:"lb_sticky_mode"`
	UseHost                   types.Bool   `tfsdk:"use_host"`
	UseH2C                    types.Bool   `tfsdk:"use_h2c"`
	UsePathPrefix             types.Bool   `tfsdk:"use_path_prefix"`
	ExtraHeaders              types.Map    `tfsdk:"extra_headers"`
	WhitelistInboundIPs       types.List   `tfsdk:"whitelist_inbound_ips"`
	AdditionalFilters         types.List   `tfsdk:"additional_filters"`
	SmartShield               types.Object `tfsdk:"smart_shield"`
}

// ---------- resource.Resource ----------

func (r *routeResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_route"
}

func (r *routeResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Cosmos Server proxy route.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Description: "Unique name of the route. Used as the resource ID.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"target": schema.StringAttribute{
				Description: "Target URL the route proxies to.",
				Required:    true,
			},
			"host": schema.StringAttribute{
				Description: "Hostname the route listens on.",
				Optional:    true,
				Computed:    true,
			},
			"mode": schema.StringAttribute{
				Description: "Proxy mode (e.g. PROXY, SPA, STATIC, REDIRECT, SERVAPP).",
				Optional:    true,
				Computed:    true,
			},
			"description": schema.StringAttribute{
				Description: "Human-readable description of the route.",
				Optional:    true,
				Computed:    true,
			},
			"auth_enabled": schema.BoolAttribute{
				Description: "Whether authentication is required.",
				Optional:    true,
				Computed:    true,
			},
			"admin_only": schema.BoolAttribute{
				Description: "Restrict access to admin users.",
				Optional:    true,
				Computed:    true,
			},
			"disabled": schema.BoolAttribute{
				Description: "Whether the route is disabled.",
				Optional:    true,
				Computed:    true,
			},
			"hide_from_dashboard": schema.BoolAttribute{
				Description: "Hide the route from the Cosmos dashboard.",
				Optional:    true,
				Computed:    true,
			},
			"restrict_to_constellation": schema.BoolAttribute{
				Description: "Only allow access from the Constellation VPN.",
				Optional:    true,
				Computed:    true,
			},
			"block_common_bots": schema.BoolAttribute{
				Description: "Block common bots and crawlers.",
				Optional:    true,
				Computed:    true,
			},
			"block_api_abuse": schema.BoolAttribute{
				Description: "Block API abuse patterns.",
				Optional:    true,
				Computed:    true,
			},
			"disable_header_hardening": schema.BoolAttribute{
				Description: "Disable security header hardening.",
				Optional:    true,
				Computed:    true,
			},
			"spoof_hostname": schema.BoolAttribute{
				Description: "Spoof the hostname in proxied requests.",
				Optional:    true,
				Computed:    true,
			},
			"accept_insecure_https_target": schema.BoolAttribute{
				Description: "Accept self-signed certificates on the target.",
				Optional:    true,
				Computed:    true,
			},
			"skip_url_clean": schema.BoolAttribute{
				Description: "Skip URL cleaning/normalization.",
				Optional:    true,
				Computed:    true,
			},
			"strip_path_prefix": schema.BoolAttribute{
				Description: "Strip the path prefix before forwarding.",
				Optional:    true,
				Computed:    true,
			},
			"cors_origin": schema.StringAttribute{
				Description: "Allowed CORS origin.",
				Optional:    true,
				Computed:    true,
			},
			"icon": schema.StringAttribute{
				Description: "Icon identifier for the dashboard.",
				Optional:    true,
				Computed:    true,
			},
			"max_bandwidth": schema.Int64Attribute{
				Description: "Maximum bandwidth in bytes per second (0 = unlimited).",
				Optional:    true,
				Computed:    true,
			},
			"overwrite_host_header": schema.StringAttribute{
				Description: "Override the Host header sent to the target.",
				Optional:    true,
				Computed:    true,
			},
			"path_prefix": schema.StringAttribute{
				Description: "Path prefix to match for this route.",
				Optional:    true,
				Computed:    true,
			},
			"throttle_per_minute": schema.Int64Attribute{
				Description: "Maximum requests per minute (0 = unlimited).",
				Optional:    true,
				Computed:    true,
			},
			"timeout": schema.Int64Attribute{
				Description: "Request timeout in nanoseconds.",
				Optional:    true,
				Computed:    true,
			},
			"tunnel": schema.StringAttribute{
				Description: "Tunnel identifier for Constellation tunneling.",
				Optional:    true,
				Computed:    true,
			},
			"tunneled_host": schema.StringAttribute{
				Description: "Hostname to use when tunneling through Constellation.",
				Optional:    true,
				Computed:    true,
			},
			"lb_mode": schema.StringAttribute{
				Description: "Load balancer mode.",
				Optional:    true,
				Computed:    true,
			},
			"lb_sticky_mode": schema.BoolAttribute{
				Description: "Enable sticky sessions for load balancing.",
				Optional:    true,
				Computed:    true,
			},
			"use_host": schema.BoolAttribute{
				Description: "Whether the route matches by hostname.",
				Optional:    true,
				Computed:    true,
			},
			"use_h2c": schema.BoolAttribute{
				Description: "Use HTTP/2 cleartext (h2c) for the upstream connection.",
				Optional:    true,
				Computed:    true,
			},
			"use_path_prefix": schema.BoolAttribute{
				Description: "Whether the route matches by path prefix.",
				Optional:    true,
				Computed:    true,
			},
			"extra_headers": schema.MapAttribute{
				Description: "Additional HTTP headers to add to proxied requests.",
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
			},
			"whitelist_inbound_ips": schema.ListAttribute{
				Description: "List of IP addresses or CIDRs allowed to access this route.",
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
			},
			"additional_filters": schema.ListNestedAttribute{
				Description: "Additional request filters.",
				Optional:    true,
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Description: "Filter name.",
							Optional:    true,
						},
						"type": schema.StringAttribute{
							Description: "Filter type.",
							Optional:    true,
						},
						"value": schema.StringAttribute{
							Description: "Filter value.",
							Optional:    true,
						},
					},
				},
			},
			"smart_shield": schema.SingleNestedAttribute{
				Description: "SmartShield DDoS protection policy.",
				Optional:    true,
				Computed:    true,
				Attributes: map[string]schema.Attribute{
					"enabled": schema.BoolAttribute{
						Description: "Enable SmartShield protection.",
						Optional:    true,
						Computed:    true,
					},
					"max_global_simultaneous": schema.Int64Attribute{
						Description: "Maximum global simultaneous connections.",
						Optional:    true,
						Computed:    true,
					},
					"per_user_byte_limit": schema.Int64Attribute{
						Description: "Per-user byte limit.",
						Optional:    true,
						Computed:    true,
					},
					"per_user_request_limit": schema.Int64Attribute{
						Description: "Per-user request limit.",
						Optional:    true,
						Computed:    true,
					},
					"per_user_simultaneous": schema.Int64Attribute{
						Description: "Per-user simultaneous connections limit.",
						Optional:    true,
						Computed:    true,
					},
					"per_user_time_budget": schema.Float32Attribute{
						Description: "Per-user time budget in seconds.",
						Optional:    true,
						Computed:    true,
					},
					"policy_strictness": schema.Int64Attribute{
						Description: "Policy strictness level.",
						Optional:    true,
						Computed:    true,
					},
					"privileged_groups": schema.Int64Attribute{
						Description: "Privileged groups bitmask.",
						Optional:    true,
						Computed:    true,
					},
				},
			},
		},
	}
}

func (r *routeResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// ---------- CRUD ----------

func (r *routeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan routeModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiReq := routeModelToSDK(ctx, plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	httpResp, err := r.client.Raw.PostApiRoutes(ctx, apiReq)
	if err != nil {
		resp.Diagnostics.AddError("Error creating route", err.Error())
		return
	}
	if err := client.CheckResponse(httpResp); err != nil {
		resp.Diagnostics.AddError("Error creating route", err.Error())
		return
	}

	if !readRouteIntoModel(ctx, r.client, plan.Name.ValueString(), &plan, &resp.Diagnostics) {
		resp.Diagnostics.AddError("Error reading route", "Route not found after creation")
		return
	}
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *routeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state routeModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	found := readRouteIntoModel(ctx, r.client, state.Name.ValueString(), &state, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	if !found {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *routeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan routeModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state routeModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiReq := routeModelToSDK(ctx, plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	httpResp, err := r.client.Raw.PutApiRoutesName(ctx, state.Name.ValueString(), apiReq)
	if err != nil {
		resp.Diagnostics.AddError("Error updating route", err.Error())
		return
	}
	if err := client.CheckResponse(httpResp); err != nil {
		resp.Diagnostics.AddError("Error updating route", err.Error())
		return
	}

	if !readRouteIntoModel(ctx, r.client, plan.Name.ValueString(), &plan, &resp.Diagnostics) {
		resp.Diagnostics.AddError("Error reading route", "Route not found after update")
		return
	}
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *routeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state routeModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	httpResp, err := r.client.Raw.DeleteApiRoutesName(ctx, state.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error deleting route", err.Error())
		return
	}
	if err := client.CheckResponse(httpResp); err != nil {
		resp.Diagnostics.AddError("Error deleting route", err.Error())
		return
	}
}

func (r *routeResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}

// ---------- helpers ----------

// readRouteIntoModel fetches a route by name from the API and populates the model.
// It returns false if the route was not found (404), in which case no diagnostic
// error is added — the caller should remove the resource from state.
func readRouteIntoModel(ctx context.Context, c *client.CosmosClient, name string, model *routeModel, diags *diag.Diagnostics) bool {
	httpResp, err := c.Raw.GetApiRoutesName(ctx, name)
	if err != nil {
		diags.AddError("Error reading route", err.Error())
		return true
	}
	route, err := client.ParseResponse[cosmossdk.UtilsProxyRouteConfig](httpResp)
	if err != nil {
		if client.IsNotFound(err) {
			return false
		}
		diags.AddError("Error reading route", err.Error())
		return true
	}
	if route == nil {
		return false
	}

	sdkToRouteModel(ctx, route, model, diags)
	return true
}

func routeModelToSDK(ctx context.Context, m routeModel, diags *diag.Diagnostics) cosmossdk.UtilsProxyRouteConfig {
	cfg := cosmossdk.UtilsProxyRouteConfig{
		Name:   m.Name.ValueString(),
		Target: m.Target.ValueString(),
	}

	if !m.Host.IsNull() && !m.Host.IsUnknown() {
		cfg.Host = client.StringPtr(m.Host.ValueString())
	}
	if !m.Mode.IsNull() && !m.Mode.IsUnknown() {
		cfg.Mode = client.StringPtr(m.Mode.ValueString())
	}
	if !m.Description.IsNull() && !m.Description.IsUnknown() {
		cfg.Description = client.StringPtr(m.Description.ValueString())
	}
	if !m.AuthEnabled.IsNull() && !m.AuthEnabled.IsUnknown() {
		cfg.AuthEnabled = client.BoolPtr(m.AuthEnabled.ValueBool())
	}
	if !m.AdminOnly.IsNull() && !m.AdminOnly.IsUnknown() {
		cfg.AdminOnly = client.BoolPtr(m.AdminOnly.ValueBool())
	}
	if !m.Disabled.IsNull() && !m.Disabled.IsUnknown() {
		cfg.Disabled = client.BoolPtr(m.Disabled.ValueBool())
	}
	if !m.HideFromDashboard.IsNull() && !m.HideFromDashboard.IsUnknown() {
		cfg.HideFromDashboard = client.BoolPtr(m.HideFromDashboard.ValueBool())
	}
	if !m.RestrictToConstellation.IsNull() && !m.RestrictToConstellation.IsUnknown() {
		cfg.RestrictToConstellation = client.BoolPtr(m.RestrictToConstellation.ValueBool())
	}
	if !m.BlockCommonBots.IsNull() && !m.BlockCommonBots.IsUnknown() {
		cfg.BlockCommonBots = client.BoolPtr(m.BlockCommonBots.ValueBool())
	}
	if !m.BlockAPIAbuse.IsNull() && !m.BlockAPIAbuse.IsUnknown() {
		cfg.BlockAPIAbuse = client.BoolPtr(m.BlockAPIAbuse.ValueBool())
	}
	if !m.DisableHeaderHardening.IsNull() && !m.DisableHeaderHardening.IsUnknown() {
		cfg.DisableHeaderHardening = client.BoolPtr(m.DisableHeaderHardening.ValueBool())
	}
	if !m.SpoofHostname.IsNull() && !m.SpoofHostname.IsUnknown() {
		cfg.SpoofHostname = client.BoolPtr(m.SpoofHostname.ValueBool())
	}
	if !m.AcceptInsecureHTTPSTarget.IsNull() && !m.AcceptInsecureHTTPSTarget.IsUnknown() {
		cfg.AcceptInsecureHTTPSTarget = client.BoolPtr(m.AcceptInsecureHTTPSTarget.ValueBool())
	}
	if !m.SkipURLClean.IsNull() && !m.SkipURLClean.IsUnknown() {
		cfg.SkipURLClean = client.BoolPtr(m.SkipURLClean.ValueBool())
	}
	if !m.StripPathPrefix.IsNull() && !m.StripPathPrefix.IsUnknown() {
		cfg.StripPathPrefix = client.BoolPtr(m.StripPathPrefix.ValueBool())
	}
	if !m.Corsorigin.IsNull() && !m.Corsorigin.IsUnknown() {
		cfg.Corsorigin = client.StringPtr(m.Corsorigin.ValueString())
	}
	if !m.Icon.IsNull() && !m.Icon.IsUnknown() {
		cfg.Icon = client.StringPtr(m.Icon.ValueString())
	}
	if !m.MaxBandwith.IsNull() && !m.MaxBandwith.IsUnknown() {
		cfg.MaxBandwith = client.IntPtr(int(m.MaxBandwith.ValueInt64()))
	}
	if !m.OverwriteHostHeader.IsNull() && !m.OverwriteHostHeader.IsUnknown() {
		cfg.OverwriteHostHeader = client.StringPtr(m.OverwriteHostHeader.ValueString())
	}
	if !m.PathPrefix.IsNull() && !m.PathPrefix.IsUnknown() {
		cfg.PathPrefix = client.StringPtr(m.PathPrefix.ValueString())
	}
	if !m.ThrottlePerMinute.IsNull() && !m.ThrottlePerMinute.IsUnknown() {
		cfg.ThrottlePerMinute = client.IntPtr(int(m.ThrottlePerMinute.ValueInt64()))
	}
	if !m.Timeout.IsNull() && !m.Timeout.IsUnknown() {
		t := cosmossdk.TimeDuration(m.Timeout.ValueInt64())
		cfg.Timeout = &t
	}
	if !m.Tunnel.IsNull() && !m.Tunnel.IsUnknown() {
		cfg.Tunnel = client.StringPtr(m.Tunnel.ValueString())
	}
	if !m.TunneledHost.IsNull() && !m.TunneledHost.IsUnknown() {
		cfg.TunneledHost = client.StringPtr(m.TunneledHost.ValueString())
	}
	if !m.LBMode.IsNull() && !m.LBMode.IsUnknown() {
		cfg.LBMode = client.StringPtr(m.LBMode.ValueString())
	}
	if !m.LBStickyMode.IsNull() && !m.LBStickyMode.IsUnknown() {
		cfg.LBStickyMode = client.BoolPtr(m.LBStickyMode.ValueBool())
	}
	if !m.UseHost.IsNull() && !m.UseHost.IsUnknown() {
		cfg.UseHost = client.BoolPtr(m.UseHost.ValueBool())
	}
	if !m.UseH2C.IsNull() && !m.UseH2C.IsUnknown() {
		cfg.UseH2C = client.BoolPtr(m.UseH2C.ValueBool())
	}
	if !m.UsePathPrefix.IsNull() && !m.UsePathPrefix.IsUnknown() {
		cfg.UsePathPrefix = client.BoolPtr(m.UsePathPrefix.ValueBool())
	}
	if !m.ExtraHeaders.IsNull() && !m.ExtraHeaders.IsUnknown() {
		headers := make(map[string]string)
		diags.Append(m.ExtraHeaders.ElementsAs(ctx, &headers, false)...)
		cfg.ExtraHeaders = &headers
	}
	if !m.WhitelistInboundIPs.IsNull() && !m.WhitelistInboundIPs.IsUnknown() {
		var ips []string
		diags.Append(m.WhitelistInboundIPs.ElementsAs(ctx, &ips, false)...)
		cfg.WhitelistInboundIPs = &ips
	}
	if !m.AdditionalFilters.IsNull() && !m.AdditionalFilters.IsUnknown() {
		var filterModels []routeAdditionalFilterModel
		diags.Append(m.AdditionalFilters.ElementsAs(ctx, &filterModels, false)...)
		filters := make([]cosmossdk.UtilsAddionalFiltersConfig, len(filterModels))
		for i, fm := range filterModels {
			if !fm.Name.IsNull() {
				filters[i].Name = client.StringPtr(fm.Name.ValueString())
			}
			if !fm.Type.IsNull() {
				filters[i].Type = client.StringPtr(fm.Type.ValueString())
			}
			if !fm.Value.IsNull() {
				filters[i].Value = client.StringPtr(fm.Value.ValueString())
			}
		}
		cfg.AddionalFilters = &filters
	}

	if !m.SmartShield.IsNull() && !m.SmartShield.IsUnknown() {
		var ssModel routeSmartShieldModel
		d := m.SmartShield.As(ctx, &ssModel, basetypes.ObjectAsOptions{})
		diags.Append(d...)
		if !d.HasError() {
			ss := &cosmossdk.UtilsSmartShieldPolicy{}
			if !ssModel.Enabled.IsNull() && !ssModel.Enabled.IsUnknown() {
				ss.Enabled = client.BoolPtr(ssModel.Enabled.ValueBool())
			}
			if !ssModel.MaxGlobalSimultaneous.IsNull() && !ssModel.MaxGlobalSimultaneous.IsUnknown() {
				ss.MaxGlobalSimultaneous = client.IntPtr(int(ssModel.MaxGlobalSimultaneous.ValueInt64()))
			}
			if !ssModel.PerUserByteLimit.IsNull() && !ssModel.PerUserByteLimit.IsUnknown() {
				ss.PerUserByteLimit = client.IntPtr(int(ssModel.PerUserByteLimit.ValueInt64()))
			}
			if !ssModel.PerUserRequestLimit.IsNull() && !ssModel.PerUserRequestLimit.IsUnknown() {
				ss.PerUserRequestLimit = client.IntPtr(int(ssModel.PerUserRequestLimit.ValueInt64()))
			}
			if !ssModel.PerUserSimultaneous.IsNull() && !ssModel.PerUserSimultaneous.IsUnknown() {
				ss.PerUserSimultaneous = client.IntPtr(int(ssModel.PerUserSimultaneous.ValueInt64()))
			}
			if !ssModel.PerUserTimeBudget.IsNull() && !ssModel.PerUserTimeBudget.IsUnknown() {
				ss.PerUserTimeBudget = client.Float32Ptr(ssModel.PerUserTimeBudget.ValueFloat32())
			}
			if !ssModel.PolicyStrictness.IsNull() && !ssModel.PolicyStrictness.IsUnknown() {
				ss.PolicyStrictness = client.IntPtr(int(ssModel.PolicyStrictness.ValueInt64()))
			}
			if !ssModel.PrivilegedGroups.IsNull() && !ssModel.PrivilegedGroups.IsUnknown() {
				ss.PrivilegedGroups = client.IntPtr(int(ssModel.PrivilegedGroups.ValueInt64()))
			}
			cfg.SmartShield = ss
		}
	}

	return cfg
}

func sdkToRouteModel(ctx context.Context, route *cosmossdk.UtilsProxyRouteConfig, m *routeModel, diags *diag.Diagnostics) {
	m.Name = types.StringValue(route.Name)
	m.Target = types.StringValue(route.Target)
	m.Host = types.StringValue(client.StringPtrVal(route.Host))
	m.Mode = types.StringValue(client.StringPtrVal(route.Mode))
	m.Description = types.StringValue(client.StringPtrVal(route.Description))
	m.AuthEnabled = types.BoolValue(client.BoolPtrVal(route.AuthEnabled))
	m.AdminOnly = types.BoolValue(client.BoolPtrVal(route.AdminOnly))
	m.Disabled = types.BoolValue(client.BoolPtrVal(route.Disabled))
	m.HideFromDashboard = types.BoolValue(client.BoolPtrVal(route.HideFromDashboard))
	m.RestrictToConstellation = types.BoolValue(client.BoolPtrVal(route.RestrictToConstellation))
	m.BlockCommonBots = types.BoolValue(client.BoolPtrVal(route.BlockCommonBots))
	m.BlockAPIAbuse = types.BoolValue(client.BoolPtrVal(route.BlockAPIAbuse))
	m.DisableHeaderHardening = types.BoolValue(client.BoolPtrVal(route.DisableHeaderHardening))
	m.SpoofHostname = types.BoolValue(client.BoolPtrVal(route.SpoofHostname))
	m.AcceptInsecureHTTPSTarget = types.BoolValue(client.BoolPtrVal(route.AcceptInsecureHTTPSTarget))
	m.SkipURLClean = types.BoolValue(client.BoolPtrVal(route.SkipURLClean))
	m.StripPathPrefix = types.BoolValue(client.BoolPtrVal(route.StripPathPrefix))
	m.Corsorigin = types.StringValue(client.StringPtrVal(route.Corsorigin))
	m.Icon = types.StringValue(client.StringPtrVal(route.Icon))
	m.MaxBandwith = types.Int64Value(int64(client.IntPtrVal(route.MaxBandwith)))
	m.OverwriteHostHeader = types.StringValue(client.StringPtrVal(route.OverwriteHostHeader))
	m.PathPrefix = types.StringValue(client.StringPtrVal(route.PathPrefix))
	m.ThrottlePerMinute = types.Int64Value(int64(client.IntPtrVal(route.ThrottlePerMinute)))
	if route.Timeout != nil {
		m.Timeout = types.Int64Value(int64(*route.Timeout))
	} else {
		m.Timeout = types.Int64Value(0)
	}
	m.Tunnel = types.StringValue(client.StringPtrVal(route.Tunnel))
	m.TunneledHost = types.StringValue(client.StringPtrVal(route.TunneledHost))
	m.LBMode = types.StringValue(client.StringPtrVal(route.LBMode))
	m.LBStickyMode = types.BoolValue(client.BoolPtrVal(route.LBStickyMode))
	m.UseHost = types.BoolValue(client.BoolPtrVal(route.UseHost))
	m.UseH2C = types.BoolValue(client.BoolPtrVal(route.UseH2C))
	m.UsePathPrefix = types.BoolValue(client.BoolPtrVal(route.UsePathPrefix))

	// ExtraHeaders
	if route.ExtraHeaders != nil && len(*route.ExtraHeaders) > 0 {
		mapVal, d := types.MapValueFrom(ctx, types.StringType, *route.ExtraHeaders)
		diags.Append(d...)
		m.ExtraHeaders = mapVal
	} else {
		m.ExtraHeaders = types.MapNull(types.StringType)
	}

	// WhitelistInboundIPs
	if route.WhitelistInboundIPs != nil && len(*route.WhitelistInboundIPs) > 0 {
		listVal, d := types.ListValueFrom(ctx, types.StringType, *route.WhitelistInboundIPs)
		diags.Append(d...)
		m.WhitelistInboundIPs = listVal
	} else {
		m.WhitelistInboundIPs = types.ListNull(types.StringType)
	}

	// AdditionalFilters
	if route.AddionalFilters != nil && len(*route.AddionalFilters) > 0 {
		filterModels := make([]routeAdditionalFilterModel, len(*route.AddionalFilters))
		for i, f := range *route.AddionalFilters {
			filterModels[i] = routeAdditionalFilterModel{
				Name:  types.StringValue(client.StringPtrVal(f.Name)),
				Type:  types.StringValue(client.StringPtrVal(f.Type)),
				Value: types.StringValue(client.StringPtrVal(f.Value)),
			}
		}
		listVal, d := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: additionalFilterAttrTypes}, filterModels)
		diags.Append(d...)
		m.AdditionalFilters = listVal
	} else {
		m.AdditionalFilters = types.ListNull(types.ObjectType{AttrTypes: additionalFilterAttrTypes})
	}

	// SmartShield
	if route.SmartShield != nil {
		ssModel := routeSmartShieldModel{
			Enabled:               types.BoolValue(client.BoolPtrVal(route.SmartShield.Enabled)),
			MaxGlobalSimultaneous: types.Int64Value(int64(client.IntPtrVal(route.SmartShield.MaxGlobalSimultaneous))),
			PerUserByteLimit:      types.Int64Value(int64(client.IntPtrVal(route.SmartShield.PerUserByteLimit))),
			PerUserRequestLimit:   types.Int64Value(int64(client.IntPtrVal(route.SmartShield.PerUserRequestLimit))),
			PerUserSimultaneous:   types.Int64Value(int64(client.IntPtrVal(route.SmartShield.PerUserSimultaneous))),
			PerUserTimeBudget:     types.Float32Value(client.Float32PtrVal(route.SmartShield.PerUserTimeBudget)),
			PolicyStrictness:      types.Int64Value(int64(client.IntPtrVal(route.SmartShield.PolicyStrictness))),
			PrivilegedGroups:      types.Int64Value(int64(client.IntPtrVal(route.SmartShield.PrivilegedGroups))),
		}
		objVal, d := types.ObjectValueFrom(ctx, smartShieldAttrTypes, ssModel)
		diags.Append(d...)
		m.SmartShield = objVal
	} else {
		m.SmartShield = types.ObjectNull(smartShieldAttrTypes)
	}
}
