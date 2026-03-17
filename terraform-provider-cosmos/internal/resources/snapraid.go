package resources

import (
	"context"
	"encoding/json"
	"fmt"

	cosmossdk "github.com/azukaar/cosmos-server/go-sdk"
	"github.com/azukaar/terraform-provider-cosmos/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &snapRAIDResource{}
	_ resource.ResourceWithImportState = &snapRAIDResource{}
)

func NewSnapRAIDResource() resource.Resource {
	return &snapRAIDResource{}
}

type snapRAIDResource struct {
	client *client.CosmosClient
}

// ---------- Terraform model ----------

type snapRAIDModel struct {
	Name         types.String `tfsdk:"name"`
	Enabled      types.Bool   `tfsdk:"enabled"`
	Data         types.Map    `tfsdk:"data"`
	Parity       types.List   `tfsdk:"parity"`
	SyncCrontab  types.String `tfsdk:"sync_crontab"`
	ScrubCrontab types.String `tfsdk:"scrub_crontab"`
	CheckOnFix   types.Bool   `tfsdk:"check_on_fix"`
}

// ---------- resource.Resource ----------

func (r *snapRAIDResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_snapraid"
}

func (r *snapRAIDResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a SnapRAID configuration in Cosmos Server.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Description: "The unique name of the SnapRAID configuration. Used as the resource ID.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"enabled": schema.BoolAttribute{
				Description: "Whether the SnapRAID configuration is enabled.",
				Optional:    true,
			},
			"data": schema.MapAttribute{
				Description: "Map of data disk names to their paths.",
				Optional:    true,
				ElementType: types.StringType,
			},
			"parity": schema.ListAttribute{
				Description: "List of parity file paths.",
				Optional:    true,
				ElementType: types.StringType,
			},
			"sync_crontab": schema.StringAttribute{
				Description: "Cron schedule for SnapRAID sync operations.",
				Optional:    true,
			},
			"scrub_crontab": schema.StringAttribute{
				Description: "Cron schedule for SnapRAID scrub operations.",
				Optional:    true,
			},
			"check_on_fix": schema.BoolAttribute{
				Description: "Whether to run a check before fixing.",
				Optional:    true,
			},
		},
	}
}

