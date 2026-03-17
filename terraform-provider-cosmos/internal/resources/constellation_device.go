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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &constellationDeviceResource{}
	_ resource.ResourceWithImportState = &constellationDeviceResource{}
)

// NewConstellationDeviceResource returns a new constellation device resource instance.
func NewConstellationDeviceResource() resource.Resource {
	return &constellationDeviceResource{}
}

type constellationDeviceResource struct {
	client *client.CosmosClient
}

// ---------- Terraform model ----------

type constellationDeviceResourceModel struct {
	DeviceName     types.String `tfsdk:"device_name"`
	IP             types.String `tfsdk:"ip"`
	Nickname       types.String `tfsdk:"nickname"`
	IsLighthouse   types.Bool   `tfsdk:"is_lighthouse"`
	IsRelay        types.Bool   `tfsdk:"is_relay"`
	IsExitNode     types.Bool   `tfsdk:"is_exit_node"`
	IsLoadBalancer types.Bool   `tfsdk:"is_load_balancer"`
	PublicHostname types.String `tfsdk:"public_hostname"`
	Invisible      types.Bool   `tfsdk:"invisible"`
	CosmosNode     types.Int64  `tfsdk:"cosmos_node"`
	Config         types.String `tfsdk:"config"`
}

type deviceJSON struct {
	DeviceName     string `json:"deviceName"`
	IP             string `json:"ip"`
	Nickname       string `json:"nickname"`
	IsLighthouse   *bool  `json:"isLighthouse,omitempty"`
	IsRelay        *bool  `json:"isRelay,omitempty"`
	IsExitNode     *bool  `json:"isExitNode,omitempty"`
	IsLoadBalancer *bool  `json:"isLoadBalancer,omitempty"`
	PublicHostname string `json:"PublicHostname,omitempty"`
	Invisible      *bool  `json:"invisible,omitempty"`
	CosmosNode     *int   `json:"cosmosNode,omitempty"`
	Port           string `json:"port,omitempty"`
}

// ---------- Resource interface ----------

func (r *constellationDeviceResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_constellation_device"
}

