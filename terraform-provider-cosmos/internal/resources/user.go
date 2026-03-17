package resources

import (
	"context"
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
	_ resource.Resource                = &userResource{}
	_ resource.ResourceWithImportState = &userResource{}
)

func NewUserResource() resource.Resource {
	return &userResource{}
}

type userResource struct {
	client *client.CosmosClient
}

// ---------- Terraform model ----------

type userModel struct {
	Nickname  types.String `tfsdk:"nickname"`
	Email     types.String `tfsdk:"email"`
	Role      types.Int64  `tfsdk:"role"`
	Password  types.String `tfsdk:"password"`
	CreatedAt types.String `tfsdk:"created_at"`
	LastLogin types.String `tfsdk:"last_login"`
}

// ---------- resource.Resource ----------

func (r *userResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user"
}

func (r *userResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Cosmos Server user account.",
		Attributes: map[string]schema.Attribute{
			"nickname": schema.StringAttribute{
				Description: "Unique nickname of the user. Used as the resource ID.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"email": schema.StringAttribute{
				Description: "Email address of the user.",
				Optional:    true,
				Computed:    true,
			},
			"role": schema.Int64Attribute{
				Description: "Role of the user (e.g. 1 = admin, 2 = user).",
				Optional:    true,
				Computed:    true,
			},
			"password": schema.StringAttribute{
				Description: "Password for the user. Only used during creation; never read back from the API.",
				Optional:    true,
				Sensitive:   true,
			},
			"created_at": schema.StringAttribute{
				Description: "Timestamp when the user was created.",
				Computed:    true,
			},
			"last_login": schema.StringAttribute{
				Description: "Timestamp of the user's last login.",
				Computed:    true,
			},
		},
	}
}

func (r *userResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *userResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan userModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := cosmossdk.UserCreateRequestJSON{
		Nickname: plan.Nickname.ValueString(),
	}
	if !plan.Email.IsNull() && !plan.Email.IsUnknown() {
		createReq.Email = client.StringPtr(plan.Email.ValueString())
	}
	if !plan.Role.IsNull() && !plan.Role.IsUnknown() {
		createReq.Role = client.IntPtr(int(plan.Role.ValueInt64()))
	}

	httpResp, err := r.client.Raw.PostApiUsers(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError("Error creating user", err.Error())
		return
	}
	if err := client.CheckResponse(httpResp); err != nil {
		resp.Diagnostics.AddError("Error creating user", err.Error())
		return
	}

	// Read back the created user to populate computed fields.
	if !readUserIntoModel(ctx, r.client, plan.Nickname.ValueString(), &plan, &resp.Diagnostics) {
		resp.Diagnostics.AddError("Error reading user", "User not found after creation")
		return
	}
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *userResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state userModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	found := readUserIntoModel(ctx, r.client, state.Nickname.ValueString(), &state, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	if !found {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *userResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan userModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	editReq := cosmossdk.UserEditRequestJSON{}
	if !plan.Email.IsNull() && !plan.Email.IsUnknown() {
		editReq.Email = client.StringPtr(plan.Email.ValueString())
	}
	if !plan.Role.IsNull() && !plan.Role.IsUnknown() {
		editReq.Role = client.IntPtr(int(plan.Role.ValueInt64()))
	}

	httpResp, err := r.client.Raw.PatchApiUsersNickname(ctx, plan.Nickname.ValueString(), editReq)
	if err != nil {
		resp.Diagnostics.AddError("Error updating user", err.Error())
		return
	}
	if err := client.CheckResponse(httpResp); err != nil {
		resp.Diagnostics.AddError("Error updating user", err.Error())
		return
	}

	// Read back to capture any server-side changes.
	if !readUserIntoModel(ctx, r.client, plan.Nickname.ValueString(), &plan, &resp.Diagnostics) {
		resp.Diagnostics.AddError("Error reading user", "User not found after update")
		return
	}
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *userResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state userModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	httpResp, err := r.client.Raw.DeleteApiUsersNickname(ctx, state.Nickname.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error deleting user", err.Error())
		return
	}
	if err := client.CheckResponse(httpResp); err != nil {
		resp.Diagnostics.AddError("Error deleting user", err.Error())
		return
	}
}

func (r *userResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("nickname"), req, resp)
}

// ---------- helpers ----------

// readUserIntoModel fetches a user by nickname from the API and populates the model.
// It returns false if the user was not found (404), in which case no diagnostic
// error is added — the caller should remove the resource from state.
func readUserIntoModel(ctx context.Context, c *client.CosmosClient, nickname string, model *userModel, diags *diag.Diagnostics) bool {
	httpResp, err := c.Raw.GetApiUsersNickname(ctx, nickname)
	if err != nil {
		diags.AddError("Error reading user", err.Error())
		return true
	}
	user, err := client.ParseResponse[cosmossdk.UtilsUser](httpResp)
	if err != nil {
		if client.IsNotFound(err) {
			return false
		}
		diags.AddError("Error reading user", err.Error())
		return true
	}
	if user == nil {
		return false
	}

	model.Nickname = types.StringValue(user.Nickname)
	model.Email = types.StringValue(client.StringPtrVal(user.Email))
	model.Role = types.Int64Value(int64(user.Role))
	model.CreatedAt = types.StringValue(client.StringPtrVal(user.CreatedAt))
	model.LastLogin = types.StringValue(client.StringPtrVal(user.LastLogin))
	// Password is write-only; never populated from API responses.
	return true
}
