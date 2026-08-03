package resources

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

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

// retryOn503 retries an API call while the server returns 503. The
// constellation-deployments KV is created asynchronously by
// pro.ClientHeartbeatInit after the NATS cluster forms, which lags
// cosmos_install completion by ~30-60s. The deployments endpoints are the
// only ones gated on that KV, so the wait belongs here rather than in
// cosmos_install. On final timeout we return the last 503 response so the
// parser surfaces the real DP001 error.
func retryOn503(ctx context.Context, do func() (*http.Response, error)) (*http.Response, error) {
	deadline := time.Now().Add(2 * time.Minute)
	for {
		resp, err := do()
		if err != nil {
			return nil, err
		}
		if resp.StatusCode != http.StatusServiceUnavailable {
			return resp, nil
		}
		if time.Now().After(deadline) {
			return resp, nil
		}
		_, _ = io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(5 * time.Second):
		}
	}
}

var (
	_ resource.Resource                = &deploymentResource{}
	_ resource.ResourceWithImportState = &deploymentResource{}
)

func NewDeploymentResource() resource.Resource {
	return &deploymentResource{}
}

type deploymentResource struct {
	client *client.CosmosClient
}

type deploymentModel struct {
	Name     types.String `tfsdk:"name"`
	Replicas types.Int64  `tfsdk:"replicas"`
	Strategy types.String `tfsdk:"strategy"`
	Tags     types.Set    `tfsdk:"tags"`
	Storage  types.Set    `tfsdk:"storage"`
	Compose  types.String `tfsdk:"compose"`
}

func (r *deploymentResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_deployment"
}

func (r *deploymentResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Cosmos cluster deployment. Deployments are replicated across the constellation, so this resource can target any node in the cluster.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Description: "Unique deployment name (3-64 alphanumeric chars). Used as the resource ID.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"replicas": schema.Int64Attribute{
				Description: "Target number of container replicas across eligible nodes.",
				Required:    true,
			},
			"strategy": schema.StringAttribute{
				Description: "Placement strategy: \"round-robin\" (default) or \"least-busy\".",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("round-robin"),
			},
			"tags": schema.SetAttribute{
				Description: "Node tags required for placement. All tags must be present on a node (AND'd). Empty means any node is eligible.",
				Optional:    true,
				ElementType: types.StringType,
			},
			"storage": schema.SetAttribute{
				Description: "RCLONE remote names required by this deployment. ${storage.NAME} in compose fields resolves to the mount path.",
				Optional:    true,
				ElementType: types.StringType,
			},
			"compose": schema.StringAttribute{
				Description: "JSON-encoded docker.DockerServiceCreateRequest (services, volumes, networks). Use jsonencode() in HCL.",
				Required:    true,
			},
		},
	}
}

func (r *deploymentResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *deploymentResource) buildBody(ctx context.Context, m *deploymentModel) (*cosmossdk.ProDeployment, error) {
	var compose cosmossdk.DockerDockerServiceCreateRequest
	composeStr := m.Compose.ValueString()
	if composeStr == "" {
		return nil, fmt.Errorf("compose must be a non-empty JSON object")
	}
	if err := json.Unmarshal([]byte(composeStr), &compose); err != nil {
		return nil, fmt.Errorf("parsing compose JSON: %w", err)
	}

	body := &cosmossdk.ProDeployment{
		Name:     m.Name.ValueString(),
		Replicas: int(m.Replicas.ValueInt64()),
		Compose:  compose,
	}

	if !m.Strategy.IsNull() && !m.Strategy.IsUnknown() && m.Strategy.ValueString() != "" {
		strat := cosmossdk.ProDeploymentStrategy(m.Strategy.ValueString())
		body.Strategy = &strat
	}

	if !m.Tags.IsNull() && !m.Tags.IsUnknown() {
		var tags []string
		if d := m.Tags.ElementsAs(ctx, &tags, false); d.HasError() {
			return nil, fmt.Errorf("parsing tags: %s", d.Errors())
		}
		if len(tags) > 0 {
			body.Tags = &tags
		}
	}

	if !m.Storage.IsNull() && !m.Storage.IsUnknown() {
		var storage []string
		if d := m.Storage.ElementsAs(ctx, &storage, false); d.HasError() {
			return nil, fmt.Errorf("parsing storage: %s", d.Errors())
		}
		if len(storage) > 0 {
			body.Storage = &storage
		}
	}

	return body, nil
}

