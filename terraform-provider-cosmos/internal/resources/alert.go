package resources

import (
	"context"
	"fmt"

	cosmossdk "github.com/azukaar/cosmos-server/go-sdk"
	"github.com/azukaar/terraform-provider-cosmos/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

var (
	_ resource.Resource                = &alertResource{}
	_ resource.ResourceWithImportState = &alertResource{}
)

// NewAlertResource returns a new alert resource instance.
func NewAlertResource() resource.Resource {
	return &alertResource{}
}

type alertResource struct {
	client *client.CosmosClient
}

// ---------- Terraform model types ----------

type alertResourceModel struct {
	Name           types.String `tfsdk:"name"`
	Enabled        types.Bool   `tfsdk:"enabled"`
	TrackingMetric types.String `tfsdk:"tracking_metric"`
	Period         types.String `tfsdk:"period"`
	Severity       types.String `tfsdk:"severity"`
	Condition      types.Object `tfsdk:"condition"`
	Actions        types.List   `tfsdk:"actions"`
	LastTriggered  types.String `tfsdk:"last_triggered"`
	Throttled      types.Bool   `tfsdk:"throttled"`
}

type alertConditionModel struct {
	Operator types.String `tfsdk:"operator"`
	Value    types.Int64  `tfsdk:"value"`
	Percent  types.Bool   `tfsdk:"percent"`
}

type alertActionModel struct {
	Type   types.String `tfsdk:"type"`
	Target types.String `tfsdk:"target"`
}

// Attribute type maps used when constructing framework object/list values.
var alertConditionAttrTypes = map[string]attr.Type{
	"operator": types.StringType,
	"value":    types.Int64Type,
	"percent":  types.BoolType,
}

var alertActionAttrTypes = map[string]attr.Type{
	"type":   types.StringType,
	"target": types.StringType,
}

// ---------- Resource interface ----------

func (r *alertResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_alert"
}

func (r *alertResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Cosmos Server monitoring alert.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Description: "The unique name of the alert.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"enabled": schema.BoolAttribute{
				Description: "Whether the alert is enabled.",
				Optional:    true,
			},
			"tracking_metric": schema.StringAttribute{
				Description: "The metric to track for this alert.",
				Optional:    true,
			},
			"period": schema.StringAttribute{
				Description: "The time period for alert evaluation.",
				Optional:    true,
			},
			"severity": schema.StringAttribute{
				Description: "The severity level of the alert.",
				Optional:    true,
			},
			"condition": schema.SingleNestedAttribute{
				Description: "The condition that triggers the alert.",
				Optional:    true,
				Attributes: map[string]schema.Attribute{
					"operator": schema.StringAttribute{
						Description: "The comparison operator (e.g. >, <, ==).",
						Optional:    true,
					},
					"value": schema.Int64Attribute{
						Description: "The threshold value.",
						Optional:    true,
					},
					"percent": schema.BoolAttribute{
						Description: "Whether the value is a percentage.",
						Optional:    true,
					},
				},
			},
			"actions": schema.ListNestedAttribute{
				Description: "Actions to perform when the alert triggers.",
				Optional:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"type": schema.StringAttribute{
							Description: "The action type.",
							Optional:    true,
						},
						"target": schema.StringAttribute{
							Description: "The action target.",
							Optional:    true,
						},
					},
				},
			},
			"last_triggered": schema.StringAttribute{
				Description: "The last time the alert was triggered.",
				Computed:    true,
			},
			"throttled": schema.BoolAttribute{
				Description: "Whether the alert is currently throttled.",
				Computed:    true,
			},
		},
	}
}