func (r *constellationDeviceResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Cosmos Server constellation VPN device.",
		Attributes: map[string]schema.Attribute{
			"device_name": schema.StringAttribute{
				Description: "The unique name of the device.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"ip": schema.StringAttribute{
				Description: "The IP address assigned to the device in the constellation network.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"is_lighthouse": schema.BoolAttribute{
				Description: "Whether this device acts as a lighthouse.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"is_relay": schema.BoolAttribute{
				Description: "Whether this device acts as a relay.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"is_exit_node": schema.BoolAttribute{
				Description: "Whether this device acts as an exit node.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"is_load_balancer": schema.BoolAttribute{
				Description: "Whether this device acts as a load balancer.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"public_hostname": schema.StringAttribute{
				Description: "The public hostname of the device.",
				Optional:    true,
				Computed:    true,
			},
			"nickname": schema.StringAttribute{
				Description: "The device nickname. Defaults to device_name if not set.",
				Optional:    true,
				Computed:    true,
			},
			"invisible": schema.BoolAttribute{
				Description: "Whether the device is hidden from the device list.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"cosmos_node": schema.Int64Attribute{
				Description: "Cosmos node type (0 = not a cosmos node, 1 = agent, 2 = full).",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(0),
			},
			"config": schema.StringAttribute{
				Description: "The full Nebula YAML configuration for this device. Only available after creation — not returned by subsequent reads. On import this value will be empty.",
				Computed:    true,
				Sensitive:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *constellationDeviceResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *constellationDeviceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan constellationDeviceResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	deviceName := plan.DeviceName.ValueString()
	nickname := deviceName
	if !plan.Nickname.IsNull() && !plan.Nickname.IsUnknown() {
		nickname = plan.Nickname.ValueString()
	}

	body := cosmossdk.ConstellationDeviceCreateRequestJSON{
		DeviceName: deviceName,
		Ip:         plan.IP.ValueString(),
		Nickname:   nickname,
	}
	if !plan.IsLighthouse.IsNull() && !plan.IsLighthouse.IsUnknown() {
		body.IsLighthouse = client.BoolPtr(plan.IsLighthouse.ValueBool())
	}
	if !plan.IsRelay.IsNull() && !plan.IsRelay.IsUnknown() {
		body.IsRelay = client.BoolPtr(plan.IsRelay.ValueBool())
	}
	if !plan.IsExitNode.IsNull() && !plan.IsExitNode.IsUnknown() {
		body.IsExitNode = client.BoolPtr(plan.IsExitNode.ValueBool())
	}
	if !plan.IsLoadBalancer.IsNull() && !plan.IsLoadBalancer.IsUnknown() {
		body.IsLoadBalancer = client.BoolPtr(plan.IsLoadBalancer.ValueBool())
	}
	if !plan.PublicHostname.IsNull() && !plan.PublicHostname.IsUnknown() {
		body.PublicHostname = client.StringPtr(plan.PublicHostname.ValueString())
	}
	if !plan.Invisible.IsNull() && !plan.Invisible.IsUnknown() {
		body.Invisible = client.BoolPtr(plan.Invisible.ValueBool())
	}
	if !plan.CosmosNode.IsNull() && !plan.CosmosNode.IsUnknown() {
		v := int(plan.CosmosNode.ValueInt64())
		body.CosmosNode = &v
	}

	httpResp, err := r.client.Raw.PostApiConstellationDevices(ctx, body)
	if err != nil {
		resp.Diagnostics.AddError("Error creating constellation device", err.Error())
		return
	}

	// Capture the Config from the create response — it's only returned once.
	rawData, err := client.ParseRawResponse(httpResp)
	if err != nil {
		resp.Diagnostics.AddError("Error creating constellation device", err.Error())
		return
	}
	if rawData != nil {
		var createResult struct {
			Config string `json:"Config"`
		}
		if err := json.Unmarshal(rawData, &createResult); err == nil && createResult.Config != "" {
			plan.Config = types.StringValue(createResult.Config)
		}
	}

	// Read back to populate computed fields.
	if !r.readDevice(ctx, plan.DeviceName.ValueString(), &plan, &resp.Diagnostics) {
		resp.Diagnostics.AddError("Error reading constellation device", "Device not found after creation")
		return
	}
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *constellationDeviceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state constellationDeviceResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	deviceName := state.DeviceName.ValueString()

	httpResp, err := r.client.Raw.GetApiConstellationDevices(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error reading constellation devices", err.Error())
		return
	}

	rawData, err := client.ParseRawResponse(httpResp)
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error parsing constellation devices response", err.Error())
		return
	}

	if rawData == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	device, err := findDevice(rawData, deviceName)
	if err != nil {
		resp.Diagnostics.AddError("Error locating constellation device", err.Error())
		return
	}
	if device == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	populateDeviceModel(device, &state)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *constellationDeviceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan constellationDeviceResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := cosmossdk.ConstellationDeviceEditRequestJSON{}
	if !plan.IsLighthouse.IsNull() && !plan.IsLighthouse.IsUnknown() {
		body.IsLighthouse = client.BoolPtr(plan.IsLighthouse.ValueBool())
	}
	if !plan.IsRelay.IsNull() && !plan.IsRelay.IsUnknown() {
		body.IsRelay = client.BoolPtr(plan.IsRelay.ValueBool())
	}
	if !plan.IsExitNode.IsNull() && !plan.IsExitNode.IsUnknown() {
		body.IsExitNode = client.BoolPtr(plan.IsExitNode.ValueBool())
	}
	if !plan.IsLoadBalancer.IsNull() && !plan.IsLoadBalancer.IsUnknown() {
		body.IsLoadBalancer = client.BoolPtr(plan.IsLoadBalancer.ValueBool())
	}

	httpResp, err := r.client.Raw.PostApiConstellationEditDevice(ctx, body)
	if err != nil {
		resp.Diagnostics.AddError("Error updating constellation device", err.Error())
		return
	}
	if err := client.CheckResponse(httpResp); err != nil {
		resp.Diagnostics.AddError("Error updating constellation device", err.Error())
		return
	}

	// Read back to refresh state.
	if !r.readDevice(ctx, plan.DeviceName.ValueString(), &plan, &resp.Diagnostics) {
		resp.Diagnostics.AddError("Error reading constellation device", "Device not found after update")
		return
	}
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *constellationDeviceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state constellationDeviceResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := state.DeviceName.ValueString()
	httpResp, err := r.client.Raw.PostApiConstellationBlock(ctx, cosmossdk.ConstellationDeviceBlockRequestJSON{
		DeviceName: name,
		Nickname:   name,
		Block:      client.BoolPtr(true),
	})
	if err != nil {
		resp.Diagnostics.AddError("Error deleting (blocking) constellation device", err.Error())
		return
	}
	if err := client.CheckResponse(httpResp); err != nil {
		resp.Diagnostics.AddError("Error deleting (blocking) constellation device", err.Error())
		return
	}
}

func (r *constellationDeviceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("device_name"), req, resp)
}

