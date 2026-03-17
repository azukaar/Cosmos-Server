package resources

import (
	"context"
	"fmt"

	cosmossdk "github.com/azukaar/cosmos-server/go-sdk"
	"github.com/azukaar/terraform-provider-cosmos/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure interface compliance.
var (
	_ resource.Resource                = &backupResource{}
	_ resource.ResourceWithImportState = &backupResource{}
)

// NewBackupResource returns a new resource.Resource for backup configurations.
func NewBackupResource() resource.Resource {
	return &backupResource{}
}

type backupResource struct {
	client *client.CosmosClient
}

type backupResourceModel struct {
	Name               types.String `tfsdk:"name"`
	Repository         types.String `tfsdk:"repository"`
	Source             types.String `tfsdk:"source"`
	Crontab            types.String `tfsdk:"crontab"`
	CrontabForget      types.String `tfsdk:"crontab_forget"`
	RetentionPolicy    types.String `tfsdk:"retention_policy"`
	AutoStopContainers types.Bool   `tfsdk:"auto_stop_containers"`
}

func (r *backupResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_backup"
}

func (r *backupResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Cosmos Server backup configuration.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Description: "The unique name of the backup configuration (used as the resource ID). Changing this forces re-creation.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"repository": schema.StringAttribute{
				Description: "The repository path for storing backups.",
				Required:    true,
			},
			"source": schema.StringAttribute{
				Description: "The source path to back up.",
				Optional:    true,
			},
			"crontab": schema.StringAttribute{
				Description: "Cron schedule for running backups.",
				Optional:    true,
			},
			"crontab_forget": schema.StringAttribute{
				Description: "Cron schedule for running backup forget/prune.",
				Optional:    true,
				Computed:    true,
			},
				"retention_policy": schema.StringAttribute{
				Description: "Retention policy for backup snapshots.",
				Optional:    true,
				Computed:    true,
			},
			"auto_stop_containers": schema.BoolAttribute{
				Description: "Whether to automatically stop containers before backing up.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
		},
	}
}

func (r *backupResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *backupResource) backupConfigFromModel(model *backupResourceModel) cosmossdk.UtilsSingleBackupConfig {
	cfg := cosmossdk.UtilsSingleBackupConfig{
		Name:       model.Name.ValueString(),
		Repository: model.Repository.ValueString(),
	}

	if !model.Source.IsNull() && !model.Source.IsUnknown() {
		cfg.Source = client.StringPtr(model.Source.ValueString())
	}
	if !model.Crontab.IsNull() && !model.Crontab.IsUnknown() {
		cfg.Crontab = client.StringPtr(model.Crontab.ValueString())
	}
	if !model.CrontabForget.IsNull() && !model.CrontabForget.IsUnknown() {
		cfg.CrontabForget = client.StringPtr(model.CrontabForget.ValueString())
	}
	if !model.RetentionPolicy.IsNull() && !model.RetentionPolicy.IsUnknown() {
		cfg.RetentionPolicy = client.StringPtr(model.RetentionPolicy.ValueString())
	}
	if !model.AutoStopContainers.IsNull() && !model.AutoStopContainers.IsUnknown() {
		cfg.AutoStopContainers = client.BoolPtr(model.AutoStopContainers.ValueBool())
	}

	return cfg
}

func (r *backupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan backupResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	cfg := r.backupConfigFromModel(&plan)

	httpResp, err := r.client.Raw.PostApiBackups(ctx, cfg)
	if err != nil {
		resp.Diagnostics.AddError("Error creating backup", err.Error())
		return
	}
	if err := client.CheckResponse(httpResp); err != nil {
		resp.Diagnostics.AddError("Error creating backup", err.Error())
		return
	}

	// Resolve any unknown Computed fields to null so Terraform doesn't error.
	// We don't read back server-set defaults to avoid phantom diffs on next plan.
	if plan.CrontabForget.IsUnknown() {
		plan.CrontabForget = types.StringNull()
	}
	if plan.RetentionPolicy.IsUnknown() {
		plan.RetentionPolicy = types.StringNull()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *backupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state backupResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := state.Name.ValueString()

	httpResp, err := r.client.Raw.GetApiBackupsConfigName(ctx, name)
	if err != nil {
		resp.Diagnostics.AddError("Error reading backup", err.Error())
		return
	}

	backupPtr, err := client.ParseResponse[cosmossdk.UtilsSingleBackupConfig](httpResp)
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading backup", err.Error())
		return
	}
	if backupPtr == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	backup := *backupPtr

	state.Name = types.StringValue(backup.Name)
	state.Repository = types.StringValue(backup.Repository)

	if backup.Source != nil {
		state.Source = types.StringValue(*backup.Source)
	} else {
		state.Source = types.StringNull()
	}

	if backup.Crontab != nil {
		state.Crontab = types.StringValue(*backup.Crontab)
	} else {
		state.Crontab = types.StringNull()
	}

	// Only update crontab_forget and retention_policy from API if the user set them.
	// The server may populate defaults that would cause phantom diffs.
	if !state.CrontabForget.IsNull() && backup.CrontabForget != nil {
		state.CrontabForget = types.StringValue(*backup.CrontabForget)
	}

	if !state.RetentionPolicy.IsNull() && backup.RetentionPolicy != nil {
		state.RetentionPolicy = types.StringValue(*backup.RetentionPolicy)
	}

	if backup.AutoStopContainers != nil {
		state.AutoStopContainers = types.BoolValue(*backup.AutoStopContainers)
	} else {
		state.AutoStopContainers = types.BoolNull()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *backupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan backupResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	cfg := r.backupConfigFromModel(&plan)

	httpResp, err := r.client.Raw.PostApiBackupsEdit(ctx, cfg)
	if err != nil {
		resp.Diagnostics.AddError("Error updating backup", err.Error())
		return
	}
	if err := client.CheckResponse(httpResp); err != nil {
		resp.Diagnostics.AddError("Error updating backup", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *backupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state backupResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	httpResp, err := r.client.Raw.DeleteApiBackupsName(ctx, state.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error deleting backup", err.Error())
		return
	}
	if err := client.CheckResponse(httpResp); err != nil {
		resp.Diagnostics.AddError("Error deleting backup", err.Error())
		return
	}
}

func (r *backupResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}
