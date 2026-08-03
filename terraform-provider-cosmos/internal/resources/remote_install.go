package resources

import (
	"context"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"golang.org/x/crypto/ssh"
)

const (
	defaultGetScript    = "https://cosmos-cloud.io/get.sh"
	defaultGetProScript = "https://cosmos-cloud.io/get-pro.sh"
)

var (
	_ resource.Resource = &remoteInstallResource{}
)

func NewRemoteInstallResource() resource.Resource {
	return &remoteInstallResource{}
}

type remoteInstallResource struct{}

type remoteInstallModel struct {
	Host                  types.String `tfsdk:"host"`
	Port                  types.Int64  `tfsdk:"port"`
	User                  types.String `tfsdk:"user"`
	PrivateKey            types.String `tfsdk:"private_key"`
	PrivateKeyPath        types.String `tfsdk:"private_key_path"`
	Password              types.String `tfsdk:"password"`
	Pro                   types.Bool   `tfsdk:"pro"`
	ScriptURL             types.String `tfsdk:"script_url"`
	HostKeyCheck          types.Bool   `tfsdk:"host_key_check"`
	HostKey               types.String `tfsdk:"host_key"`
	ConnectTimeoutSeconds types.Int64  `tfsdk:"connect_timeout_seconds"`
	CommandTimeoutSeconds types.Int64  `tfsdk:"command_timeout_seconds"`

	Installed        types.Bool   `tfsdk:"installed"`
	ScriptExitCode   types.Int64  `tfsdk:"script_exit_code"`
	ScriptOutputTail types.String `tfsdk:"script_output_tail"`
}

func (r *remoteInstallResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_remote_install"
}

