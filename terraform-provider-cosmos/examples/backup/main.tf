# Backup configuration example.
#
# IMPORTANT: Destroying a cosmos_backup resource permanently deletes all its
# snapshots from the repository. Always use prevent_destroy to guard against
# accidental deletion.

resource "cosmos_backup" "daily" {
  name       = "db-daily"
  repository = "/backups/repo1"
  source     = "/data/db"
  crontab    = "0 2 * * *"

  lifecycle {
    prevent_destroy = true
  }
}
