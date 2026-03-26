package resources

import (
	"context"
	"fmt"

	"github.com/azukaar/terraform-provider-cosmos/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &groupResource{}
	_ resource.ResourceWithImportState = &groupResource{}
)

func NewGroupResource() resource.Resource {
	return &groupResource{}
}

type groupResource struct {
	client *client.CosmosClient
}

// ---------- Terraform model ----------

type groupModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Permissions types.List   `tfsdk:"permissions"`
}

// ---------- resource.Resource ----------

func (r *groupResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_group"
}

func (r *groupResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Cosmos Server user group (custom role). Requires Cosmos Pro.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The numeric ID of the group (assigned by the server).",
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "Display name of the group.",
				Required:    true,
			},
			"permissions": schema.ListAttribute{
				Description: "List of permission IDs assigned to this group.",
				Optional:    true,
				Computed:    true,
				ElementType: types.Int64Type,
			},
		},
	}
}

func (r *groupResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// ---------- CRUD ----------

func (r *groupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan groupModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := map[string]interface{}{
		"name": plan.Name.ValueString(),
	}
	if !plan.Permissions.IsNull() && !plan.Permissions.IsUnknown() {
		var perms []int64
		resp.Diagnostics.Append(plan.Permissions.ElementsAs(ctx, &perms, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		intPerms := make([]int, len(perms))
		for i, p := range perms {
			intPerms[i] = int(p)
		}
		body["permissions"] = intPerms
	}

	httpResp, err := r.client.Raw.PostApiGroups(ctx, body)
	if err != nil {
		resp.Diagnostics.AddError("Error creating group", err.Error())
		return
	}

	type createResp struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}
	result, err := client.ParseResponse[createResp](httpResp)
	if err != nil {
		resp.Diagnostics.AddError("Error creating group", err.Error())
		return
	}

	plan.ID = types.StringValue(fmt.Sprintf("%d", result.ID))

	// Read back to populate permissions with server-applied defaults (LOGIN, LOGIN_WEAK)
	if !r.readGroupIntoModel(ctx, plan.ID.ValueString(), &plan, &resp.Diagnostics) {
		resp.Diagnostics.AddError("Error reading group", "Group not found after creation")
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *groupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state groupModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	found := r.readGroupIntoModel(ctx, state.ID.ValueString(), &state, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	if !found {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *groupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan groupModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := map[string]interface{}{}
	if !plan.Name.IsNull() && !plan.Name.IsUnknown() {
		body["name"] = plan.Name.ValueString()
	}
	if !plan.Permissions.IsNull() && !plan.Permissions.IsUnknown() {
		var perms []int64
		resp.Diagnostics.Append(plan.Permissions.ElementsAs(ctx, &perms, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		intPerms := make([]int, len(perms))
		for i, p := range perms {
			intPerms[i] = int(p)
		}
		body["permissions"] = intPerms
	}

	// SDK uses int for the id path param
	id := plan.ID.ValueString()
	var idInt int
	fmt.Sscanf(id, "%d", &idInt)

	httpResp, err := r.client.Raw.PutApiGroupsId(ctx, idInt, body)
	if err != nil {
		resp.Diagnostics.AddError("Error updating group", err.Error())
		return
	}
	if err := client.CheckResponse(httpResp); err != nil {
		resp.Diagnostics.AddError("Error updating group", err.Error())
		return
	}

	if !r.readGroupIntoModel(ctx, id, &plan, &resp.Diagnostics) {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *groupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state groupModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var idInt int
	fmt.Sscanf(state.ID.ValueString(), "%d", &idInt)

	httpResp, err := r.client.Raw.DeleteApiGroupsId(ctx, idInt)
	if err != nil {
		resp.Diagnostics.AddError("Error deleting group", err.Error())
		return
	}
	if err := client.CheckResponse(httpResp); err != nil {
		resp.Diagnostics.AddError("Error deleting group", err.Error())
		return
	}
}

func (r *groupResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// ---------- helpers ----------

type groupData struct {
	Name        string `json:"name"`
	Permissions []int  `json:"permissions"`
}

func (r *groupResource) readGroupIntoModel(ctx context.Context, id string, model *groupModel, diags *diag.Diagnostics) bool {
	httpResp, err := r.client.Raw.GetApiGroups(ctx)
	if err != nil {
		diags.AddError("Error reading groups", err.Error())
		return true
	}

	groups, err := client.ParseResponse[map[string]groupData](httpResp)
	if err != nil {
		if client.IsNotFound(err) {
			return false
		}
		diags.AddError("Error reading groups", err.Error())
		return true
	}

	if groups == nil {
		return false
	}

	group, exists := (*groups)[id]
	if !exists {
		return false
	}

	model.ID = types.StringValue(id)
	model.Name = types.StringValue(group.Name)

	if len(group.Permissions) > 0 {
		perms := make([]int64, len(group.Permissions))
		for i, p := range group.Permissions {
			perms[i] = int64(p)
		}
		listVal, d := types.ListValueFrom(ctx, types.Int64Type, perms)
		diags.Append(d...)
		model.Permissions = listVal
	} else {
		model.Permissions = types.ListNull(types.Int64Type)
	}

	return true
}