func (r *snapRAIDResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// ---------- CRUD ----------

func (r *snapRAIDResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan snapRAIDModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload := r.modelToSDK(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	httpResp, err := r.client.Raw.PostApiSnapraid(ctx, payload)
	if err != nil {
		resp.Diagnostics.AddError("Error creating SnapRAID config", err.Error())
		return
	}
	if err := client.CheckResponse(httpResp); err != nil {
		resp.Diagnostics.AddError("Error creating SnapRAID config", err.Error())
		return
	}

	// Read back to populate any server-set defaults.
	if !r.readSnapRAID(ctx, plan.Name.ValueString(), &plan, &resp.Diagnostics) {
		resp.Diagnostics.AddError("Error reading SnapRAID config", "SnapRAID config not found after creation")
		return
	}
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *snapRAIDResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state snapRAIDModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	found := r.readSnapRAID(ctx, state.Name.ValueString(), &state, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	if !found {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *snapRAIDResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan snapRAIDModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := plan.Name.ValueString()
	payload := r.modelToSDK(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	httpResp, err := r.client.Raw.PostApiSnapraidName(ctx, name, payload)
	if err != nil {
		resp.Diagnostics.AddError("Error updating SnapRAID config", err.Error())
		return
	}
	if err := client.CheckResponse(httpResp); err != nil {
		resp.Diagnostics.AddError("Error updating SnapRAID config", err.Error())
		return
	}

	// Read back to capture any server-side changes.
	if !r.readSnapRAID(ctx, name, &plan, &resp.Diagnostics) {
		resp.Diagnostics.AddError("Error reading SnapRAID config", "SnapRAID config not found after update")
		return
	}
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *snapRAIDResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state snapRAIDModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := state.Name.ValueString()
	body := cosmossdk.UtilsSnapRAIDConfig{
		Name: client.StringPtr(name),
	}

	httpResp, err := r.client.Raw.DeleteApiSnapraidName(ctx, name, body)
	if err != nil {
		resp.Diagnostics.AddError("Error deleting SnapRAID config", err.Error())
		return
	}
	if err := client.CheckResponse(httpResp); err != nil {
		resp.Diagnostics.AddError("Error deleting SnapRAID config", err.Error())
		return
	}
}

func (r *snapRAIDResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}

// ---------- helpers ----------

// readSnapRAID fetches the SnapRAID config by name and populates the model.
// It returns false if the config was not found (404), in which case no diagnostic
// error is added — the caller should remove the resource from state.
func (r *snapRAIDResource) readSnapRAID(ctx context.Context, name string, model *snapRAIDModel, diags interface {
	AddError(string, string)
	HasError() bool
}) bool {
	body := cosmossdk.UtilsSnapRAIDConfig{
		Name: client.StringPtr(name),
	}

	httpResp, err := r.client.Raw.GetApiSnapraid(ctx, body)
	if err != nil {
		diags.AddError("Error reading SnapRAID config", err.Error())
		return true
	}

	rawData, err := client.ParseRawResponse(httpResp)
	if err != nil {
		if client.IsNotFound(err) {
			return false
		}
		diags.AddError("Error reading SnapRAID config", err.Error())
		return true
	}
	if rawData == nil {
		return false
	}

	var config cosmossdk.UtilsSnapRAIDConfig
	if err := json.Unmarshal(rawData, &config); err != nil {
		diags.AddError("Error parsing SnapRAID config response", err.Error())
		return true
	}

	r.sdkToModel(&config, model)
	return true
}

// modelToSDK converts the Terraform model to an SDK UtilsSnapRAIDConfig.
func (r *snapRAIDResource) modelToSDK(ctx context.Context, model *snapRAIDModel, diags *diag.Diagnostics) cosmossdk.UtilsSnapRAIDConfig {
	config := cosmossdk.UtilsSnapRAIDConfig{
		Name: client.StringPtr(model.Name.ValueString()),
	}

	if !model.Enabled.IsNull() && !model.Enabled.IsUnknown() {
		config.Enabled = client.BoolPtr(model.Enabled.ValueBool())
	}
	if !model.CheckOnFix.IsNull() && !model.CheckOnFix.IsUnknown() {
		config.CheckOnFix = client.BoolPtr(model.CheckOnFix.ValueBool())
	}
	if !model.SyncCrontab.IsNull() && !model.SyncCrontab.IsUnknown() {
		config.SyncCrontab = client.StringPtr(model.SyncCrontab.ValueString())
	}
	if !model.ScrubCrontab.IsNull() && !model.ScrubCrontab.IsUnknown() {
		config.ScrubCrontab = client.StringPtr(model.ScrubCrontab.ValueString())
	}

	if !model.Data.IsNull() && !model.Data.IsUnknown() {
		dataMap := make(map[string]string)
		for k, v := range model.Data.Elements() {
			if sv, ok := v.(types.String); ok {
				dataMap[k] = sv.ValueString()
			}
		}
		config.Data = &dataMap
	}

	if !model.Parity.IsNull() && !model.Parity.IsUnknown() {
		var parityList []string
		for _, v := range model.Parity.Elements() {
			if sv, ok := v.(types.String); ok {
				parityList = append(parityList, sv.ValueString())
			}
		}
		config.Parity = &parityList
	}

	return config
}

// sdkToModel converts an SDK UtilsSnapRAIDConfig to the Terraform model.
func (r *snapRAIDResource) sdkToModel(config *cosmossdk.UtilsSnapRAIDConfig, model *snapRAIDModel) {
	model.Name = types.StringValue(client.StringPtrVal(config.Name))

	if config.Enabled != nil {
		model.Enabled = types.BoolValue(*config.Enabled)
	} else {
		model.Enabled = types.BoolNull()
	}

	if config.CheckOnFix != nil {
		model.CheckOnFix = types.BoolValue(*config.CheckOnFix)
	} else {
		model.CheckOnFix = types.BoolNull()
	}

	if config.SyncCrontab != nil {
		model.SyncCrontab = types.StringValue(*config.SyncCrontab)
	} else {
		model.SyncCrontab = types.StringNull()
	}

	if config.ScrubCrontab != nil {
		model.ScrubCrontab = types.StringValue(*config.ScrubCrontab)
	} else {
		model.ScrubCrontab = types.StringNull()
	}

	if config.Data != nil {
		mapVal, diags := types.MapValueFrom(context.Background(), types.StringType, *config.Data)
		if !diags.HasError() {
			model.Data = mapVal
		} else {
			model.Data = types.MapNull(types.StringType)
		}
	} else {
		model.Data = types.MapNull(types.StringType)
	}

	if config.Parity != nil {
		listVal, diags := types.ListValueFrom(context.Background(), types.StringType, *config.Parity)
		if !diags.HasError() {
			model.Parity = listVal
		} else {
			model.Parity = types.ListNull(types.StringType)
		}
	} else {
		model.Parity = types.ListNull(types.StringType)
	}
}
