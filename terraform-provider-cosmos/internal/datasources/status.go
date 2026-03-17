package datasources

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/azukaar/terraform-provider-cosmos/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &statusDataSource{}

func NewStatusDataSource() datasource.DataSource {
	return &statusDataSource{}
}

type statusDataSource struct {
	client *client.CosmosClient
}

// ---------- Terraform model ----------

type statusDataSourceModel struct {
	Hostname            types.String `tfsdk:"hostname"`
	Version             types.String `tfsdk:"version"`
	NeedsUpdate         types.Bool   `tfsdk:"needs_update"`
	DockerStatus        types.String `tfsdk:"docker_status"`
	ConstellationStatus types.String `tfsdk:"constellation_status"`
}

// ---------- datasource.DataSource ----------

func (d *statusDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_status"
}

func (d *statusDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads the current status of the Cosmos Server instance.",
		Attributes: map[string]schema.Attribute{
			"hostname": schema.StringAttribute{
				Description: "The hostname of the Cosmos Server.",
				Computed:    true,
			},
			"version": schema.StringAttribute{
				Description: "The current version of Cosmos Server.",
				Computed:    true,
			},
			"needs_update": schema.BoolAttribute{
				Description: "Whether an update is available for Cosmos Server.",
				Computed:    true,
			},
			"docker_status": schema.StringAttribute{
				Description: "The status of the Docker daemon.",
				Computed:    true,
			},
			"constellation_status": schema.StringAttribute{
				Description: "The status of the Constellation VPN.",
				Computed:    true,
			},
		},
	}
}

func (d *statusDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.CosmosClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected DataSource Configure Type",
			fmt.Sprintf("Expected *client.CosmosClient, got: %T", req.ProviderData),
		)
		return
	}
	d.client = c
}

func (d *statusDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	httpResp, err := d.client.Raw.GetApiStatus(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error reading Cosmos status", err.Error())
		return
	}

	rawData, err := client.ParseRawResponse(httpResp)
	if err != nil {
		resp.Diagnostics.AddError("Error reading Cosmos status", err.Error())
		return
	}

	// Parse the status response as a generic map since the structure may vary.
	var status map[string]interface{}
	if rawData != nil {
		if err := json.Unmarshal(rawData, &status); err != nil {
			resp.Diagnostics.AddError("Error parsing Cosmos status response", err.Error())
			return
		}
	}

	var model statusDataSourceModel

	model.Hostname = types.StringValue(extractString(status, "hostname"))
	model.Version = types.StringValue(extractString(status, "version"))
	model.NeedsUpdate = types.BoolValue(extractBool(status, "needsUpdate"))
	model.DockerStatus = types.StringValue(extractString(status, "dockerStatus"))
	model.ConstellationStatus = types.StringValue(extractString(status, "constellationStatus"))

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

// ---------- helpers ----------

// extractString safely extracts a string value from a map, returning "" if not found.
func extractString(m map[string]interface{}, key string) string {
	if m == nil {
		return ""
	}
	if v, ok := m[key]; ok {
		switch val := v.(type) {
		case string:
			return val
		case float64:
			return fmt.Sprintf("%v", val)
		default:
			// Try JSON marshalling for complex types.
			b, err := json.Marshal(val)
			if err != nil {
				return fmt.Sprintf("%v", val)
			}
			return string(b)
		}
	}
	return ""
}

// extractBool safely extracts a boolean value from a map, returning false if not found.
func extractBool(m map[string]interface{}, key string) bool {
	if m == nil {
		return false
	}
	if v, ok := m[key]; ok {
		if b, ok := v.(bool); ok {
			return b
		}
	}
	return false
}