func (r *alertResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *alertResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan alertResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	alert := r.modelToSDK(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	httpResp, err := r.client.Raw.PostApiAlerts(ctx, alert)
	if err != nil {
		resp.Diagnostics.AddError("Error creating alert", err.Error())
		return
	}
	if err := client.CheckResponse(httpResp); err != nil {
		resp.Diagnostics.AddError("Error creating alert", err.Error())
		return
	}

	// Read back to populate computed fields.
	if !r.readAlert(ctx, plan.Name.ValueString(), &plan, &resp.Diagnostics) {
		resp.Diagnostics.AddError("Error reading alert", "Alert not found after creation")
		return
	}
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *alertResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state alertResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := state.Name.ValueString()
	httpResp, err := r.client.Raw.GetApiAlertsName(ctx, name, cosmossdk.UtilsAlert{Name: name})
	if err != nil {
		resp.Diagnostics.AddError("Error reading alert", err.Error())
		return
	}

	alert, err := client.ParseResponse[cosmossdk.UtilsAlert](httpResp)
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error parsing alert response", err.Error())
		return
	}
	if alert == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	r.sdkToModel(alert, &state)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *alertResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan alertResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	alert := r.modelToSDK(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	name := plan.Name.ValueString()
	httpResp, err := r.client.Raw.PutApiAlertsName(ctx, name, alert)
	if err != nil {
		resp.Diagnostics.AddError("Error updating alert", err.Error())
		return
	}
	if err := client.CheckResponse(httpResp); err != nil {
		resp.Diagnostics.AddError("Error updating alert", err.Error())
		return
	}

	if !r.readAlert(ctx, name, &plan, &resp.Diagnostics) {
		resp.Diagnostics.AddError("Error reading alert", "Alert not found after update")
		return
	}
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *alertResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state alertResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := state.Name.ValueString()
	httpResp, err := r.client.Raw.DeleteApiAlertsName(ctx, name, cosmossdk.UtilsAlert{Name: name})
	if err != nil {
		resp.Diagnostics.AddError("Error deleting alert", err.Error())
		return
	}
	if err := client.CheckResponse(httpResp); err != nil {
		resp.Diagnostics.AddError("Error deleting alert", err.Error())
		return
	}
}

func (r *alertResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}

// ---------- helpers ----------

// readAlert fetches the alert from the API and updates the Terraform model.
// It returns false if the alert was not found (404), in which case no diagnostic
// error is added — the caller should remove the resource from state.
func (r *alertResource) readAlert(ctx context.Context, name string, model *alertResourceModel, diags *diag.Diagnostics) bool {
	httpResp, err := r.client.Raw.GetApiAlertsName(ctx, name, cosmossdk.UtilsAlert{Name: name})
	if err != nil {
		diags.AddError("Error reading alert", err.Error())
		return true
	}

	alert, err := client.ParseResponse[cosmossdk.UtilsAlert](httpResp)
	if err != nil {
		if client.IsNotFound(err) {
			return false
		}
		diags.AddError("Error parsing alert response", err.Error())
		return true
	}
	if alert == nil {
		return false
	}

	r.sdkToModel(alert, model)
	return true
}

// modelToSDK converts the Terraform resource model into an SDK UtilsAlert.
func (r *alertResource) modelToSDK(ctx context.Context, model *alertResourceModel, diags *diag.Diagnostics) cosmossdk.UtilsAlert {
	alert := cosmossdk.UtilsAlert{
		Name: model.Name.ValueString(),
	}

	if !model.Enabled.IsNull() && !model.Enabled.IsUnknown() {
		alert.Enabled = client.BoolPtr(model.Enabled.ValueBool())
	}
	if !model.TrackingMetric.IsNull() && !model.TrackingMetric.IsUnknown() {
		alert.TrackingMetric = client.StringPtr(model.TrackingMetric.ValueString())
	}
	if !model.Period.IsNull() && !model.Period.IsUnknown() {
		alert.Period = client.StringPtr(model.Period.ValueString())
	}
	if !model.Severity.IsNull() && !model.Severity.IsUnknown() {
		alert.Severity = client.StringPtr(model.Severity.ValueString())
	}

	// Condition
	if !model.Condition.IsNull() && !model.Condition.IsUnknown() {
		var condModel alertConditionModel
		d := model.Condition.As(ctx, &condModel, basetypes.ObjectAsOptions{})
		diags.Append(d...)
		if d.HasError() {
			return alert
		}
		cond := cosmossdk.UtilsAlertCondition{}
		if !condModel.Operator.IsNull() && !condModel.Operator.IsUnknown() {
			cond.Operator = client.StringPtr(condModel.Operator.ValueString())
		}
		if !condModel.Value.IsNull() && !condModel.Value.IsUnknown() {
			v := int(condModel.Value.ValueInt64())
			cond.Value = client.IntPtr(v)
		}
		if !condModel.Percent.IsNull() && !condModel.Percent.IsUnknown() {
			cond.Percent = client.BoolPtr(condModel.Percent.ValueBool())
		}
		alert.Condition = &cond
	}

	// Actions
	if !model.Actions.IsNull() && !model.Actions.IsUnknown() {
		var actionModels []alertActionModel
		d := model.Actions.ElementsAs(ctx, &actionModels, false)
		diags.Append(d...)
		if d.HasError() {
			return alert
		}
		actions := make([]cosmossdk.UtilsAlertAction, len(actionModels))
		for i, am := range actionModels {
			a := cosmossdk.UtilsAlertAction{}
			if !am.Type.IsNull() && !am.Type.IsUnknown() {
				a.Type = client.StringPtr(am.Type.ValueString())
			}
			if !am.Target.IsNull() && !am.Target.IsUnknown() {
				a.Target = client.StringPtr(am.Target.ValueString())
			}
			actions[i] = a
		}
		alert.Actions = &actions
	}

	return alert
}

// sdkToModel populates the Terraform resource model from an SDK UtilsAlert.
func (r *alertResource) sdkToModel(alert *cosmossdk.UtilsAlert, model *alertResourceModel) {
	model.Name = types.StringValue(alert.Name)

	if alert.Enabled != nil {
		model.Enabled = types.BoolValue(*alert.Enabled)
	}
	if alert.TrackingMetric != nil {
		model.TrackingMetric = types.StringValue(*alert.TrackingMetric)
	}
	if alert.Period != nil {
		model.Period = types.StringValue(*alert.Period)
	}
	if alert.Severity != nil {
		model.Severity = types.StringValue(*alert.Severity)
	}

	// Computed fields always get a value.
	if alert.LastTriggered != nil {
		model.LastTriggered = types.StringValue(*alert.LastTriggered)
	} else {
		model.LastTriggered = types.StringValue("")
	}
	if alert.Throttled != nil {
		model.Throttled = types.BoolValue(*alert.Throttled)
	} else {
		model.Throttled = types.BoolValue(false)
	}

	// Condition
	if alert.Condition != nil {
		c := alert.Condition
		condObj, _ := types.ObjectValue(alertConditionAttrTypes, map[string]attr.Value{
			"operator": alertStringOrNull(c.Operator),
			"value":    alertIntOrNull(c.Value),
			"percent":  alertBoolOrNull(c.Percent),
		})
		model.Condition = condObj
	} else {
		model.Condition = types.ObjectNull(alertConditionAttrTypes)
	}

	// Actions
	if alert.Actions != nil && len(*alert.Actions) > 0 {
		actionValues := make([]attr.Value, len(*alert.Actions))
		for i, a := range *alert.Actions {
			obj, _ := types.ObjectValue(alertActionAttrTypes, map[string]attr.Value{
				"type":   alertStringOrNull(a.Type),
				"target": alertStringOrNull(a.Target),
			})
			actionValues[i] = obj
		}
		actionList, _ := types.ListValue(types.ObjectType{AttrTypes: alertActionAttrTypes}, actionValues)
		model.Actions = actionList
	} else {
		model.Actions = types.ListNull(types.ObjectType{AttrTypes: alertActionAttrTypes})
	}
}

// Small helpers to convert SDK pointers into framework attr.Value.

func alertStringOrNull(s *string) attr.Value {
	if s == nil || *s == "" {
		return types.StringNull()
	}
	return types.StringValue(*s)
}

func alertIntOrNull(i *int) attr.Value {
	if i == nil {
		return types.Int64Null()
	}
	return types.Int64Value(int64(*i))
}

func alertBoolOrNull(b *bool) attr.Value {
	if b == nil || !*b {
		return types.BoolNull()
	}
	return types.BoolValue(*b)
}
