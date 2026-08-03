# 3-node cluster on AWS EC2

Provisions three EC2 instances and bootstraps them as an equal-footing Cosmos cluster.

## What it does

1. Creates an EC2 security group (22, 80, 443 tcp, 4242 udp) and 3 `aws_instance.node` (Ubuntu).
2. `cosmos_remote_install` runs `get-pro.sh` on each instance.
3. Bootstraps node 0 via `cosmos_install` (the "founder") — Let's Encrypt + Create-mode MongoDB.
4. Creates the constellation on node 0 (`cosmos_constellation`).
5. Mints two device join blobs with `cosmos_constellation_device`, **`cosmos_node = 2`** (full server peers, not agents).
6. Bootstraps node 1 and node 2 via `cosmos_install` with `constellation_config`.
7. Creates one `cosmos_deployment` with `replicas = 3`, placed cluster-wide via `least-busy`.

After step 4, all three nodes operate as equal Cosmos servers. The "founder" label only describes which node generated the CA.

## Prerequisites

- AWS credentials in your environment (`AWS_ACCESS_KEY_ID` / `AWS_SECRET_ACCESS_KEY` or a profile).
- An existing AWS key pair (`var.key_name`) and the matching private key on disk.
- An Ubuntu AMI ID for your region (e.g. find with `aws ec2 describe-images --owners 099720109477 --filters Name=name,Values='ubuntu/images/hvm-ssd-gp3/ubuntu-noble-24.04-amd64-server-*'`).
- Three DNS A records (one per node), each pointing at the matching `aws_instance.node[i].public_ip`, so Let's Encrypt can issue per-node certs. Run a two-stage apply: first `terraform apply -target=aws_instance.node` to allocate the IPs, point your DNS records at them, then full `terraform apply`. Pass them in as `-var 'node_hostnames=["a.example.com","b.example.com","c.example.com"]'`.
- A Cosmos Pro licence key.

## Run it

```bash
cd examples/aws-3node

terraform init
terraform apply \
  -var 'region=us-east-1' \
  -var 'key_name=my-aws-key' \
  -var 'ssh_private_key_path=~/.ssh/my-aws-key.pem' \
  -var 'ami=ami-0abcdef…' \
  -var 'node_hostnames=["cosmos-0.example.com","cosmos-1.example.com","cosmos-2.example.com"]' \
  -var 'admin_password=…' \
  -var 'cosmos_licence=…'
```

Outputs: `node_ips` (list of 3) and `admin_token` (sensitive).

## Notes

- All three nodes use `LETSENCRYPT`. Each `node_hostnames[i]` must resolve to the matching `aws_instance.node[i].public_ip` before the corresponding `cosmos_install` runs, otherwise the LE HTTP-01 challenge will fail. The two-stage apply above handles the chicken-and-egg.
- The `cosmos_deployment.web` lives in NATS KV and is cluster-replicated, so the `cosmos.cluster` provider could equally point at any of the three nodes.
