import wrap from './wrap';

interface BackupConfig {
  name: string;
  source: string;
  repository: string;
  password: string;
  crontab?: string;
  tags?: string[];
  exclude?: string[];
  autoStopContainers?: boolean;
}

interface RestoreConfig {
  snapshotId: string;
  target: string;
  include?: string[];
}

function listSnapshots(name: string) {
  return new Promise((resolve, reject) => {
    resolve({
      "data": [
        {
          "gid": 1000,
          "hostname": "DESKTOP-U8NKOPC",
          "id": "dd6829376d5ff357b8f2a26fda8e889f07315b0508f30938425a815c0b593d8a",
          "paths": [
            "/mnt/e/work/Cosmos-Server/build"
          ],
          "program_version": "restic 0.17.3",
          "short_id": "dd682937",
          "summary": {
            "backup_end": "2025-02-20T09:52:15.9215347Z",
            "backup_start": "2025-02-20T09:52:13.0886693Z",
            "data_added": 465466773,
            "data_added_packed": 174222327,
            "data_blobs": 635,
            "dirs_changed": 0,
            "dirs_new": 8,
            "dirs_unmodified": 0,
            "files_changed": 0,
            "files_new": 342,
            "files_unmodified": 0,
            "total_bytes_processed": 468036069,
            "total_files_processed": 342,
            "tree_blobs": 9
          },
          "tags": [
            "New Backup"
          ],
          "time": "2025-02-20T09:52:13.0886693Z",
          "tree": "ac64752c2df0e9e520dbdeada075510c7ee8163ee5a79a7ccbf152fb71cc7620",
          "uid": 1000,
          "username": "yann"
        },
        {
          "gid": 1000,
          "hostname": "DESKTOP-U8NKOPC",
          "id": "6e457ddfe4f7da64eb7a63a610be5785dd679e9fcf2a2a1e718a341af00af387",
          "parent": "dd6829376d5ff357b8f2a26fda8e889f07315b0508f30938425a815c0b593d8a",
          "paths": [
            "/mnt/e/work/Cosmos-Server/build"
          ],
          "program_version": "restic 0.17.3",
          "short_id": "6e457ddf",
          "summary": {
            "backup_end": "2025-02-20T09:52:19.3953924Z",
            "backup_start": "2025-02-20T09:52:18.6098852Z",
            "data_added": 0,
            "data_added_packed": 0,
            "data_blobs": 0,
            "dirs_changed": 0,
            "dirs_new": 0,
            "dirs_unmodified": 8,
            "files_changed": 0,
            "files_new": 0,
            "files_unmodified": 342,
            "total_bytes_processed": 468036069,
            "total_files_processed": 342,
            "tree_blobs": 0
          },
          "tags": [
            "New Backup"
          ],
          "time": "2025-02-20T09:52:18.6098852Z",
          "tree": "ac64752c2df0e9e520dbdeada075510c7ee8163ee5a79a7ccbf152fb71cc7620",
          "uid": 1000,
          "username": "yann"
        },
        {
          "gid": 1000,
          "hostname": "DESKTOP-U8NKOPC",
          "id": "5846cbc568e7c1f6954f2019aa8d9ffdb1282cde3f03323f43fce336be97d306",
          "parent": "6e457ddfe4f7da64eb7a63a610be5785dd679e9fcf2a2a1e718a341af00af387",
          "paths": [
            "/mnt/e/work/Cosmos-Server/build"
          ],
          "program_version": "restic 0.17.3",
          "short_id": "5846cbc5",
          "summary": {
            "backup_end": "2025-02-20T10:36:31.2479732Z",
            "backup_start": "2025-02-20T10:36:30.3198455Z",
            "data_added": 0,
            "data_added_packed": 0,
            "data_blobs": 0,
            "dirs_changed": 0,
            "dirs_new": 0,
            "dirs_unmodified": 8,
            "files_changed": 0,
            "files_new": 0,
            "files_unmodified": 342,
            "total_bytes_processed": 468036069,
            "total_files_processed": 342,
            "tree_blobs": 0
          },
          "tags": [
            "New Backup"
          ],
          "time": "2025-02-20T10:36:30.3198455Z",
          "tree": "ac64752c2df0e9e520dbdeada075510c7ee8163ee5a79a7ccbf152fb71cc7620",
          "uid": 1000,
          "username": "yann"
        }
      ],
      "status": "OK"
    })
  });
}

