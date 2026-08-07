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

var (
	_ resource.Resource                = &constellationDNSResource{}
	_ resource.ResourceWithImportState = &constellationDNSResource{}
)

type constellationDNSResource struct {
	client *client.CosmosClient
}

type constellationDNSResourceModel struct {
	Key   types.String `tfsdk:"key"`
	Type  types.String `tfsdk:"type"`
	Value types.String `tfsdk:"value"`
}

func NewConstellationDNSResource() resource.Resource {
	return &constellationDNSResource{}
}

func (r *constellationDNSResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_constellation_dns"
}

func (r *constellationDNSResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Cosmos Constellation DNS entry.",
		Attributes: map[string]schema.Attribute{
			"key": schema.StringAttribute{
				Description: "The DNS entry key (used as the resource ID).",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"type": schema.StringAttribute{
				Description: "The DNS record type (e.g. \"A\", \"CNAME\").",
				Required:    true,
			},
			"value": schema.StringAttribute{
				Description: "The DNS record value.",
				Required:    true,
			},
		},
	}
}

func (r *constellationDNSResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	cosmosClient, ok := req.ProviderData.(*client.CosmosClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.CosmosClient, got: %T", req.ProviderData),
		)
		return
	}

	r.client = cosmosClient
}

func (r *constellationDNSResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan constellationDNSResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	entry := cosmossdk.UtilsConstellationDNSEntry{
		Key:   plan.Key.ValueString(),
		Type:  client.StringPtr(plan.Type.ValueString()),
		Value: client.StringPtr(plan.Value.ValueString()),
	}

	httpResp, err := r.client.Raw.PostApiConstellationDns(ctx, entry)
	if err != nil {
		resp.Diagnostics.AddError("Error creating constellation DNS entry", err.Error())
		return
	}

	if err := client.CheckResponse(httpResp); err != nil {
		resp.Diagnostics.AddError("Error creating constellation DNS entry", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *constellationDNSResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state constellationDNSResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	key := state.Key.ValueString()

	httpResp, err := r.client.Raw.GetApiConstellationDnsKey(ctx, key)
	if err != nil {
		resp.Diagnostics.AddError("Error reading constellation DNS entry", err.Error())
		return
	}

	entry, err := client.ParseResponse[cosmossdk.UtilsConstellationDNSEntry](httpResp)
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading constellation DNS entry", err.Error())
		return
	}

	if entry == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	state.Key = types.StringValue(entry.Key)
	state.Type = types.StringValue(client.StringPtrVal(entry.Type))
	state.Value = types.StringValue(client.StringPtrVal(entry.Value))

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *constellationDNSResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan constellationDNSResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	key := plan.Key.ValueString()

	entry := cosmossdk.UtilsConstellationDNSEntry{
		Key:   key,
		Type:  client.StringPtr(plan.Type.ValueString()),
		Value: client.StringPtr(plan.Value.ValueString()),
	}

	httpResp, err := r.client.Raw.PutApiConstellationDnsKey(ctx, key, entry)
	if err != nil {
		resp.Diagnostics.AddError("Error updating constellation DNS entry", err.Error())
		return
	}

	if err := client.CheckResponse(httpResp); err != nil {
		resp.Diagnostics.AddError("Error updating constellation DNS entry", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *constellationDNSResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state constellationDNSResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	key := state.Key.ValueString()

	httpResp, err := r.client.Raw.DeleteApiConstellationDnsKey(ctx, key)
	if err != nil {
		resp.Diagnostics.AddError("Error deleting constellation DNS entry", err.Error())
		return
	}

	if err := client.CheckResponse(httpResp); err != nil {
		resp.Diagnostics.AddError("Error deleting constellation DNS entry", err.Error())
		return
	}
}

func (r *constellationDNSResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("key"), req, resp)
}
