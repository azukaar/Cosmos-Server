// https://github.com/rclone/rclone-webui-react/blob/master/src/views/RemoteManagement/NewDrive/config.js

/*
The MIT License (MIT)

Copyright (c) 2018 creativeLabs Łukasz Holeczek.

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/

export const ProvConfig = [
  {
      "Name": "alias",
      "Description": "Alias for a existing remote",
      "Prefix": "alias",
      "Options": [
          {
              "Name": "remote",
              "Help": "Remote or path to alias.\nCan be \"myremote:path/to/dir\", \"myremote:bucket\", \"myremote:\" or \"/local/path\".",
              "Provider": "",
              "Default": "",
              "Value": null,
              "ShortOpt": "",
              "Hide": 0,
              "Required": true,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": false
          }
      ]
  },
  {
      "Name": "amazon cloud drive",
      "Description": "Amazon Drive",
      "Prefix": "acd",
      "Options": [
          {
              "Name": "client_id",
              "Help": "Amazon Application Client ID.",
              "Provider": "",
              "Default": "",
              "Value": null,
              "ShortOpt": "",
              "Hide": 0,
              "Required": true,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": false
          },
          {
              "Name": "client_secret",
              "Help": "Amazon Application Client Secret.",
              "Provider": "",
              "Default": "",
              "Value": null,
              "ShortOpt": "",
              "Hide": 0,
              "Required": true,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": false
          },
          {
              "Name": "auth_url",
              "Help": "Auth server URL.\nLeave blank to use Amazon's.",
              "Provider": "",
              "Default": "",
              "Value": null,
              "ShortOpt": "",
              "Hide": 0,
              "Required": false,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": true
          },
          {
              "Name": "token_url",
              "Help": "Token server url.\nleave blank to use Amazon's.",
              "Provider": "",
              "Default": "",
              "Value": null,
              "ShortOpt": "",
              "Hide": 0,
              "Required": false,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": true
          },
          {
              "Name": "checkpoint",
              "Help": "Checkpoint for internal polling (debug).",
              "Provider": "",
              "Default": "",
              "Value": null,
              "ShortOpt": "",
              "Hide": 3,
              "Required": false,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": true
          },
          {
              "Name": "upload_wait_per_gb",
              "Help": "Additional time per GB to wait after a failed complete upload to see if it appears.\n\nSometimes Amazon Drive gives an error when a file has been fully\nuploaded but the file appears anyway after a little while.  This\nhappens sometimes for files over 1GB in size and nearly every time for\nfiles bigger than 10GB. This parameter controls the time rclone waits\nfor the file to appear.\n\nThe default value for this parameter is 3 minutes per GB, so by\ndefault it will wait 3 minutes for every GB uploaded to see if the\nfile appears.\n\nYou can disable this feature by setting it to 0. This may cause\nconflict errors as rclone retries the failed upload but the file will\nmost likely appear correctly eventually.\n\nThese values were determined empirically by observing lots of uploads\nof big files for a range of file sizes.\n\nUpload with the \"-v\" flag to see more info about what rclone is doing\nin this situation.",
              "Provider": "",
              "Default": 180000000000,
              "Value": null,
              "ShortOpt": "",
              "Hide": 0,
              "Required": false,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": true
          },
          {
              "Name": "templink_threshold",
              "Help": "Files \u003e= this size will be downloaded via their tempLink.\n\nFiles this size or more will be downloaded via their \"tempLink\". This\nis to work around a problem with Amazon Drive which blocks downloads\nof files bigger than about 10GB.  The default for this is 9GB which\nshouldn't need to be changed.\n\nTo download files above this threshold, rclone requests a \"tempLink\"\nwhich downloads the file through a temporary URL directly from the\nunderlying S3 storage.",
              "Provider": "",
              "Default": 9663676416,
              "Value": null,
              "ShortOpt": "",
              "Hide": 0,
              "Required": false,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": true
          }
      ]
  },
  {
      "Name": "azureblob",
      "Description": "Microsoft Azure Blob Storage",
      "Prefix": "azureblob",
      "Options": [
          {
              "Name": "account",
              "Help": "Storage Account Name (leave blank to use connection string or SAS URL)",
              "Provider": "",
              "Default": "",
              "Value": null,
              "ShortOpt": "",
              "Hide": 0,
              "Required": false,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": false
          },
          {
              "Name": "key",
              "Help": "Storage Account Key (leave blank to use connection string or SAS URL)",
              "Provider": "",
              "Default": "",
              "Value": null,
              "ShortOpt": "",
              "Hide": 0,
              "Required": false,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": false
          },
          {
              "Name": "sas_url",
              "Help": "SAS URL for container level access only\n(leave blank if using account/key or connection string)",
              "Provider": "",
              "Default": "",
              "Value": null,
              "ShortOpt": "",
              "Hide": 0,
              "Required": false,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": false
          },
          {
              "Name": "endpoint",
              "Help": "Endpoint for the service\nLeave blank normally.",
              "Provider": "",
              "Default": "",
              "Value": null,
              "ShortOpt": "",
              "Hide": 0,
              "Required": false,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": true
          },
          {
              "Name": "upload_cutoff",
              "Help": "Cutoff for switching to chunked upload (\u003c= 256MB).",
              "Provider": "",
              "Default": 268435456,
              "Value": null,
              "ShortOpt": "",
              "Hide": 0,
              "Required": false,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": true
          },
          {
              "Name": "chunk_size",
              "Help": "Upload chunk size (\u003c= 100MB).\n\nNote that this is stored in memory and there may be up to\n\"--transfers\" chunks stored at once in memory.",
              "Provider": "",
              "Default": 4194304,
              "Value": null,
              "ShortOpt": "",
              "Hide": 0,
              "Required": false,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": true
          },
          {
              "Name": "list_chunk",
              "Help": "Size of blob list.\n\nThis sets the number of blobs requested in each listing chunk. Default\nis the maximum, 5000. \"List blobs\" requests are permitted 2 minutes\nper megabyte to complete. If an operation is taking longer than 2\nminutes per megabyte on average, it will time out (\n[source](https://docs.microsoft.com/en-us/rest/api/storageservices/setting-timeouts-for-blob-service-operations#exceptions-to-default-timeout-interval)\n). This can be used to limit the number of blobs items to return, to\navoid the time out.",
              "Provider": "",
              "Default": 5000,
              "Value": null,
              "ShortOpt": "",
              "Hide": 0,
              "Required": false,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": true
          },
          {
              "Name": "access_tier",
              "Help": "Access tier of blob: hot, cool or archive.\n\nArchived blobs can be restored by setting access tier to hot or\ncool. Leave blank if you intend to use default access tier, which is\nset at account level\n\nIf there is no \"access tier\" specified, rclone doesn't apply any tier.\nrclone performs \"Set Tier\" operation on blobs while uploading, if objects\nare not modified, specifying \"access tier\" to new one will have no effect.\nIf blobs are in \"archive tier\" at remote, trying to perform data transfer\noperations from remote will not be allowed. User should first restore by\ntiering blob to \"Hot\" or \"Cool\".",
              "Provider": "",
              "Default": "",
              "Value": null,
              "ShortOpt": "",
              "Hide": 0,
              "Required": false,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": true
          }
      ]
  },
  {
      "Name": "b2",
      "Description": "Backblaze B2",
      "Prefix": "b2",
      "Options": [
          {
              "Name": "account",
              "Help": "Account ID or Application Key ID",
              "Provider": "",
              "Default": "",
              "Value": null,
              "ShortOpt": "",
              "Hide": 0,
              "Required": true,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": false
          },
          {
              "Name": "key",
              "Help": "Application Key",
              "Provider": "",
              "Default": "",
              "Value": null,
              "ShortOpt": "",
              "Hide": 0,
              "Required": true,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": false
          },
          {
              "Name": "endpoint",
              "Help": "Endpoint for the service.\nLeave blank normally.",
              "Provider": "",
              "Default": "",
              "Value": null,
              "ShortOpt": "",
              "Hide": 0,
              "Required": false,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": true
          },
          {
              "Name": "test_mode",
              "Help": "A flag string for X-Bz-Test-Mode header for debugging.\n\nThis is for debugging purposes only. Setting it to one of the strings\nbelow will cause b2 to return specific errors:\n\n  * \"fail_some_uploads\"\n  * \"expire_some_account_authorization_tokens\"\n  * \"force_cap_exceeded\"\n\nThese will be set in the \"X-Bz-Test-Mode\" header which is documented\nin the [b2 integrations checklist](https://www.backblaze.com/b2/docs/integration_checklist.html).",
              "Provider": "",
              "Default": "",
              "Value": null,
              "ShortOpt": "",
              "Hide": 2,
              "Required": false,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": true
          },
          {
              "Name": "versions",
              "Help": "Include old versions in directory listings.\nNote that when using this no file write operations are permitted,\nso you can't upload files or delete them.",
              "Provider": "",
              "Default": false,
              "Value": null,
              "ShortOpt": "",
              "Hide": 0,
              "Required": false,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": true
          },
          {
              "Name": "hard_delete",
              "Help": "Permanently delete files on remote removal, otherwise hide files.",
              "Provider": "",
              "Default": false,
              "Value": null,
              "ShortOpt": "",
              "Hide": 0,
              "Required": false,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": false
          },
          {
              "Name": "upload_cutoff",
              "Help": "Cutoff for switching to chunked upload.\n\nFiles above this size will be uploaded in chunks of \"--b2-chunk-size\".\n\nThis value should be set no larger than 4.657GiB (== 5GB).",
              "Provider": "",
              "Default": 209715200,
              "Value": null,
              "ShortOpt": "",
              "Hide": 0,
              "Required": false,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": true
          },
          {
              "Name": "chunk_size",
              "Help": "Upload chunk size. Must fit in memory.\n\nWhen uploading large files, chunk the file into this size.  Note that\nthese chunks are buffered in memory and there might a maximum of\n\"--transfers\" chunks in progress at once.  5,000,000 Bytes is the\nminimum size.",
              "Provider": "",
              "Default": 100663296,
              "Value": null,
              "ShortOpt": "",
              "Hide": 0,
              "Required": false,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": true
          },
          {
              "Name": "disable_checksum",
              "Help": "Disable checksums for large (\u003e upload cutoff) files",
              "Provider": "",
              "Default": false,
              "Value": null,
              "ShortOpt": "",
              "Hide": 0,
              "Required": false,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": true
          }
      ]
  },
  {
      "Name": "box",
      "Description": "Box",
      "Prefix": "box",
      "Options": [
          {
              "Name": "client_id",
              "Help": "Box App Client Id.\nLeave blank normally.",
              "Provider": "",
              "Default": "",
              "Value": null,
              "ShortOpt": "",
              "Hide": 0,
              "Required": false,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": false
          },
          {
              "Name": "client_secret",
              "Help": "Box App Client Secret\nLeave blank normally.",
              "Provider": "",
              "Default": "",
              "Value": null,
              "ShortOpt": "",
              "Hide": 0,
              "Required": false,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": false
          },
          {
              "Name": "upload_cutoff",
              "Help": "Cutoff for switching to multipart upload (\u003e= 50MB).",
              "Provider": "",
              "Default": 52428800,
              "Value": null,
              "ShortOpt": "",
              "Hide": 0,
              "Required": false,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": true
          },
          {
              "Name": "commit_retries",
              "Help": "Max number of times to try committing a multipart file.",
              "Provider": "",
              "Default": 100,
              "Value": null,
              "ShortOpt": "",
              "Hide": 0,
              "Required": false,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": true
          }
      ]
  },
  {
      "Name": "crypt",
      "Description": "Encrypt/Decrypt a remote",
      "Prefix": "crypt",
      "Options": [
          {
              "Name": "remote",
              "Help": "Remote to encrypt/decrypt.\nNormally should contain a ':' and a path, eg \"myremote:path/to/dir\",\n\"myremote:bucket\" or maybe \"myremote:\" (not recommended).",
              "Provider": "",
              "Default": "",
              "Value": null,
              "ShortOpt": "",
              "Hide": 0,
              "Required": true,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": false
          },
          {
              "Name": "filename_encryption",
              "Help": "How to encrypt the filenames.",
              "Provider": "",
              "Default": "standard",
              "Value": null,
              "Examples": [
                  {
                      "Value": "off",
                      "Help": "Don't encrypt the file names.  Adds a \".bin\" extension only.",
                      "Provider": ""
                  },
                  {
                      "Value": "standard",
                      "Help": "Encrypt the filenames see the docs for the details.",
                      "Provider": ""
                  },
                  {
                      "Value": "obfuscate",
                      "Help": "Very simple filename obfuscation.",
                      "Provider": ""
                  }
              ],
              "ShortOpt": "",
              "Hide": 0,
              "Required": false,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": false
          },
          {
              "Name": "directory_name_encryption",
              "Help": "Option to either encrypt directory names or leave them intact.",
              "Provider": "",
              "Default": true,
              "Value": null,
              "Examples": [
                  {
                      "Value": "true",
                      "Help": "Encrypt directory names.",
                      "Provider": ""
                  },
                  {
                      "Value": "false",
                      "Help": "Don't encrypt directory names, leave them intact.",
                      "Provider": ""
                  }
              ],
              "ShortOpt": "",
              "Hide": 0,
              "Required": false,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": false
          },
          {
              "Name": "password",
              "Help": "Password or pass phrase for encryption.",
              "Provider": "",
              "Default": "",
              "Value": null,
              "ShortOpt": "",
              "Hide": 0,
              "Required": false,
              "IsPassword": true,
              "NoPrefix": false,
              "Advanced": false
          },
          {
              "Name": "password2",
              "Help": "Password or pass phrase for salt. Optional but recommended.\nShould be different to the previous password.",
              "Provider": "",
              "Default": "",
              "Value": null,
              "ShortOpt": "",
              "Hide": 0,
              "Required": false,
              "IsPassword": true,
              "NoPrefix": false,
              "Advanced": false
          },
          {
              "Name": "show_mapping",
              "Help": "For all files listed show how the names encrypt.\n\nIf this flag is set then for each file that the remote is asked to\nlist, it will log (at level INFO) a line stating the decrypted file\nname and the encrypted file name.\n\nThis is so you can work out which encrypted names are which decrypted\nnames just in case you need to do something with the encrypted file\nnames, or for debugging purposes.",
              "Provider": "",
              "Default": false,
              "Value": null,
              "ShortOpt": "",
              "Hide": 2,
              "Required": false,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": true
          }
      ]
  },
  {
      "Name": "cache",
      "Description": "Cache a remote",
      "Prefix": "cache",
      "Options": [
          {
              "Name": "remote",
              "Help": "Remote to cache.\nNormally should contain a ':' and a path, eg \"myremote:path/to/dir\",\n\"myremote:bucket\" or maybe \"myremote:\" (not recommended).",
              "Provider": "",
              "Default": "",
              "Value": null,
              "ShortOpt": "",
              "Hide": 0,
              "Required": true,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": false
          },
          {
              "Name": "plex_url",
              "Help": "The URL of the Plex server",
              "Provider": "",
              "Default": "",
              "Value": null,
              "ShortOpt": "",
              "Hide": 0,
              "Required": false,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": false
          },
          {
              "Name": "plex_username",
              "Help": "The username of the Plex user",
              "Provider": "",
              "Default": "",
              "Value": null,
              "ShortOpt": "",
              "Hide": 0,
              "Required": false,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": false
          },
          {
              "Name": "plex_password",
              "Help": "The password of the Plex user",
              "Provider": "",
              "Default": "",
              "Value": null,
              "ShortOpt": "",
              "Hide": 0,
              "Required": false,
              "IsPassword": true,
              "NoPrefix": false,
              "Advanced": false
          },
          {
              "Name": "plex_token",
              "Help": "The plex token for authentication - auto set normally",
              "Provider": "",
              "Default": "",
              "Value": null,
              "ShortOpt": "",
              "Hide": 3,
              "Required": false,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": true
          },
          {
              "Name": "plex_insecure",
              "Help": "Skip all certificate verifications when connecting to the Plex server",
              "Provider": "",
              "Default": "",
              "Value": null,
              "ShortOpt": "",
              "Hide": 0,
              "Required": false,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": true
          },
          {
              "Name": "chunk_size",
              "Help": "The size of a chunk (partial file data).\n\nUse lower numbers for slower connections. If the chunk size is\nchanged, any downloaded chunks will be invalid and cache-chunk-path\nwill need to be cleared or unexpected EOF errors will occur.",
              "Provider": "",
              "Default": 5242880,
              "Value": null,
              "Examples": [
                  {
                      "Value": "1m",
                      "Help": "1MB",
                      "Provider": ""
                  },
                  {
                      "Value": "5M",
                      "Help": "5 MB",
                      "Provider": ""
                  },
                  {
                      "Value": "10M",
                      "Help": "10 MB",
                      "Provider": ""
                  }
              ],
              "ShortOpt": "",
              "Hide": 0,
              "Required": false,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": false
          },
          {
              "Name": "info_age",
              "Help": "How long to cache file structure information (directory listings, file size, times etc). \nIf all write operations are done through the cache then you can safely make\nthis value very large as the cache store will also be updated in real time.",
              "Provider": "",
              "Default": 21600000000000,
              "Value": null,
              "Examples": [
                  {
                      "Value": "1h",
                      "Help": "1 hour",
                      "Provider": ""
                  },
                  {
                      "Value": "24h",
                      "Help": "24 hours",
                      "Provider": ""
                  },
                  {
                      "Value": "48h",
                      "Help": "48 hours",
                      "Provider": ""
                  }
              ],
              "ShortOpt": "",
              "Hide": 0,
              "Required": false,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": false
          },
          {
              "Name": "chunk_total_size",
              "Help": "The total size that the chunks can take up on the local disk.\n\nIf the cache exceeds this value then it will start to delete the\noldest chunks until it goes under this value.",
              "Provider": "",
              "Default": 10737418240,
              "Value": null,
              "Examples": [
                  {
                      "Value": "500M",
                      "Help": "500 MB",
                      "Provider": ""
                  },
                  {
                      "Value": "1G",
                      "Help": "1 GB",
                      "Provider": ""
                  },
                  {
                      "Value": "10G",
                      "Help": "10 GB",
                      "Provider": ""
                  }
              ],
              "ShortOpt": "",
              "Hide": 0,
              "Required": false,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": false
          },
          {
              "Name": "db_path",
              "Help": "Directory to store file structure metadata DB.\nThe remote name is used as the DB file name.",
              "Provider": "",
              "Default": "/home/negative0/.cache/rclone/cache-backend",
              "Value": null,
              "ShortOpt": "",
              "Hide": 0,
              "Required": false,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": true
          },
          {
              "Name": "chunk_path",
              "Help": "Directory to cache chunk files.\n\nPath to where partial file data (chunks) are stored locally. The remote\nname is appended to the final path.\n\nThis config follows the \"--cache-db-path\". If you specify a custom\nlocation for \"--cache-db-path\" and don't specify one for \"--cache-chunk-path\"\nthen \"--cache-chunk-path\" will use the same path as \"--cache-db-path\".",
              "Provider": "",
              "Default": "/home/negative0/.cache/rclone/cache-backend",
              "Value": null,
              "ShortOpt": "",
              "Hide": 0,
              "Required": false,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": true
          },
          {
              "Name": "db_purge",
              "Help": "Clear all the cached data for this remote on start.",
              "Provider": "",
              "Default": false,
              "Value": null,
              "ShortOpt": "",
              "Hide": 2,
              "Required": false,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": true
          },
          {
              "Name": "chunk_clean_interval",
              "Help": "How often should the cache perform cleanups of the chunk storage.\nThe default value should be ok for most people. If you find that the\ncache goes over \"cache-chunk-total-size\" too often then try to lower\nthis value to force it to perform cleanups more often.",
              "Provider": "",
              "Default": 60000000000,
              "Value": null,
              "ShortOpt": "",
              "Hide": 0,
              "Required": false,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": true
          },
          {
              "Name": "read_retries",
              "Help": "How many times to retry a read from a cache storage.\n\nSince reading from a cache stream is independent from downloading file\ndata, readers can get to a point where there's no more data in the\ncache.  Most of the times this can indicate a connectivity issue if\ncache isn't able to provide file data anymore.\n\nFor really slow connections, increase this to a point where the stream is\nable to provide data but your experience will be very stuttering.",
              "Provider": "",
              "Default": 10,
              "Value": null,
              "ShortOpt": "",
              "Hide": 0,
              "Required": false,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": true
          },
          {
              "Name": "workers",
              "Help": "How many workers should run in parallel to download chunks.\n\nHigher values will mean more parallel processing (better CPU needed)\nand more concurrent requests on the cloud provider.  This impacts\nseveral aspects like the cloud provider API limits, more stress on the\nhardware that rclone runs on but it also means that streams will be\nmore fluid and data will be available much more faster to readers.\n\n**Note**: If the optional Plex integration is enabled then this\nsetting will adapt to the type of reading performed and the value\nspecified here will be used as a maximum number of workers to use.",
              "Provider": "",
              "Default": 4,
              "Value": null,
              "ShortOpt": "",
              "Hide": 0,
              "Required": false,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": true
          },
          {
              "Name": "chunk_no_memory",
              "Help": "Disable the in-memory cache for storing chunks during streaming.\n\nBy default, cache will keep file data during streaming in RAM as well\nto provide it to readers as fast as possible.\n\nThis transient data is evicted as soon as it is read and the number of\nchunks stored doesn't exceed the number of workers. However, depending\non other settings like \"cache-chunk-size\" and \"cache-workers\" this footprint\ncan increase if there are parallel streams too (multiple files being read\nat the same time).\n\nIf the hardware permits it, use this feature to provide an overall better\nperformance during streaming but it can also be disabled if RAM is not\navailable on the local machine.",
              "Provider": "",
              "Default": false,
              "Value": null,
              "ShortOpt": "",
              "Hide": 0,
              "Required": false,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": true
          },
          {
              "Name": "rps",
              "Help": "Limits the number of requests per second to the source FS (-1 to disable)\n\nThis setting places a hard limit on the number of requests per second\nthat cache will be doing to the cloud provider remote and try to\nrespect that value by setting waits between reads.\n\nIf you find that you're getting banned or limited on the cloud\nprovider through cache and know that a smaller number of requests per\nsecond will allow you to work with it then you can use this setting\nfor that.\n\nA good balance of all the other settings should make this setting\nuseless but it is available to set for more special cases.\n\n**NOTE**: This will limit the number of requests during streams but\nother API calls to the cloud provider like directory listings will\nstill pass.",
              "Provider": "",
              "Default": -1,
              "Value": null,
              "ShortOpt": "",
              "Hide": 0,
              "Required": false,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": true
          },
          {
              "Name": "writes",
              "Help": "Cache file data on writes through the FS\n\nIf you need to read files immediately after you upload them through\ncache you can enable this flag to have their data stored in the\ncache store at the same time during upload.",
              "Provider": "",
              "Default": false,
              "Value": null,
              "ShortOpt": "",
              "Hide": 0,
              "Required": false,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": true
          },
          {
              "Name": "tmp_upload_path",
              "Help": "Directory to keep temporary files until they are uploaded.\n\nThis is the path where cache will use as a temporary storage for new\nfiles that need to be uploaded to the cloud provider.\n\nSpecifying a value will enable this feature. Without it, it is\ncompletely disabled and files will be uploaded directly to the cloud\nprovider",
              "Provider": "",
              "Default": "",
              "Value": null,
              "ShortOpt": "",
              "Hide": 0,
              "Required": false,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": true
          },
          {
              "Name": "tmp_wait_time",
              "Help": "How long should files be stored in local cache before being uploaded\n\nThis is the duration that a file must wait in the temporary location\n_cache-tmp-upload-path_ before it is selected for upload.\n\nNote that only one file is uploaded at a time and it can take longer\nto start the upload if a queue formed for this purpose.",
              "Provider": "",
              "Default": 15000000000,
              "Value": null,
              "ShortOpt": "",
              "Hide": 0,
              "Required": false,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": true
          },
          {
              "Name": "db_wait_time",
              "Help": "How long to wait for the DB to be available - 0 is unlimited\n\nOnly one process can have the DB open at any one time, so rclone waits\nfor this duration for the DB to become available before it gives an\nerror.\n\nIf you set it to 0 then it will wait forever.",
              "Provider": "",
              "Default": 1000000000,
              "Value": null,
              "ShortOpt": "",
              "Hide": 0,
              "Required": false,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": true
          }
      ]
  },
  {
      "Name": "drive",
      "Description": "Google Drive",
      "Prefix": "drive",
      "Options": [
          {
              "Name": "client_id",
              "Help": "Google Application Client Id\nLeave blank normally.",
              "Provider": "",
              "Default": "",
              "Value": null,
              "ShortOpt": "",
              "Hide": 0,
              "Required": false,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": false
          },
          {
              "Name": "client_secret",
              "Help": "Google Application Client Secret\nLeave blank normally.",
              "Provider": "",
              "Default": "",
              "Value": null,
              "ShortOpt": "",
              "Hide": 0,
              "Required": false,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": false
          },
          {
              "Name": "scope",
              "Help": "Scope that rclone should use when requesting access from drive.",
              "Provider": "",
              "Default": "",
              "Value": null,
              "Examples": [
                  {
                      "Value": "drive",
                      "Help": "Full access all files, excluding Application Data Folder.",
                      "Provider": ""
                  },
                  {
                      "Value": "drive.readonly",
                      "Help": "Read-only access to file metadata and file contents.",
                      "Provider": ""
                  },
                  {
                      "Value": "drive.file",
                      "Help": "Access to files created by rclone only.\nThese are visible in the drive website.\nFile authorization is revoked when the user deauthorizes the app.",
                      "Provider": ""
                  },
                  {
                      "Value": "drive.appfolder",
                      "Help": "Allows read and write access to the Application Data folder.\nThis is not visible in the drive website.",
                      "Provider": ""
                  },
                  {
                      "Value": "drive.metadata.readonly",
                      "Help": "Allows read-only access to file metadata but\ndoes not allow any access to read or download file content.",
                      "Provider": ""
                  }
              ],
              "ShortOpt": "",
              "Hide": 0,
              "Required": false,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": false
          },
          {
              "Name": "root_folder_id",
              "Help": "ID of the root folder\nLeave blank normally.\nFill in to access \"Computers\" folders. (see docs).",
              "Provider": "",
              "Default": "",
              "Value": null,
              "ShortOpt": "",
              "Hide": 0,
              "Required": false,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": false
          },
          {
              "Name": "service_account_file",
              "Help": "Service Account Credentials JSON file path \nLeave blank normally.\nNeeded only if you want use SA instead of interactive login.",
              "Provider": "",
              "Default": "",
              "Value": null,
              "ShortOpt": "",
              "Hide": 0,
              "Required": false,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": false
          },
          {
              "Name": "service_account_credentials",
              "Help": "Service Account Credentials JSON blob\nLeave blank normally.\nNeeded only if you want use SA instead of interactive login.",
              "Provider": "",
              "Default": "",
              "Value": null,
              "ShortOpt": "",
              "Hide": 2,
              "Required": false,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": true
          },
          {
              "Name": "team_drive",
              "Help": "ID of the Team Drive",
              "Provider": "",
              "Default": "",
              "Value": null,
              "ShortOpt": "",
              "Hide": 2,
              "Required": false,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": true
          },
          {
              "Name": "auth_owner_only",
              "Help": "Only consider files owned by the authenticated user.",
              "Provider": "",
              "Default": false,
              "Value": null,
              "ShortOpt": "",
              "Hide": 0,
              "Required": false,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": true
          },
          {
              "Name": "use_trash",
              "Help": "Send files to the trash instead of deleting permanently.\nDefaults to true, namely sending files to the trash.\nUse `--drive-use-trash=false` to delete files permanently instead.",
              "Provider": "",
              "Default": true,
              "Value": null,
              "ShortOpt": "",
              "Hide": 0,
              "Required": false,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": true
          },
          {
              "Name": "skip_gdocs",
              "Help": "Skip google documents in all listings.\nIf given, gdocs practically become invisible to rclone.",
              "Provider": "",
              "Default": false,
              "Value": null,
              "ShortOpt": "",
              "Hide": 0,
              "Required": false,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": true
          },
          {
              "Name": "shared_with_me",
              "Help": "Only show files that are shared with me.\n\nInstructs rclone to operate on your \"Shared with me\" folder (where\nGoogle Drive lets you access the files and folders others have shared\nwith you).\n\nThis works both with the \"list\" (lsd, lsl, etc) and the \"copy\"\ncommands (copy, sync, etc), and with all other commands too.",
              "Provider": "",
              "Default": false,
              "Value": null,
              "ShortOpt": "",
              "Hide": 0,
              "Required": false,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": true
          },
          {
              "Name": "trashed_only",
              "Help": "Only show files that are in the trash.\nThis will show trashed files in their original directory structure.",
              "Provider": "",
              "Default": false,
              "Value": null,
              "ShortOpt": "",
              "Hide": 0,
              "Required": false,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": true
          },
          {
              "Name": "formats",
              "Help": "Deprecated: see export_formats",
              "Provider": "",
              "Default": "",
              "Value": null,
              "ShortOpt": "",
              "Hide": 2,
              "Required": false,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": true
          },
          {
              "Name": "export_formats",
              "Help": "Comma separated list of preferred formats for downloading Google docs.",
              "Provider": "",
              "Default": "docx,xlsx,pptx,svg",
              "Value": null,
              "ShortOpt": "",
              "Hide": 0,
              "Required": false,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": true
          },
          {
              "Name": "import_formats",
              "Help": "Comma separated list of preferred formats for uploading Google docs.",
              "Provider": "",
              "Default": "",
              "Value": null,
              "ShortOpt": "",
              "Hide": 0,
              "Required": false,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": true
          },
          {
              "Name": "allow_import_name_change",
              "Help": "Allow the filetype to change when uploading Google docs (e.g. file.doc to file.docx). This will confuse sync and reupload every time.",
              "Provider": "",
              "Default": false,
              "Value": null,
              "ShortOpt": "",
              "Hide": 0,
              "Required": false,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": true
          },
          {
              "Name": "use_created_date",
              "Help": "Use file created date instead of modified date.,\n\nUseful when downloading data and you want the creation date used in\nplace of the last modified date.\n\n**WARNING**: This flag may have some unexpected consequences.\n\nWhen uploading to your drive all files will be overwritten unless they\nhaven't been modified since their creation. And the inverse will occur\nwhile downloading.  This side effect can be avoided by using the\n\"--checksum\" flag.\n\nThis feature was implemented to retain photos capture date as recorded\nby google photos. You will first need to check the \"Create a Google\nPhotos folder\" option in your google drive settings. You can then copy\nor move the photos locally and use the date the image was taken\n(created) set as the modification date.",
              "Provider": "",
              "Default": false,
              "Value": null,
              "ShortOpt": "",
              "Hide": 0,
              "Required": false,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": true
          },
          {
              "Name": "list_chunk",
              "Help": "Size of listing chunk 100-1000. 0 to disable.",
              "Provider": "",
              "Default": 1000,
              "Value": null,
              "ShortOpt": "",
              "Hide": 0,
              "Required": false,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": true
          },
          {
              "Name": "impersonate",
              "Help": "Impersonate this user when using a service account.",
              "Provider": "",
              "Default": "",
              "Value": null,
              "ShortOpt": "",
              "Hide": 0,
              "Required": false,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": true
          },
          {
              "Name": "alternate_export",
              "Help": "Use alternate export URLs for google documents export.,\n\nIf this option is set this instructs rclone to use an alternate set of\nexport URLs for drive documents.  Users have reported that the\nofficial export URLs can't export large documents, whereas these\nunofficial ones can.\n\nSee rclone issue [#2243](https://github.com/ncw/rclone/issues/2243) for background,\n[this google drive issue](https://issuetracker.google.com/issues/36761333) and\n[this helpful post](https://www.labnol.org/internet/direct-links-for-google-drive/28356/).",
              "Provider": "",
              "Default": false,
              "Value": null,
              "ShortOpt": "",
              "Hide": 0,
              "Required": false,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": true
          },
          {
              "Name": "upload_cutoff",
              "Help": "Cutoff for switching to chunked upload",
              "Provider": "",
              "Default": 8388608,
              "Value": null,
              "ShortOpt": "",
              "Hide": 0,
              "Required": false,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": true
          },
          {
              "Name": "chunk_size",
              "Help": "Upload chunk size. Must a power of 2 \u003e= 256k.\n\nMaking this larger will improve performance, but note that each chunk\nis buffered in memory one per transfer.\n\nReducing this will reduce memory usage but decrease performance.",
              "Provider": "",
              "Default": 8388608,
              "Value": null,
              "ShortOpt": "",
              "Hide": 0,
              "Required": false,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": true
          },
          {
              "Name": "acknowledge_abuse",
              "Help": "Set to allow files which return cannotDownloadAbusiveFile to be downloaded.\n\nIf downloading a file returns the error \"This file has been identified\nas malware or spam and cannot be downloaded\" with the error code\n\"cannotDownloadAbusiveFile\" then supply this flag to rclone to\nindicate you acknowledge the risks of downloading the file and rclone\nwill download it anyway.",
              "Provider": "",
              "Default": false,
              "Value": null,
              "ShortOpt": "",
              "Hide": 0,
              "Required": false,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": true
          },
          {
              "Name": "keep_revision_forever",
              "Help": "Keep new head revision of each file forever.",
              "Provider": "",
              "Default": false,
              "Value": null,
              "ShortOpt": "",
              "Hide": 0,
              "Required": false,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": true
          },
          {
              "Name": "v2_download_min_size",
              "Help": "If Object's are greater, use drive v2 API to download.",
              "Provider": "",
              "Default": -1,
              "Value": null,
              "ShortOpt": "",
              "Hide": 0,
              "Required": false,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": true
          },
          {
              "Name": "pacer_min_sleep",
              "Help": "Minimum time to sleep between API calls.",
              "Provider": "",
              "Default": 100000000,
              "Value": null,
              "ShortOpt": "",
              "Hide": 0,
              "Required": false,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": true
          },
          {
              "Name": "pacer_burst",
              "Help": "Number of API calls to allow without sleeping.",
              "Provider": "",
              "Default": 100,
              "Value": null,
              "ShortOpt": "",
              "Hide": 0,
              "Required": false,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": true
          }
      ]
  },
  {
      "Name": "dropbox",
      "Description": "Dropbox",
      "Prefix": "dropbox",
      "Options": [
          {
              "Name": "client_id",
              "Help": "Dropbox App Client Id\nLeave blank normally.",
              "Provider": "",
              "Default": "",
              "Value": null,
              "ShortOpt": "",
              "Hide": 0,
              "Required": false,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": false
          },
          {
              "Name": "client_secret",
              "Help": "Dropbox App Client Secret\nLeave blank normally.",
              "Provider": "",
              "Default": "",
              "Value": null,
              "ShortOpt": "",
              "Hide": 0,
              "Required": false,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": false
          },
          {
              "Name": "chunk_size",
              "Help": "Upload chunk size. (\u003c 150M).\n\nAny files larger than this will be uploaded in chunks of this size.\n\nNote that chunks are buffered in memory (one at a time) so rclone can\ndeal with retries.  Setting this larger will increase the speed\nslightly (at most 10% for 128MB in tests) at the cost of using more\nmemory.  It can be set smaller if you are tight on memory.",
              "Provider": "",
              "Default": 50331648,
              "Value": null,
              "ShortOpt": "",
              "Hide": 0,
              "Required": false,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": true
          },
          {
              "Name": "impersonate",
              "Help": "Impersonate this user when using a business account.",
              "Provider": "",
              "Default": "",
              "Value": null,
              "ShortOpt": "",
              "Hide": 0,
              "Required": false,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": true
          }
      ]
  },
  {
      "Name": "ftp",
      "Description": "FTP Connection",
      "Prefix": "ftp",
      "Options": [
          {
              "Name": "host",
              "Help": "FTP host to connect to",
              "Provider": "",
              "Default": "",
              "Value": null,
              "Examples": [
                  {
                      "Value": "ftp.example.com",
                      "Help": "Connect to ftp.example.com",
                      "Provider": ""
                  }
              ],
              "ShortOpt": "",
              "Hide": 0,
              "Required": true,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": false
          },
          {
              "Name": "user",
              "Help": "FTP username, leave blank for current username, negative0",
              "Provider": "",
              "Default": "",
              "Value": null,
              "ShortOpt": "",
              "Hide": 0,
              "Required": false,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": false
          },
          {
              "Name": "port",
              "Help": "FTP port, leave blank to use default (21)",
              "Provider": "",
              "Default": "",
              "Value": null,
              "ShortOpt": "",
              "Hide": 0,
              "Required": false,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": false
          },
          {
              "Name": "pass",
              "Help": "FTP password",
              "Provider": "",
              "Default": "",
              "Value": null,
              "ShortOpt": "",
              "Hide": 0,
              "Required": true,
              "IsPassword": true,
              "NoPrefix": false,
              "Advanced": false
          }
      ]
  },
  {
      "Name": "google cloud storage",
      "Description": "Google Cloud Storage (this is not Google Drive)",
      "Prefix": "gcs",
      "Options": [
          {
              "Name": "client_id",
              "Help": "Google Application Client Id\nLeave blank normally.",
              "Provider": "",
              "Default": "",
              "Value": null,
              "ShortOpt": "",
              "Hide": 0,
              "Required": false,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": false
          },
          {
              "Name": "client_secret",
              "Help": "Google Application Client Secret\nLeave blank normally.",
              "Provider": "",
              "Default": "",
              "Value": null,
              "ShortOpt": "",
              "Hide": 0,
              "Required": false,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": false
          },
          {
              "Name": "project_number",
              "Help": "Project number.\nOptional - needed only for list/create/delete buckets - see your developer console.",
              "Provider": "",
              "Default": "",
              "Value": null,
              "ShortOpt": "",
              "Hide": 0,
              "Required": false,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": false
          },
          {
              "Name": "service_account_file",
              "Help": "Service Account Credentials JSON file path\nLeave blank normally.\nNeeded only if you want use SA instead of interactive login.",
              "Provider": "",
              "Default": "",
              "Value": null,
              "ShortOpt": "",
              "Hide": 0,
              "Required": false,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": false
          },
          {
              "Name": "service_account_credentials",
              "Help": "Service Account Credentials JSON blob\nLeave blank normally.\nNeeded only if you want use SA instead of interactive login.",
              "Provider": "",
              "Default": "",
              "Value": null,
              "ShortOpt": "",
              "Hide": 3,
              "Required": false,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": false
          },
          {
              "Name": "object_acl",
              "Help": "Access Control List for new objects.",
              "Provider": "",
              "Default": "",
              "Value": null,
              "Examples": [
                  {
                      "Value": "authenticatedRead",
                      "Help": "Object owner gets OWNER access, and all Authenticated Users get READER access.",
                      "Provider": ""
                  },
                  {
                      "Value": "bucketOwnerFullControl",
                      "Help": "Object owner gets OWNER access, and project team owners get OWNER access.",
                      "Provider": ""
                  },
                  {
                      "Value": "bucketOwnerRead",
                      "Help": "Object owner gets OWNER access, and project team owners get READER access.",
                      "Provider": ""
                  },
                  {
                      "Value": "private",
                      "Help": "Object owner gets OWNER access [default if left blank].",
                      "Provider": ""
                  },
                  {
                      "Value": "projectPrivate",
                      "Help": "Object owner gets OWNER access, and project team members get access according to their roles.",
                      "Provider": ""
                  },
                  {
                      "Value": "publicRead",
                      "Help": "Object owner gets OWNER access, and all Users get READER access.",
                      "Provider": ""
                  }
              ],
              "ShortOpt": "",
              "Hide": 0,
              "Required": false,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": false
          },
          {
              "Name": "bucket_acl",
              "Help": "Access Control List for new buckets.",
              "Provider": "",
              "Default": "",
              "Value": null,
              "Examples": [
                  {
                      "Value": "authenticatedRead",
                      "Help": "Project team owners get OWNER access, and all Authenticated Users get READER access.",
                      "Provider": ""
                  },
                  {
                      "Value": "private",
                      "Help": "Project team owners get OWNER access [default if left blank].",
                      "Provider": ""
                  },
                  {
                      "Value": "projectPrivate",
                      "Help": "Project team members get access according to their roles.",
                      "Provider": ""
                  },
                  {
                      "Value": "publicRead",
                      "Help": "Project team owners get OWNER access, and all Users get READER access.",
                      "Provider": ""
                  },
                  {
                      "Value": "publicReadWrite",
                      "Help": "Project team owners get OWNER access, and all Users get WRITER access.",
                      "Provider": ""
                  }
              ],
              "ShortOpt": "",
              "Hide": 0,
              "Required": false,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": false
          },
          {
              "Name": "location",
              "Help": "Location for the newly created buckets.",
              "Provider": "",
              "Default": "",
              "Value": null,
              "Examples": [
                  {
                      "Value": "",
                      "Help": "Empty for default location (US).",
                      "Provider": ""
                  },
                  {
                      "Value": "asia",
                      "Help": "Multi-regional location for Asia.",
                      "Provider": ""
                  },
                  {
                      "Value": "eu",
                      "Help": "Multi-regional location for Europe.",
                      "Provider": ""
                  },
                  {
                      "Value": "us",
                      "Help": "Multi-regional location for United States.",
                      "Provider": ""
                  },
                  {
                      "Value": "asia-east1",
                      "Help": "Taiwan.",
                      "Provider": ""
                  },
                  {
                      "Value": "asia-east2",
                      "Help": "Hong Kong.",
                      "Provider": ""
                  },
                  {
                      "Value": "asia-northeast1",
                      "Help": "Tokyo.",
                      "Provider": ""
                  },
                  {
                      "Value": "asia-south1",
                      "Help": "Mumbai.",
                      "Provider": ""
                  },
                  {
                      "Value": "asia-southeast1",
                      "Help": "Singapore.",
                      "Provider": ""
                  },
                  {
                      "Value": "australia-southeast1",
                      "Help": "Sydney.",
                      "Provider": ""
                  },
                  {
                      "Value": "europe-north1",
                      "Help": "Finland.",
                      "Provider": ""
                  },
                  {
                      "Value": "europe-west1",
                      "Help": "Belgium.",
                      "Provider": ""
                  },
                  {
                      "Value": "europe-west2",
                      "Help": "London.",
                      "Provider": ""
                  },
                  {
                      "Value": "europe-west3",
                      "Help": "Frankfurt.",
                      "Provider": ""
                  },
                  {
                      "Value": "europe-west4",
                      "Help": "Netherlands.",
                      "Provider": ""
                  },
                  {
                      "Value": "us-central1",
                      "Help": "Iowa.",
                      "Provider": ""
                  },
                  {
                      "Value": "us-east1",
                      "Help": "South Carolina.",
                      "Provider": ""
                  },
                  {
                      "Value": "us-east4",
                      "Help": "Northern Virginia.",
                      "Provider": ""
                  },
                  {
                      "Value": "us-west1",
                      "Help": "Oregon.",
                      "Provider": ""
                  },
                  {
                      "Value": "us-west2",
                      "Help": "California.",
                      "Provider": ""
                  }
              ],
              "ShortOpt": "",
              "Hide": 0,
              "Required": false,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": false
          },
          {
              "Name": "storage_class",
              "Help": "The storage class to use when storing objects in Google Cloud Storage.",
              "Provider": "",
              "Default": "",
              "Value": null,
              "Examples": [
                  {
                      "Value": "",
                      "Help": "Default",
                      "Provider": ""
                  },
                  {
                      "Value": "MULTI_REGIONAL",
                      "Help": "Multi-regional storage class",
                      "Provider": ""
                  },
                  {
                      "Value": "REGIONAL",
                      "Help": "Regional storage class",
                      "Provider": ""
                  },
                  {
                      "Value": "NEARLINE",
                      "Help": "Nearline storage class",
                      "Provider": ""
                  },
                  {
                      "Value": "COLDLINE",
                      "Help": "Coldline storage class",
                      "Provider": ""
                  },
                  {
                      "Value": "DURABLE_REDUCED_AVAILABILITY",
                      "Help": "Durable reduced availability storage class",
                      "Provider": ""
                  }
              ],
              "ShortOpt": "",
              "Hide": 0,
              "Required": false,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": false
          }
      ]
  },
  {
      "Name": "http",
      "Description": "http Connection",
      "Prefix": "http",
      "Options": [
          {
              "Name": "url",
              "Help": "URL of http host to connect to",
              "Provider": "",
              "Default": "",
              "Value": null,
              "Examples": [
                  {
                      "Value": "https://example.com",
                      "Help": "Connect to example.com",
                      "Provider": ""
                  },
                  {
                      "Value": "https://user:pass@example.com",
                      "Help": "Connect to example.com using a username and password",
                      "Provider": ""
                  }
              ],
              "ShortOpt": "",
              "Hide": 0,
              "Required": true,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": false
          }
      ]
  },
  {
      "Name": "swift",
      "Description": "Openstack Swift (Rackspace Cloud Files, Memset Memstore, OVH)",
      "Prefix": "swift",
      "Options": [
          {
              "Name": "env_auth",
              "Help": "Get swift credentials from environment variables in standard OpenStack form.",
              "Provider": "",
              "Default": false,
              "Value": null,
              "Examples": [
                  {
                      "Value": "false",
                      "Help": "Enter swift credentials in the next step",
                      "Provider": ""
                  },
                  {
                      "Value": "true",
                      "Help": "Get swift credentials from environment vars. Leave other fields blank if using this.",
                      "Provider": ""
                  }
              ],
              "ShortOpt": "",
              "Hide": 0,
              "Required": false,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": false
          },
          {
              "Name": "user",
              "Help": "User name to log in (OS_USERNAME).",
              "Provider": "",
              "Default": "",
              "Value": null,
              "ShortOpt": "",
              "Hide": 0,
              "Required": false,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": false
          },
          {
              "Name": "key",
              "Help": "API key or password (OS_PASSWORD).",
              "Provider": "",
              "Default": "",
              "Value": null,
              "ShortOpt": "",
              "Hide": 0,
              "Required": false,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": false
          },
          {
              "Name": "auth",
              "Help": "Authentication URL for server (OS_AUTH_URL).",
              "Provider": "",
              "Default": "",
              "Value": null,
              "Examples": [
                  {
                      "Value": "https://auth.api.rackspacecloud.com/v1.0",
                      "Help": "Rackspace US",
                      "Provider": ""
                  },
                  {
                      "Value": "https://lon.auth.api.rackspacecloud.com/v1.0",
                      "Help": "Rackspace UK",
                      "Provider": ""
                  },
                  {
                      "Value": "https://identity.api.rackspacecloud.com/v2.0",
                      "Help": "Rackspace v2",
                      "Provider": ""
                  },
                  {
                      "Value": "https://auth.storage.memset.com/v1.0",
                      "Help": "Memset Memstore UK",
                      "Provider": ""
                  },
                  {
                      "Value": "https://auth.storage.memset.com/v2.0",
                      "Help": "Memset Memstore UK v2",
                      "Provider": ""
                  },
                  {
                      "Value": "https://auth.cloud.ovh.net/v2.0",
                      "Help": "OVH",
                      "Provider": ""
                  }
              ],
              "ShortOpt": "",
              "Hide": 0,
              "Required": false,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": false
          },
          {
              "Name": "user_id",
              "Help": "User ID to log in - optional - most swift systems use user and leave this blank (v3 auth) (OS_USER_ID).",
              "Provider": "",
              "Default": "",
              "Value": null,
              "ShortOpt": "",
              "Hide": 0,
              "Required": false,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": false
          },
          {
              "Name": "domain",
              "Help": "User domain - optional (v3 auth) (OS_USER_DOMAIN_NAME)",
              "Provider": "",
              "Default": "",
              "Value": null,
              "ShortOpt": "",
              "Hide": 0,
              "Required": false,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": false
          },
          {
              "Name": "tenant",
              "Help": "Tenant name - optional for v1 auth, this or tenant_id required otherwise (OS_TENANT_NAME or OS_PROJECT_NAME)",
              "Provider": "",
              "Default": "",
              "Value": null,
              "ShortOpt": "",
              "Hide": 0,
              "Required": false,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": false
          },
          {
              "Name": "tenant_id",
              "Help": "Tenant ID - optional for v1 auth, this or tenant required otherwise (OS_TENANT_ID)",
              "Provider": "",
              "Default": "",
              "Value": null,
              "ShortOpt": "",
              "Hide": 0,
              "Required": false,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": false
          },
          {
              "Name": "tenant_domain",
              "Help": "Tenant domain - optional (v3 auth) (OS_PROJECT_DOMAIN_NAME)",
              "Provider": "",
              "Default": "",
              "Value": null,
              "ShortOpt": "",
              "Hide": 0,
              "Required": false,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": false
          },
          {
              "Name": "region",
              "Help": "Region name - optional (OS_REGION_NAME)",
              "Provider": "",
              "Default": "",
              "Value": null,
              "ShortOpt": "",
              "Hide": 0,
              "Required": false,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": false
          },
          {
              "Name": "storage_url",
              "Help": "Storage URL - optional (OS_STORAGE_URL)",
              "Provider": "",
              "Default": "",
              "Value": null,
              "ShortOpt": "",
              "Hide": 0,
              "Required": false,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": false
          },
          {
              "Name": "auth_token",
              "Help": "Auth Token from alternate authentication - optional (OS_AUTH_TOKEN)",
              "Provider": "",
              "Default": "",
              "Value": null,
              "ShortOpt": "",
              "Hide": 0,
              "Required": false,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": false
          },
          {
              "Name": "application_credential_id",
              "Help": "Application Credential ID (OS_APPLICATION_CREDENTIAL_ID)",
              "Provider": "",
              "Default": "",
              "Value": null,
              "ShortOpt": "",
              "Hide": 0,
              "Required": false,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": false
          },
          {
              "Name": "application_credential_name",
              "Help": "Application Credential Name (OS_APPLICATION_CREDENTIAL_NAME)",
              "Provider": "",
              "Default": "",
              "Value": null,
              "ShortOpt": "",
              "Hide": 0,
              "Required": false,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": false
          },
          {
              "Name": "application_credential_secret",
              "Help": "Application Credential Secret (OS_APPLICATION_CREDENTIAL_SECRET)",
              "Provider": "",
              "Default": "",
              "Value": null,
              "ShortOpt": "",
              "Hide": 0,
              "Required": false,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": false
          },
          {
              "Name": "auth_version",
              "Help": "AuthVersion - optional - set to (1,2,3) if your auth URL has no version (ST_AUTH_VERSION)",
              "Provider": "",
              "Default": 0,
              "Value": null,
              "ShortOpt": "",
              "Hide": 0,
              "Required": false,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": false
          },
          {
              "Name": "endpoint_type",
              "Help": "Endpoint type to choose from the service catalogue (OS_ENDPOINT_TYPE)",
              "Provider": "",
              "Default": "public",
              "Value": null,
              "Examples": [
                  {
                      "Value": "public",
                      "Help": "Public (default, choose this if not sure)",
                      "Provider": ""
                  },
                  {
                      "Value": "internal",
                      "Help": "Internal (use internal service net)",
                      "Provider": ""
                  },
                  {
                      "Value": "admin",
                      "Help": "Admin",
                      "Provider": ""
                  }
              ],
              "ShortOpt": "",
              "Hide": 0,
              "Required": false,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": false
          },
          {
              "Name": "storage_policy",
              "Help": "The storage policy to use when creating a new container\n\nThis applies the specified storage policy when creating a new\ncontainer. The policy cannot be changed afterwards. The allowed\nconfiguration values and their meaning depend on your Swift storage\nprovider.",
              "Provider": "",
              "Default": "",
              "Value": null,
              "Examples": [
                  {
                      "Value": "",
                      "Help": "Default",
                      "Provider": ""
                  },
                  {
                      "Value": "pcs",
                      "Help": "OVH Public Cloud Storage",
                      "Provider": ""
                  },
                  {
                      "Value": "pca",
                      "Help": "OVH Public Cloud Archive",
                      "Provider": ""
                  }
              ],
              "ShortOpt": "",
              "Hide": 0,
              "Required": false,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": false
          },
          {
              "Name": "chunk_size",
              "Help": "Above this size files will be chunked into a _segments container.\n\nAbove this size files will be chunked into a _segments container.  The\ndefault for this is 5GB which is its maximum value.",
              "Provider": "",
              "Default": 5368709120,
              "Value": null,
              "ShortOpt": "",
              "Hide": 0,
              "Required": false,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": true
          },
          {
              "Name": "no_chunk",
              "Help": "Don't chunk files during streaming upload.\n\nWhen doing streaming uploads (eg using rcat or mount) setting this\nflag will cause the swift backend to not upload chunked files.\n\nThis will limit the maximum upload size to 5GB. However non chunked\nfiles are easier to deal with and have an MD5SUM.\n\nRclone will still chunk files bigger than chunk_size when doing normal\ncopy operations.",
              "Provider": "",
              "Default": false,
              "Value": null,
              "ShortOpt": "",
              "Hide": 0,
              "Required": false,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": true
          }
      ]
  },
  {
      "Name": "hubic",
      "Description": "Hubic",
      "Prefix": "hubic",
      "Options": [
          {
              "Name": "client_id",
              "Help": "Hubic Client Id\nLeave blank normally.",
              "Provider": "",
              "Default": "",
              "Value": null,
              "ShortOpt": "",
              "Hide": 0,
              "Required": false,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": false
          },
          {
              "Name": "client_secret",
              "Help": "Hubic Client Secret\nLeave blank normally.",
              "Provider": "",
              "Default": "",
              "Value": null,
              "ShortOpt": "",
              "Hide": 0,
              "Required": false,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": false
          },
          {
              "Name": "chunk_size",
              "Help": "Above this size files will be chunked into a _segments container.\n\nAbove this size files will be chunked into a _segments container.  The\ndefault for this is 5GB which is its maximum value.",
              "Provider": "",
              "Default": 5368709120,
              "Value": null,
              "ShortOpt": "",
              "Hide": 0,
              "Required": false,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": true
          },
          {
              "Name": "no_chunk",
              "Help": "Don't chunk files during streaming upload.\n\nWhen doing streaming uploads (eg using rcat or mount) setting this\nflag will cause the swift backend to not upload chunked files.\n\nThis will limit the maximum upload size to 5GB. However non chunked\nfiles are easier to deal with and have an MD5SUM.\n\nRclone will still chunk files bigger than chunk_size when doing normal\ncopy operations.",
              "Provider": "",
              "Default": false,
              "Value": null,
              "ShortOpt": "",
              "Hide": 0,
              "Required": false,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": true
          }
      ]
  },
  {
      "Name": "jottacloud",
      "Description": "JottaCloud",
      "Prefix": "jottacloud",
      "Options": [
          {
              "Name": "user",
              "Help": "User Name:",
              "Provider": "",
              "Default": "",
              "Value": null,
              "ShortOpt": "",
              "Hide": 0,
              "Required": false,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": false
          },
          {
              "Name": "mountpoint",
              "Help": "The mountpoint to use.",
              "Provider": "",
              "Default": "",
              "Value": null,
              "Examples": [
                  {
                      "Value": "Sync",
                      "Help": "Will be synced by the official client.",
                      "Provider": ""
                  },
                  {
                      "Value": "Archive",
                      "Help": "Archive",
                      "Provider": ""
                  }
              ],
              "ShortOpt": "",
              "Hide": 0,
              "Required": true,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": false
          },
          {
              "Name": "md5_memory_limit",
              "Help": "Files bigger than this will be cached on disk to calculate the MD5 if required.",
              "Provider": "",
              "Default": 10485760,
              "Value": null,
              "ShortOpt": "",
              "Hide": 0,
              "Required": false,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": true
          },
          {
              "Name": "hard_delete",
              "Help": "Delete files permanently rather than putting them into the trash.",
              "Provider": "",
              "Default": false,
              "Value": null,
              "ShortOpt": "",
              "Hide": 0,
              "Required": false,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": true
          },
          {
              "Name": "unlink",
              "Help": "Remove existing public link to file/folder with link command rather than creating.\nDefault is false, meaning link command will create or retrieve public link.",
              "Provider": "",
              "Default": false,
              "Value": null,
              "ShortOpt": "",
              "Hide": 0,
              "Required": false,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": true
          },
          {
              "Name": "upload_resume_limit",
              "Help": "Files bigger than this can be resumed if the upload fail's.",
              "Provider": "",
              "Default": 10485760,
              "Value": null,
              "ShortOpt": "",
              "Hide": 0,
              "Required": false,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": true
          }
      ]
  },
  {
      "Name": "local",
      "Description": "Local Disk",
      "Prefix": "local",
      "Options": [
          {
              "Name": "nounc",
              "Help": "Disable UNC (long path names) conversion on Windows",
              "Provider": "",
              "Default": "",
              "Value": null,
              "Examples": [
                  {
                      "Value": "true",
                      "Help": "Disables long file names",
                      "Provider": ""
                  }
              ],
              "ShortOpt": "",
              "Hide": 0,
              "Required": false,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": false
          },
          {
              "Name": "copy_links",
              "Help": "Follow symlinks and copy the pointed to item.",
              "Provider": "",
              "Default": false,
              "Value": null,
              "ShortOpt": "L",
              "Hide": 0,
              "Required": false,
              "IsPassword": false,
              "NoPrefix": true,
              "Advanced": true
          },
          {
              "Name": "links",
              "Help": "Translate symlinks to/from regular files with a '.rclonelink' extension",
              "Provider": "",
              "Default": false,
              "Value": null,
              "ShortOpt": "l",
              "Hide": 0,
              "Required": false,
              "IsPassword": false,
              "NoPrefix": true,
              "Advanced": true
          },
          {
              "Name": "skip_links",
              "Help": "Don't warn about skipped symlinks.\nThis flag disables warning messages on skipped symlinks or junction\npoints, as you explicitly acknowledge that they should be skipped.",
              "Provider": "",
              "Default": false,
              "Value": null,
              "ShortOpt": "",
              "Hide": 0,
              "Required": false,
              "IsPassword": false,
              "NoPrefix": true,
              "Advanced": true
          },
          {
              "Name": "no_unicode_normalization",
              "Help": "Don't apply unicode normalization to paths and filenames (Deprecated)\n\nThis flag is deprecated now.  Rclone no longer normalizes unicode file\nnames, but it compares them with unicode normalization in the sync\nroutine instead.",
              "Provider": "",
              "Default": false,
              "Value": null,
              "ShortOpt": "",
              "Hide": 0,
              "Required": false,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": true
          },
          {
              "Name": "no_check_updated",
              "Help": "Don't check to see if the files change during upload\n\nNormally rclone checks the size and modification time of files as they\nare being uploaded and aborts with a message which starts \"can't copy\n- source file is being updated\" if the file changes during upload.\n\nHowever on some file systems this modification time check may fail (eg\n[Glusterfs #2206](https://github.com/ncw/rclone/issues/2206)) so this\ncheck can be disabled with this flag.",
              "Provider": "",
              "Default": false,
              "Value": null,
              "ShortOpt": "",
              "Hide": 0,
              "Required": false,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": true
          },
          {
              "Name": "one_file_system",
              "Help": "Don't cross filesystem boundaries (unix/macOS only).",
              "Provider": "",
              "Default": false,
              "Value": null,
              "ShortOpt": "x",
              "Hide": 0,
              "Required": false,
              "IsPassword": false,
              "NoPrefix": true,
              "Advanced": true
          }
      ]
  },
  {
      "Name": "mega",
      "Description": "Mega",
      "Prefix": "mega",
      "Options": [
          {
              "Name": "user",
              "Help": "User name",
              "Provider": "",
              "Default": "",
              "Value": null,
              "ShortOpt": "",
              "Hide": 0,
              "Required": true,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": false
          },
          {
              "Name": "pass",
              "Help": "Password.",
              "Provider": "",
              "Default": "",
              "Value": null,
              "ShortOpt": "",
              "Hide": 0,
              "Required": true,
              "IsPassword": true,
              "NoPrefix": false,
              "Advanced": false
          },
          {
              "Name": "debug",
              "Help": "Output more debug from Mega.\n\nIf this flag is set (along with -vv) it will print further debugging\ninformation from the mega backend.",
              "Provider": "",
              "Default": false,
              "Value": null,
              "ShortOpt": "",
              "Hide": 0,
              "Required": false,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": true
          },
          {
              "Name": "hard_delete",
              "Help": "Delete files permanently rather than putting them into the trash.\n\nNormally the mega backend will put all deletions into the trash rather\nthan permanently deleting them.  If you specify this then rclone will\npermanently delete objects instead.",
              "Provider": "",
              "Default": false,
              "Value": null,
              "ShortOpt": "",
              "Hide": 0,
              "Required": false,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": true
          }
      ]
  },
  {
      "Name": "onedrive",
      "Description": "Microsoft OneDrive",
      "Prefix": "onedrive",
      "Options": [
          {
              "Name": "client_id",
              "Help": "Microsoft App Client Id\nLeave blank normally.",
              "Provider": "",
              "Default": "",
              "Value": null,
              "ShortOpt": "",
              "Hide": 0,
              "Required": false,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": false
          },
          {
              "Name": "client_secret",
              "Help": "Microsoft App Client Secret\nLeave blank normally.",
              "Provider": "",
              "Default": "",
              "Value": null,
              "ShortOpt": "",
              "Hide": 0,
              "Required": false,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": false
          },
          {
              "Name": "chunk_size",
              "Help": "Chunk size to upload files with - must be multiple of 320k.\n\nAbove this size files will be chunked - must be multiple of 320k. Note\nthat the chunks will be buffered into memory.",
              "Provider": "",
              "Default": 10485760,
              "Value": null,
              "ShortOpt": "",
              "Hide": 0,
              "Required": false,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": true
          },
          {
              "Name": "drive_id",
              "Help": "The ID of the drive to use",
              "Provider": "",
              "Default": "",
              "Value": null,
              "ShortOpt": "",
              "Hide": 0,
              "Required": false,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": true
          },
          {
              "Name": "drive_type",
              "Help": "The type of the drive ( personal | business | documentLibrary )",
              "Provider": "",
              "Default": "",
              "Value": null,
              "ShortOpt": "",
              "Hide": 0,
              "Required": false,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": true
          },
          {
              "Name": "expose_onenote_files",
              "Help": "Set to make OneNote files show up in directory listings.\n\nBy default rclone will hide OneNote files in directory listings because\noperations like \"Open\" and \"Update\" won't work on them.  But this\nbehaviour may also prevent you from deleting them.  If you want to\ndelete OneNote files or otherwise want them to show up in directory\nlisting, set this option.",
              "Provider": "",
              "Default": false,
              "Value": null,
              "ShortOpt": "",
              "Hide": 0,
              "Required": false,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": true
          }
      ]
  },
  {
      "Name": "opendrive",
      "Description": "OpenDrive",
      "Prefix": "opendrive",
      "Options": [
          {
              "Name": "username",
              "Help": "Username",
              "Provider": "",
              "Default": "",
              "Value": null,
              "ShortOpt": "",
              "Hide": 0,
              "Required": true,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": false
          },
          {
              "Name": "password",
              "Help": "Password.",
              "Provider": "",
              "Default": "",
              "Value": null,
              "ShortOpt": "",
              "Hide": 0,
              "Required": true,
              "IsPassword": true,
              "NoPrefix": false,
              "Advanced": false
          }
      ]
  },
  {
      "Name": "pcloud",
      "Description": "Pcloud",
      "Prefix": "pcloud",
      "Options": [
          {
              "Name": "client_id",
              "Help": "Pcloud App Client Id\nLeave blank normally.",
              "Provider": "",
              "Default": "",
              "Value": null,
              "ShortOpt": "",
              "Hide": 0,
              "Required": false,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": false
          },
          {
              "Name": "client_secret",
              "Help": "Pcloud App Client Secret\nLeave blank normally.",
              "Provider": "",
              "Default": "",
              "Value": null,
              "ShortOpt": "",
              "Hide": 0,
              "Required": false,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": false
          }
      ]
  },
  {
      "Name": "qingstor",
      "Description": "QingCloud Object Storage",
      "Prefix": "qingstor",
      "Options": [
          {
              "Name": "env_auth",
              "Help": "Get QingStor credentials from runtime. Only applies if access_key_id and secret_access_key is blank.",
              "Provider": "",
              "Default": false,
              "Value": null,
              "Examples": [
                  {
                      "Value": "false",
                      "Help": "Enter QingStor credentials in the next step",
                      "Provider": ""
                  },
                  {
                      "Value": "true",
                      "Help": "Get QingStor credentials from the environment (env vars or IAM)",
                      "Provider": ""
                  }
              ],
              "ShortOpt": "",
              "Hide": 0,
              "Required": false,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": false
          },
          {
              "Name": "access_key_id",
              "Help": "QingStor Access Key ID\nLeave blank for anonymous access or runtime credentials.",
              "Provider": "",
              "Default": "",
              "Value": null,
              "ShortOpt": "",
              "Hide": 0,
              "Required": false,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": false
          },
          {
              "Name": "secret_access_key",
              "Help": "QingStor Secret Access Key (password)\nLeave blank for anonymous access or runtime credentials.",
              "Provider": "",
              "Default": "",
              "Value": null,
              "ShortOpt": "",
              "Hide": 0,
              "Required": false,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": false
          },
          {
              "Name": "endpoint",
              "Help": "Enter a endpoint URL to connection QingStor API.\nLeave blank will use the default value \"https://qingstor.com:443\"",
              "Provider": "",
              "Default": "",
              "Value": null,
              "ShortOpt": "",
              "Hide": 0,
              "Required": false,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": false
          },
          {
              "Name": "zone",
              "Help": "Zone to connect to.\nDefault is \"pek3a\".",
              "Provider": "",
              "Default": "",
              "Value": null,
              "Examples": [
                  {
                      "Value": "pek3a",
                      "Help": "The Beijing (China) Three Zone\nNeeds location constraint pek3a.",
                      "Provider": ""
                  },
                  {
                      "Value": "sh1a",
                      "Help": "The Shanghai (China) First Zone\nNeeds location constraint sh1a.",
                      "Provider": ""
                  },
                  {
                      "Value": "gd2a",
                      "Help": "The Guangdong (China) Second Zone\nNeeds location constraint gd2a.",
                      "Provider": ""
                  }
              ],
              "ShortOpt": "",
              "Hide": 0,
              "Required": false,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": false
          },
          {
              "Name": "connection_retries",
              "Help": "Number of connection retries.",
              "Provider": "",
              "Default": 3,
              "Value": null,
              "ShortOpt": "",
              "Hide": 0,
              "Required": false,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": true
          },
          {
              "Name": "upload_cutoff",
              "Help": "Cutoff for switching to chunked upload\n\nAny files larger than this will be uploaded in chunks of chunk_size.\nThe minimum is 0 and the maximum is 5GB.",
              "Provider": "",
              "Default": 209715200,
              "Value": null,
              "ShortOpt": "",
              "Hide": 0,
              "Required": false,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": true
          },
          {
              "Name": "chunk_size",
              "Help": "Chunk size to use for uploading.\n\nWhen uploading files larger than upload_cutoff they will be uploaded\nas multipart uploads using this chunk size.\n\nNote that \"--qingstor-upload-concurrency\" chunks of this size are buffered\nin memory per transfer.\n\nIf you are transferring large files over high speed links and you have\nenough memory, then increasing this will speed up the transfers.",
              "Provider": "",
              "Default": 4194304,
              "Value": null,
              "ShortOpt": "",
              "Hide": 0,
              "Required": false,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": true
          },
          {
              "Name": "upload_concurrency",
              "Help": "Concurrency for multipart uploads.\n\nThis is the number of chunks of the same file that are uploaded\nconcurrently.\n\nNB if you set this to \u003e 1 then the checksums of multpart uploads\nbecome corrupted (the uploads themselves are not corrupted though).\n\nIf you are uploading small numbers of large file over high speed link\nand these uploads do not fully utilize your bandwidth, then increasing\nthis may help to speed up the transfers.",
              "Provider": "",
              "Default": 1,
              "Value": null,
              "ShortOpt": "",
              "Hide": 0,
              "Required": false,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": true
          }
      ]
  },
  {
      "Name": "s3",
      "Description": "Amazon S3 Compliant Storage Provider (AWS, Alibaba, Ceph, Digital Ocean, Dreamhost, IBM COS, Minio, etc)",
      "Prefix": "s3",
      "Options": [
          {
              "Name": "provider",
              "Help": "Choose your S3 provider.",
              "Provider": "",
              "Default": "",
              "Value": null,
              "Examples": [
                  {
                      "Value": "AWS",
                      "Help": "Amazon Web Services (AWS) S3",
                      "Provider": ""
                  },
                  {
                      "Value": "Alibaba",
                      "Help": "Alibaba Cloud Object Storage System (OSS) formerly Aliyun",
                      "Provider": ""
                  },
                  {
                      "Value": "Ceph",
                      "Help": "Ceph Object Storage",
                      "Provider": ""
                  },
                  {
                      "Value": "DigitalOcean",
                      "Help": "Digital Ocean Spaces",
                      "Provider": ""
                  },
                  {
                      "Value": "Dreamhost",
                      "Help": "Dreamhost DreamObjects",
                      "Provider": ""
                  },
                  {
                      "Value": "IBMCOS",
                      "Help": "IBM COS S3",
                      "Provider": ""
                  },
                  {
                      "Value": "Minio",
                      "Help": "Minio Object Storage",
                      "Provider": ""
                  },
                  {
                      "Value": "Netease",
                      "Help": "Netease Object Storage (NOS)",
                      "Provider": ""
                  },
                  {
                      "Value": "Wasabi",
                      "Help": "Wasabi Object Storage",
                      "Provider": ""
                  },
                  {
                      "Value": "Other",
                      "Help": "Any other S3 compatible provider",
                      "Provider": ""
                  }
              ],
              "ShortOpt": "",
              "Hide": 0,
              "Required": false,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": false
          },
          {
              "Name": "env_auth",
              "Help": "Get AWS credentials from runtime (environment variables or EC2/ECS meta data if no env vars).\nOnly applies if access_key_id and secret_access_key is blank.",
              "Provider": "",
              "Default": false,
              "Value": null,
              "Examples": [
                  {
                      "Value": "false",
                      "Help": "Enter AWS credentials in the next step",
                      "Provider": ""
                  },
                  {
                      "Value": "true",
                      "Help": "Get AWS credentials from the environment (env vars or IAM)",
                      "Provider": ""
                  }
              ],
              "ShortOpt": "",
              "Hide": 0,
              "Required": false,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": false
          },
          {
              "Name": "access_key_id",
              "Help": "AWS Access Key ID.\nLeave blank for anonymous access or runtime credentials.",
              "Provider": "",
              "Default": "",
              "Value": null,
              "ShortOpt": "",
              "Hide": 0,
              "Required": false,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": false
          },
          {
              "Name": "secret_access_key",
              "Help": "AWS Secret Access Key (password)\nLeave blank for anonymous access or runtime credentials.",
              "Provider": "",
              "Default": "",
              "Value": null,
              "ShortOpt": "",
              "Hide": 0,
              "Required": false,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": false
          },
          {
              "Name": "region",
              "Help": "Region to connect to.",
              "Provider": "AWS",
              "Default": "",
              "Value": null,
              "Examples": [
                  {
                      "Value": "us-east-1",
                      "Help": "The default endpoint - a good choice if you are unsure.\nUS Region, Northern Virginia or Pacific Northwest.\nLeave location constraint empty.",
                      "Provider": ""
                  },
                  {
                      "Value": "us-east-2",
                      "Help": "US East (Ohio) Region\nNeeds location constraint us-east-2.",
                      "Provider": ""
                  },
                  {
                      "Value": "us-west-2",
                      "Help": "US West (Oregon) Region\nNeeds location constraint us-west-2.",
                      "Provider": ""
                  },
                  {
                      "Value": "us-west-1",
                      "Help": "US West (Northern California) Region\nNeeds location constraint us-west-1.",
                      "Provider": ""
                  },
                  {
                      "Value": "ca-central-1",
                      "Help": "Canada (Central) Region\nNeeds location constraint ca-central-1.",
                      "Provider": ""
                  },
                  {
                      "Value": "eu-west-1",
                      "Help": "EU (Ireland) Region\nNeeds location constraint EU or eu-west-1.",
                      "Provider": ""
                  },
                  {
                      "Value": "eu-west-2",
                      "Help": "EU (London) Region\nNeeds location constraint eu-west-2.",
                      "Provider": ""
                  },
                  {
                      "Value": "eu-north-1",
                      "Help": "EU (Stockholm) Region\nNeeds location constraint eu-north-1.",
                      "Provider": ""
                  },
                  {
                      "Value": "eu-central-1",
                      "Help": "EU (Frankfurt) Region\nNeeds location constraint eu-central-1.",
                      "Provider": ""
                  },
                  {
                      "Value": "ap-southeast-1",
                      "Help": "Asia Pacific (Singapore) Region\nNeeds location constraint ap-southeast-1.",
                      "Provider": ""
                  },
                  {
                      "Value": "ap-southeast-2",
                      "Help": "Asia Pacific (Sydney) Region\nNeeds location constraint ap-southeast-2.",
                      "Provider": ""
                  },
                  {
                      "Value": "ap-northeast-1",
                      "Help": "Asia Pacific (Tokyo) Region\nNeeds location constraint ap-northeast-1.",
                      "Provider": ""
                  },
                  {
                      "Value": "ap-northeast-2",
                      "Help": "Asia Pacific (Seoul)\nNeeds location constraint ap-northeast-2.",
                      "Provider": ""
                  },
                  {
                      "Value": "ap-south-1",
                      "Help": "Asia Pacific (Mumbai)\nNeeds location constraint ap-south-1.",
                      "Provider": ""
                  },
                  {
                      "Value": "sa-east-1",
                      "Help": "South America (Sao Paulo) Region\nNeeds location constraint sa-east-1.",
                      "Provider": ""
                  }
              ],
              "ShortOpt": "",
              "Hide": 0,
              "Required": false,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": false
          },
          {
              "Name": "region",
              "Help": "Region to connect to.\nLeave blank if you are using an S3 clone and you don't have a region.",
              "Provider": "!AWS,Alibaba",
              "Default": "",
              "Value": null,
              "Examples": [
                  {
                      "Value": "",
                      "Help": "Use this if unsure. Will use v4 signatures and an empty region.",
                      "Provider": ""
                  },
                  {
                      "Value": "other-v2-signature",
                      "Help": "Use this only if v4 signatures don't work, eg pre Jewel/v10 CEPH.",
                      "Provider": ""
                  }
              ],
              "ShortOpt": "",
              "Hide": 0,
              "Required": false,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": false
          },
          {
              "Name": "endpoint",
              "Help": "Endpoint for S3 API.\nLeave blank if using AWS to use the default endpoint for the region.",
              "Provider": "AWS",
              "Default": "",
              "Value": null,
              "ShortOpt": "",
              "Hide": 0,
              "Required": false,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": false
          },
          {
              "Name": "endpoint",
              "Help": "Endpoint for IBM COS S3 API.\nSpecify if using an IBM COS On Premise.",
              "Provider": "IBMCOS",
              "Default": "",
              "Value": null,
              "Examples": [
                  {
                      "Value": "s3-api.us-geo.objectstorage.softlayer.net",
                      "Help": "US Cross Region Endpoint",
                      "Provider": ""
                  },
                  {
                      "Value": "s3-api.dal.us-geo.objectstorage.softlayer.net",
                      "Help": "US Cross Region Dallas Endpoint",
                      "Provider": ""
                  },
                  {
                      "Value": "s3-api.wdc-us-geo.objectstorage.softlayer.net",
                      "Help": "US Cross Region Washington DC Endpoint",
                      "Provider": ""
                  },
                  {
                      "Value": "s3-api.sjc-us-geo.objectstorage.softlayer.net",
                      "Help": "US Cross Region San Jose Endpoint",
                      "Provider": ""
                  },
                  {
                      "Value": "s3-api.us-geo.objectstorage.service.networklayer.com",
                      "Help": "US Cross Region Private Endpoint",
                      "Provider": ""
                  },
                  {
                      "Value": "s3-api.dal-us-geo.objectstorage.service.networklayer.com",
                      "Help": "US Cross Region Dallas Private Endpoint",
                      "Provider": ""
                  },
                  {
                      "Value": "s3-api.wdc-us-geo.objectstorage.service.networklayer.com",
                      "Help": "US Cross Region Washington DC Private Endpoint",
                      "Provider": ""
                  },
                  {
                      "Value": "s3-api.sjc-us-geo.objectstorage.service.networklayer.com",
                      "Help": "US Cross Region San Jose Private Endpoint",
                      "Provider": ""
                  },
                  {
                      "Value": "s3.us-east.objectstorage.softlayer.net",
                      "Help": "US Region East Endpoint",
                      "Provider": ""
                  },
                  {
                      "Value": "s3.us-east.objectstorage.service.networklayer.com",
                      "Help": "US Region East Private Endpoint",
                      "Provider": ""
                  },
                  {
                      "Value": "s3.us-south.objectstorage.softlayer.net",
                      "Help": "US Region South Endpoint",
                      "Provider": ""
                  },
                  {
                      "Value": "s3.us-south.objectstorage.service.networklayer.com",
                      "Help": "US Region South Private Endpoint",
                      "Provider": ""
                  },
                  {
                      "Value": "s3.eu-geo.objectstorage.softlayer.net",
                      "Help": "EU Cross Region Endpoint",
                      "Provider": ""
                  },
                  {
                      "Value": "s3.fra-eu-geo.objectstorage.softlayer.net",
                      "Help": "EU Cross Region Frankfurt Endpoint",
                      "Provider": ""
                  },
                  {
                      "Value": "s3.mil-eu-geo.objectstorage.softlayer.net",
                      "Help": "EU Cross Region Milan Endpoint",
                      "Provider": ""
                  },
                  {
                      "Value": "s3.ams-eu-geo.objectstorage.softlayer.net",
                      "Help": "EU Cross Region Amsterdam Endpoint",
                      "Provider": ""
                  },
                  {
                      "Value": "s3.eu-geo.objectstorage.service.networklayer.com",
                      "Help": "EU Cross Region Private Endpoint",
                      "Provider": ""
                  },
                  {
                      "Value": "s3.fra-eu-geo.objectstorage.service.networklayer.com",
                      "Help": "EU Cross Region Frankfurt Private Endpoint",
                      "Provider": ""
                  },
                  {
                      "Value": "s3.mil-eu-geo.objectstorage.service.networklayer.com",
                      "Help": "EU Cross Region Milan Private Endpoint",
                      "Provider": ""
                  },
                  {
                      "Value": "s3.ams-eu-geo.objectstorage.service.networklayer.com",
                      "Help": "EU Cross Region Amsterdam Private Endpoint",
                      "Provider": ""
                  },
                  {
                      "Value": "s3.eu-gb.objectstorage.softlayer.net",
                      "Help": "Great Britain Endpoint",
                      "Provider": ""
                  },
                  {
                      "Value": "s3.eu-gb.objectstorage.service.networklayer.com",
                      "Help": "Great Britain Private Endpoint",
                      "Provider": ""
                  },
                  {
                      "Value": "s3.ap-geo.objectstorage.softlayer.net",
                      "Help": "APAC Cross Regional Endpoint",
                      "Provider": ""
                  },
                  {
                      "Value": "s3.tok-ap-geo.objectstorage.softlayer.net",
                      "Help": "APAC Cross Regional Tokyo Endpoint",
                      "Provider": ""
                  },
                  {
                      "Value": "s3.hkg-ap-geo.objectstorage.softlayer.net",
                      "Help": "APAC Cross Regional HongKong Endpoint",
                      "Provider": ""
                  },
                  {
                      "Value": "s3.seo-ap-geo.objectstorage.softlayer.net",
                      "Help": "APAC Cross Regional Seoul Endpoint",
                      "Provider": ""
                  },
                  {
                      "Value": "s3.ap-geo.objectstorage.service.networklayer.com",
                      "Help": "APAC Cross Regional Private Endpoint",
                      "Provider": ""
                  },
                  {
                      "Value": "s3.tok-ap-geo.objectstorage.service.networklayer.com",
                      "Help": "APAC Cross Regional Tokyo Private Endpoint",
                      "Provider": ""
                  },
                  {
                      "Value": "s3.hkg-ap-geo.objectstorage.service.networklayer.com",
                      "Help": "APAC Cross Regional HongKong Private Endpoint",
                      "Provider": ""
                  },
                  {
                      "Value": "s3.seo-ap-geo.objectstorage.service.networklayer.com",
                      "Help": "APAC Cross Regional Seoul Private Endpoint",
                      "Provider": ""
                  },
                  {
                      "Value": "s3.mel01.objectstorage.softlayer.net",
                      "Help": "Melbourne Single Site Endpoint",
                      "Provider": ""
                  },
                  {
                      "Value": "s3.mel01.objectstorage.service.networklayer.com",
                      "Help": "Melbourne Single Site Private Endpoint",
                      "Provider": ""
                  },
                  {
                      "Value": "s3.tor01.objectstorage.softlayer.net",
                      "Help": "Toronto Single Site Endpoint",
                      "Provider": ""
                  },
                  {
                      "Value": "s3.tor01.objectstorage.service.networklayer.com",
                      "Help": "Toronto Single Site Private Endpoint",
                      "Provider": ""
                  }
              ],
              "ShortOpt": "",
              "Hide": 0,
              "Required": false,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": false
          },
          {
              "Name": "endpoint",
              "Help": "Endpoint for OSS API.",
              "Provider": "Alibaba",
              "Default": "",
              "Value": null,
              "Examples": [
                  {
                      "Value": "oss-cn-hangzhou.aliyuncs.com",
                      "Help": "East China 1 (Hangzhou)",
                      "Provider": ""
                  },
                  {
                      "Value": "oss-cn-shanghai.aliyuncs.com",
                      "Help": "East China 2 (Shanghai)",
                      "Provider": ""
                  },
                  {
                      "Value": "oss-cn-qingdao.aliyuncs.com",
                      "Help": "North China 1 (Qingdao)",
                      "Provider": ""
                  },
                  {
                      "Value": "oss-cn-beijing.aliyuncs.com",
                      "Help": "North China 2 (Beijing)",
                      "Provider": ""
                  },
                  {
                      "Value": "oss-cn-zhangjiakou.aliyuncs.com",
                      "Help": "North China 3 (Zhangjiakou)",
                      "Provider": ""
                  },
                  {
                      "Value": "oss-cn-huhehaote.aliyuncs.com",
                      "Help": "North China 5 (Huhehaote)",
                      "Provider": ""
                  },
                  {
                      "Value": "oss-cn-shenzhen.aliyuncs.com",
                      "Help": "South China 1 (Shenzhen)",
                      "Provider": ""
                  },
                  {
                      "Value": "oss-cn-hongkong.aliyuncs.com",
                      "Help": "Hong Kong (Hong Kong)",
                      "Provider": ""
                  },
                  {
                      "Value": "oss-us-west-1.aliyuncs.com",
                      "Help": "US West 1 (Silicon Valley)",
                      "Provider": ""
                  },
                  {
                      "Value": "oss-us-east-1.aliyuncs.com",
                      "Help": "US East 1 (Virginia)",
                      "Provider": ""
                  },
                  {
                      "Value": "oss-ap-southeast-1.aliyuncs.com",
                      "Help": "Southeast Asia Southeast 1 (Singapore)",
                      "Provider": ""
                  },
                  {
                      "Value": "oss-ap-southeast-2.aliyuncs.com",
                      "Help": "Asia Pacific Southeast 2 (Sydney)",
                      "Provider": ""
                  },
                  {
                      "Value": "oss-ap-southeast-3.aliyuncs.com",
                      "Help": "Southeast Asia Southeast 3 (Kuala Lumpur)",
                      "Provider": ""
                  },
                  {
                      "Value": "oss-ap-southeast-5.aliyuncs.com",
                      "Help": "Asia Pacific Southeast 5 (Jakarta)",
                      "Provider": ""
                  },
                  {
                      "Value": "oss-ap-northeast-1.aliyuncs.com",
                      "Help": "Asia Pacific Northeast 1 (Japan)",
                      "Provider": ""
                  },
                  {
                      "Value": "oss-ap-south-1.aliyuncs.com",
                      "Help": "Asia Pacific South 1 (Mumbai)",
                      "Provider": ""
                  },
                  {
                      "Value": "oss-eu-central-1.aliyuncs.com",
                      "Help": "Central Europe 1 (Frankfurt)",
                      "Provider": ""
                  },
                  {
                      "Value": "oss-eu-west-1.aliyuncs.com",
                      "Help": "West Europe (London)",
                      "Provider": ""
                  },
                  {
                      "Value": "oss-me-east-1.aliyuncs.com",
                      "Help": "Middle East 1 (Dubai)",
                      "Provider": ""
                  }
              ],
              "ShortOpt": "",
              "Hide": 0,
              "Required": false,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": false
          },
          {
              "Name": "endpoint",
              "Help": "Endpoint for S3 API.\nRequired when using an S3 clone.",
              "Provider": "!AWS,IBMCOS,Alibaba",
              "Default": "",
              "Value": null,
              "Examples": [
                  {
                      "Value": "objects-us-west-1.dream.io",
                      "Help": "Dream Objects endpoint",
                      "Provider": "Dreamhost"
                  },
                  {
                      "Value": "nyc3.digitaloceanspaces.com",
                      "Help": "Digital Ocean Spaces New York 3",
                      "Provider": "DigitalOcean"
                  },
                  {
                      "Value": "ams3.digitaloceanspaces.com",
                      "Help": "Digital Ocean Spaces Amsterdam 3",
                      "Provider": "DigitalOcean"
                  },
                  {
                      "Value": "sgp1.digitaloceanspaces.com",
                      "Help": "Digital Ocean Spaces Singapore 1",
                      "Provider": "DigitalOcean"
                  },
                  {
                      "Value": "s3.wasabisys.com",
                      "Help": "Wasabi US East endpoint",
                      "Provider": "Wasabi"
                  },
                  {
                      "Value": "s3.us-west-1.wasabisys.com",
                      "Help": "Wasabi US West endpoint",
                      "Provider": "Wasabi"
                  }
              ],
              "ShortOpt": "",
              "Hide": 0,
              "Required": false,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": false
          },
          {
              "Name": "location_constraint",
              "Help": "Location constraint - must be set to match the Region.\nUsed when creating buckets only.",
              "Provider": "AWS",
              "Default": "",
              "Value": null,
              "Examples": [
                  {
                      "Value": "",
                      "Help": "Empty for US Region, Northern Virginia or Pacific Northwest.",
                      "Provider": ""
                  },
                  {
                      "Value": "us-east-2",
                      "Help": "US East (Ohio) Region.",
                      "Provider": ""
                  },
                  {
                      "Value": "us-west-2",
                      "Help": "US West (Oregon) Region.",
                      "Provider": ""
                  },
                  {
                      "Value": "us-west-1",
                      "Help": "US West (Northern California) Region.",
                      "Provider": ""
                  },
                  {
                      "Value": "ca-central-1",
                      "Help": "Canada (Central) Region.",
                      "Provider": ""
                  },
                  {
                      "Value": "eu-west-1",
                      "Help": "EU (Ireland) Region.",
                      "Provider": ""
                  },
                  {
                      "Value": "eu-west-2",
                      "Help": "EU (London) Region.",
                      "Provider": ""
                  },
                  {
                      "Value": "eu-north-1",
                      "Help": "EU (Stockholm) Region.",
                      "Provider": ""
                  },
                  {
                      "Value": "EU",
                      "Help": "EU Region.",
                      "Provider": ""
                  },
                  {
                      "Value": "ap-southeast-1",
                      "Help": "Asia Pacific (Singapore) Region.",
                      "Provider": ""
                  },
                  {
                      "Value": "ap-southeast-2",
                      "Help": "Asia Pacific (Sydney) Region.",
                      "Provider": ""
                  },
                  {
                      "Value": "ap-northeast-1",
                      "Help": "Asia Pacific (Tokyo) Region.",
                      "Provider": ""
                  },
                  {
                      "Value": "ap-northeast-2",
                      "Help": "Asia Pacific (Seoul)",
                      "Provider": ""
                  },
                  {
                      "Value": "ap-south-1",
                      "Help": "Asia Pacific (Mumbai)",
                      "Provider": ""
                  },
                  {
                      "Value": "sa-east-1",
                      "Help": "South America (Sao Paulo) Region.",
                      "Provider": ""
                  }
              ],
              "ShortOpt": "",
              "Hide": 0,
              "Required": false,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": false
          },
          {
              "Name": "location_constraint",
              "Help": "Location constraint - must match endpoint when using IBM Cloud Public.\nFor on-prem COS, do not make a selection from this list, hit enter",
              "Provider": "IBMCOS",
              "Default": "",
              "Value": null,
              "Examples": [
                  {
                      "Value": "us-standard",
                      "Help": "US Cross Region Standard",
                      "Provider": ""
                  },
                  {
                      "Value": "us-vault",
                      "Help": "US Cross Region Vault",
                      "Provider": ""
                  },
                  {
                      "Value": "us-cold",
                      "Help": "US Cross Region Cold",
                      "Provider": ""
                  },
                  {
                      "Value": "us-flex",
                      "Help": "US Cross Region Flex",
                      "Provider": ""
                  },
                  {
                      "Value": "us-east-standard",
                      "Help": "US East Region Standard",
                      "Provider": ""
                  },
                  {
                      "Value": "us-east-vault",
                      "Help": "US East Region Vault",
                      "Provider": ""
                  },
                  {
                      "Value": "us-east-cold",
                      "Help": "US East Region Cold",
                      "Provider": ""
                  },
                  {
                      "Value": "us-east-flex",
                      "Help": "US East Region Flex",
                      "Provider": ""
                  },
                  {
                      "Value": "us-south-standard",
                      "Help": "US South Region Standard",
                      "Provider": ""
                  },
                  {
                      "Value": "us-south-vault",
                      "Help": "US South Region Vault",
                      "Provider": ""
                  },
                  {
                      "Value": "us-south-cold",
                      "Help": "US South Region Cold",
                      "Provider": ""
                  },
                  {
                      "Value": "us-south-flex",
                      "Help": "US South Region Flex",
                      "Provider": ""
                  },
                  {
                      "Value": "eu-standard",
                      "Help": "EU Cross Region Standard",
                      "Provider": ""
                  },
                  {
                      "Value": "eu-vault",
                      "Help": "EU Cross Region Vault",
                      "Provider": ""
                  },
                  {
                      "Value": "eu-cold",
                      "Help": "EU Cross Region Cold",
                      "Provider": ""
                  },
                  {
                      "Value": "eu-flex",
                      "Help": "EU Cross Region Flex",
                      "Provider": ""
                  },
                  {
                      "Value": "eu-gb-standard",
                      "Help": "Great Britain Standard",
                      "Provider": ""
                  },
                  {
                      "Value": "eu-gb-vault",
                      "Help": "Great Britain Vault",
                      "Provider": ""
                  },
                  {
                      "Value": "eu-gb-cold",
                      "Help": "Great Britain Cold",
                      "Provider": ""
                  },
                  {
                      "Value": "eu-gb-flex",
                      "Help": "Great Britain Flex",
                      "Provider": ""
                  },
                  {
                      "Value": "ap-standard",
                      "Help": "APAC Standard",
                      "Provider": ""
                  },
                  {
                      "Value": "ap-vault",
                      "Help": "APAC Vault",
                      "Provider": ""
                  },
                  {
                      "Value": "ap-cold",
                      "Help": "APAC Cold",
                      "Provider": ""
                  },
                  {
                      "Value": "ap-flex",
                      "Help": "APAC Flex",
                      "Provider": ""
                  },
                  {
                      "Value": "mel01-standard",
                      "Help": "Melbourne Standard",
                      "Provider": ""
                  },
                  {
                      "Value": "mel01-vault",
                      "Help": "Melbourne Vault",
                      "Provider": ""
                  },
                  {
                      "Value": "mel01-cold",
                      "Help": "Melbourne Cold",
                      "Provider": ""
                  },
                  {
                      "Value": "mel01-flex",
                      "Help": "Melbourne Flex",
                      "Provider": ""
                  },
                  {
                      "Value": "tor01-standard",
                      "Help": "Toronto Standard",
                      "Provider": ""
                  },
                  {
                      "Value": "tor01-vault",
                      "Help": "Toronto Vault",
                      "Provider": ""
                  },
                  {
                      "Value": "tor01-cold",
                      "Help": "Toronto Cold",
                      "Provider": ""
                  },
                  {
                      "Value": "tor01-flex",
                      "Help": "Toronto Flex",
                      "Provider": ""
                  }
              ],
              "ShortOpt": "",
              "Hide": 0,
              "Required": false,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": false
          },
          {
              "Name": "location_constraint",
              "Help": "Location constraint - must be set to match the Region.\nLeave blank if not sure. Used when creating buckets only.",
              "Provider": "!AWS,IBMCOS,Alibaba",
              "Default": "",
              "Value": null,
              "ShortOpt": "",
              "Hide": 0,
              "Required": false,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": false
          },
          {
              "Name": "acl",
              "Help": "Canned ACL used when creating buckets and storing or copying objects.\n\nThis ACL is used for creating objects and if bucket_acl isn't set, for creating buckets too.\n\nFor more info visit https://docs.aws.amazon.com/AmazonS3/latest/dev/acl-overview.html#canned-acl\n\nNote that this ACL is applied when server side copying objects as S3\ndoesn't copy the ACL from the source but rather writes a fresh one.",
              "Provider": "",
              "Default": "",
              "Value": null,
              "Examples": [
                  {
                      "Value": "private",
                      "Help": "Owner gets FULL_CONTROL. No one else has access rights (default).",
                      "Provider": "!IBMCOS"
                  },
                  {
                      "Value": "public-read",
                      "Help": "Owner gets FULL_CONTROL. The AllUsers group gets READ access.",
                      "Provider": "!IBMCOS"
                  },
                  {
                      "Value": "public-read-write",
                      "Help": "Owner gets FULL_CONTROL. The AllUsers group gets READ and WRITE access.\nGranting this on a bucket is generally not recommended.",
                      "Provider": "!IBMCOS"
                  },
                  {
                      "Value": "authenticated-read",
                      "Help": "Owner gets FULL_CONTROL. The AuthenticatedUsers group gets READ access.",
                      "Provider": "!IBMCOS"
                  },
                  {
                      "Value": "bucket-owner-read",
                      "Help": "Object owner gets FULL_CONTROL. Bucket owner gets READ access.\nIf you specify this canned ACL when creating a bucket, Amazon S3 ignores it.",
                      "Provider": "!IBMCOS"
                  },
                  {
                      "Value": "bucket-owner-full-control",
                      "Help": "Both the object owner and the bucket owner get FULL_CONTROL over the object.\nIf you specify this canned ACL when creating a bucket, Amazon S3 ignores it.",
                      "Provider": "!IBMCOS"
                  },
                  {
                      "Value": "private",
                      "Help": "Owner gets FULL_CONTROL. No one else has access rights (default). This acl is available on IBM Cloud (Infra), IBM Cloud (Storage), On-Premise COS",
                      "Provider": "IBMCOS"
                  },
                  {
                      "Value": "public-read",
                      "Help": "Owner gets FULL_CONTROL. The AllUsers group gets READ access. This acl is available on IBM Cloud (Infra), IBM Cloud (Storage), On-Premise IBM COS",
                      "Provider": "IBMCOS"
                  },
                  {
                      "Value": "public-read-write",
                      "Help": "Owner gets FULL_CONTROL. The AllUsers group gets READ and WRITE access. This acl is available on IBM Cloud (Infra), On-Premise IBM COS",
                      "Provider": "IBMCOS"
                  },
                  {
                      "Value": "authenticated-read",
                      "Help": "Owner gets FULL_CONTROL. The AuthenticatedUsers group gets READ access. Not supported on Buckets. This acl is available on IBM Cloud (Infra) and On-Premise IBM COS",
                      "Provider": "IBMCOS"
                  }
              ],
              "ShortOpt": "",
              "Hide": 0,
              "Required": false,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": false
          },
          {
              "Name": "bucket_acl",
              "Help": "Canned ACL used when creating buckets.\n\nFor more info visit https://docs.aws.amazon.com/AmazonS3/latest/dev/acl-overview.html#canned-acl\n\nNote that this ACL is applied when only when creating buckets.  If it\nisn't set then \"acl\" is used instead.",
              "Provider": "",
              "Default": "",
              "Value": null,
              "Examples": [
                  {
                      "Value": "private",
                      "Help": "Owner gets FULL_CONTROL. No one else has access rights (default).",
                      "Provider": ""
                  },
                  {
                      "Value": "public-read",
                      "Help": "Owner gets FULL_CONTROL. The AllUsers group gets READ access.",
                      "Provider": ""
                  },
                  {
                      "Value": "public-read-write",
                      "Help": "Owner gets FULL_CONTROL. The AllUsers group gets READ and WRITE access.\nGranting this on a bucket is generally not recommended.",
                      "Provider": ""
                  },
                  {
                      "Value": "authenticated-read",
                      "Help": "Owner gets FULL_CONTROL. The AuthenticatedUsers group gets READ access.",
                      "Provider": ""
                  }
              ],
              "ShortOpt": "",
              "Hide": 0,
              "Required": false,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": true
          },
          {
              "Name": "server_side_encryption",
              "Help": "The server-side encryption algorithm used when storing this object in S3.",
              "Provider": "AWS",
              "Default": "",
              "Value": null,
              "Examples": [
                  {
                      "Value": "",
                      "Help": "None",
                      "Provider": ""
                  },
                  {
                      "Value": "AES256",
                      "Help": "AES256",
                      "Provider": ""
                  },
                  {
                      "Value": "aws:kms",
                      "Help": "aws:kms",
                      "Provider": ""
                  }
              ],
              "ShortOpt": "",
              "Hide": 0,
              "Required": false,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": false
          },
          {
              "Name": "sse_kms_key_id",
              "Help": "If using KMS ID you must provide the ARN of Key.",
              "Provider": "AWS",
              "Default": "",
              "Value": null,
              "Examples": [
                  {
                      "Value": "",
                      "Help": "None",
                      "Provider": ""
                  },
                  {
                      "Value": "arn:aws:kms:us-east-1:*",
                      "Help": "arn:aws:kms:*",
                      "Provider": ""
                  }
              ],
              "ShortOpt": "",
              "Hide": 0,
              "Required": false,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": false
          },
          {
              "Name": "storage_class",
              "Help": "The storage class to use when storing new objects in S3.",
              "Provider": "AWS",
              "Default": "",
              "Value": null,
              "Examples": [
                  {
                      "Value": "",
                      "Help": "Default",
                      "Provider": ""
                  },
                  {
                      "Value": "STANDARD",
                      "Help": "Standard storage class",
                      "Provider": ""
                  },
                  {
                      "Value": "REDUCED_REDUNDANCY",
                      "Help": "Reduced redundancy storage class",
                      "Provider": ""
                  },
                  {
                      "Value": "STANDARD_IA",
                      "Help": "Standard Infrequent Access storage class",
                      "Provider": ""
                  },
                  {
                      "Value": "ONEZONE_IA",
                      "Help": "One Zone Infrequent Access storage class",
                      "Provider": ""
                  },
                  {
                      "Value": "GLACIER",
                      "Help": "Glacier storage class",
                      "Provider": ""
                  }
              ],
              "ShortOpt": "",
              "Hide": 0,
              "Required": false,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": false
          },
          {
              "Name": "storage_class",
              "Help": "The storage class to use when storing new objects in OSS.",
              "Provider": "Alibaba",
              "Default": "",
              "Value": null,
              "Examples": [
                  {
                      "Value": "",
                      "Help": "Default",
                      "Provider": ""
                  },
                  {
                      "Value": "STANDARD",
                      "Help": "Standard storage class",
                      "Provider": ""
                  },
                  {
                      "Value": "GLACIER",
                      "Help": "Archive storage mode.",
                      "Provider": ""
                  },
                  {
                      "Value": "STANDARD_IA",
                      "Help": "Infrequent access storage mode.",
                      "Provider": ""
                  }
              ],
              "ShortOpt": "",
              "Hide": 0,
              "Required": false,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": false
          },
          {
              "Name": "upload_cutoff",
              "Help": "Cutoff for switching to chunked upload\n\nAny files larger than this will be uploaded in chunks of chunk_size.\nThe minimum is 0 and the maximum is 5GB.",
              "Provider": "",
              "Default": 209715200,
              "Value": null,
              "ShortOpt": "",
              "Hide": 0,
              "Required": false,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": true
          },
          {
              "Name": "chunk_size",
              "Help": "Chunk size to use for uploading.\n\nWhen uploading files larger than upload_cutoff they will be uploaded\nas multipart uploads using this chunk size.\n\nNote that \"--s3-upload-concurrency\" chunks of this size are buffered\nin memory per transfer.\n\nIf you are transferring large files over high speed links and you have\nenough memory, then increasing this will speed up the transfers.",
              "Provider": "",
              "Default": 5242880,
              "Value": null,
              "ShortOpt": "",
              "Hide": 0,
              "Required": false,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": true
          },
          {
              "Name": "disable_checksum",
              "Help": "Don't store MD5 checksum with object metadata",
              "Provider": "",
              "Default": false,
              "Value": null,
              "ShortOpt": "",
              "Hide": 0,
              "Required": false,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": true
          },
          {
              "Name": "session_token",
              "Help": "An AWS session token",
              "Provider": "",
              "Default": "",
              "Value": null,
              "ShortOpt": "",
              "Hide": 0,
              "Required": false,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": true
          },
          {
              "Name": "upload_concurrency",
              "Help": "Concurrency for multipart uploads.\n\nThis is the number of chunks of the same file that are uploaded\nconcurrently.\n\nIf you are uploading small numbers of large file over high speed link\nand these uploads do not fully utilize your bandwidth, then increasing\nthis may help to speed up the transfers.",
              "Provider": "",
              "Default": 4,
              "Value": null,
              "ShortOpt": "",
              "Hide": 0,
              "Required": false,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": true
          },
          {
              "Name": "force_path_style",
              "Help": "If true use path style access if false use virtual hosted style.\n\nIf this is true (the default) then rclone will use path style access,\nif false then rclone will use virtual path style. See [the AWS S3\ndocs](https://docs.aws.amazon.com/AmazonS3/latest/dev/UsingBucket.html#access-bucket-intro)\nfor more info.\n\nSome providers (eg Aliyun OSS or Netease COS) require this set to false.",
              "Provider": "",
              "Default": true,
              "Value": null,
              "ShortOpt": "",
              "Hide": 0,
              "Required": false,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": true
          },
          {
              "Name": "v2_auth",
              "Help": "If true use v2 authentication.\n\nIf this is false (the default) then rclone will use v4 authentication.\nIf it is set then rclone will use v2 authentication.\n\nUse this only if v4 signatures don't work, eg pre Jewel/v10 CEPH.",
              "Provider": "",
              "Default": false,
              "Value": null,
              "ShortOpt": "",
              "Hide": 0,
              "Required": false,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": true
          }
      ]
  },
  {
      "Name": "sftp",
      "Description": "SSH/SFTP Connection",
      "Prefix": "sftp",
      "Options": [
          {
              "Name": "host",
              "Help": "SSH host to connect to",
              "Provider": "",
              "Default": "",
              "Value": null,
              "Examples": [
                  {
                      "Value": "example.com",
                      "Help": "Connect to example.com",
                      "Provider": ""
                  }
              ],
              "ShortOpt": "",
              "Hide": 0,
              "Required": true,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": false
          },
          {
              "Name": "user",
              "Help": "SSH username, leave blank for current username, negative0",
              "Provider": "",
              "Default": "",
              "Value": null,
              "ShortOpt": "",
              "Hide": 0,
              "Required": false,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": false
          },
          {
              "Name": "port",
              "Help": "SSH port, leave blank to use default (22)",
              "Provider": "",
              "Default": "",
              "Value": null,
              "ShortOpt": "",
              "Hide": 0,
              "Required": false,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": false
          },
          {
              "Name": "pass",
              "Help": "SSH password, leave blank to use ssh-agent.",
              "Provider": "",
              "Default": "",
              "Value": null,
              "ShortOpt": "",
              "Hide": 0,
              "Required": false,
              "IsPassword": true,
              "NoPrefix": false,
              "Advanced": false
          },
          {
              "Name": "key_file",
              "Help": "Path to PEM-encoded private key file, leave blank or set key-use-agent to use ssh-agent.",
              "Provider": "",
              "Default": "",
              "Value": null,
              "ShortOpt": "",
              "Hide": 0,
              "Required": false,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": false
          },
          {
              "Name": "key_file_pass",
              "Help": "The passphrase to decrypt the PEM-encoded private key file.\n\nOnly PEM encrypted key files (old OpenSSH format) are supported. Encrypted keys\nin the new OpenSSH format can't be used.",
              "Provider": "",
              "Default": "",
              "Value": null,
              "ShortOpt": "",
              "Hide": 0,
              "Required": false,
              "IsPassword": true,
              "NoPrefix": false,
              "Advanced": false
          },
          {
              "Name": "key_use_agent",
              "Help": "When set forces the usage of the ssh-agent.\n\nWhen key-file is also set, the \".pub\" file of the specified key-file is read and only the associated key is\nrequested from the ssh-agent. This allows to avoid `Too many authentication failures for *username*` errors\nwhen the ssh-agent contains many keys.",
              "Provider": "",
              "Default": false,
              "Value": null,
              "ShortOpt": "",
              "Hide": 0,
              "Required": false,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": false
          },
          {
              "Name": "use_insecure_cipher",
              "Help": "Enable the use of the aes128-cbc cipher. This cipher is insecure and may allow plaintext data to be recovered by an attacker.",
              "Provider": "",
              "Default": false,
              "Value": null,
              "Examples": [
                  {
                      "Value": "false",
                      "Help": "Use default Cipher list.",
                      "Provider": ""
                  },
                  {
                      "Value": "true",
                      "Help": "Enables the use of the aes128-cbc cipher.",
                      "Provider": ""
                  }
              ],
              "ShortOpt": "",
              "Hide": 0,
              "Required": false,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": false
          },
          {
              "Name": "disable_hashcheck",
              "Help": "Disable the execution of SSH commands to determine if remote file hashing is available.\nLeave blank or set to false to enable hashing (recommended), set to true to disable hashing.",
              "Provider": "",
              "Default": false,
              "Value": null,
              "ShortOpt": "",
              "Hide": 0,
              "Required": false,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": false
          },
          {
              "Name": "ask_password",
              "Help": "Allow asking for SFTP password when needed.",
              "Provider": "",
              "Default": false,
              "Value": null,
              "ShortOpt": "",
              "Hide": 0,
              "Required": false,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": true
          },
          {
              "Name": "path_override",
              "Help": "Override path used by SSH connection.\n\nThis allows checksum calculation when SFTP and SSH paths are\ndifferent. This issue affects among others Synology NAS boxes.\n\nShared folders can be found in directories representing volumes\n\n    rclone sync /home/local/directory remote:/directory --ssh-path-override /volume2/directory\n\nHome directory can be found in a shared folder called \"home\"\n\n    rclone sync /home/local/directory remote:/home/directory --ssh-path-override /volume1/homes/USER/directory",
              "Provider": "",
              "Default": "",
              "Value": null,
              "ShortOpt": "",
              "Hide": 0,
              "Required": false,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": true
          },
          {
              "Name": "set_modtime",
              "Help": "Set the modified time on the remote if set.",
              "Provider": "",
              "Default": true,
              "Value": null,
              "ShortOpt": "",
              "Hide": 0,
              "Required": false,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": true
          }
      ]
  },
  {
      "Name": "union",
      "Description": "A stackable unification remote, which can appear to merge the contents of several remotes",
      "Prefix": "union",
      "Options": [
          {
              "Name": "remotes",
              "Help": "List of space separated remotes.\nCan be 'remotea:test/dir remoteb:', '\"remotea:test/space dir\" remoteb:', etc.\nThe last remote is used to write to.",
              "Provider": "",
              "Default": "",
              "Value": null,
              "ShortOpt": "",
              "Hide": 0,
              "Required": true,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": false
          }
      ]
  },
  {
      "Name": "webdav",
      "Description": "Webdav",
      "Prefix": "webdav",
      "Options": [
          {
              "Name": "url",
              "Help": "URL of http host to connect to",
              "Provider": "",
              "Default": "",
              "Value": null,
              "Examples": [
                  {
                      "Value": "https://example.com",
                      "Help": "Connect to example.com",
                      "Provider": ""
                  }
              ],
              "ShortOpt": "",
              "Hide": 0,
              "Required": true,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": false
          },
          {
              "Name": "vendor",
              "Help": "Name of the Webdav site/service/software you are using",
              "Provider": "",
              "Default": "",
              "Value": null,
              "Examples": [
                  {
                      "Value": "nextcloud",
                      "Help": "Nextcloud",
                      "Provider": ""
                  },
                  {
                      "Value": "owncloud",
                      "Help": "Owncloud",
                      "Provider": ""
                  },
                  {
                      "Value": "sharepoint",
                      "Help": "Sharepoint",
                      "Provider": ""
                  },
                  {
                      "Value": "other",
                      "Help": "Other site/service or software",
                      "Provider": ""
                  }
              ],
              "ShortOpt": "",
              "Hide": 0,
              "Required": false,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": false
          },
          {
              "Name": "user",
              "Help": "User name",
              "Provider": "",
              "Default": "",
              "Value": null,
              "ShortOpt": "",
              "Hide": 0,
              "Required": false,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": false
          },
          {
              "Name": "pass",
              "Help": "Password.",
              "Provider": "",
              "Default": "",
              "Value": null,
              "ShortOpt": "",
              "Hide": 0,
              "Required": false,
              "IsPassword": true,
              "NoPrefix": false,
              "Advanced": false
          },
          {
              "Name": "bearer_token",
              "Help": "Bearer token instead of user/pass (eg a Macaroon)",
              "Provider": "",
              "Default": "",
              "Value": null,
              "ShortOpt": "",
              "Hide": 0,
              "Required": false,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": false
          }
      ]
  },
  {
      "Name": "yandex",
      "Description": "Yandex Disk",
      "Prefix": "yandex",
      "Options": [
          {
              "Name": "client_id",
              "Help": "Yandex Client Id\nLeave blank normally.",
              "Provider": "",
              "Default": "",
              "Value": null,
              "ShortOpt": "",
              "Hide": 0,
              "Required": false,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": false
          },
          {
              "Name": "client_secret",
              "Help": "Yandex Client Secret\nLeave blank normally.",
              "Provider": "",
              "Default": "",
              "Value": null,
              "ShortOpt": "",
              "Hide": 0,
              "Required": false,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": false
          },
          {
              "Name": "unlink",
              "Help": "Remove existing public link to file/folder with link command rather than creating.\nDefault is false, meaning link command will create or retrieve public link.",
              "Provider": "",
              "Default": false,
              "Value": null,
              "ShortOpt": "",
              "Hide": 0,
              "Required": false,
              "IsPassword": false,
              "NoPrefix": false,
              "Advanced": true
          }
      ]
  }
];

export const ProvAuth = [
  "acd",
  "box",
  "drive",
  "dropbox",
  "gcs",
  "gphotos",
  "hubic",
  "jottacloud",
  "onedrive",
  "pcloud",
  "premiumizeme",
  "putio",
  "sharefile",
  "sugarsync",
  "yandex",
]