// ---------- helpers ----------

// readDevice fetches the device list from the API, finds the device by name,
// and populates the Terraform model. The response data structure may be a map
// keyed by device name or a list — we try both.
// It returns false if the device was not found (404), in which case no diagnostic
// error is added — the caller should remove the resource from state.
func (r *constellationDeviceResource) readDevice(ctx context.Context, deviceName string, model *constellationDeviceResourceModel, diags *diag.Diagnostics) bool {
	httpResp, err := r.client.Raw.GetApiConstellationDevices(ctx)
	if err != nil {
		diags.AddError("Error reading constellation devices", err.Error())
		return true
	}

	rawData, err := client.ParseRawResponse(httpResp)
	if err != nil {
		if client.IsNotFound(err) {
			return false
		}
		diags.AddError("Error parsing constellation devices response", err.Error())
		return true
	}
	if rawData == nil {
		return false
	}

	device, err := findDevice(rawData, deviceName)
	if err != nil {
		diags.AddError("Error locating constellation device", err.Error())
		return true
	}
	if device == nil {
		return false
	}

	populateDeviceModel(device, model)
	return true
}

// populateDeviceModel maps the raw device JSON fields to the Terraform model.
func populateDeviceModel(device *deviceJSON, model *constellationDeviceResourceModel) {
	model.DeviceName = types.StringValue(device.DeviceName)
	model.IP = types.StringValue(device.IP)

	if device.Nickname != "" {
		model.Nickname = types.StringValue(device.Nickname)
	} else {
		model.Nickname = types.StringValue("")
	}

	model.IsLighthouse = types.BoolValue(client.BoolPtrVal(device.IsLighthouse))
	model.IsRelay = types.BoolValue(client.BoolPtrVal(device.IsRelay))
	model.IsExitNode = types.BoolValue(client.BoolPtrVal(device.IsExitNode))
	model.IsLoadBalancer = types.BoolValue(client.BoolPtrVal(device.IsLoadBalancer))
	model.PublicHostname = types.StringValue(device.PublicHostname)
	model.Invisible = types.BoolValue(client.BoolPtrVal(device.Invisible))
	if device.CosmosNode != nil {
		model.CosmosNode = types.Int64Value(int64(*device.CosmosNode))
	} else {
		model.CosmosNode = types.Int64Value(0)
	}
}

// findDevice tries to locate a device by name in the raw JSON data. The API
// may return either a map[string]Device (keyed by device name) or a []Device.
func findDevice(rawData json.RawMessage, deviceName string) (*deviceJSON, error) {
	// Try map keyed by device name first.
	var deviceMap map[string]json.RawMessage
	if err := json.Unmarshal(rawData, &deviceMap); err == nil {
		// Look for an exact key match.
		if raw, ok := deviceMap[deviceName]; ok {
			var d deviceJSON
			if err := json.Unmarshal(raw, &d); err != nil {
				return nil, fmt.Errorf("unmarshalling device from map: %w", err)
			}
			return &d, nil
		}
		// The map may contain devices where the key differs from the
		// deviceName field. Iterate values and check the field.
		for _, raw := range deviceMap {
			var d deviceJSON
			if err := json.Unmarshal(raw, &d); err != nil {
				continue
			}
			if d.DeviceName == deviceName {
				return &d, nil
			}
		}
	}

	// Try as a list.
	var deviceList []json.RawMessage
	if err := json.Unmarshal(rawData, &deviceList); err == nil {
		for _, raw := range deviceList {
			var d deviceJSON
			if err := json.Unmarshal(raw, &d); err != nil {
				continue
			}
			if d.DeviceName == deviceName {
				return &d, nil
			}
		}
	}

	return nil, nil
}