function listSnapshotsFromRepo(name: string) {
  return new Promise((resolve, reject) => {
    resolve({
      "data": [
        {
          "gid": 1000,
          "hostname": "DESKTOP-U8NKOPC",
          "id": "dd6829376d5ff357b8f2a26fda8e889f07315b0508f30938425a815c0b593d8a",
          "paths": [
            "/mnt/e/work/Cosmos-Server/build"
          ],
          "program_version": "restic 0.17.3",
          "short_id": "dd682937",
          "summary": {
            "backup_end": "2025-02-20T09:52:15.9215347Z",
            "backup_start": "2025-02-20T09:52:13.0886693Z",
            "data_added": 465466773,
            "data_added_packed": 174222327,
            "data_blobs": 635,
            "dirs_changed": 0,
            "dirs_new": 8,
            "dirs_unmodified": 0,
            "files_changed": 0,
            "files_new": 342,
            "files_unmodified": 0,
            "total_bytes_processed": 468036069,
            "total_files_processed": 342,
            "tree_blobs": 9
          },
          "tags": [
            "New Backup"
          ],
          "time": "2025-02-20T09:52:13.0886693Z",
          "tree": "ac64752c2df0e9e520dbdeada075510c7ee8163ee5a79a7ccbf152fb71cc7620",
          "uid": 1000,
          "username": "yann"
        },
        {
          "gid": 1000,
          "hostname": "DESKTOP-U8NKOPC",
          "id": "6e457ddfe4f7da64eb7a63a610be5785dd679e9fcf2a2a1e718a341af00af387",
          "parent": "dd6829376d5ff357b8f2a26fda8e889f07315b0508f30938425a815c0b593d8a",
          "paths": [
            "/mnt/e/work/Cosmos-Server/build"
          ],
          "program_version": "restic 0.17.3",
          "short_id": "6e457ddf",
          "summary": {
            "backup_end": "2025-02-20T09:52:19.3953924Z",
            "backup_start": "2025-02-20T09:52:18.6098852Z",
            "data_added": 0,
            "data_added_packed": 0,
            "data_blobs": 0,
            "dirs_changed": 0,
            "dirs_new": 0,
            "dirs_unmodified": 8,
            "files_changed": 0,
            "files_new": 0,
            "files_unmodified": 342,
            "total_bytes_processed": 468036069,
            "total_files_processed": 342,
            "tree_blobs": 0
          },
          "tags": [
            "New Backup"
          ],
          "time": "2025-02-20T09:52:18.6098852Z",
          "tree": "ac64752c2df0e9e520dbdeada075510c7ee8163ee5a79a7ccbf152fb71cc7620",
          "uid": 1000,
          "username": "yann"
        },
        {
          "gid": 1000,
          "hostname": "DESKTOP-U8NKOPC",
          "id": "5846cbc568e7c1f6954f2019aa8d9ffdb1282cde3f03323f43fce336be97d306",
          "parent": "6e457ddfe4f7da64eb7a63a610be5785dd679e9fcf2a2a1e718a341af00af387",
          "paths": [
            "/mnt/e/work/Cosmos-Server/build"
          ],
          "program_version": "restic 0.17.3",
          "short_id": "5846cbc5",
          "summary": {
            "backup_end": "2025-02-20T10:36:31.2479732Z",
            "backup_start": "2025-02-20T10:36:30.3198455Z",
            "data_added": 0,
            "data_added_packed": 0,
            "data_blobs": 0,
            "dirs_changed": 0,
            "dirs_new": 0,
            "dirs_unmodified": 8,
            "files_changed": 0,
            "files_new": 0,
            "files_unmodified": 342,
            "total_bytes_processed": 468036069,
            "total_files_processed": 342,
            "tree_blobs": 0
          },
          "tags": [
            "New Backup"
          ],
          "time": "2025-02-20T10:36:30.3198455Z",
          "tree": "ac64752c2df0e9e520dbdeada075510c7ee8163ee5a79a7ccbf152fb71cc7620",
          "uid": 1000,
          "username": "yann"
        }
      ],
      "status": "OK"
    })
  });
}

