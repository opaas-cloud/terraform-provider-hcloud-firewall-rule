package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hetznercloud/hcloud-go/v2/hcloud"
	"net"
	"slices"
	"terraform-provider-hcloud-firewall-rule/tools"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource              = &firewallRuleResource{}
	_ resource.ResourceWithConfigure = &firewallRuleResource{}
)

// NewRepoResource is a helper function to simplify the provider implementation.
func NewFirewallRuleResource() resource.Resource {
	return &firewallRuleResource{}
}

// repoResource is the resource implementation.
type firewallRuleResource struct {
	client *tools.FirewallRule
}

// Configure adds the provider configured client to the resource.
func (r *firewallRuleResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {

	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*tools.FirewallRule)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *hashicups.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}

// Metadata returns the resource type name.
func (r *firewallRuleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_rule"
}

// Schema defines the schema for the resource.
func (r *firewallRuleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"firewall_name": schema.StringAttribute{
				Required: true,
			},
			"token": schema.StringAttribute{
				Required: true,
			},
			"source_ip": schema.StringAttribute{
				Required: true,
			},
			"port": schema.StringAttribute{
				Required: true,
			},
			"description": schema.StringAttribute{
				Required: true,
			},
		},
	}
}

// Create a new resource.
func (r *firewallRuleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan tools.FirewallRuleModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	client := hcloud.NewClient(hcloud.WithToken(plan.Token.ValueString()))
	firewall, _, _ := client.Firewall.GetByName(ctx, plan.Name.ValueString())

	rules := firewall.Rules

	sourceIp, _, _ := net.ParseCIDR(plan.SourceIP.ValueString())

	rule := hcloud.FirewallRule{
		Direction: hcloud.FirewallRuleDirectionIn,
		Protocol:  hcloud.FirewallRuleProtocolTCP,
		SourceIPs: []net.IPNet{
			{
				IP:   sourceIp,
				Mask: net.IPv4Mask(255, 255, 255, 255),
			},
		},
		Port:           hcloud.Ptr(plan.Port.ValueString()),
		DestinationIPs: []net.IPNet{},
		Description:    hcloud.Ptr(plan.Description.ValueString()),
	}
	rules = append(rules, rule)

	opts := hcloud.FirewallSetRulesOpts{
		Rules: rules,
	}

	_, _, _ = client.Firewall.SetRules(ctx, firewall, opts)

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *firewallRuleResource) Read(_ context.Context, _ resource.ReadRequest, _ *resource.ReadResponse) {
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *firewallRuleResource) Update(_ context.Context, _ resource.UpdateRequest, _ *resource.UpdateResponse) {
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *firewallRuleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state tools.FirewallRuleModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)

	client := hcloud.NewClient(hcloud.WithToken(state.Token.ValueString()))
	firewall, _, _ := client.Firewall.GetByName(ctx, state.Name.ValueString())

	test := mapModels(firewall.Rules, func(i hcloud.FirewallRule) string {
		return *i.Description
	})

	newRules := slices.Delete(firewall.Rules, slices.Index(test, state.Description.ValueString()), slices.Index(test, state.Description.ValueString())+1)
	opts := hcloud.FirewallSetRulesOpts{
		Rules: newRules,
	}

	_, _, _ = client.Firewall.SetRules(ctx, firewall, opts)

	if resp.Diagnostics.HasError() {
		return
	}
}

func mapModels(data []hcloud.FirewallRule, f func(model hcloud.FirewallRule) string) []string {

	mapped := make([]string, len(data))

	for i, e := range data {
		mapped[i] = f(e)
	}

	return mapped
}
