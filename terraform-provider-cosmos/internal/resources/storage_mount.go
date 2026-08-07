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

var (
	_ resource.Resource                = &storageMountResource{}
	_ resource.ResourceWithImportState = &storageMountResource{}
)

func NewStorageMountResource() resource.Resource {
	return &storageMountResource{}
}

type storageMountResource struct {
	client *client.CosmosClient
}

// ---------- Terraform model ----------

type storageMountModel struct {
	Path       types.String `tfsdk:"path"`
	MountPoint types.String `tfsdk:"mount_point"`
	Permanent  types.Bool   `tfsdk:"permanent"`
	NetDisk    types.Bool   `tfsdk:"net_disk"`
	Chown      types.String `tfsdk:"chown"`
}

// ---------- resource.Resource ----------

func (r *storageMountResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_storage_mount"
}

func (r *storageMountResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a storage mount in Cosmos Server.",
		Attributes: map[string]schema.Attribute{
			"path": schema.StringAttribute{
				Description: "The source path to mount.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"mount_point": schema.StringAttribute{
				Description: "The mount point destination. Used as the resource ID.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"permanent": schema.BoolAttribute{
				Description: "Whether the mount should persist across reboots.",
				Optional:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"net_disk": schema.BoolAttribute{
				Description: "Whether this is a network disk mount.",
				Optional:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"chown": schema.StringAttribute{
				Description: "Change ownership of the mount point (e.g. \"1000:1000\").",
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *storageMountResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *storageMountResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan storageMountModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload := r.modelToSDK(&plan)

	httpResp, err := r.client.Raw.PostApiMount(ctx, payload)
	if err != nil {
		resp.Diagnostics.AddError("Error creating storage mount", err.Error())
		return
	}
	if err := client.CheckResponse(httpResp); err != nil {
		resp.Diagnostics.AddError("Error creating storage mount", err.Error())
		return
	}

	// Read back to verify mount was created and populate state.
	r.readMount(ctx, plan.MountPoint.ValueString(), &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *storageMountResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state storageMountModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	mountPoint := state.MountPoint.ValueString()
	found := r.readMount(ctx, mountPoint, &state, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	if !found {
		// Mount no longer exists; remove from state.
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *storageMountResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	// All attributes have RequiresReplace, so Update should never be called.
	resp.Diagnostics.AddError(
		"Update not supported",
		"All storage mount attributes require replacement. This is an internal error.",
	)
}

func (r *storageMountResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state storageMountModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload := cosmossdk.StorageMountRequest{
		Path:       state.Path.ValueString(),
		MountPoint: state.MountPoint.ValueString(),
	}

	httpResp, err := r.client.Raw.PostApiUnmount(ctx, payload)
	if err != nil {
		resp.Diagnostics.AddError("Error deleting storage mount", err.Error())
		return
	}
	if err := client.CheckResponse(httpResp); err != nil {
		resp.Diagnostics.AddError("Error deleting storage mount", err.Error())
		return
	}
}

func (r *storageMountResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("mount_point"), req, resp)
}

// ---------- helpers ----------

// readMount fetches all mounts and finds the one matching the given mount_point.
// Returns true if the mount was found, false otherwise.
func (r *storageMountResource) readMount(ctx context.Context, mountPoint string, model *storageMountModel, diags interface {
	AddError(string, string)
	HasError() bool
}) bool {
	httpResp, err := r.client.Raw.GetApiMounts(ctx)
	if err != nil {
		diags.AddError("Error reading storage mounts", err.Error())
		return false
	}

	rawData, err := client.ParseRawResponse(httpResp)
	if err != nil {
		if client.IsNotFound(err) {
			return false
		}
		diags.AddError("Error reading storage mounts", err.Error())
		return false
	}
	if rawData == nil {
		return false
	}

	// The response may be an array of mount objects or a map. Try array first.
	var mounts []map[string]interface{}
	if err := json.Unmarshal(rawData, &mounts); err != nil {
		// Try as a map keyed by mount point.
		var mountMap map[string]map[string]interface{}
		if err2 := json.Unmarshal(rawData, &mountMap); err2 != nil {
			diags.AddError("Error parsing storage mounts response", fmt.Sprintf("array: %s, map: %s", err.Error(), err2.Error()))
			return false
		}
		// Convert map to slice for uniform handling.
		for mp, m := range mountMap {
			m["mountPoint"] = mp
			mounts = append(mounts, m)
		}
	}

	for _, m := range mounts {
		mp, _ := m["mountPoint"].(string)
		if mp == "" {
			mp, _ = m["mount_point"].(string)
		}
		if mp != mountPoint {
			continue
		}

		// Found the mount. Populate path if available.
		if p, ok := m["path"].(string); ok {
			model.Path = types.StringValue(p)
		}
		model.MountPoint = types.StringValue(mountPoint)

		if perm, ok := m["permanent"].(bool); ok {
			model.Permanent = types.BoolValue(perm)
		} else if !model.Permanent.IsNull() {
			// Preserve existing state value if not returned.
		}

		if nd, ok := m["netDisk"].(bool); ok {
			model.NetDisk = types.BoolValue(nd)
		} else if ndAlt, ok := m["net_disk"].(bool); ok {
			model.NetDisk = types.BoolValue(ndAlt)
		}

		if ch, ok := m["chown"].(string); ok && ch != "" {
			model.Chown = types.StringValue(ch)
		}

		return true
	}

	return false
}

// modelToSDK converts the Terraform model to an SDK StorageMountRequest.
func (r *storageMountResource) modelToSDK(model *storageMountModel) cosmossdk.StorageMountRequest {
	req := cosmossdk.StorageMountRequest{
		Path:       model.Path.ValueString(),
		MountPoint: model.MountPoint.ValueString(),
	}

	if !model.Permanent.IsNull() && !model.Permanent.IsUnknown() {
		req.Permanent = client.BoolPtr(model.Permanent.ValueBool())
	}
	if !model.NetDisk.IsNull() && !model.NetDisk.IsUnknown() {
		req.NetDisk = client.BoolPtr(model.NetDisk.ValueBool())
	}
	if !model.Chown.IsNull() && !model.Chown.IsUnknown() {
		req.Chown = client.StringPtr(model.Chown.ValueString())
	}

	return req
}