function listRepo() {
  return new Promise((resolve, reject) => {
    setTimeout(() => {
      resolve({
        "data": {
          "rclone:dropbox/myBackup": {
            "id": "My First Backup",
            "path": "rclone:dropbox/myBackup",
            "stats": {
              "compression_progress": 100,
              "compression_ratio": 1.485109909317082,
              "compression_space_saving": 32.664916332028014,
              "snapshots_count": 6,
              "total_blob_count": 83971,
              "total_size": 1731022265,
              "total_uncompressed_size": 2570758319
            },
            "status": "ok"
          }
        },
        "status": "OK"
      })},
      1000
    );
  });
}

function listFolders(name: string, snapshot: string, path?: string) {
  return new Promise((resolve, reject) => {
    setTimeout(() => {
      resolve({
        "data": [
          {
            "gid": 1000,
            "hostname": "DESKTOP-U8NKOPC",
            "id": "5846cbc568e7c1f6954f2019aa8d9ffdb1282cde3f03323f43fce336be97d306",
            "message_type": "snapshot",
            "parent": "6e457ddfe4f7da64eb7a63a610be5785dd679e9fcf2a2a1e718a341af00af387",
            "paths": [
              "/mnt/e/work/Cosmos-Server/build"
            ],
            "program_version": "restic 0.17.3",
            "short_id": "5846cbc5",
            "struct_type": "snapshot",
            "summary": {
              "backup_end": "2025-02-20T10:36:31.2479732Z",
              "backup_start": "2025-02-20T10:36:30.3198455Z",
              "data_added": 0,
              "data_added_packed": 0,
              "data_blobs": 0,
              "dirs_changed": 0,
              "dirs_new": 0,
              "dirs_unmodified": 8,
              "files_changed": 0,
              "files_new": 0,
              "files_unmodified": 342,
              "total_bytes_processed": 468036069,
              "total_files_processed": 342,
              "tree_blobs": 0
            },
            "tags": [
              "New Backup"
            ],
            "time": "2025-02-20T10:36:30.3198455Z",
            "tree": "ac64752c2df0e9e520dbdeada075510c7ee8163ee5a79a7ccbf152fb71cc7620",
            "uid": 1000,
            "username": "yann"
          },
          {
            "atime": "2025-02-20T09:44:24.3988013Z",
            "ctime": "2025-02-20T09:44:24.3988013Z",
            "gid": 1000,
            "inode": 63331869759897910,
            "message_type": "node",
            "mode": 2147484159,
            "mtime": "2025-02-20T09:44:24.3988013Z",
            "name": "build",
            "path": "/mnt/e/work/Cosmos-Server/build",
            "permissions": "drwxrwxrwx",
            "struct_type": "node",
            "type": "dir",
            "uid": 1000
          },
          {
            "atime": "2025-02-20T09:44:23.9911535Z",
            "ctime": "2025-02-20T09:44:23.9911535Z",
            "gid": 1000,
            "inode": 9570149208794440,
            "message_type": "node",
            "mode": 511,
            "mtime": "2025-02-20T09:44:23.9911535Z",
            "name": "GeoLite2-Country.mmdb",
            "path": "/mnt/e/work/Cosmos-Server/build/GeoLite2-Country.mmdb",
            "permissions": "-rwxrwxrwx",
            "size": 5845254,
            "struct_type": "node",
            "type": "file",
            "uid": 1000
          },
          {
            "atime": "2025-02-20T09:44:24.261641Z",
            "ctime": "2025-02-20T09:44:24.261641Z",
            "gid": 1000,
            "inode": 9851624185505104,
            "message_type": "node",
            "mode": 511,
            "mtime": "2025-02-20T09:44:24.261641Z",
            "name": "Logo.png",
            "path": "/mnt/e/work/Cosmos-Server/build/Logo.png",
            "permissions": "-rwxrwxrwx",
            "size": 356775,
            "struct_type": "node",
            "type": "file",
            "uid": 1000
          },
          {
            "atime": "2025-02-20T09:44:15.7563772Z",
            "ctime": "2025-02-20T09:44:15.7563772Z",
            "gid": 1000,
            "inode": 49821070877788024,
            "message_type": "node",
            "mode": 511,
            "mtime": "2025-02-20T09:44:15.7563772Z",
            "name": "cosmos",
            "path": "/mnt/e/work/Cosmos-Server/build/cosmos",
            "permissions": "-rwxrwxrwx",
            "size": 300710080,
            "struct_type": "node",
            "type": "file",
            "uid": 1000
          },
          {
            "atime": "2025-02-20T09:44:16.5374615Z",
            "ctime": "2025-02-20T09:44:16.5374615Z",
            "gid": 1000,
            "inode": 62205969853057070,
            "message_type": "node",
            "mode": 511,
            "mtime": "2025-02-20T09:44:16.5374615Z",
            "name": "cosmos-launcher",
            "path": "/mnt/e/work/Cosmos-Server/build/cosmos-launcher",
            "permissions": "-rwxrwxrwx",
            "size": 8313258,
            "struct_type": "node",
            "type": "file",
            "uid": 1000
          },
          {
            "atime": "2025-02-20T09:44:24.3912673Z",
            "ctime": "2025-02-20T09:44:24.3912673Z",
            "gid": 1000,
            "inode": 9570149208794452,
            "message_type": "node",
            "mode": 511,
            "mtime": "2025-02-20T09:44:24.3912673Z",
            "name": "cosmos_gray.png",
            "path": "/mnt/e/work/Cosmos-Server/build/cosmos_gray.png",
            "permissions": "-rwxrwxrwx",
            "size": 144765,
            "struct_type": "node",
            "type": "file",
            "uid": 1000
          },
          {
            "atime": "2025-02-20T09:44:24.3789064Z",
            "ctime": "2025-02-20T09:44:24.3789064Z",
            "gid": 1000,
            "inode": 9288674232083796,
            "message_type": "node",
            "mode": 2147484159,
            "mtime": "2025-02-20T09:44:24.3789064Z",
            "name": "images",
            "path": "/mnt/e/work/Cosmos-Server/build/images",
            "permissions": "drwxrwxrwx",
            "struct_type": "node",
            "type": "dir",
            "uid": 1000
          },
          {
            "atime": "2025-02-20T09:44:24.4288281Z",
            "ctime": "2025-02-20T09:44:24.4288281Z",
            "gid": 1000,
            "inode": 9288674232083798,
            "message_type": "node",
            "mode": 511,
            "mtime": "2025-02-20T09:44:24.4288281Z",
            "name": "meta.json",
            "path": "/mnt/e/work/Cosmos-Server/build/meta.json",
            "permissions": "-rwxrwxrwx",
            "size": 119,
            "struct_type": "node",
            "type": "file",
            "uid": 1000
          },
          {
            "atime": "2025-02-20T09:44:24.2481053Z",
            "ctime": "2025-02-20T09:44:24.2481053Z",
            "gid": 1000,
            "inode": 9851624185505104,
            "message_type": "node",
            "mode": 511,
            "mtime": "2025-02-20T09:44:24.2481053Z",
            "name": "nebula",
            "path": "/mnt/e/work/Cosmos-Server/build/nebula",
            "permissions": "-rwxrwxrwx",
            "size": 18986052,
            "struct_type": "node",
            "type": "file",
            "uid": 1000
          },
          {
            "atime": "2025-02-20T09:44:24.1670615Z",
            "ctime": "2025-02-20T09:44:24.1670615Z",
            "gid": 1000,
            "inode": 9570149208794444,
            "message_type": "node",
            "mode": 511,
            "mtime": "2025-02-20T09:44:24.1670615Z",
            "name": "nebula-arm",
            "path": "/mnt/e/work/Cosmos-Server/build/nebula-arm",
            "permissions": "-rwxrwxrwx",
            "size": 17814937,
            "struct_type": "node",
            "type": "file",
            "uid": 1000
          },
          {
            "atime": "2025-02-20T09:44:24.0332778Z",
            "ctime": "2025-02-20T09:44:24.0332778Z",
            "gid": 1000,
            "inode": 9288674232083784,
            "message_type": "node",
            "mode": 511,
            "mtime": "2025-02-20T09:44:24.0332778Z",
            "name": "nebula-arm-cert",
            "path": "/mnt/e/work/Cosmos-Server/build/nebula-arm-cert",
            "permissions": "-rwxrwxrwx",
            "size": 7263711,
            "struct_type": "node",
            "type": "file",
            "uid": 1000
          },
          {
            "atime": "2025-02-20T09:44:24.1067802Z",
            "ctime": "2025-02-20T09:44:24.1067802Z",
            "gid": 1000,
            "inode": 9570149208794444,
            "message_type": "node",
            "mode": 511,
            "mtime": "2025-02-20T09:44:24.1067802Z",
            "name": "nebula-cert",
            "path": "/mnt/e/work/Cosmos-Server/build/nebula-cert",
            "permissions": "-rwxrwxrwx",
            "size": 7675668,
            "struct_type": "node",
            "type": "file",
            "uid": 1000
          },
          {
            "atime": "2025-02-20T09:44:24.3687777Z",
            "ctime": "2025-02-20T09:44:24.3687777Z",
            "gid": 1000,
            "inode": 9570149208794450,
            "message_type": "node",
            "mode": 511,
            "mtime": "2025-02-20T09:44:24.3687777Z",
            "name": "restic",
            "path": "/mnt/e/work/Cosmos-Server/build/restic",
            "permissions": "-rwxrwxrwx",
            "size": 26501272,
            "struct_type": "node",
            "type": "file",
            "uid": 1000
          },
          {
            "atime": "2025-02-20T09:44:16.559491Z",
            "ctime": "2025-02-20T09:44:16.559491Z",
            "gid": 1000,
            "inode": 60235645016082504,
            "message_type": "node",
            "mode": 511,
            "mtime": "2025-02-20T09:44:16.559491Z",
            "name": "start.sh",
            "path": "/mnt/e/work/Cosmos-Server/build/start.sh",
            "permissions": "-rwxrwxrwx",
            "size": 84,
            "struct_type": "node",
            "type": "file",
            "uid": 1000
          },
          {
            "atime": "2025-02-20T09:44:16.603078Z",
            "ctime": "2025-02-20T09:44:16.603078Z",
            "gid": 1000,
            "inode": 39687971716204620,
            "message_type": "node",
            "mode": 2147484159,
            "mtime": "2025-02-20T09:44:16.603078Z",
            "name": "static",
            "path": "/mnt/e/work/Cosmos-Server/build/static",
            "permissions": "drwxrwxrwx",
            "struct_type": "node",
            "type": "dir",
            "uid": 1000
          }
        ],
        "status": "OK"
      })},
      500
    );
  });
}

