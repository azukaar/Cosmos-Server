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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure interface compliance.
var (
	_ resource.Resource                = &dockerVolumeResource{}
	_ resource.ResourceWithImportState = &dockerVolumeResource{}
)

// NewDockerVolumeResource returns a new resource.Resource for Docker volumes.
func NewDockerVolumeResource() resource.Resource {
	return &dockerVolumeResource{}
}

type dockerVolumeResource struct {
	client *client.CosmosClient
}

type dockerVolumeResourceModel struct {
	Name   types.String `tfsdk:"name"`
	Driver types.String `tfsdk:"driver"`
}

func (r *dockerVolumeResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_docker_volume"
}

func (r *dockerVolumeResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Docker volume through Cosmos Server.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Description: "The name of the Docker volume (used as the resource ID). Changing this forces re-creation.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"driver": schema.StringAttribute{
				Description: "The volume driver to use. Default: local. Changing this forces re-creation.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("local"),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *dockerVolumeResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *dockerVolumeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan dockerVolumeResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload := cosmossdk.DockerVolumeCreateRequest{
		Name: plan.Name.ValueString(),
	}
	if !plan.Driver.IsNull() && !plan.Driver.IsUnknown() {
		payload.Driver = client.StringPtr(plan.Driver.ValueString())
	}

	httpResp, err := r.client.Raw.PostApiVolumes(ctx, payload)
	if err != nil {
		resp.Diagnostics.AddError("Error creating Docker volume", err.Error())
		return
	}
	if err := client.CheckResponse(httpResp); err != nil {
		resp.Diagnostics.AddError("Error creating Docker volume", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *dockerVolumeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state dockerVolumeResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := state.Name.ValueString()

	httpResp, err := r.client.Raw.GetApiVolumes(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error reading Docker volumes", err.Error())
		return
	}

	rawData, err := client.ParseRawResponse(httpResp)
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading Docker volumes", err.Error())
		return
	}

	// The API returns {"Volumes": [...], "Warnings": ...}
	var volumeList struct {
		Volumes []volumeInfo `json:"Volumes"`
	}
	if err := json.Unmarshal(rawData, &volumeList); err != nil {
		resp.Diagnostics.AddError("Error parsing Docker volumes response", err.Error())
		return
	}

	var found *volumeInfo
	for i := range volumeList.Volumes {
		if volumeList.Volumes[i].Name == name {
			found = &volumeList.Volumes[i]
			break
		}
	}

	if found == nil {
		// Volume no longer exists; remove from state.
		resp.State.RemoveResource(ctx)
		return
	}

	state.Name = types.StringValue(found.Name)

	// Only update driver from API if the user originally set it.
	// Otherwise leave it as null to avoid spurious replacements
	// (e.g. API returns "local" but user didn't specify driver).
	if !state.Driver.IsNull() {
		if found.Driver != "" {
			state.Driver = types.StringValue(found.Driver)
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *dockerVolumeResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	// All attributes require replacement; no in-place update is supported.
	resp.Diagnostics.AddError(
		"Update not supported",
		"All attributes of cosmos_docker_volume require replacement. This method should never be called.",
	)
}

func (r *dockerVolumeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state dockerVolumeResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	httpResp, err := r.client.Raw.DeleteApiVolumeVolumeName(ctx, state.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error deleting Docker volume", err.Error())
		return
	}
	if err := client.CheckResponse(httpResp); err != nil {
		resp.Diagnostics.AddError("Error deleting Docker volume", err.Error())
		return
	}
}

func (r *dockerVolumeResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}

// ---------- helpers ----------

// volumeInfo represents the minimal fields returned by the volume list API.
type volumeInfo struct {
	Name   string `json:"Name"`
	Driver string `json:"Driver"`
}
