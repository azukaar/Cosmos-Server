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

var _ datasource.DataSource = &containersDataSource{}

func NewContainersDataSource() datasource.DataSource {
	return &containersDataSource{}
}

type containersDataSource struct {
	client *client.CosmosClient
}

// ---------- Terraform model ----------

type containersDataSourceModel struct {
	Containers []containerModel `tfsdk:"containers"`
}

type containerModel struct {
	ID    types.String `tfsdk:"id"`
	Name  types.String `tfsdk:"name"`
	Image types.String `tfsdk:"image"`
	State types.String `tfsdk:"state"`
}

// ---------- datasource.DataSource ----------

func (d *containersDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_containers"
}

func (d *containersDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Lists all Docker containers managed by Cosmos Server.",
		Attributes: map[string]schema.Attribute{
			"containers": schema.ListNestedAttribute{
				Description: "List of containers.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "The container ID.",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "The container name.",
							Computed:    true,
						},
						"image": schema.StringAttribute{
							Description: "The container image.",
							Computed:    true,
						},
						"state": schema.StringAttribute{
							Description: "The current state of the container (e.g. running, stopped).",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (d *containersDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *containersDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	httpResp, err := d.client.Raw.GetApiServapps(ctx, nil)
	if err != nil {
		resp.Diagnostics.AddError("Error reading containers", err.Error())
		return
	}

	rawData, err := client.ParseRawResponse(httpResp)
	if err != nil {
		resp.Diagnostics.AddError("Error reading containers", err.Error())
		return
	}

	var model containersDataSourceModel

	if rawData == nil || string(rawData) == "null" {
		model.Containers = []containerModel{}
		resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
		return
	}

	var containers []map[string]interface{}
	if err := json.Unmarshal(rawData, &containers); err != nil {
		resp.Diagnostics.AddError("Error parsing containers response", err.Error())
		return
	}

	model.Containers = make([]containerModel, 0, len(containers))
	for _, c := range containers {
		cm := containerModel{
			ID:    types.StringValue(containerFieldString(c, "Id", "id", "ID")),
			Name:  types.StringValue(containerFieldString(c, "Name", "name", "Names")),
			Image: types.StringValue(containerFieldString(c, "Image", "image")),
			State: types.StringValue(containerFieldString(c, "State", "state", "Status")),
		}
		model.Containers = append(model.Containers, cm)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

// ---------- helpers ----------

// containerFieldString extracts a string field from a container map, trying multiple key variants.
func containerFieldString(m map[string]interface{}, keys ...string) string {
	for _, key := range keys {
		if v, ok := m[key]; ok {
			switch val := v.(type) {
			case string:
				return val
			case []interface{}:
				// Docker sometimes returns Names as an array like ["/container_name"].
				if len(val) > 0 {
					if s, ok := val[0].(string); ok {
						// Strip leading slash if present.
						if len(s) > 0 && s[0] == '/' {
							return s[1:]
						}
						return s
					}
				}
			default:
				return fmt.Sprintf("%v", val)
			}
		}
	}
	return ""
}
