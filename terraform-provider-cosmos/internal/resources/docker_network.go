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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure interface compliance.
var (
	_ resource.Resource                = &dockerNetworkResource{}
	_ resource.ResourceWithImportState = &dockerNetworkResource{}
)

// NewDockerNetworkResource returns a new resource.Resource for Docker networks.
func NewDockerNetworkResource() resource.Resource {
	return &dockerNetworkResource{}
}

type dockerNetworkResource struct {
	client *client.CosmosClient
}

type dockerNetworkResourceModel struct {
	ID           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	Driver       types.String `tfsdk:"driver"`
	Subnet       types.String `tfsdk:"subnet"`
	AttachCosmos types.Bool   `tfsdk:"attach_cosmos"`
}

func (r *dockerNetworkResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_docker_network"
}

func (r *dockerNetworkResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Docker network through Cosmos Server.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The Docker network ID assigned on creation.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the Docker network. Changing this forces re-creation.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"driver": schema.StringAttribute{
				Description: "The network driver to use (e.g. bridge, overlay). Changing this forces re-creation.",
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"subnet": schema.StringAttribute{
				Description: "The subnet in CIDR format for the network. Changing this forces re-creation.",
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"attach_cosmos": schema.BoolAttribute{
				Description: "Whether to attach the Cosmos container to this network. Changing this forces re-creation.",
				Optional:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *dockerNetworkResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *dockerNetworkResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan dockerNetworkResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload := cosmossdk.DockerCreateNetworkPayload{
		Name: plan.Name.ValueString(),
	}
	if !plan.Driver.IsNull() && !plan.Driver.IsUnknown() {
		payload.Driver = client.StringPtr(plan.Driver.ValueString())
	}
	if !plan.Subnet.IsNull() && !plan.Subnet.IsUnknown() {
		payload.Subnet = client.StringPtr(plan.Subnet.ValueString())
	}
	if !plan.AttachCosmos.IsNull() && !plan.AttachCosmos.IsUnknown() {
		payload.AttachCosmos = client.BoolPtr(plan.AttachCosmos.ValueBool())
	}

	httpResp, err := r.client.Raw.PostApiNetworks(ctx, payload)
	if err != nil {
		resp.Diagnostics.AddError("Error creating Docker network", err.Error())
		return
	}

	rawData, err := client.ParseRawResponse(httpResp)
	if err != nil {
		resp.Diagnostics.AddError("Error creating Docker network", err.Error())
		return
	}

	// The response data may contain the network ID in various formats.
	// Try to extract it as a JSON object with an "Id" field first,
	// then fall back to a plain string.
	var networkID string
	var idObj struct {
		ID string `json:"Id"`
	}
	if err := json.Unmarshal(rawData, &idObj); err == nil && idObj.ID != "" {
		networkID = idObj.ID
	} else {
		// Try as a plain string.
		if err := json.Unmarshal(rawData, &networkID); err != nil || networkID == "" {
			// Fall back: read the network list to find by name.
			foundID, findErr := r.findNetworkIDByName(ctx, plan.Name.ValueString())
			if findErr != nil {
				resp.Diagnostics.AddError("Error determining Docker network ID after creation", findErr.Error())
				return
			}
			networkID = foundID
		}
	}

	plan.ID = types.StringValue(networkID)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *dockerNetworkResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state dockerNetworkResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	networkID := state.ID.ValueString()
	networkName := state.Name.ValueString()

	httpResp, err := r.client.Raw.GetApiNetworks(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error reading Docker networks", err.Error())
		return
	}

	rawData, err := client.ParseRawResponse(httpResp)
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading Docker networks", err.Error())
		return
	}

	var networks []networkInfo
	if err := json.Unmarshal(rawData, &networks); err != nil {
		resp.Diagnostics.AddError("Error parsing Docker networks response", err.Error())
		return
	}

	var found *networkInfo
	for i := range networks {
		if networks[i].ID == networkID {
			found = &networks[i]
			break
		}
	}
	// Fall back to matching by name if ID was not found.
	if found == nil {
		for i := range networks {
			if networks[i].Name == networkName {
				found = &networks[i]
				break
			}
		}
	}

	if found == nil {
		// Network no longer exists; remove from state.
		resp.State.RemoveResource(ctx)
		return
	}

	state.ID = types.StringValue(found.ID)
	state.Name = types.StringValue(found.Name)

	if found.Driver != "" {
		state.Driver = types.StringValue(found.Driver)
	} else if !state.Driver.IsNull() {
		state.Driver = types.StringNull()
	}

	// Subnet and AttachCosmos are creation-time only and are not returned
	// by the list endpoint, so we preserve the existing state values.

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *dockerNetworkResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	// All attributes require replacement; no in-place update is supported.
	resp.Diagnostics.AddError(
		"Update not supported",
		"All attributes of cosmos_docker_network require replacement. This method should never be called.",
	)
}

func (r *dockerNetworkResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state dockerNetworkResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	httpResp, err := r.client.Raw.DeleteApiNetworkNetworkID(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error deleting Docker network", err.Error())
		return
	}
	if err := client.CheckResponse(httpResp); err != nil {
		resp.Diagnostics.AddError("Error deleting Docker network", err.Error())
		return
	}
}

func (r *dockerNetworkResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// ---------- helpers ----------

// networkInfo represents the minimal fields returned by the network list API.
type networkInfo struct {
	ID     string `json:"Id"`
	Name   string `json:"Name"`
	Driver string `json:"Driver"`
}

// findNetworkIDByName queries the network list and returns the ID for the
// given network name.
func (r *dockerNetworkResource) findNetworkIDByName(ctx context.Context, name string) (string, error) {
	httpResp, err := r.client.Raw.GetApiNetworks(ctx)
	if err != nil {
		return "", fmt.Errorf("listing networks: %w", err)
	}

	rawData, err := client.ParseRawResponse(httpResp)
	if err != nil {
		return "", fmt.Errorf("parsing network list: %w", err)
	}

	var networks []networkInfo
	if err := json.Unmarshal(rawData, &networks); err != nil {
		return "", fmt.Errorf("unmarshalling network list: %w", err)
	}

	for _, n := range networks {
		if n.Name == name {
			return n.ID, nil
		}
	}

	return "", fmt.Errorf("network %q not found after creation", name)
}
