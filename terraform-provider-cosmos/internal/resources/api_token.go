package resources

import (
	"context"
	"encoding/json"
	"fmt"

	cosmossdk "github.com/azukaar/cosmos-server/go-sdk"
	"github.com/azukaar/terraform-provider-cosmos/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure interface compliance.
var (
	_ resource.Resource                = &apiTokenResource{}
	_ resource.ResourceWithImportState = &apiTokenResource{}
)

// NewAPITokenResource returns a new resource.Resource for API tokens.
func NewAPITokenResource() resource.Resource {
	return &apiTokenResource{}
}

type apiTokenResource struct {
	client *client.CosmosClient
}

// apiTokenModel maps the resource schema to a Go struct.
type apiTokenModel struct {
	Name                    types.String `tfsdk:"name"`
	Description             types.String `tfsdk:"description"`
	ReadOnly                types.Bool   `tfsdk:"read_only"`
	ExpiryDays              types.Int64  `tfsdk:"expiry_days"`
	IpWhitelist             types.List   `tfsdk:"ip_whitelist"`
	RestrictToConstellation types.Bool   `tfsdk:"restrict_to_constellation"`
	Token                   types.String `tfsdk:"token"`
	TokenSuffix             types.String `tfsdk:"token_suffix"`
	ExpiresAt               types.String `tfsdk:"expires_at"`
}

func (r *apiTokenResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_api_token"
}

func (r *apiTokenResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Cosmos Server API token.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Description: "The unique name of the API token. Changing this forces re-creation.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				Description: "A human-readable description of the API token.",
				Optional:    true,
			},
			"read_only": schema.BoolAttribute{
				Description: "Whether the token is read-only. Can only be set at creation time.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
				PlanModifiers: []planmodifier.Bool{
					// read_only is only sent on create; after that it is immutable.
				},
			},
			"expiry_days": schema.Int64Attribute{
				Description: "Number of days until the token expires (0 = never). Changing this forces re-creation.",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(0),
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"ip_whitelist": schema.ListAttribute{
				Description: "List of IP addresses or CIDRs allowed to use this token.",
				Optional:    true,
				ElementType: types.StringType,
			},
			"restrict_to_constellation": schema.BoolAttribute{
				Description: "Whether this token is restricted to the Constellation network.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"token": schema.StringAttribute{
				Description: "The raw API token string. Only available after creation; it is not returned by subsequent reads. On import this value will be empty.",
				Computed:    true,
				Sensitive:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"token_suffix": schema.StringAttribute{
				Description: "The last 5 characters of the token, for identification purposes.",
				Computed:    true,
			},
			"expires_at": schema.StringAttribute{
				Description: "The absolute expiration time of the token in ISO 8601 format. Empty if the token never expires.",
				Computed:    true,
			},
		},
	}
}

func (r *apiTokenResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.CosmosClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.CosmosClient, got: %T", req.ProviderData),
		)
		return
	}
	r.client = c
}