function restoreBackup(name: string, config: RestoreConfig) {
  return new Promise((resolve, reject) => {
    setTimeout(() => {
      resolve({
        "status": "ok",
      })},
      500
    );
  });
}

function addBackup(config: BackupConfig) {
  return new Promise((resolve, reject) => {
    setTimeout(() => {
      resolve({
        "status": "ok",
      })},
      500
    );
  });
}

function editBackup(config: BackupConfig) {
  return new Promise((resolve, reject) => {
    setTimeout(() => {
      resolve({
        "status": "ok",
      })},
      500
    );
  });
}

function removeBackup(name: string, deleteRepo: boolean = false) {
  return new Promise((resolve, reject) => {
    setTimeout(() => {
      resolve({
        "status": "ok",
      })},
      500
    );
  });
}

function forgetSnapshot(name: string, snapshot: string, deleteRepo: boolean = false) {
  return new Promise((resolve, reject) => {
    setTimeout(() => {
      resolve({
        "status": "ok",
      })},
      500
    );
  });
}

function backupNow(name: string) {
  return new Promise((resolve, reject) => {
    setTimeout(() => {
      resolve({
        "status": "ok",
      })},
      500
    );
  });
}

function forgetNow(name: string) {
  return new Promise((resolve, reject) => {
    setTimeout(() => {
      resolve({
        "status": "ok",
      })},
      500
    );
  });
}

export {
  listSnapshots,
  listFolders,
  restoreBackup,
  addBackup,
  removeBackup,
  BackupConfig,
  RestoreConfig,
  listRepo,
  listSnapshotsFromRepo,
  forgetSnapshot,
  editBackup,
  backupNow,
  forgetNow,
};