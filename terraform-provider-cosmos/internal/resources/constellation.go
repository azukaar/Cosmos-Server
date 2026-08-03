package resources

import (
	"context"
	"fmt"

	"github.com/azukaar/terraform-provider-cosmos/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &constellationResource{}
	_ resource.ResourceWithImportState = &constellationResource{}
)

func NewConstellationResource() resource.Resource {
	return &constellationResource{}
}

type constellationResource struct {
	client *client.CosmosClient
}

type constellationModel struct {
	DeviceName   types.String `tfsdk:"device_name"`
	Hostname     types.String `tfsdk:"hostname"`
	IPRange      types.String `tfsdk:"ip_range"`
	IsLighthouse types.Bool   `tfsdk:"is_lighthouse"`
	Enabled      types.Bool   `tfsdk:"enabled"`
	NATSReplicas types.Int64  `tfsdk:"nats_replicas"`
}

func (r *constellationResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_constellation"
}

func (r *constellationResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Initializes a Cosmos Constellation VPN network. This is a one-time setup that generates CA certificates and creates the lighthouse device. Changing any field triggers a full reset and re-initialization.",
		Attributes: map[string]schema.Attribute{
			"device_name": schema.StringAttribute{
				Description: "Name for the lighthouse device on this server.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"hostname": schema.StringAttribute{
				Description: "Public hostname for the Constellation VPN (e.g. vpn.example.com).",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"ip_range": schema.StringAttribute{
				Description: "IP range for the VPN network in CIDR notation.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("192.168.201.0/24"),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"is_lighthouse": schema.BoolAttribute{
				Description: "Whether this device acts as a lighthouse (should be true for the initial server).",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"enabled": schema.BoolAttribute{
				Description: "Whether the Constellation VPN is currently enabled.",
				Computed:    true,
			},
			"nats_replicas": schema.Int64Attribute{
				Description: "Number of NATS JetStream replicas for Raft consensus. Set during creation, cannot be changed after. Use odd numbers (1, 3, 5).",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *constellationResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *constellationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan constellationModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := map[string]interface{}{
		"deviceName":   plan.DeviceName.ValueString(),
		"hostname":     plan.Hostname.ValueString(),
		"ipRange":      plan.IPRange.ValueString(),
		"isLighthouse": plan.IsLighthouse.ValueBool(),
	}
	if !plan.NATSReplicas.IsNull() && !plan.NATSReplicas.IsUnknown() {
		body["natsReplicas"] = plan.NATSReplicas.ValueInt64()
	}

	httpResp, err := r.client.Raw.PostApiConstellationCreate(ctx, body)
	if err != nil {
		resp.Diagnostics.AddError("Error creating constellation", err.Error())
		return
	}
	if err := client.CheckResponse(httpResp); err != nil {
		resp.Diagnostics.AddError("Error creating constellation", err.Error())
		return
	}

	plan.Enabled = types.BoolValue(true)
	// nats_replicas is Optional+Computed. If the user didn't set it, the
	// server keeps the existing/default value (effectively 0 — single-node).
	// Materialize that into state so terraform doesn't choke on a still-unknown
	// computed value after apply.
	if plan.NATSReplicas.IsNull() || plan.NATSReplicas.IsUnknown() {
		plan.NATSReplicas = types.Int64Value(0)
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *constellationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state constellationModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Check if constellation is still active by calling the config endpoint.
	httpResp, err := r.client.Raw.GetApiConstellationConfig(ctx)
	if err != nil {
		// If we can't reach the API, preserve state.
		return
	}

	_, err = client.ParseRawResponse(httpResp)
	if err != nil {
		if client.IsNotFound(err) {
			// Constellation was reset/disabled externally — remove from state.
			resp.State.RemoveResource(ctx)
			return
		}
		// Constellation was reset/disabled externally — remove from state.
		resp.State.RemoveResource(ctx)
		return
	}

	// Constellation is active. User-set fields are preserved from state.
	state.Enabled = types.BoolValue(true)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *constellationResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	// All fields are RequiresReplace, so this should never be called.
	resp.Diagnostics.AddError("Update not supported", "All constellation fields require replacement. This is a provider bug.")
}

func (r *constellationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	httpResp, err := r.client.Raw.GetApiConstellationReset(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error resetting constellation", err.Error())
		return
	}
	if err := client.CheckResponse(httpResp); err != nil {
		resp.Diagnostics.AddError("Error resetting constellation", err.Error())
		return
	}
}

func (r *constellationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("device_name"), req, resp)
}
