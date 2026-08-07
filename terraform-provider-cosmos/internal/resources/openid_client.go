package resources

import (
	"context"
	"fmt"

	cosmossdk "github.com/azukaar/cosmos-server/go-sdk"
	"github.com/azukaar/terraform-provider-cosmos/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure interface compliance.
var (
	_ resource.Resource                = &openIDClientResource{}
	_ resource.ResourceWithImportState = &openIDClientResource{}
)

// NewOpenIDClientResource returns a new resource.Resource for OpenID clients.
func NewOpenIDClientResource() resource.Resource {
	return &openIDClientResource{}
}

type openIDClientResource struct {
	client *client.CosmosClient
}

// openIDClientModel maps the resource schema to a Go struct.
type openIDClientModel struct {
	ID       types.String `tfsdk:"id"`
	Secret   types.String `tfsdk:"secret"`
	Redirect types.String `tfsdk:"redirect"`
}

func (r *openIDClientResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_openid_client"
}

func (r *openIDClientResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Cosmos Server OpenID Connect client.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the OpenID client. Changing this forces re-creation.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"secret": schema.StringAttribute{
				Description: "The client secret for this OpenID client. Returned on creation only — subsequent reads will not return it.",
				Optional:    true,
				Computed:    true,
				Sensitive:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"redirect": schema.StringAttribute{
				Description: "The redirect URI for this OpenID client.",
				Optional:    true,
			},
		},
	}
}

func (r *openIDClientResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *openIDClientResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan openIDClientModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := cosmossdk.UtilsOpenIDClient{
		Id: plan.ID.ValueString(),
	}

	if !plan.Secret.IsNull() && !plan.Secret.IsUnknown() {
		createReq.Secret = client.StringPtr(plan.Secret.ValueString())
	}
	if !plan.Redirect.IsNull() && !plan.Redirect.IsUnknown() {
		createReq.Redirect = client.StringPtr(plan.Redirect.ValueString())
	}

	httpResp, err := r.client.Raw.PostApiOpenid(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError("Error creating OpenID client", err.Error())
		return
	}

	// Capture the secret from the create response — the GET endpoint strips it.
	result, err := client.ParseResponse[cosmossdk.UtilsOpenIDClient](httpResp)
	if err != nil {
		resp.Diagnostics.AddError("Error creating OpenID client", err.Error())
		return
	}
	if result != nil && result.Secret != nil && *result.Secret != "" {
		plan.Secret = types.StringValue(*result.Secret)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *openIDClientResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state openIDClientModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()

	httpResp, err := r.client.Raw.GetApiOpenidId(ctx, id)
	if err != nil {
		resp.Diagnostics.AddError("Error reading OpenID client", err.Error())
		return
	}

	result, err := client.ParseResponse[cosmossdk.UtilsOpenIDClient](httpResp)
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error parsing OpenID client response", err.Error())
		return
	}

	if result == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	state.ID = types.StringValue(result.Id)

	// Don't read secret back from API — preserve whatever is in state.
	// The secret should ideally not be returned by the server at all.

	if result.Redirect != nil {
		state.Redirect = types.StringValue(*result.Redirect)
	} else {
		state.Redirect = types.StringNull()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *openIDClientResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan openIDClientModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := plan.ID.ValueString()

	updateReq := cosmossdk.UtilsOpenIDClient{
		Id: id,
	}

	if !plan.Secret.IsNull() && !plan.Secret.IsUnknown() {
		updateReq.Secret = client.StringPtr(plan.Secret.ValueString())
	}
	if !plan.Redirect.IsNull() && !plan.Redirect.IsUnknown() {
		updateReq.Redirect = client.StringPtr(plan.Redirect.ValueString())
	}

	httpResp, err := r.client.Raw.PutApiOpenidId(ctx, id, updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Error updating OpenID client", err.Error())
		return
	}
	if err := client.CheckResponse(httpResp); err != nil {
		resp.Diagnostics.AddError("Error updating OpenID client", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *openIDClientResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state openIDClientModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()

	httpResp, err := r.client.Raw.DeleteApiOpenidId(ctx, id)
	if err != nil {
		resp.Diagnostics.AddError("Error deleting OpenID client", err.Error())
		return
	}
	if err := client.CheckResponse(httpResp); err != nil {
		resp.Diagnostics.AddError("Error deleting OpenID client", err.Error())
		return
	}
}

func (r *openIDClientResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