func (r *remoteInstallResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	replaceStr := []planmodifier.String{stringplanmodifier.RequiresReplace()}

	resp.Schema = schema.Schema{
		Description: "One-shot resource that SSHes into a freshly provisioned VM and runs the public Cosmos installer (get.sh or get-pro.sh). Read is a no-op; Delete does not auto-uninstall.",
		Attributes: map[string]schema.Attribute{
			"host": schema.StringAttribute{
				Description:   "Target host (IP or DNS).",
				Required:      true,
				PlanModifiers: replaceStr,
			},
			"port": schema.Int64Attribute{
				Description: "SSH port. Defaults to 22.",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(22),
			},
			"user": schema.StringAttribute{
				Description:   "SSH username.",
				Required:      true,
				PlanModifiers: replaceStr,
			},
			"private_key": schema.StringAttribute{
				Description: "PEM-encoded SSH private key. Mutually exclusive with private_key_path and password.",
				Optional:    true,
				Sensitive:   true,
			},
			"private_key_path": schema.StringAttribute{
				Description: "Path to a PEM-encoded SSH private key on disk. Mutually exclusive with private_key and password.",
				Optional:    true,
			},
			"password": schema.StringAttribute{
				Description: "SSH password. Discouraged — prefer key auth.",
				Optional:    true,
				Sensitive:   true,
			},
			"pro": schema.BoolAttribute{
				Description: "When true, defaults script_url to https://cosmos-cloud.io/get-pro.sh. Ignored if script_url is set explicitly.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"script_url": schema.StringAttribute{
				Description: "Override installer URL. Defaults to get.sh (or get-pro.sh when pro=true).",
				Optional:    true,
				Computed:    true,
			},
			"host_key_check": schema.BoolAttribute{
				Description: "Whether to verify the SSH host key. Default true. When false, host key is ignored (insecure — not recommended for production).",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"host_key": schema.StringAttribute{
				Description: "Expected SSH host public key in authorized_keys format (e.g. \"ssh-ed25519 AAAA...\"). When set, takes precedence over host_key_check.",
				Optional:    true,
			},
			"connect_timeout_seconds": schema.Int64Attribute{
				Description: "Timeout for the initial SSH dial.",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(60),
			},
			"command_timeout_seconds": schema.Int64Attribute{
				Description: "Timeout for the installer command.",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(600),
			},
			"installed": schema.BoolAttribute{
				Description: "True once the installer script ran with exit code 0.",
				Computed:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"script_exit_code": schema.Int64Attribute{
				Description: "Exit code reported by the installer script.",
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"script_output_tail": schema.StringAttribute{
				Description: "Last 4 KiB of the installer's combined stdout+stderr.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *remoteInstallResource) Configure(_ context.Context, _ resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	// No client needed.
}

func (r *remoteInstallResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan remoteInstallModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	scriptURL := plan.ScriptURL.ValueString()
	if scriptURL == "" {
		if plan.Pro.ValueBool() {
			scriptURL = defaultGetProScript
		} else {
			scriptURL = defaultGetScript
		}
	}

	cfg, err := buildSSHConfig(&plan)
	if err != nil {
		resp.Diagnostics.AddError("Invalid SSH configuration", err.Error())
		return
	}

	if !plan.HostKeyCheck.ValueBool() && plan.HostKey.ValueString() == "" {
		resp.Diagnostics.AddWarning(
			"SSH host key check disabled",
			"host_key_check=false skips host key verification. Set host_key for production.",
		)
	}

	addr := net.JoinHostPort(plan.Host.ValueString(), strconv.FormatInt(plan.Port.ValueInt64(), 10))
	connectTimeout := time.Duration(plan.ConnectTimeoutSeconds.ValueInt64()) * time.Second

	dialer := &net.Dialer{Timeout: connectTimeout}
	conn, err := dialer.DialContext(ctx, "tcp", addr)
	if err != nil {
		resp.Diagnostics.AddError("SSH dial failed", fmt.Sprintf("connecting to %s: %s", addr, err))
		return
	}

	sshConn, chans, reqs, err := ssh.NewClientConn(conn, addr, cfg)
	if err != nil {
		conn.Close()
		resp.Diagnostics.AddError("SSH handshake failed", err.Error())
		return
	}
	sshClient := ssh.NewClient(sshConn, chans, reqs)
	defer sshClient.Close()

	session, err := sshClient.NewSession()
	if err != nil {
		resp.Diagnostics.AddError("SSH session failed", err.Error())
		return
	}
	defer session.Close()

	// get-pro.sh uses bashisms ([[, process substitution, etc.) so we must
	// pipe to bash (not sh — which is dash on Debian/Ubuntu). Wrap the whole
	// pipe in `bash -c` so set -o pipefail applies to the curl|bash pipeline.
	//
	// Freshly provisioned Debian/Ubuntu VMs run cloud-init plus apt-daily /
	// unattended-upgrades in the background. Those hold /var/lib/dpkg/lock-*
	// and race the installer's apt-get calls. Wait for cloud-init to drain,
	// stop the apt-daily timers, then poll until no process holds an apt or
	// dpkg lock before kicking off the installer.
	innerCmd := fmt.Sprintf(`set -o pipefail
if [ "$(id -u)" -eq 0 ]; then SUDO=""; else SUDO="sudo"; fi
if command -v cloud-init >/dev/null 2>&1; then
  cloud-init status --wait >/dev/null 2>&1 || true
fi
$SUDO systemctl stop apt-daily.timer apt-daily-upgrade.timer 2>/dev/null || true
$SUDO systemctl stop apt-daily.service apt-daily-upgrade.service unattended-upgrades.service 2>/dev/null || true
n=0
while $SUDO fuser /var/lib/dpkg/lock-frontend /var/lib/apt/lists/lock /var/lib/dpkg/lock >/dev/null 2>&1; do
  n=$((n+1))
  if [ "$n" -gt 120 ]; then
    echo "Timed out waiting for apt/dpkg locks to clear (10 min)" >&2
    exit 1
  fi
  sleep 5
done
curl -fsSL %s | bash`, shellQuote(scriptURL))
	command := fmt.Sprintf("bash -c %s", shellQuote(innerCmd))

	commandTimeout := time.Duration(plan.CommandTimeoutSeconds.ValueInt64()) * time.Second
	cmdCtx, cancel := context.WithTimeout(ctx, commandTimeout)
	defer cancel()

	type runResult struct {
		out []byte
		err error
	}
	resultCh := make(chan runResult, 1)
	go func() {
		out, err := session.CombinedOutput(command)
		resultCh <- runResult{out: out, err: err}
	}()

	var output []byte
	var runErr error
	select {
	case res := <-resultCh:
		output = res.out
		runErr = res.err
	case <-cmdCtx.Done():
		_ = session.Signal(ssh.SIGKILL)
		resp.Diagnostics.AddError("Installer timed out", cmdCtx.Err().Error())
		return
	}

	tail := tailString(string(output), 4096)
	exitCode := 0
	if runErr != nil {
		if exitErr, ok := runErr.(*ssh.ExitError); ok {
			exitCode = exitErr.ExitStatus()
		} else {
			resp.Diagnostics.AddError(
				"Installer command failed",
				fmt.Sprintf("%s\n\n--- output tail ---\n%s", runErr, tail),
			)
			return
		}
	}

	plan.ScriptURL = types.StringValue(scriptURL)
	plan.ScriptExitCode = types.Int64Value(int64(exitCode))
	plan.ScriptOutputTail = types.StringValue(tail)
	plan.Installed = types.BoolValue(exitCode == 0)

	if exitCode != 0 {
		resp.Diagnostics.AddError(
			"Installer exited non-zero",
			fmt.Sprintf("exit code %d\n\n--- output tail ---\n%s", exitCode, tail),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *remoteInstallResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// One-shot side effect; nothing to refresh.
	var state remoteInstallModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *remoteInstallResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Update not supported",
		"All cosmos_remote_install input fields require replacement. This is a provider bug if reached.",
	)
}

func (r *remoteInstallResource) Delete(_ context.Context, _ resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.Diagnostics.AddWarning(
		"cosmos_remote_install does not auto-uninstall",
		"Removing this resource only drops it from state. Cosmos remains installed on the remote host.",
	)
}

func buildSSHConfig(plan *remoteInstallModel) (*ssh.ClientConfig, error) {
	auths := []ssh.AuthMethod{}

	if !plan.PrivateKey.IsNull() && !plan.PrivateKey.IsUnknown() && plan.PrivateKey.ValueString() != "" {
		signer, err := ssh.ParsePrivateKey([]byte(plan.PrivateKey.ValueString()))
		if err != nil {
			return nil, fmt.Errorf("parsing private_key: %w", err)
		}
		auths = append(auths, ssh.PublicKeys(signer))
	}

	if !plan.PrivateKeyPath.IsNull() && !plan.PrivateKeyPath.IsUnknown() && plan.PrivateKeyPath.ValueString() != "" {
		data, err := os.ReadFile(plan.PrivateKeyPath.ValueString())
		if err != nil {
			return nil, fmt.Errorf("reading private_key_path: %w", err)
		}
		signer, err := ssh.ParsePrivateKey(data)
		if err != nil {
			return nil, fmt.Errorf("parsing private_key_path: %w", err)
		}
		auths = append(auths, ssh.PublicKeys(signer))
	}

	if !plan.Password.IsNull() && !plan.Password.IsUnknown() && plan.Password.ValueString() != "" {
		auths = append(auths, ssh.Password(plan.Password.ValueString()))
	}

	if len(auths) == 0 {
		return nil, fmt.Errorf("at least one of private_key, private_key_path, or password must be set")
	}

	hostKeyCallback, err := buildHostKeyCallback(plan)
	if err != nil {
		return nil, err
	}

	return &ssh.ClientConfig{
		User:            plan.User.ValueString(),
		Auth:            auths,
		HostKeyCallback: hostKeyCallback,
		Timeout:         time.Duration(plan.ConnectTimeoutSeconds.ValueInt64()) * time.Second,
	}, nil
}

func buildHostKeyCallback(plan *remoteInstallModel) (ssh.HostKeyCallback, error) {
	if !plan.HostKey.IsNull() && !plan.HostKey.IsUnknown() && plan.HostKey.ValueString() != "" {
		expected, _, _, _, err := ssh.ParseAuthorizedKey([]byte(plan.HostKey.ValueString()))
		if err != nil {
			return nil, fmt.Errorf("parsing host_key: %w", err)
		}
		return ssh.FixedHostKey(expected), nil
	}
	if plan.HostKeyCheck.ValueBool() {
		return nil, fmt.Errorf("host_key_check=true requires host_key to be set")
	}
	return ssh.InsecureIgnoreHostKey(), nil
}

func tailString(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return "..." + s[len(s)-max:]
}

func shellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}
