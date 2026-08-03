package provider

import (
	"context"
	"os"

	"github.com/azukaar/terraform-provider-cosmos/internal/client"
	"github.com/azukaar/terraform-provider-cosmos/internal/datasources"
	"github.com/azukaar/terraform-provider-cosmos/internal/resources"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ provider.Provider = &CosmosProvider{}

type CosmosProvider struct {
	version string
}

type CosmosProviderModel struct {
	BaseURL  types.String `tfsdk:"base_url"`
	Token    types.String `tfsdk:"token"`
	Insecure types.Bool   `tfsdk:"insecure"`
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &CosmosProvider{version: version}
	}
}

func (p *CosmosProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "cosmos"
	resp.Version = p.version
}

func (p *CosmosProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Terraform provider for managing Cosmos Server resources.",
		Attributes: map[string]schema.Attribute{
			"base_url": schema.StringAttribute{
				Description: "The base URL of the Cosmos Server (e.g. https://cosmos.example.com). Can also be set with COSMOS_BASE_URL env var.",
				Optional:    true,
			},
			"token": schema.StringAttribute{
				Description: "API token for authenticating with Cosmos Server. Can also be set with COSMOS_TOKEN env var.",
				Optional:    true,
				Sensitive:   true,
			},
			"insecure": schema.BoolAttribute{
				Description: "Skip TLS certificate verification. Can also be set with COSMOS_INSECURE env var. Default: false.",
				Optional:    true,
			},
		},
	}
}

func (p *CosmosProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config CosmosProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// If any field references something not yet known (e.g. a public IP from
	// an aws_instance still being created), defer Configure. Terraform will
	// retry once the dependency resolves during apply.
	if config.BaseURL.IsUnknown() || config.Token.IsUnknown() || config.Insecure.IsUnknown() {
		return
	}

	baseURL := os.Getenv("COSMOS_BASE_URL")
	if !config.BaseURL.IsNull() {
		baseURL = config.BaseURL.ValueString()
	}

	token := os.Getenv("COSMOS_TOKEN")
	if !config.Token.IsNull() {
		token = config.Token.ValueString()
	}
	// token may be empty: cosmos_install runs against an unconfigured server
	// where /api/setup is unauthenticated. Resources that require auth will
	// surface a 401 from the server with a clear error.

	insecure := os.Getenv("COSMOS_INSECURE") == "true"
	if !config.Insecure.IsNull() {
		insecure = config.Insecure.ValueBool()
	}

	// base_url may be empty: a default `provider "cosmos" {}` block can exist
	// purely to satisfy required_providers when only aliased providers are
	// actually used (or when cosmos_remote_install — which never touches the
	// Cosmos API — is the only resource on the default provider). Build the
	// client anyway; resources that try to make a real request without a
	// configured base_url will get a clear connection error.
	cosmosClient, err := client.NewCosmosClient(baseURL, token, insecure)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create Cosmos client", err.Error())
		return
	}

	resp.DataSourceData = cosmosClient
	resp.ResourceData = cosmosClient
}

func (p *CosmosProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		resources.NewRouteResource,
		resources.NewUserResource,
		resources.NewAPITokenResource,
		resources.NewOpenIDClientResource,
		resources.NewConstellationResource,
		resources.NewConstellationDNSResource,
		resources.NewCronJobResource,
		resources.NewAlertResource,
		resources.NewConstellationDeviceResource,
		resources.NewBackupResource,
		resources.NewDockerNetworkResource,
		resources.NewDockerVolumeResource,
		resources.NewDockerServiceResource,
		resources.NewSnapRAIDResource,
		resources.NewStorageMountResource,
		resources.NewGroupResource,
		resources.NewDeploymentResource,
		resources.NewInstallResource,
		resources.NewRemoteInstallResource,
	}
}

func (p *CosmosProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		datasources.NewStatusDataSource,
		datasources.NewContainersDataSource,
	}
}