func (r *apiTokenResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan apiTokenModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build the SDK request.
	createReq := cosmossdk.ConfigapiCreateAPITokenRequest{
		Name: plan.Name.ValueString(),
	}

	if !plan.Description.IsNull() && !plan.Description.IsUnknown() {
		createReq.Description = client.StringPtr(plan.Description.ValueString())
	}
	if !plan.ReadOnly.IsNull() && !plan.ReadOnly.IsUnknown() {
		createReq.ReadOnly = client.BoolPtr(plan.ReadOnly.ValueBool())
	}
	if !plan.RestrictToConstellation.IsNull() && !plan.RestrictToConstellation.IsUnknown() {
		createReq.RestrictToConstellation = client.BoolPtr(plan.RestrictToConstellation.ValueBool())
	}
	if !plan.ExpiryDays.IsNull() && !plan.ExpiryDays.IsUnknown() {
		ed := int(plan.ExpiryDays.ValueInt64())
		createReq.ExpiryDays = &ed
	}
	if !plan.IpWhitelist.IsNull() && !plan.IpWhitelist.IsUnknown() {
		var ips []string
		resp.Diagnostics.Append(plan.IpWhitelist.ElementsAs(ctx, &ips, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		createReq.IpWhitelist = &ips
	}

	// Call the API.
	httpResp, err := r.client.Raw.PostApiApiTokens(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError("Error creating API token", err.Error())
		return
	}

	// The create response data is {"token":"..."} — use ParseRawResponse so we
	// can extract the token value which is only returned on create.
	rawData, err := client.ParseRawResponse(httpResp)
	if err != nil {
		resp.Diagnostics.AddError("Error parsing create API token response", err.Error())
		return
	}

	var createResult struct {
		Token       string `json:"token"`
		TokenSuffix string `json:"tokenSuffix"`
		ExpiresAt   string `json:"expiresAt"`
	}
	if err := json.Unmarshal(rawData, &createResult); err != nil {
		resp.Diagnostics.AddError("Error unmarshalling create API token response", err.Error())
		return
	}

	plan.Token = types.StringValue(createResult.Token)
	plan.TokenSuffix = types.StringValue(createResult.TokenSuffix)
	plan.ExpiresAt = types.StringValue(createResult.ExpiresAt)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *apiTokenResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state apiTokenModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := state.Name.ValueString()

	httpResp, err := r.client.Raw.GetApiApiTokens(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error reading API tokens", err.Error())
		return
	}

	rawData, err := client.ParseRawResponse(httpResp)
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error parsing API tokens response", err.Error())
		return
	}

	// The API returns a map keyed by token name.
	var tokensMap map[string]cosmossdk.UtilsAPITokenConfig
	if err := json.Unmarshal(rawData, &tokensMap); err != nil {
		resp.Diagnostics.AddError("Error parsing API tokens response", err.Error())
		return
	}

	tokenData, ok := tokensMap[name]
	if !ok {
		resp.State.RemoveResource(ctx)
		return
	}

	found := &tokenData

	// Map API response back to state.
	state.Name = types.StringValue(client.StringPtrVal(found.Name))
	state.Description = types.StringValue(client.StringPtrVal(found.Description))

	if found.RestrictToConstellation != nil {
		state.RestrictToConstellation = types.BoolValue(*found.RestrictToConstellation)
	} else {
		state.RestrictToConstellation = types.BoolValue(false)
	}

	if found.IpWhitelist != nil {
		listVal, diags := types.ListValueFrom(ctx, types.StringType, *found.IpWhitelist)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		state.IpWhitelist = listVal
	} else {
		state.IpWhitelist = types.ListNull(types.StringType)
	}

	if found.TokenSuffix != nil {
		state.TokenSuffix = types.StringValue(*found.TokenSuffix)
	} else {
		state.TokenSuffix = types.StringValue("")
	}
	if found.ExpiresAt != nil && *found.ExpiresAt != "" {
		state.ExpiresAt = types.StringValue(*found.ExpiresAt)
	} else {
		state.ExpiresAt = types.StringValue("")
	}

	// The token field is never returned by Read. Preserve whatever is already
	// in state (set during Create). On import it will be empty/unknown.
	// state.Token is left unchanged from what was in req.State.

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *apiTokenResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan apiTokenModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Preserve the token from current state since it is not part of the update.
	var state apiTokenModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := plan.Name.ValueString()

	updateReq := cosmossdk.ConfigapiUpdateAPITokenRequest{}

	if !plan.Description.IsNull() && !plan.Description.IsUnknown() {
		updateReq.Description = client.StringPtr(plan.Description.ValueString())
	}
	if !plan.RestrictToConstellation.IsNull() && !plan.RestrictToConstellation.IsUnknown() {
		updateReq.RestrictToConstellation = client.BoolPtr(plan.RestrictToConstellation.ValueBool())
	}
	if !plan.IpWhitelist.IsNull() && !plan.IpWhitelist.IsUnknown() {
		var ips []string
		resp.Diagnostics.Append(plan.IpWhitelist.ElementsAs(ctx, &ips, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		updateReq.IpWhitelist = &ips
	}

	httpResp, err := r.client.Raw.PutApiApiTokensName(ctx, name, updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Error updating API token", err.Error())
		return
	}
	if err := client.CheckResponse(httpResp); err != nil {
		resp.Diagnostics.AddError("Error updating API token", err.Error())
		return
	}

	// Preserve the token value from state — it is never returned after create.
	plan.Token = state.Token

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *apiTokenResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state apiTokenModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := state.Name.ValueString()

	httpResp, err := r.client.Raw.DeleteApiApiTokens(ctx, cosmossdk.ConfigapiDeleteAPITokenRequest{
		Name: name,
	})
	if err != nil {
		resp.Diagnostics.AddError("Error deleting API token", err.Error())
		return
	}
	if err := client.CheckResponse(httpResp); err != nil {
		resp.Diagnostics.AddError("Error deleting API token", err.Error())
		return
	}
}

func (r *apiTokenResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import by name. Note: the token value will be empty after import because
	// it is only returned during creation.
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}
