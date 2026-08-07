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

var (
	_ resource.Resource                = &cronJobResource{}
	_ resource.ResourceWithImportState = &cronJobResource{}
)

type cronJobResource struct {
	client *client.CosmosClient
}

type cronJobResourceModel struct {
	Name      types.String `tfsdk:"name"`
	Crontab   types.String `tfsdk:"crontab"`
	Command   types.String `tfsdk:"command"`
	Container types.String `tfsdk:"container"`
	Enabled   types.Bool   `tfsdk:"enabled"`
}

func NewCronJobResource() resource.Resource {
	return &cronJobResource{}
}

func (r *cronJobResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cron_job"
}

func (r *cronJobResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Cosmos cron job.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Description: "The cron job name (used as the resource ID).",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"crontab": schema.StringAttribute{
				Description: "The cron schedule expression (e.g. \"0 * * * *\").",
				Required:    true,
			},
			"command": schema.StringAttribute{
				Description: "The command to execute.",
				Optional:    true,
			},
			"container": schema.StringAttribute{
				Description: "The container to run the command in.",
				Optional:    true,
			},
			"enabled": schema.BoolAttribute{
				Description: "Whether the cron job is enabled. Default: true.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
		},
	}
}

func (r *cronJobResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *cronJobResource) cronConfigFromModel(model *cronJobResourceModel) cosmossdk.UtilsCRONConfig {
	cfg := cosmossdk.UtilsCRONConfig{
		Name:    model.Name.ValueString(),
		Crontab: client.StringPtr(model.Crontab.ValueString()),
		Enabled: client.BoolPtr(model.Enabled.ValueBool()),
	}

	if !model.Command.IsNull() && !model.Command.IsUnknown() {
		cfg.Command = client.StringPtr(model.Command.ValueString())
	}

	if !model.Container.IsNull() && !model.Container.IsUnknown() {
		cfg.Container = client.StringPtr(model.Container.ValueString())
	}

	return cfg
}

func (r *cronJobResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan cronJobResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	cfg := r.cronConfigFromModel(&plan)

	httpResp, err := r.client.Raw.PostApiCron(ctx, cfg)
	if err != nil {
		resp.Diagnostics.AddError("Error creating cron job", err.Error())
		return
	}

	if err := client.CheckResponse(httpResp); err != nil {
		resp.Diagnostics.AddError("Error creating cron job", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *cronJobResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state cronJobResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := state.Name.ValueString()

	httpResp, err := r.client.Raw.GetApiCronName(ctx, name, cosmossdk.UtilsCRONConfig{Name: name})
	if err != nil {
		resp.Diagnostics.AddError("Error reading cron job", err.Error())
		return
	}

	cronCfg, err := client.ParseResponse[cosmossdk.UtilsCRONConfig](httpResp)
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading cron job", err.Error())
		return
	}

	if cronCfg == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	state.Name = types.StringValue(cronCfg.Name)
	state.Crontab = types.StringValue(client.StringPtrVal(cronCfg.Crontab))
	state.Enabled = types.BoolValue(client.BoolPtrVal(cronCfg.Enabled))

	if cronCfg.Command != nil {
		state.Command = types.StringValue(client.StringPtrVal(cronCfg.Command))
	} else {
		state.Command = types.StringNull()
	}

	if cronCfg.Container != nil {
		state.Container = types.StringValue(client.StringPtrVal(cronCfg.Container))
	} else {
		state.Container = types.StringNull()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *cronJobResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan cronJobResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := plan.Name.ValueString()
	cfg := r.cronConfigFromModel(&plan)

	httpResp, err := r.client.Raw.PutApiCronName(ctx, name, cfg)
	if err != nil {
		resp.Diagnostics.AddError("Error updating cron job", err.Error())
		return
	}

	if err := client.CheckResponse(httpResp); err != nil {
		resp.Diagnostics.AddError("Error updating cron job", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *cronJobResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state cronJobResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := state.Name.ValueString()

	httpResp, err := r.client.Raw.DeleteApiCronName(ctx, name, cosmossdk.UtilsCRONConfig{Name: name})
	if err != nil {
		resp.Diagnostics.AddError("Error deleting cron job", err.Error())
		return
	}

	if err := client.CheckResponse(httpResp); err != nil {
		resp.Diagnostics.AddError("Error deleting cron job", err.Error())
		return
	}
}

func (r *cronJobResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}
