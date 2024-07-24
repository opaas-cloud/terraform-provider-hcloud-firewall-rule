package tools

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type FirewallRule struct {
	Name        string
	Token       string
	SourceIP    string
	Port        string
	Description string
}

type FirewallRuleModel struct {
	Name        types.String `tfsdk:"firewall_name"`
	Token       types.String `tfsdk:"token"`
	SourceIP    types.String `tfsdk:"source_ip"`
	Port        types.String `tfsdk:"port"`
	Description types.String `tfsdk:"description"`
}