func (r *deploymentResource) populateState(ctx context.Context, m *deploymentModel, dep *cosmossdk.ProDeployment) error {
	m.Name = types.StringValue(dep.Name)
	m.Replicas = types.Int64Value(int64(dep.Replicas))

	if dep.Strategy != nil && *dep.Strategy != "" {
		m.Strategy = types.StringValue(string(*dep.Strategy))
	} else {
		m.Strategy = types.StringValue("round-robin")
	}

	if dep.Tags != nil && len(*dep.Tags) > 0 {
		setVal, d := types.SetValueFrom(ctx, types.StringType, *dep.Tags)
		if d.HasError() {
			return fmt.Errorf("setting tags: %s", d.Errors())
		}
		m.Tags = setVal
	} else {
		m.Tags = types.SetNull(types.StringType)
	}

	if dep.Storage != nil && len(*dep.Storage) > 0 {
		setVal, d := types.SetValueFrom(ctx, types.StringType, *dep.Storage)
		if d.HasError() {
			return fmt.Errorf("setting storage: %s", d.Errors())
		}
		m.Storage = setVal
	} else {
		m.Storage = types.SetNull(types.StringType)
	}

	// Compose is intentionally not refreshed from the server. The server
	// roundtrips it through docker.DockerServiceCreateRequest (zero-value
	// fields without omitempty re-emit on response) and further mutates it
	// at deploy time, so the response can't byte-match what the user wrote.
	// Terraform's consistent-result check compares strings, so we keep the
	// user's HCL-authored value as the source of truth in state. Trade-off:
	// out-of-band edits to compose won't be detected by `terraform plan`.

	return nil
}

func (r *deploymentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan deploymentModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body, err := r.buildBody(ctx, &plan)
	if err != nil {
		resp.Diagnostics.AddError("Invalid deployment configuration", err.Error())
		return
	}

	httpResp, err := retryOn503(ctx, func() (*http.Response, error) {
		return r.client.Raw.PostApiConstellationDeployments(ctx, *body)
	})
	if err != nil {
		resp.Diagnostics.AddError("Error creating deployment", err.Error())
		return
	}

	dep, err := client.ParseResponse[cosmossdk.ProDeployment](httpResp)
	if err != nil {
		resp.Diagnostics.AddError("Error parsing deployment create response", err.Error())
		return
	}
	if dep == nil {
		resp.Diagnostics.AddError("Empty deployment create response", "API returned no deployment data")
		return
	}

	if err := r.populateState(ctx, &plan, dep); err != nil {
		resp.Diagnostics.AddError("Error mapping deployment to state", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *deploymentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state deploymentModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := state.Name.ValueString()
	httpResp, err := retryOn503(ctx, func() (*http.Response, error) {
		return r.client.Raw.GetApiConstellationDeploymentsName(ctx, name)
	})
	if err != nil {
		resp.Diagnostics.AddError("Error reading deployment", err.Error())
		return
	}

	dep, err := client.ParseResponse[cosmossdk.ProDeployment](httpResp)
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error parsing deployment response", err.Error())
		return
	}
	if dep == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	if err := r.populateState(ctx, &state, dep); err != nil {
		resp.Diagnostics.AddError("Error mapping deployment to state", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *deploymentResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan deploymentModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body, err := r.buildBody(ctx, &plan)
	if err != nil {
		resp.Diagnostics.AddError("Invalid deployment configuration", err.Error())
		return
	}

	name := plan.Name.ValueString()
	httpResp, err := retryOn503(ctx, func() (*http.Response, error) {
		return r.client.Raw.PutApiConstellationDeploymentsName(ctx, name, *body)
	})
	if err != nil {
		resp.Diagnostics.AddError("Error updating deployment", err.Error())
		return
	}

	dep, err := client.ParseResponse[cosmossdk.ProDeployment](httpResp)
	if err != nil {
		resp.Diagnostics.AddError("Error parsing deployment update response", err.Error())
		return
	}
	if dep == nil {
		resp.Diagnostics.AddError("Empty deployment update response", "API returned no deployment data")
		return
	}

	if err := r.populateState(ctx, &plan, dep); err != nil {
		resp.Diagnostics.AddError("Error mapping deployment to state", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *deploymentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state deploymentModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := state.Name.ValueString()
	httpResp, err := retryOn503(ctx, func() (*http.Response, error) {
		return r.client.Raw.DeleteApiConstellationDeploymentsName(ctx, name)
	})
	if err != nil {
		resp.Diagnostics.AddError("Error deleting deployment", err.Error())
		return
	}
	if err := client.CheckResponse(httpResp); err != nil {
		if client.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Error deleting deployment", err.Error())
		return
	}
}

func (r *deploymentResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}
