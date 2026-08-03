# Single-node install

End-to-end bootstrap of one Cosmos node from a freshly provisioned VM you already own.

## What it does

1. SSHes into the VM and runs `https://cosmos-cloud.io/get-pro.sh`.
2. Calls `/api/setup` to configure database, HTTPS, and the admin user; captures an admin API token.
3. Re-creates the provider with that token aliased as `cosmos.configured` and creates one `cosmos_deployment`.

Resources used: `cosmos_remote_install`, `cosmos_install`, `cosmos_deployment`.

## Prerequisites

- A reachable VM (any cloud or bare metal) with sshd running.
- The matching SSH private key on disk.
- A DNS name pointing at the VM's public IP, if you want Let's Encrypt — otherwise drop `https_certificate_mode = "LETSENCRYPT"` and use `SELFSIGNED`.
- Optional: a Cosmos Pro licence key.

## Run it

```bash
cd examples/install

terraform init
terraform apply \
  -var 'vm_host=203.0.113.10' \
  -var 'ssh_private_key_path=~/.ssh/id_ed25519' \
  -var 'hostname=cosmos.example.com' \
  -var 'admin_password=…' \
  -var 'cosmos_licence=…'
```

The admin token surfaces as a sensitive output; read it with `terraform output -raw admin_token`.

## Notes

- `cosmos_remote_install` is one-shot: re-running `terraform apply` is a no-op. Re-creating it (e.g. `taint`) would re-run `get-pro.sh` on the host, which is generally safe but not idempotent.
- `cosmos_install` `Delete` revokes the issued admin token (best-effort) but leaves the server up.
