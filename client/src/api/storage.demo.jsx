import wrap from './wrap';

const mounts = {
  list: () => {
    return new Promise((resolve, reject) => {
      setTimeout(() => {
        resolve({
          "data": [
            {
              "path": "/var/mnt/cosmos",
              "permenant": false,
              "device": "Cosmos-Server",
              "type": "vboxsf",
              "opts": [
                "rw",
                "nodev",
                "relatime"
              ]
            },
            {
              "path": "/var/mnt/sda1",
              "permenant": false,
              "device": "/dev/sda1",
              "type": "ext4",
              "opts": [
                "rw",
                "seclabel",
                "relatime"
              ]
            },
            {
              "path": "/var/mnt/sdb1",
              "permenant": false,
              "device": "/dev/sdb1",
              "type": "ext4",
              "opts": [
                "rw",
                "seclabel",
                "relatime"
              ]
            },
            {
              "path": "/var/mnt/sdc1",
              "permenant": false,
              "device": "/dev/sdc1",
              "type": "ext4",
              "opts": [
                "rw",
                "seclabel",
                "relatime"
              ]
            },
            {
              "path": "/var/mnt/sdd1",
              "permenant": false,
              "device": "/dev/sdd1",
              "type": "ext4",
              "opts": [
                "rw",
                "seclabel",
                "relatime"
              ]
            }
          ],
          "status": "OK"
        })},
        500
      );
    });
  },

  mount: ({path, mountPoint, permanent, chown}) => {
    return new Promise((resolve, reject) => {
      setTimeout(() => {
        resolve({
          "status": "ok",
        })},
        1000
      );
    });
  },

  unmount: ({mountPoint, permanent, chown}) => {
    return new Promise((resolve, reject) => {
      setTimeout(() => {
        resolve({
          "status": "ok",
        })},
        1000
      );
    });
  },

  merge: (args) => {
    return new Promise((resolve, reject) => {
      setTimeout(() => {
        resolve({
          "status": "ok",
        })},
        2000
      );
    });
  },
};


const disks = {
  list: () => {
    return new Promise((resolve, reject) => {
      setTimeout(() => {
        resolve({
          "data": [
            {
              "name": "/dev/sda",
              "type": "disk",
              "size": 16000900661248,
              "rota": true,
              "serial": "ZR7025DW",
              "wwn": "0x5000c500dc27359d",
              "vendor": "ATA     ",
              "model": "ST16000NM001J-2T",
              "rev": "SS02",
              "children": [
                {
                  "name": "/dev/sda1",
                  "type": "part",
                  "size": 16000899595776,
                  "rota": true,
                  "wwn": "0x5000c500dc27359d",
                  "fstype": "linux_raid_member",
                  "partuuid": "ba287f4b-e094-df4f-ac78-07d2abc13b04",
                  "children": [
                    {
                      "name": "/dev/md2",
                      "type": "raid6",
                      "size": 128006110576640,
                      "rota": true,
                      "fstype": "crypto_LUKS",
                      "children": [
                        {
                          "name": "/dev/mapper/securehdd",
                          "type": "crypt",
                          "size": 128006093799424,
                          "rota": true,
                          "mountpoint": "/mnt/secure",
                          "fstype": "ext4",
                          "children": [],
                          "usage": 110085240667504,
                          "smart": {
                            "Temperature": 0,
                            "Read": 0,
                            "Written": 0,
                            "PowerOnHours": 0,
                            "PowerCycles": 0,
                            "AdditionalData": null,
                            "Thresholds": null
                          }
                        }
                      ],
                      "usage": 0,
                      "smart": {
                        "Temperature": 0,
                        "Read": 0,
                        "Written": 0,
                        "PowerOnHours": 0,
                        "PowerCycles": 0,
                        "AdditionalData": null,
                        "Thresholds": null
                      }
                    }
                  ],
                  "usage": 0,
                  "smart": {
                    "Temperature": 0,
                    "Read": 0,
                    "Written": 0,
                    "PowerOnHours": 0,
                    "PowerCycles": 0,
                    "AdditionalData": null,
                    "Thresholds": null
                  }
                }
              ],
              "usage": 0,
              "smart": {
                "Temperature": 35,
                "Read": 20824670227222,
                "Written": 211307803078,
                "PowerOnHours": 19921,
                "PowerCycles": 4,
                "AdditionalData": {
                  "Version": 10,
                  "Attrs": {
                    "1": {
                      "Id": 1,
                      "Flags": 15,
                      "Current": 77,
                      "Worst": 64,
                      "VendorBytes": [
                        182,
                        180,
                        185,
                        2,
                        0,
                        0
                      ],
                      "Name": "Raw_Read_Error_Rate",
                      "Type": 3,
                      "ValueRaw": 45724854
                    },
                    "3": {
                      "Id": 3,
                      "Flags": 3,
                      "Current": 96,
                      "Worst": 96,
                      "VendorBytes": [
                        0,
                        0,
                        0,
                        0,
                        0,
                        0
                      ],
                      "Name": "Spin_Up_Time",
                      "Type": 10,
                      "ValueRaw": 0
                    },
                    "4": {
                      "Id": 4,
                      "Flags": 50,
                      "Current": 100,
                      "Worst": 100,
                      "VendorBytes": [
                        4,
                        0,
                        0,
                        0,
                        0,
                        0
                      ],
                      "Name": "Start_Stop_Count",
                      "Type": 3,
                      "ValueRaw": 4
                    },
                    "5": {
                      "Id": 5,
                      "Flags": 51,
                      "Current": 100,
                      "Worst": 100,
                      "VendorBytes": [
                        0,
                        0,
                        0,
                        0,
                        0,
                        0
                      ],
                      "Name": "Reallocated_Sector_Ct",
                      "Type": 9,
                      "ValueRaw": 0
                    },
                    "7": {
                      "Id": 7,
                      "Flags": 15,
                      "Current": 88,
                      "Worst": 60,
                      "VendorBytes": [
                        150,
                        104,
                        190,
                        33,
                        0,
                        0
                      ],
                      "Name": "Seek_Error_Rate",
                      "Type": 3,
                      "ValueRaw": 566126742
                    },
                    "9": {
                      "Id": 9,
                      "Flags": 50,
                      "Current": 78,
                      "Worst": 78,
                      "VendorBytes": [
                        209,
                        77,
                        0,
                        0,
                        0,
                        0
                      ],
                      "Name": "Power_On_Hours",
                      "Type": 11,
                      "ValueRaw": 19921
                    },
                    "10": {
                      "Id": 10,
                      "Flags": 19,
                      "Current": 100,
                      "Worst": 100,
                      "VendorBytes": [
                        0,
                        0,
                        0,
                        0,
                        0,
                        0
                      ],
                      "Name": "Spin_Retry_Count",
                      "Type": 3,
                      "ValueRaw": 0
                    },
                    "12": {
                      "Id": 12,
                      "Flags": 50,
                      "Current": 100,
                      "Worst": 100,
                      "VendorBytes": [
                        4,
                        0,
                        0,
                        0,
                        0,
                        0
                      ],
                      "Name": "Power_Cycle_Count",
                      "Type": 3,
                      "ValueRaw": 4
                    },
                    "18": {
                      "Id": 18,
                      "Flags": 11,
                      "Current": 100,
                      "Worst": 100,
                      "VendorBytes": [
                        0,
                        0,
                        0,
                        0,
                        0,
                        0
                      ],
                      "Name": "",
                      "Type": 0,
                      "ValueRaw": 0
                    },
                    "187": {
                      "Id": 187,
                      "Flags": 50,
                      "Current": 100,
                      "Worst": 100,
                      "VendorBytes": [
                        0,
                        0,
                        0,
                        0,
                        0,
                        0
                      ],
                      "Name": "Reported_Uncorrect",
                      "Type": 3,
                      "ValueRaw": 0
                    },
                    "188": {
                      "Id": 188,
                      "Flags": 50,
                      "Current": 100,
                      "Worst": 100,
                      "VendorBytes": [
                        0,
                        0,
                        0,
                        0,
                        0,
                        0
                      ],
                      "Name": "Command_Timeout",
                      "Type": 3,
                      "ValueRaw": 0
                    },
                    "190": {
                      "Id": 190,
                      "Flags": 34,
                      "Current": 65,
                      "Worst": 48,
                      "VendorBytes": [
                        35,
                        0,
                        32,
                        41,
                        0,
                        0
                      ],
                      "Name": "Airflow_Temperature_Cel",
                      "Type": 18,
                      "ValueRaw": 689963043
                    },
                    "192": {
                      "Id": 192,
                      "Flags": 50,
                      "Current": 100,
                      "Worst": 100,
                      "VendorBytes": [
                        1,
                        0,
                        0,
                        0,
                        0,
                        0
                      ],
                      "Name": "Power-Off_Retract_Count",
                      "Type": 3,
                      "ValueRaw": 1
                    },
                    "193": {
                      "Id": 193,
                      "Flags": 50,
                      "Current": 100,
                      "Worst": 100,
                      "VendorBytes": [
                        85,
                        3,
                        0,
                        0,
                        0,
                        0
                      ],
                      "Name": "Load_Cycle_Count",
                      "Type": 3,
                      "ValueRaw": 853
                    },
                    "194": {
                      "Id": 194,
                      "Flags": 34,
                      "Current": 35,
                      "Worst": 41,
                      "VendorBytes": [
                        35,
                        0,
                        0,
                        0,
                        23,
                        0
                      ],
                      "Name": "Temperature_Celsius",
                      "Type": 18,
                      "ValueRaw": 98784247843
                    },
                    "197": {
                      "Id": 197,
                      "Flags": 18,
                      "Current": 100,
                      "Worst": 100,
                      "VendorBytes": [
                        0,
                        0,
                        0,
                        0,
                        0,
                        0
                      ],
                      "Name": "Current_Pending_Sector",
                      "Type": 3,
                      "ValueRaw": 0
                    },
                    "198": {
                      "Id": 198,
                      "Flags": 16,
                      "Current": 100,
                      "Worst": 100,
                      "VendorBytes": [
                        0,
                        0,
                        0,
                        0,
                        0,
                        0
                      ],
                      "Name": "Offline_Uncorrectable",
                      "Type": 3,
                      "ValueRaw": 0
                    },
                    "199": {
                      "Id": 199,
                      "Flags": 62,
                      "Current": 200,
                      "Worst": 200,
                      "VendorBytes": [
                        0,
                        0,
                        0,
                        0,
                        0,
                        0
                      ],
                      "Name": "UDMA_CRC_Error_Count",
                      "Type": 3,
                      "ValueRaw": 0
                    },
                    "200": {
                      "Id": 200,
                      "Flags": 35,
                      "Current": 100,
                      "Worst": 100,
                      "VendorBytes": [
                        0,
                        0,
                        0,
                        0,
                        0,
                        0
                      ],
                      "Name": "Multi_Zone_Error_Rate",
                      "Type": 3,
                      "ValueRaw": 0
                    },
                    "240": {
                      "Id": 240,
                      "Flags": 0,
                      "Current": 100,
                      "Worst": 253,
                      "VendorBytes": [
                        123,
                        77,
                        0,
                        0,
                        38,
                        110
                      ],
                      "Name": "Head_Flying_Hours",
                      "Type": 11,
                      "ValueRaw": 121109487832443
                    },
                    "241": {
                      "Id": 241,
                      "Flags": 0,
                      "Current": 100,
                      "Worst": 253,
                      "VendorBytes": [
                        198,
                        49,
                        237,
                        50,
                        49,
                        0
                      ],
                      "Name": "Total_LBAs_Written",
                      "Type": 3,
                      "ValueRaw": 211307803078
                    },
                    "242": {
                      "Id": 242,
                      "Flags": 0,
                      "Current": 100,
                      "Worst": 253,
                      "VendorBytes": [
                        22,
                        75,
                        18,
                        159,
                        240,
                        18
                      ],
                      "Name": "Total_LBAs_Read",
                      "Type": 3,
                      "ValueRaw": 20824670227222
                    }
                  }
                },
                "Thresholds": {
                  "Revnumber": 1,
                  "Thresholds": {
                    "1": 44,
                    "3": 0,
                    "4": 20,
                    "5": 10,
                    "7": 45,
                    "9": 0,
                    "10": 97,
                    "12": 20,
                    "18": 50,
                    "187": 0,
                    "188": 0,
                    "190": 0,
                    "192": 0,
                    "193": 0,
                    "194": 0,
                    "197": 0,
                    "198": 0,
                    "199": 0,
                    "200": 1,
                    "240": 0,
                    "241": 0,
                    "242": 0
                  }
                }
              }
            },
            {
              "name": "/dev/sdb",
              "type": "disk",
              "size": 16000900661248,
              "rota": true,
              "serial": "ZR701TEX",
              "wwn": "0x5000c500dc2714c2",
              "vendor": "ATA     ",
              "model": "ST16000NM001J-2T",
              "rev": "SS02",
              "children": [
                {
                  "name": "/dev/sdb1",
                  "type": "part",
                  "size": 16000899595776,
                  "rota": true,
                  "wwn": "0x5000c500dc2714c2",
                  "fstype": "linux_raid_member",
                  "partuuid": "99c8bba5-9af9-464a-b21b-c4d4d321054b",
                  "children": [
                    {
                      "name": "/dev/md2",
                      "type": "raid6",
                      "size": 128006110576640,
                      "rota": true,
                      "fstype": "crypto_LUKS",
                      "children": [
                        {
                          "name": "/dev/mapper/securehdd",
                          "type": "crypt",
                          "size": 128006093799424,
                          "rota": true,
                          "mountpoint": "/mnt/secure",
                          "fstype": "ext4",
                          "children": [],
                          "usage": 110085240667504,
                          "smart": {
                            "Temperature": 0,
                            "Read": 0,
                            "Written": 0,
                            "PowerOnHours": 0,
                            "PowerCycles": 0,
                            "AdditionalData": null,
                            "Thresholds": null
                          }
                        }
                      ],
                      "usage": 0,
                      "smart": {
                        "Temperature": 0,
                        "Read": 0,
                        "Written": 0,
                        "PowerOnHours": 0,
                        "PowerCycles": 0,
                        "AdditionalData": null,
                        "Thresholds": null
                      }
                    }
                  ],
                  "usage": 0,
                  "smart": {
                    "Temperature": 0,
                    "Read": 0,
                    "Written": 0,
                    "PowerOnHours": 0,
                    "PowerCycles": 0,
                    "AdditionalData": null,
                    "Thresholds": null
                  }
                }
              ],
              "usage": 0,
              "smart": {
                "Temperature": 37,
                "Read": 14895563227712,
                "Written": 207552354038,
                "PowerOnHours": 19922,
                "PowerCycles": 4,
                "AdditionalData": {
                  "Version": 10,
                  "Attrs": {
                    "1": {
                      "Id": 1,
                      "Flags": 15,
                      "Current": 84,
                      "Worst": 64,
                      "VendorBytes": [
                        13,
                        93,
                        204,
                        13,
                        0,
                        0
                      ],
                      "Name": "Raw_Read_Error_Rate",
                      "Type": 3,
                      "ValueRaw": 231496973
                    },
                    "3": {
                      "Id": 3,
                      "Flags": 3,
                      "Current": 96,
                      "Worst": 96,
                      "VendorBytes": [
                        0,
                        0,
                        0,
                        0,
                        0,
                        0
                      ],
                      "Name": "Spin_Up_Time",
                      "Type": 10,
                      "ValueRaw": 0
                    },
                    "4": {
                      "Id": 4,
                      "Flags": 50,
                      "Current": 100,
                      "Worst": 100,
                      "VendorBytes": [
                        4,
                        0,
                        0,
                        0,
                        0,
                        0
                      ],
                      "Name": "Start_Stop_Count",
                      "Type": 3,
                      "ValueRaw": 4
                    },
                    "5": {
                      "Id": 5,
                      "Flags": 51,
                      "Current": 100,
                      "Worst": 100,
                      "VendorBytes": [
                        0,
                        0,
                        0,
                        0,
                        0,
                        0
                      ],
                      "Name": "Reallocated_Sector_Ct",
                      "Type": 9,
                      "ValueRaw": 0
                    },
                    "7": {
                      "Id": 7,
                      "Flags": 15,
                      "Current": 87,
                      "Worst": 60,
                      "VendorBytes": [
                        78,
                        136,
                        76,
                        32,
                        0,
                        0
                      ],
                      "Name": "Seek_Error_Rate",
                      "Type": 3,
                      "ValueRaw": 541886542
                    },
                    "9": {
                      "Id": 9,
                      "Flags": 50,
                      "Current": 78,
                      "Worst": 78,
                      "VendorBytes": [
                        210,
                        77,
                        0,
                        0,
                        0,
                        0
                      ],
                      "Name": "Power_On_Hours",
                      "Type": 11,
                      "ValueRaw": 19922
                    },
                    "10": {
                      "Id": 10,
                      "Flags": 19,
                      "Current": 100,
                      "Worst": 100,
                      "VendorBytes": [
                        0,
                        0,
                        0,
                        0,
                        0,
                        0
                      ],
                      "Name": "Spin_Retry_Count",
                      "Type": 3,
                      "ValueRaw": 0
                    },
                    "12": {
                      "Id": 12,
                      "Flags": 50,
                      "Current": 100,
                      "Worst": 100,
                      "VendorBytes": [
                        4,
                        0,
                        0,
                        0,
                        0,
                        0
                      ],
                      "Name": "Power_Cycle_Count",
                      "Type": 3,
                      "ValueRaw": 4
                    },
                    "18": {
                      "Id": 18,
                      "Flags": 11,
                      "Current": 100,
                      "Worst": 100,
                      "VendorBytes": [
                        0,
                        0,
                        0,
                        0,
                        0,
                        0
                      ],
                      "Name": "",
                      "Type": 0,
                      "ValueRaw": 0
                    },
                    "187": {
                      "Id": 187,
                      "Flags": 50,
                      "Current": 100,
                      "Worst": 100,
                      "VendorBytes": [
                        0,
                        0,
                        0,
                        0,
                        0,
                        0
                      ],
                      "Name": "Reported_Uncorrect",
                      "Type": 3,
                      "ValueRaw": 0
                    },
                    "188": {
                      "Id": 188,
                      "Flags": 50,
                      "Current": 100,
                      "Worst": 100,
                      "VendorBytes": [
                        0,
                        0,
                        0,
                        0,
                        0,
                        0
                      ],
                      "Name": "Command_Timeout",
                      "Type": 3,
                      "ValueRaw": 0
                    },
                    "190": {
                      "Id": 190,
                      "Flags": 34,
                      "Current": 63,
                      "Worst": 49,
                      "VendorBytes": [
                        37,
                        0,
                        32,
                        42,
                        0,
                        0
                      ],
                      "Name": "Airflow_Temperature_Cel",
                      "Type": 18,
                      "ValueRaw": 706740261
                    },
                    "192": {
                      "Id": 192,
                      "Flags": 50,
                      "Current": 100,
                      "Worst": 100,
                      "VendorBytes": [
                        1,
                        0,
                        0,
                        0,
                        0,
                        0
                      ],
                      "Name": "Power-Off_Retract_Count",
                      "Type": 3,
                      "ValueRaw": 1
                    },
                    "193": {
                      "Id": 193,
                      "Flags": 50,
                      "Current": 100,
                      "Worst": 100,
                      "VendorBytes": [
                        86,
                        3,
                        0,
                        0,
                        0,
                        0
                      ],
                      "Name": "Load_Cycle_Count",
                      "Type": 3,
                      "ValueRaw": 854
                    },
                    "194": {
                      "Id": 194,
                      "Flags": 34,
                      "Current": 37,
                      "Worst": 42,
                      "VendorBytes": [
                        37,
                        0,
                        0,
                        0,
                        23,
                        0
                      ],
                      "Name": "Temperature_Celsius",
                      "Type": 18,
                      "ValueRaw": 98784247845
                    },
                    "197": {
                      "Id": 197,
                      "Flags": 18,
                      "Current": 100,
                      "Worst": 100,
                      "VendorBytes": [
                        0,
                        0,
                        0,
                        0,
                        0,
                        0
                      ],
                      "Name": "Current_Pending_Sector",
                      "Type": 3,
                      "ValueRaw": 0
                    },
                    "198": {
                      "Id": 198,
                      "Flags": 16,
                      "Current": 100,
                      "Worst": 100,
                      "VendorBytes": [
                        0,
                        0,
                        0,
                        0,
                        0,
                        0
                      ],
                      "Name": "Offline_Uncorrectable",
                      "Type": 3,
                      "ValueRaw": 0
                    },
                    "199": {
                      "Id": 199,
                      "Flags": 62,
                      "Current": 200,
                      "Worst": 200,
                      "VendorBytes": [
                        0,
                        0,
                        0,
                        0,
                        0,
                        0
                      ],
                      "Name": "UDMA_CRC_Error_Count",
                      "Type": 3,
                      "ValueRaw": 0
                    },
                    "200": {
                      "Id": 200,
                      "Flags": 35,
                      "Current": 100,
                      "Worst": 100,
                      "VendorBytes": [
                        0,
                        0,
                        0,
                        0,
                        0,
                        0
                      ],
                      "Name": "Multi_Zone_Error_Rate",
                      "Type": 3,
                      "ValueRaw": 0
                    },
                    "240": {
                      "Id": 240,
                      "Flags": 0,
                      "Current": 100,
                      "Worst": 253,
                      "VendorBytes": [
                        122,
                        77,
                        0,
                        0,
                        144,
                        233
                      ],
                      "Name": "Head_Flying_Hours",
                      "Type": 11,
                      "ValueRaw": 256804684582266
                    },
                    "241": {
                      "Id": 241,
                      "Flags": 0,
                      "Current": 100,
                      "Worst": 253,
                      "VendorBytes": [
                        246,
                        150,
                        21,
                        83,
                        48,
                        0
                      ],
                      "Name": "Total_LBAs_Written",
                      "Type": 3,
                      "ValueRaw": 207552354038
                    },
                    "242": {
                      "Id": 242,
                      "Flags": 0,
                      "Current": 100,
                      "Worst": 253,
                      "VendorBytes": [
                        64,
                        66,
                        193,
                        36,
                        140,
                        13
                      ],
                      "Name": "Total_LBAs_Read",
                      "Type": 3,
                      "ValueRaw": 14895563227712
                    }
                  }
                },
                "Thresholds": {
                  "Revnumber": 1,
                  "Thresholds": {
                    "1": 44,
                    "3": 0,
                    "4": 20,
                    "5": 10,
                    "7": 45,
                    "9": 0,
                    "10": 97,
                    "12": 20,
                    "18": 50,
                    "187": 0,
                    "188": 0,
                    "190": 0,
                    "192": 0,
                    "193": 0,
                    "194": 0,
                    "197": 0,
                    "198": 0,
                    "199": 0,
                    "200": 1,
                    "240": 0,
                    "241": 0,
                    "242": 0
                  }
                }
              }
            },
            {
              "name": "/dev/sdc",
              "type": "disk",
              "size": 16000900661248,
              "rota": true,
              "serial": "ZR602277",
              "wwn": "0x5000c500dc27124c",
              "vendor": "ATA     ",
              "model": "ST16000NM001J-2T",
              "rev": "SS02",
              "children": [
                {
                  "name": "/dev/sdc1",
                  "type": "part",
                  "size": 16000899595776,
                  "rota": true,
                  "wwn": "0x5000c500dc27124c",
                  "fstype": "linux_raid_member",
                  "partuuid": "264ec357-569e-f046-b355-17112452b740",
                  "children": [
                    {
                      "name": "/dev/md2",
                      "type": "raid6",
                      "size": 128006110576640,
                      "rota": true,
                      "fstype": "crypto_LUKS",
                      "children": [
                        {
                          "name": "/dev/mapper/securehdd",
                          "type": "crypt",
                          "size": 128006093799424,
                          "rota": true,
                          "mountpoint": "/mnt/secure",
                          "fstype": "ext4",
                          "children": [],
                          "usage": 110085240667504,
                          "smart": {
                            "Temperature": 0,
                            "Read": 0,
                            "Written": 0,
                            "PowerOnHours": 0,
                            "PowerCycles": 0,
                            "AdditionalData": null,
                            "Thresholds": null
                          }
                        }
                      ],
                      "usage": 0,
                      "smart": {
                        "Temperature": 0,
                        "Read": 0,
                        "Written": 0,
                        "PowerOnHours": 0,
                        "PowerCycles": 0,
                        "AdditionalData": null,
                        "Thresholds": null
                      }
                    }
                  ],
                  "usage": 0,
                  "smart": {
                    "Temperature": 0,
                    "Read": 0,
                    "Written": 0,
                    "PowerOnHours": 0,
                    "PowerCycles": 0,
                    "AdditionalData": null,
                    "Thresholds": null
                  }
                }
              ],
              "usage": 0,
              "smart": {
                "Temperature": 35,
                "Read": 21314056445816,
                "Written": 207888072134,
                "PowerOnHours": 19922,
                "PowerCycles": 4,
                "AdditionalData": {
                  "Version": 10,
                  "Attrs": {
                    "1": {
                      "Id": 1,
                      "Flags": 15,
                      "Current": 84,
                      "Worst": 64,
                      "VendorBytes": [
                        160,
                        48,
                        106,
                        14,
                        0,
                        0
                      ],
                      "Name": "Raw_Read_Error_Rate",
                      "Type": 3,
                      "ValueRaw": 241840288
                    },
                    "3": {
                      "Id": 3,
                      "Flags": 3,
                      "Current": 96,
                      "Worst": 96,
                      "VendorBytes": [
                        0,
                        0,
                        0,
                        0,
                        0,
                        0
                      ],
                      "Name": "Spin_Up_Time",
                      "Type": 10,
                      "ValueRaw": 0
                    },
                    "4": {
                      "Id": 4,
                      "Flags": 50,
                      "Current": 100,
                      "Worst": 100,
                      "VendorBytes": [
                        4,
                        0,
                        0,
                        0,
                        0,
                        0
                      ],
                      "Name": "Start_Stop_Count",
                      "Type": 3,
                      "ValueRaw": 4
                    },
                    "5": {
                      "Id": 5,
                      "Flags": 51,
                      "Current": 100,
                      "Worst": 100,
                      "VendorBytes": [
                        0,
                        0,
                        0,
                        0,
                        0,
                        0
                      ],
                      "Name": "Reallocated_Sector_Ct",
                      "Type": 9,
                      "ValueRaw": 0
                    },
                    "7": {
                      "Id": 7,
                      "Flags": 15,
                      "Current": 87,
                      "Worst": 60,
                      "VendorBytes": [
                        119,
                        55,
                        43,
                        32,
                        0,
                        0
                      ],
                      "Name": "Seek_Error_Rate",
                      "Type": 3,
                      "ValueRaw": 539703159
                    },
                    "9": {
                      "Id": 9,
                      "Flags": 50,
                      "Current": 78,
                      "Worst": 78,
                      "VendorBytes": [
                        210,
                        77,
                        0,
                        0,
                        0,
                        0
                      ],
                      "Name": "Power_On_Hours",
                      "Type": 11,
                      "ValueRaw": 19922
                    },
                    "10": {
                      "Id": 10,
                      "Flags": 19,
                      "Current": 100,
                      "Worst": 100,
                      "VendorBytes": [
                        0,
                        0,
                        0,
                        0,
                        0,
                        0
                      ],
                      "Name": "Spin_Retry_Count",
                      "Type": 3,
                      "ValueRaw": 0
                    },
                    "12": {
                      "Id": 12,
                      "Flags": 50,
                      "Current": 100,
                      "Worst": 100,
                      "VendorBytes": [
                        4,
                        0,
                        0,
                        0,
                        0,
                        0
                      ],
                      "Name": "Power_Cycle_Count",
                      "Type": 3,
                      "ValueRaw": 4
                    },
                    "18": {
                      "Id": 18,
                      "Flags": 11,
                      "Current": 100,
                      "Worst": 100,
                      "VendorBytes": [
                        0,
                        0,
                        0,
                        0,
                        0,
                        0
                      ],
                      "Name": "",
                      "Type": 0,
                      "ValueRaw": 0
                    },
                    "187": {
                      "Id": 187,
                      "Flags": 50,
                      "Current": 100,
                      "Worst": 100,
                      "VendorBytes": [
                        0,
                        0,
                        0,
                        0,
                        0,
                        0
                      ],
                      "Name": "Reported_Uncorrect",
                      "Type": 3,
                      "ValueRaw": 0
                    },
                    "188": {
                      "Id": 188,
                      "Flags": 50,
                      "Current": 100,
                      "Worst": 100,
                      "VendorBytes": [
                        0,
                        0,
                        0,
                        0,
                        0,
                        0
                      ],
                      "Name": "Command_Timeout",
                      "Type": 3,
                      "ValueRaw": 0
                    },
                    "190": {
                      "Id": 190,
                      "Flags": 34,
                      "Current": 65,
                      "Worst": 49,
                      "VendorBytes": [
                        35,
                        0,
                        32,
                        41,
                        0,
                        0
                      ],
                      "Name": "Airflow_Temperature_Cel",
                      "Type": 18,
                      "ValueRaw": 689963043
                    },
                    "192": {
                      "Id": 192,
                      "Flags": 50,
                      "Current": 100,
                      "Worst": 100,
                      "VendorBytes": [
                        1,
                        0,
                        0,
                        0,
                        0,
                        0
                      ],
                      "Name": "Power-Off_Retract_Count",
                      "Type": 3,
                      "ValueRaw": 1
                    },
                    "193": {
                      "Id": 193,
                      "Flags": 50,
                      "Current": 100,
                      "Worst": 100,
                      "VendorBytes": [
                        85,
                        3,
                        0,
                        0,
                        0,
                        0
                      ],
                      "Name": "Load_Cycle_Count",
                      "Type": 3,
                      "ValueRaw": 853
                    },
                    "194": {
                      "Id": 194,
                      "Flags": 34,
                      "Current": 35,
                      "Worst": 41,
                      "VendorBytes": [
                        35,
                        0,
                        0,
                        0,
                        23,
                        0
                      ],
                      "Name": "Temperature_Celsius",
                      "Type": 18,
                      "ValueRaw": 98784247843
                    },
                    "197": {
                      "Id": 197,
                      "Flags": 18,
                      "Current": 100,
                      "Worst": 100,
                      "VendorBytes": [
                        0,
                        0,
                        0,
                        0,
                        0,
                        0
                      ],
                      "Name": "Current_Pending_Sector",
                      "Type": 3,
                      "ValueRaw": 0
                    },
                    "198": {
                      "Id": 198,
                      "Flags": 16,
                      "Current": 100,
                      "Worst": 100,
                      "VendorBytes": [
                        0,
                        0,
                        0,
                        0,
                        0,
                        0
                      ],
                      "Name": "Offline_Uncorrectable",
                      "Type": 3,
                      "ValueRaw": 0
                    },
                    "199": {
                      "Id": 199,
                      "Flags": 62,
                      "Current": 200,
                      "Worst": 200,
                      "VendorBytes": [
                        0,
                        0,
                        0,
                        0,
                        0,
                        0
                      ],
                      "Name": "UDMA_CRC_Error_Count",
                      "Type": 3,
                      "ValueRaw": 0
                    },
                    "200": {
                      "Id": 200,
                      "Flags": 35,
                      "Current": 100,
                      "Worst": 100,
                      "VendorBytes": [
                        0,
                        0,
                        0,
                        0,
                        0,
                        0
                      ],
                      "Name": "Multi_Zone_Error_Rate",
                      "Type": 3,
                      "ValueRaw": 0
                    },
                    "240": {
                      "Id": 240,
                      "Flags": 0,
                      "Current": 100,
                      "Worst": 253,
                      "VendorBytes": [
                        124,
                        77,
                        0,
                        0,
                        86,
                        123
                      ],
                      "Name": "Head_Flying_Hours",
                      "Type": 11,
                      "ValueRaw": 135609297423740
                    },
                    "241": {
                      "Id": 241,
                      "Flags": 0,
                      "Current": 100,
                      "Worst": 253,
                      "VendorBytes": [
                        198,
                        61,
                        24,
                        103,
                        48,
                        0
                      ],
                      "Name": "Total_LBAs_Written",
                      "Type": 3,
                      "ValueRaw": 207888072134
                    },
                    "242": {
                      "Id": 242,
                      "Flags": 0,
                      "Current": 100,
                      "Worst": 253,
                      "VendorBytes": [
                        120,
                        95,
                        195,
                        144,
                        98,
                        19
                      ],
                      "Name": "Total_LBAs_Read",
                      "Type": 3,
                      "ValueRaw": 21314056445816
                    }
                  }
                },
                "Thresholds": {
                  "Revnumber": 1,
                  "Thresholds": {
                    "1": 44,
                    "3": 0,
                    "4": 20,
                    "5": 10,
                    "7": 45,
                    "9": 0,
                    "10": 97,
                    "12": 20,
                    "18": 50,
                    "187": 0,
                    "188": 0,
                    "190": 0,
                    "192": 0,
                    "193": 0,
                    "194": 0,
                    "197": 0,
                    "198": 0,
                    "199": 0,
                    "200": 1,
                    "240": 0,
                    "241": 0,
                    "242": 0
                  }
                }
              }
            },
            {
              "name": "/dev/sdd",
              "type": "disk",
              "size": 16000900661248,
              "rota": true,
              "serial": "ZR602233",
              "wwn": "0x5000c500dc27a3e8",
              "vendor": "ATA     ",
              "model": "ST16000NM001J-2T",
              "rev": "SS02",
              "children": [
                {
                  "name": "/dev/sdd1",
                  "type": "part",
                  "size": 16000899595776,
                  "rota": true,
                  "wwn": "0x5000c500dc27a3e8",
                  "fstype": "linux_raid_member",
                  "partuuid": "dc8accb8-3ef1-824b-ae72-a2a018f8c93e",
                  "children": [
                    {
                      "name": "/dev/md2",
                      "type": "raid6",
                      "size": 128006110576640,
                      "rota": true,
                      "fstype": "crypto_LUKS",
                      "children": [
                        {
                          "name": "/dev/mapper/securehdd",
                          "type": "crypt",
                          "size": 128006093799424,
                          "rota": true,
                          "mountpoint": "/mnt/secure",
                          "fstype": "ext4",
                          "children": [],
                          "usage": 110085240667504,
                          "smart": {
                            "Temperature": 0,
                            "Read": 0,
                            "Written": 0,
                            "PowerOnHours": 0,
                            "PowerCycles": 0,
                            "AdditionalData": null,
                            "Thresholds": null
                          }
                        }
                      ],
                      "usage": 0,
                      "smart": {
                        "Temperature": 0,
                        "Read": 0,
                        "Written": 0,
                        "PowerOnHours": 0,
                        "PowerCycles": 0,
                        "AdditionalData": null,
                        "Thresholds": null
                      }
                    }
                  ],
                  "usage": 0,
                  "smart": {
                    "Temperature": 0,
                    "Read": 0,
                    "Written": 0,
                    "PowerOnHours": 0,
                    "PowerCycles": 0,
                    "AdditionalData": null,
                    "Thresholds": null
                  }
                }
              ],
              "usage": 0,
              "smart": {
                "Temperature": 37,
                "Read": 14698169801961,
                "Written": 208620021928,
                "PowerOnHours": 19922,
                "PowerCycles": 4,
                "AdditionalData": {
                  "Version": 10,
                  "Attrs": {
                    "1": {
                      "Id": 1,
                      "Flags": 15,
                      "Current": 75,
                      "Worst": 64,
                      "VendorBytes": [
                        226,
                        153,
                        191,
                        1,
                        0,
                        0
                      ],
                      "Name": "Raw_Read_Error_Rate",
                      "Type": 3,
                      "ValueRaw": 29333986
                    },
                    "3": {
                      "Id": 3,
                      "Flags": 3,
                      "Current": 96,
                      "Worst": 96,
                      "VendorBytes": [
                        0,
                        0,
                        0,
                        0,
                        0,
                        0
                      ],
                      "Name": "Spin_Up_Time",
                      "Type": 10,
                      "ValueRaw": 0
                    },
                    "4": {
                      "Id": 4,
                      "Flags": 50,
                      "Current": 100,
                      "Worst": 100,
                      "VendorBytes": [
                        4,
                        0,
                        0,
                        0,
                        0,
                        0
                      ],
                      "Name": "Start_Stop_Count",
                      "Type": 3,
                      "ValueRaw": 4
                    },
                    "5": {
                      "Id": 5,
                      "Flags": 51,
                      "Current": 100,
                      "Worst": 100,
                      "VendorBytes": [
                        0,
                        0,
                        0,
                        0,
                        0,
                        0
                      ],
                      "Name": "Reallocated_Sector_Ct",
                      "Type": 9,
                      "ValueRaw": 0
                    },
                    "7": {
                      "Id": 7,
                      "Flags": 15,
                      "Current": 88,
                      "Worst": 60,
                      "VendorBytes": [
                        129,
                        21,
                        158,
                        33,
                        0,
                        0
                      ],
                      "Name": "Seek_Error_Rate",
                      "Type": 3,
                      "ValueRaw": 564008321
                    },
                    "9": {
                      "Id": 9,
                      "Flags": 50,
                      "Current": 78,
                      "Worst": 78,
                      "VendorBytes": [
                        210,
                        77,
                        0,
                        0,
                        0,
                        0
                      ],
                      "Name": "Power_On_Hours",
                      "Type": 11,
                      "ValueRaw": 19922
                    },
                    "10": {
                      "Id": 10,
                      "Flags": 19,
                      "Current": 100,
                      "Worst": 100,
                      "VendorBytes": [
                        0,
                        0,
                        0,
                        0,
                        0,
                        0
                      ],
                      "Name": "Spin_Retry_Count",
                      "Type": 3,
                      "ValueRaw": 0
                    },
                    "12": {
                      "Id": 12,
                      "Flags": 50,
                      "Current": 100,
                      "Worst": 100,
                      "VendorBytes": [
                        4,
                        0,
                        0,
                        0,
                        0,
                        0
                      ],
                      "Name": "Power_Cycle_Count",
                      "Type": 3,
                      "ValueRaw": 4
                    },
                    "18": {
                      "Id": 18,
                      "Flags": 11,
                      "Current": 100,
                      "Worst": 100,
                      "VendorBytes": [
                        0,
                        0,
                        0,
                        0,
                        0,
                        0
                      ],
                      "Name": "",
                      "Type": 0,
                      "ValueRaw": 0
                    },
                    "187": {
                      "Id": 187,
                      "Flags": 50,
                      "Current": 100,
                      "Worst": 100,
                      "VendorBytes": [
                        0,
                        0,
                        0,
                        0,
                        0,
                        0
                      ],
                      "Name": "Reported_Uncorrect",
                      "Type": 3,
                      "ValueRaw": 0
                    },
                    "188": {
                      "Id": 188,
                      "Flags": 50,
                      "Current": 100,
                      "Worst": 100,
                      "VendorBytes": [
                        0,
                        0,
                        0,
                        0,
                        0,
                        0
                      ],
                      "Name": "Command_Timeout",
                      "Type": 3,
                      "ValueRaw": 0
                    },
                    "190": {
                      "Id": 190,
                      "Flags": 34,
                      "Current": 63,
                      "Worst": 48,
                      "VendorBytes": [
                        37,
                        0,
                        32,
                        42,
                        0,
                        0
                      ],
                      "Name": "Airflow_Temperature_Cel",
                      "Type": 18,
                      "ValueRaw": 706740261
                    },
                    "192": {
                      "Id": 192,
                      "Flags": 50,
                      "Current": 100,
                      "Worst": 100,
                      "VendorBytes": [
                        1,
                        0,
                        0,
                        0,
                        0,
                        0
                      ],
                      "Name": "Power-Off_Retract_Count",
                      "Type": 3,
                      "ValueRaw": 1
                    },
                    "193": {
                      "Id": 193,
                      "Flags": 50,
                      "Current": 100,
                      "Worst": 100,
                      "VendorBytes": [
                        85,
                        3,
                        0,
                        0,
                        0,
                        0
                      ],
                      "Name": "Load_Cycle_Count",
                      "Type": 3,
                      "ValueRaw": 853
                    },
                    "194": {
                      "Id": 194,
                      "Flags": 34,
                      "Current": 37,
                      "Worst": 42,
                      "VendorBytes": [
                        37,
                        0,
                        0,
                        0,
                        22,
                        0
                      ],
                      "Name": "Temperature_Celsius",
                      "Type": 18,
                      "ValueRaw": 94489280549
                    },
                    "197": {
                      "Id": 197,
                      "Flags": 18,
                      "Current": 100,
                      "Worst": 100,
                      "VendorBytes": [
                        0,
                        0,
                        0,
                        0,
                        0,
                        0
                      ],
                      "Name": "Current_Pending_Sector",
                      "Type": 3,
                      "ValueRaw": 0
                    },
                    "198": {
                      "Id": 198,
                      "Flags": 16,
                      "Current": 100,
                      "Worst": 100,
                      "VendorBytes": [
                        0,
                        0,
                        0,
                        0,
                        0,
                        0
                      ],
                      "Name": "Offline_Uncorrectable",
                      "Type": 3,
                      "ValueRaw": 0
                    },
                    "199": {
                      "Id": 199,
                      "Flags": 62,
                      "Current": 200,
                      "Worst": 200,
                      "VendorBytes": [
                        0,
                        0,
                        0,
                        0,
                        0,
                        0
                      ],
                      "Name": "UDMA_CRC_Error_Count",
                      "Type": 3,
                      "ValueRaw": 0
                    },
                    "200": {
                      "Id": 200,
                      "Flags": 35,
                      "Current": 100,
                      "Worst": 100,
                      "VendorBytes": [
                        0,
                        0,
                        0,
                        0,
                        0,
                        0
                      ],
                      "Name": "Multi_Zone_Error_Rate",
                      "Type": 3,
                      "ValueRaw": 0
                    },
                    "240": {
                      "Id": 240,
                      "Flags": 0,
                      "Current": 100,
                      "Worst": 253,
                      "VendorBytes": [
                        123,
                        77,
                        0,
                        0,
                        184,
                        173
                      ],
                      "Name": "Head_Flying_Hours",
                      "Type": 11,
                      "ValueRaw": 191005785607547
                    },
                    "241": {
                      "Id": 241,
                      "Flags": 0,
                      "Current": 100,
                      "Worst": 253,
                      "VendorBytes": [
                        168,
                        232,
                        184,
                        146,
                        48,
                        0
                      ],
                      "Name": "Total_LBAs_Written",
                      "Type": 3,
                      "ValueRaw": 208620021928
                    },
                    "242": {
                      "Id": 242,
                      "Flags": 0,
                      "Current": 100,
                      "Worst": 253,
                      "VendorBytes": [
                        233,
                        156,
                        48,
                        47,
                        94,
                        13
                      ],
                      "Name": "Total_LBAs_Read",
                      "Type": 3,
                      "ValueRaw": 14698169801961
                    }
                  }
                },
                "Thresholds": {
                  "Revnumber": 1,
                  "Thresholds": {
                    "1": 44,
                    "3": 0,
                    "4": 20,
                    "5": 10,
                    "7": 45,
                    "9": 0,
                    "10": 97,
                    "12": 20,
                    "18": 50,
                    "187": 0,
                    "188": 0,
                    "190": 0,
                    "192": 0,
                    "193": 0,
                    "194": 0,
                    "197": 0,
                    "198": 0,
                    "199": 0,
                    "200": 1,
                    "240": 0,
                    "241": 0,
                    "242": 0
                  }
                }
              }
            },
            {
              "name": "/dev/sde",
              "type": "disk",
              "size": 16000900661248,
              "rota": true,
              "serial": "71A0A063FVNG",
              "wwn": "0x5000039b08d057c3",
              "vendor": "ATA     ",
              "model": "TOSHIBA_MG08ACA1",
              "rev": "0103",
              "children": [
                {
                  "name": "/dev/sde1",
                  "type": "part",
                  "size": 16000899595776,
                  "rota": true,
                  "wwn": "0x5000039b08d057c3",
                  "fstype": "linux_raid_member",
                  "partuuid": "c459597c-45ef-c24d-83a5-f382b43893cf",
                  "children": [
                    {
                      "name": "/dev/md2",
                      "type": "raid6",
                      "size": 128006110576640,
                      "rota": true,
                      "fstype": "crypto_LUKS",
                      "children": [
                        {
                          "name": "/dev/mapper/securehdd",
                          "type": "crypt",
                          "size": 128006093799424,
                          "rota": true,
                          "mountpoint": "/mnt/secure",
                          "fstype": "ext4",
                          "children": [],
                          "usage": 110085240667504,
                          "smart": {
                            "Temperature": 0,
                            "Read": 0,
                            "Written": 0,
                            "PowerOnHours": 0,
                            "PowerCycles": 0,
                            "AdditionalData": null,
                            "Thresholds": null
                          }
                        }
                      ],
                      "usage": 0,
                      "smart": {
                        "Temperature": 0,
                        "Read": 0,
                        "Written": 0,
                        "PowerOnHours": 0,
                        "PowerCycles": 0,
                        "AdditionalData": null,
                        "Thresholds": null
                      }
                    }
                  ],
                  "usage": 0,
                  "smart": {
                    "Temperature": 0,
                    "Read": 0,
                    "Written": 0,
                    "PowerOnHours": 0,
                    "PowerCycles": 0,
                    "AdditionalData": null,
                    "Thresholds": null
                  }
                }
              ],
              "usage": 0,
              "smart": {
                "Temperature": 34,
                "Read": 0,
                "Written": 0,
                "PowerOnHours": 20390,
                "PowerCycles": 14,
                "AdditionalData": {
                  "Version": 16,
                  "Attrs": {
                    "1": {
                      "Id": 1,
                      "Flags": 11,
                      "Current": 100,
                      "Worst": 100,
                      "VendorBytes": [
                        0,
                        0,
                        0,
                        0,
                        0,
                        0
                      ],
                      "Name": "Raw_Read_Error_Rate",
                      "Type": 3,
                      "ValueRaw": 0
                    },
                    "2": {
                      "Id": 2,
                      "Flags": 5,
                      "Current": 100,
                      "Worst": 100,
                      "VendorBytes": [
                        0,
                        0,
                        0,
                        0,
                        0,
                        0
                      ],
                      "Name": "Throughput_Performance",
                      "Type": 3,
                      "ValueRaw": 0
                    },
                    "3": {
                      "Id": 3,
                      "Flags": 39,
                      "Current": 100,
                      "Worst": 100,
                      "VendorBytes": [
                        19,
                        31,
                        0,
                        0,
                        0,
                        0
                      ],
                      "Name": "Spin_Up_Time",
                      "Type": 10,
                      "ValueRaw": 7955
                    },
                    "4": {
                      "Id": 4,
                      "Flags": 50,
                      "Current": 100,
                      "Worst": 100,
                      "VendorBytes": [
                        14,
                        0,
                        0,
                        0,
                        0,
                        0
                      ],
                      "Name": "Start_Stop_Count",
                      "Type": 3,
                      "ValueRaw": 14
                    },
                    "5": {
                      "Id": 5,
                      "Flags": 51,
                      "Current": 100,
                      "Worst": 100,
                      "VendorBytes": [
                        0,
                        0,
                        0,
                        0,
                        0,
                        0
                      ],
                      "Name": "Reallocated_Sector_Ct",
                      "Type": 9,
                      "ValueRaw": 0
                    },
                    "7": {
                      "Id": 7,
                      "Flags": 11,
                      "Current": 100,
                      "Worst": 100,
                      "VendorBytes": [
                        0,
                        0,
                        0,
                        0,
                        0,
                        0
                      ],
                      "Name": "Seek_Error_Rate",
                      "Type": 3,
                      "ValueRaw": 0
                    },
                    "8": {
                      "Id": 8,
                      "Flags": 5,
                      "Current": 100,
                      "Worst": 100,
                      "VendorBytes": [
                        0,
                        0,
                        0,
                        0,
                        0,
                        0
                      ],
                      "Name": "Seek_Time_Performance",
                      "Type": 3,
                      "ValueRaw": 0
                    },
                    "9": {
                      "Id": 9,
                      "Flags": 50,
                      "Current": 50,
                      "Worst": 50,
                      "VendorBytes": [
                        166,
                        79,
                        0,
                        0,
                        0,
                        0
                      ],
                      "Name": "Power_On_Hours",
                      "Type": 11,
                      "ValueRaw": 20390
                    },
                    "10": {
                      "Id": 10,
                      "Flags": 51,
                      "Current": 100,
                      "Worst": 100,
                      "VendorBytes": [
                        0,
                        0,
                        0,
                        0,
                        0,
                        0
                      ],
                      "Name": "Spin_Retry_Count",
                      "Type": 3,
                      "ValueRaw": 0
                    },
                    "12": {
                      "Id": 12,
                      "Flags": 50,
                      "Current": 100,
                      "Worst": 100,
                      "VendorBytes": [
                        14,
                        0,
                        0,
                        0,
                        0,
                        0
                      ],
                      "Name": "Power_Cycle_Count",
                      "Type": 3,
                      "ValueRaw": 14
                    },
                    "23": {
                      "Id": 23,
                      "Flags": 35,
                      "Current": 100,
                      "Worst": 100,
                      "VendorBytes": [
                        0,
                        0,
                        0,
                        0,
                        0,
                        0
                      ],
                      "Name": "Helium_Condition_Lower",
                      "Type": 3,
                      "ValueRaw": 0
                    },
                    "24": {
                      "Id": 24,
                      "Flags": 35,
                      "Current": 100,
                      "Worst": 100,
                      "VendorBytes": [
                        0,
                        0,
                        0,
                        0,
                        0,
                        0
                      ],
                      "Name": "Helium_Condition_Upper",
                      "Type": 3,
                      "ValueRaw": 0
                    },
                    "191": {
                      "Id": 191,
                      "Flags": 50,
                      "Current": 100,
                      "Worst": 100,
                      "VendorBytes": [
                        1,
                        0,
                        0,
                        0,
                        0,
                        0
                      ],
                      "Name": "G-Sense_Error_Rate",
                      "Type": 3,
                      "ValueRaw": 1
                    },
                    "192": {
                      "Id": 192,
                      "Flags": 50,
                      "Current": 100,
                      "Worst": 100,
                      "VendorBytes": [
                        9,
                        0,
                        0,
                        0,
                        0,
                        0
                      ],
                      "Name": "Power-Off_Retract_Count",
                      "Type": 3,
                      "ValueRaw": 9
                    },
                    "193": {
                      "Id": 193,
                      "Flags": 50,
                      "Current": 100,
                      "Worst": 100,
                      "VendorBytes": [
                        34,
                        0,
                        0,
                        0,
                        0,
                        0
                      ],
                      "Name": "Load_Cycle_Count",
                      "Type": 3,
                      "ValueRaw": 34
                    },
                    "194": {
                      "Id": 194,
                      "Flags": 34,
                      "Current": 100,
                      "Worst": 100,
                      "VendorBytes": [
                        34,
                        0,
                        21,
                        0,
                        47,
                        0
                      ],
                      "Name": "Temperature_Celsius",
                      "Type": 18,
                      "ValueRaw": 201864839202
                    },
                    "196": {
                      "Id": 196,
                      "Flags": 50,
                      "Current": 100,
                      "Worst": 100,
                      "VendorBytes": [
                        0,
                        0,
                        0,
                        0,
                        0,
                        0
                      ],
                      "Name": "Reallocated_Event_Count",
                      "Type": 9,
                      "ValueRaw": 0
                    },
                    "197": {
                      "Id": 197,
                      "Flags": 50,
                      "Current": 100,
                      "Worst": 100,
                      "VendorBytes": [
                        0,
                        0,
                        0,
                        0,
                        0,
                        0
                      ],
                      "Name": "Current_Pending_Sector",
                      "Type": 3,
                      "ValueRaw": 0
                    },
                    "198": {
                      "Id": 198,
                      "Flags": 48,
                      "Current": 100,
                      "Worst": 100,
                      "VendorBytes": [
                        0,
                        0,
                        0,
                        0,
                        0,
                        0
                      ],
                      "Name": "Offline_Uncorrectable",
                      "Type": 3,
                      "ValueRaw": 0
                    },
                    "199": {
                      "Id": 199,
                      "Flags": 50,
                      "Current": 200,
                      "Worst": 200,
                      "VendorBytes": [
                        0,
                        0,
                        0,
                        0,
                        0,
                        0
                      ],
                      "Name": "UDMA_CRC_Error_Count",
                      "Type": 3,
                      "ValueRaw": 0
                    },
                    "220": {
                      "Id": 220,
                      "Flags": 2,
                      "Current": 100,
                      "Worst": 100,
                      "VendorBytes": [
                        2,
                        0,
                        6,
                        9,
                        0,
                        0
                      ],
                      "Name": "Disk_Shift",
                      "Type": 3,
                      "ValueRaw": 151388162
                    },
                    "222": {
                      "Id": 222,
                      "Flags": 50,
                      "Current": 51,
                      "Worst": 51,
                      "VendorBytes": [
                        144,
                        77,
                        0,
                        0,
                        0,
                        0
                      ],
                      "Name": "Loaded_Hours",
                      "Type": 3,
                      "ValueRaw": 19856
                    },
                    "223": {
                      "Id": 223,
                      "Flags": 50,
                      "Current": 100,
                      "Worst": 100,
                      "VendorBytes": [
                        0,
                        0,
                        0,
                        0,
                        0,
                        0
                      ],
                      "Name": "Load_Retry_Count",
                      "Type": 3,
                      "ValueRaw": 0
                    },
                    "224": {
                      "Id": 224,
                      "Flags": 34,
                      "Current": 100,
                      "Worst": 100,
                      "VendorBytes": [
                        0,
                        0,
                        0,
                        0,
                        0,
                        0
                      ],
                      "Name": "Load_Friction",
                      "Type": 3,
                      "ValueRaw": 0
                    },
                    "226": {
                      "Id": 226,
                      "Flags": 38,
                      "Current": 100,
                      "Worst": 100,
                      "VendorBytes": [
                        18,
                        2,
                        0,
                        0,
                        0,
                        0
                      ],
                      "Name": "Load-in_Time",
                      "Type": 3,
                      "ValueRaw": 530
                    },
                    "240": {
                      "Id": 240,
                      "Flags": 1,
                      "Current": 100,
                      "Worst": 100,
                      "VendorBytes": [
                        0,
                        0,
                        0,
                        0,
                        0,
                        0
                      ],
                      "Name": "Head_Flying_Hours",
                      "Type": 11,
                      "ValueRaw": 0
                    }
                  }
                },
                "Thresholds": {
                  "Revnumber": 16,
                  "Thresholds": {
                    "1": 50,
                    "2": 50,
                    "3": 1,
                    "4": 0,
                    "5": 10,
                    "7": 50,
                    "8": 50,
                    "9": 0,
                    "10": 30,
                    "12": 0,
                    "23": 75,
                    "24": 75,
                    "191": 0,
                    "192": 0,
                    "193": 0,
                    "194": 0,
                    "196": 0,
                    "197": 0,
                    "198": 0,
                    "199": 0,
                    "220": 0,
                    "222": 0,
                    "223": 0,
                    "224": 0,
                    "226": 0,
                    "240": 1
                  }
                }
              }
            },
            {
              "name": "/dev/sdf",
              "type": "disk",
              "size": 16000900661248,
              "rota": true,
              "serial": "ZR7025EB",
              "wwn": "0x5000c500dc26fe9e",
              "vendor": "ATA     ",
              "model": "ST16000NM001J-2T",
              "rev": "SS02",
              "children": [
                {
                  "name": "/dev/sdf1",
                  "type": "part",
                  "size": 16000899595776,
                  "rota": true,
                  "wwn": "0x5000c500dc26fe9e",
                  "fstype": "linux_raid_member",
                  "partuuid": "f0d37b26-86c0-de42-afb6-0b6f573ef05a",
                  "children": [
                    {
                      "name": "/dev/md2",
                      "type": "raid6",
                      "size": 128006110576640,
                      "rota": true,
                      "fstype": "crypto_LUKS",
                      "children": [
                        {
                          "name": "/dev/mapper/securehdd",
                          "type": "crypt",
                          "size": 128006093799424,
                          "rota": true,
                          "mountpoint": "/mnt/secure",
                          "fstype": "ext4",
                          "children": [],
                          "usage": 110085240667504,
                          "smart": {
                            "Temperature": 0,
                            "Read": 0,
                            "Written": 0,
                            "PowerOnHours": 0,
                            "PowerCycles": 0,
                            "AdditionalData": null,
                            "Thresholds": null
                          }
                        }
                      ],
                      "usage": 0,
                      "smart": {
                        "Temperature": 0,
                        "Read": 0,
                        "Written": 0,
                        "PowerOnHours": 0,
                        "PowerCycles": 0,
                        "AdditionalData": null,
                        "Thresholds": null
                      }
                    }
                  ],
                  "usage": 0,
                  "smart": {
                    "Temperature": 0,
                    "Read": 0,
                    "Written": 0,
                    "PowerOnHours": 0,
                    "PowerCycles": 0,
                    "AdditionalData": null,
                    "Thresholds": null
                  }
                }
              ],
              "usage": 0,
              "smart": {
                "Temperature": 37,
                "Read": 14315903696107,
                "Written": 208003444746,
                "PowerOnHours": 19921,
                "PowerCycles": 4,
                "AdditionalData": {
                  "Version": 10,
                  "Attrs": {
                    "1": {
                      "Id": 1,
                      "Flags": 15,
                      "Current": 79,
                      "Worst": 64,
                      "VendorBytes": [
                        134,
                        73,
                        36,
                        5,
                        0,
                        0
                      ],
                      "Name": "Raw_Read_Error_Rate",
                      "Type": 3,
                      "ValueRaw": 86264198
                    },
                    "3": {
                      "Id": 3,
                      "Flags": 3,
                      "Current": 96,
                      "Worst": 96,
                      "VendorBytes": [
                        0,
                        0,
                        0,
                        0,
                        0,
                        0
                      ],
                      "Name": "Spin_Up_Time",
                      "Type": 10,
                      "ValueRaw": 0
                    },
                    "4": {
                      "Id": 4,
                      "Flags": 50,
                      "Current": 100,
                      "Worst": 100,
                      "VendorBytes": [
                        4,
                        0,
                        0,
                        0,
                        0,
                        0
                      ],
                      "Name": "Start_Stop_Count",
                      "Type": 3,
                      "ValueRaw": 4
                    },
                    "5": {
                      "Id": 5,
                      "Flags": 51,
                      "Current": 100,
                      "Worst": 100,
                      "VendorBytes": [
                        0,
                        0,
                        0,
                        0,
                        0,
                        0
                      ],
                      "Name": "Reallocated_Sector_Ct",
                      "Type": 9,
                      "ValueRaw": 0
                    },
                    "7": {
                      "Id": 7,
                      "Flags": 15,
                      "Current": 88,
                      "Worst": 60,
                      "VendorBytes": [
                        191,
                        239,
                        118,
                        34,
                        0,
                        0
                      ],
                      "Name": "Seek_Error_Rate",
                      "Type": 3,
                      "ValueRaw": 578219967
                    },
                    "9": {
                      "Id": 9,
                      "Flags": 50,
                      "Current": 78,
                      "Worst": 78,
                      "VendorBytes": [
                        209,
                        77,
                        0,
                        0,
                        0,
                        0
                      ],
                      "Name": "Power_On_Hours",
                      "Type": 11,
                      "ValueRaw": 19921
                    },
                    "10": {
                      "Id": 10,
                      "Flags": 19,
                      "Current": 100,
                      "Worst": 100,
                      "VendorBytes": [
                        0,
                        0,
                        0,
                        0,
                        0,
                        0
                      ],
                      "Name": "Spin_Retry_Count",
                      "Type": 3,
                      "ValueRaw": 0
                    },
                    "12": {
                      "Id": 12,
                      "Flags": 50,
                      "Current": 100,
                      "Worst": 100,
                      "VendorBytes": [
                        4,
                        0,
                        0,
                        0,
                        0,
                        0
                      ],
                      "Name": "Power_Cycle_Count",
                      "Type": 3,
                      "ValueRaw": 4
                    },
                    "18": {
                      "Id": 18,
                      "Flags": 11,
                      "Current": 100,
                      "Worst": 100,
                      "VendorBytes": [
                        0,
                        0,
                        0,
                        0,
                        0,
                        0
                      ],
                      "Name": "",
                      "Type": 0,
                      "ValueRaw": 0
                    },
                    "187": {
                      "Id": 187,
                      "Flags": 50,
                      "Current": 100,
                      "Worst": 100,
                      "VendorBytes": [
                        0,
                        0,
                        0,
                        0,
                        0,
                        0
                      ],
                      "Name": "Reported_Uncorrect",
                      "Type": 3,
                      "ValueRaw": 0
                    },
                    "188": {
                      "Id": 188,
                      "Flags": 50,
                      "Current": 100,
                      "Worst": 100,
                      "VendorBytes": [
                        0,
                        0,
                        0,
                        0,
                        0,
                        0
                      ],
                      "Name": "Command_Timeout",
                      "Type": 3,
                      "ValueRaw": 0
                    },
                    "190": {
                      "Id": 190,
                      "Flags": 34,
                      "Current": 63,
                      "Worst": 48,
                      "VendorBytes": [
                        37,
                        0,
                        31,
                        42,
                        0,
                        0
                      ],
                      "Name": "Airflow_Temperature_Cel",
                      "Type": 18,
                      "ValueRaw": 706674725
                    },
                    "192": {
                      "Id": 192,
                      "Flags": 50,
                      "Current": 100,
                      "Worst": 100,
                      "VendorBytes": [
                        1,
                        0,
                        0,
                        0,
                        0,
                        0
                      ],
                      "Name": "Power-Off_Retract_Count",
                      "Type": 3,
                      "ValueRaw": 1
                    },
                    "193": {
                      "Id": 193,
                      "Flags": 50,
                      "Current": 100,
                      "Worst": 100,
                      "VendorBytes": [
                        86,
                        3,
                        0,
                        0,
                        0,
                        0
                      ],
                      "Name": "Load_Cycle_Count",
                      "Type": 3,
                      "ValueRaw": 854
                    },
                    "194": {
                      "Id": 194,
                      "Flags": 34,
                      "Current": 37,
                      "Worst": 42,
                      "VendorBytes": [
                        37,
                        0,
                        0,
                        0,
                        22,
                        0
                      ],
                      "Name": "Temperature_Celsius",
                      "Type": 18,
                      "ValueRaw": 94489280549
                    },
                    "197": {
                      "Id": 197,
                      "Flags": 18,
                      "Current": 100,
                      "Worst": 100,
                      "VendorBytes": [
                        0,
                        0,
                        0,
                        0,
                        0,
                        0
                      ],
                      "Name": "Current_Pending_Sector",
                      "Type": 3,
                      "ValueRaw": 0
                    },
                    "198": {
                      "Id": 198,
                      "Flags": 16,
                      "Current": 100,
                      "Worst": 100,
                      "VendorBytes": [
                        0,
                        0,
                        0,
                        0,
                        0,
                        0
                      ],
                      "Name": "Offline_Uncorrectable",
                      "Type": 3,
                      "ValueRaw": 0
                    },
                    "199": {
                      "Id": 199,
                      "Flags": 62,
                      "Current": 200,
                      "Worst": 200,
                      "VendorBytes": [
                        0,
                        0,
                        0,
                        0,
                        0,
                        0
                      ],
                      "Name": "UDMA_CRC_Error_Count",
                      "Type": 3,
                      "ValueRaw": 0
                    },
                    "200": {
                      "Id": 200,
                      "Flags": 35,
                      "Current": 100,
                      "Worst": 100,
                      "VendorBytes": [
                        0,
                        0,
                        0,
                        0,
                        0,
                        0
                      ],
                      "Name": "Multi_Zone_Error_Rate",
                      "Type": 3,
                      "ValueRaw": 0
                    },
                    "240": {
                      "Id": 240,
                      "Flags": 0,
                      "Current": 100,
                      "Worst": 253,
                      "VendorBytes": [
                        122,
                        77,
                        0,
                        0,
                        77,
                        142
                      ],
                      "Name": "Head_Flying_Hours",
                      "Type": 11,
                      "ValueRaw": 156461363645818
                    },
                    "241": {
                      "Id": 241,
                      "Flags": 0,
                      "Current": 100,
                      "Worst": 253,
                      "VendorBytes": [
                        10,
                        176,
                        248,
                        109,
                        48,
                        0
                      ],
                      "Name": "Total_LBAs_Written",
                      "Type": 3,
                      "ValueRaw": 208003444746
                    },
                    "242": {
                      "Id": 242,
                      "Flags": 0,
                      "Current": 100,
                      "Worst": 253,
                      "VendorBytes": [
                        235,
                        188,
                        90,
                        46,
                        5,
                        13
                      ],
                      "Name": "Total_LBAs_Read",
                      "Type": 3,
                      "ValueRaw": 14315903696107
                    }
                  }
                },
                "Thresholds": {
                  "Revnumber": 1,
                  "Thresholds": {
                    "1": 44,
                    "3": 0,
                    "4": 20,
                    "5": 10,
                    "7": 45,
                    "9": 0,
                    "10": 97,
                    "12": 20,
                    "18": 50,
                    "187": 0,
                    "188": 0,
                    "190": 0,
                    "192": 0,
                    "193": 0,
                    "194": 0,
                    "197": 0,
                    "198": 0,
                    "199": 0,
                    "200": 1,
                    "240": 0,
                    "241": 0,
                    "242": 0
                  }
                }
              }
            },
            {
              "name": "/dev/sdg",
              "type": "disk",
              "size": 16000900661248,
              "rota": true,
              "serial": "X1E0A1XGFVNG",
              "wwn": "0x5000039b38d188bf",
              "vendor": "ATA     ",
              "model": "TOSHIBA_MG08ACA1",
              "rev": "0103",
              "children": [
                {
                  "name": "/dev/sdg1",
                  "type": "part",
                  "size": 16000899595776,
                  "rota": true,
                  "wwn": "0x5000039b38d188bf",
                  "fstype": "linux_raid_member",
                  "partuuid": "0e38f1c6-7f3a-da48-bd7c-2c498b0c0445",
                  "children": [
                    {
                      "name": "/dev/md2",
                      "type": "raid6",
                      "size": 128006110576640,
                      "rota": true,
                      "fstype": "crypto_LUKS",
                      "children": [
                        {
                          "name": "/dev/mapper/securehdd",
                          "type": "crypt",
                          "size": 128006093799424,
                          "rota": true,
                          "mountpoint": "/mnt/secure",
                          "fstype": "ext4",
                          "children": [],
                          "usage": 110085240667504,
                          "smart": {
                            "Temperature": 0,
                            "Read": 0,
                            "Written": 0,
                            "PowerOnHours": 0,
                            "PowerCycles": 0,
                            "AdditionalData": null,
                            "Thresholds": null
                          }
                        }
                      ],
                      "usage": 0,
                      "smart": {
                        "Temperature": 0,
                        "Read": 0,
                        "Written": 0,
                        "PowerOnHours": 0,
                        "PowerCycles": 0,
                        "AdditionalData": null,
                        "Thresholds": null
                      }
                    }
                  ],
                  "usage": 0,
                  "smart": {
                    "Temperature": 0,
                    "Read": 0,
                    "Written": 0,
                    "PowerOnHours": 0,
                    "PowerCycles": 0,
                    "AdditionalData": null,
                    "Thresholds": null
                  }
                }
              ],
              "usage": 0,
              "smart": {
                "Temperature": 32,
                "Read": 0,
                "Written": 0,
                "PowerOnHours": 16206,
                "PowerCycles": 6,
                "AdditionalData": {
                  "Version": 16,
                  "Attrs": {
                    "1": {
                      "Id": 1,
                      "Flags": 11,
                      "Current": 100,
                      "Worst": 100,
                      "VendorBytes": [
                        0,
                        0,
                        0,
                        0,
                        0,
                        0
                      ],
                      "Name": "Raw_Read_Error_Rate",
                      "Type": 3,
                      "ValueRaw": 0
                    },
                    "2": {
                      "Id": 2,
                      "Flags": 5,
                      "Current": 100,
                      "Worst": 100,
                      "VendorBytes": [
                        0,
                        0,
                        0,
                        0,
                        0,
                        0
                      ],
                      "Name": "Throughput_Performance",
                      "Type": 3,
                      "ValueRaw": 0
                    },
                    "3": {
                      "Id": 3,
                      "Flags": 39,
                      "Current": 100,
                      "Worst": 100,
                      "VendorBytes": [
                        141,
                        30,
                        0,
                        0,
                        0,
                        0
                      ],
                      "Name": "Spin_Up_Time",
                      "Type": 10,
                      "ValueRaw": 7821
                    },
                    "4": {
                      "Id": 4,
                      "Flags": 50,
                      "Current": 100,
                      "Worst": 100,
                      "VendorBytes": [
                        7,
                        0,
                        0,
                        0,
                        0,
                        0
                      ],
                      "Name": "Start_Stop_Count",
                      "Type": 3,
                      "ValueRaw": 7
                    },
                    "5": {
                      "Id": 5,
                      "Flags": 51,
                      "Current": 100,
                      "Worst": 100,
                      "VendorBytes": [
                        0,
                        0,
                        0,
                        0,
                        0,
                        0
                      ],
                      "Name": "Reallocated_Sector_Ct",
                      "Type": 9,
                      "ValueRaw": 0
                    },
                    "7": {
                      "Id": 7,
                      "Flags": 11,
                      "Current": 100,
                      "Worst": 100,
                      "VendorBytes": [
                        0,
                        0,
                        0,
                        0,
                        0,
                        0
                      ],
                      "Name": "Seek_Error_Rate",
                      "Type": 3,
                      "ValueRaw": 0
                    },
                    "8": {
                      "Id": 8,
                      "Flags": 5,
                      "Current": 100,
                      "Worst": 100,
                      "VendorBytes": [
                        0,
                        0,
                        0,
                        0,
                        0,
                        0
                      ],
                      "Name": "Seek_Time_Performance",
                      "Type": 3,
                      "ValueRaw": 0
                    },
                    "9": {
                      "Id": 9,
                      "Flags": 50,
                      "Current": 60,
                      "Worst": 60,
                      "VendorBytes": [
                        78,
                        63,
                        0,
                        0,
                        0,
                        0
                      ],
                      "Name": "Power_On_Hours",
                      "Type": 11,
                      "ValueRaw": 16206
                    },
                    "10": {
                      "Id": 10,
                      "Flags": 51,
                      "Current": 100,
                      "Worst": 100,
                      "VendorBytes": [
                        0,
                        0,
                        0,
                        0,
                        0,
                        0
                      ],
                      "Name": "Spin_Retry_Count",
                      "Type": 3,
                      "ValueRaw": 0
                    },
                    "12": {
                      "Id": 12,
                      "Flags": 50,
                      "Current": 100,
                      "Worst": 100,
                      "VendorBytes": [
                        6,
                        0,
                        0,
                        0,
                        0,
                        0
                      ],
                      "Name": "Power_Cycle_Count",
                      "Type": 3,
                      "ValueRaw": 6
                    },
                    "23": {
                      "Id": 23,
                      "Flags": 35,
                      "Current": 100,
                      "Worst": 100,
                      "VendorBytes": [
                        0,
                        0,
                        0,
                        0,
                        0,
                        0
                      ],
                      "Name": "Helium_Condition_Lower",
                      "Type": 3,
                      "ValueRaw": 0
                    },
                    "24": {
                      "Id": 24,
                      "Flags": 35,
                      "Current": 100,
                      "Worst": 100,
                      "VendorBytes": [
                        0,
                        0,
                        0,
                        0,
                        0,
                        0
                      ],
                      "Name": "Helium_Condition_Upper",
                      "Type": 3,
                      "ValueRaw": 0
                    },
                    "191": {
                      "Id": 191,
                      "Flags": 50,
                      "Current": 100,
                      "Worst": 100,
                      "VendorBytes": [
                        0,
                        0,
                        0,
                        0,
                        0,
                        0
                      ],
                      "Name": "G-Sense_Error_Rate",
                      "Type": 3,
                      "ValueRaw": 0
                    },
                    "192": {
                      "Id": 192,
                      "Flags": 50,
                      "Current": 100,
                      "Worst": 100,
                      "VendorBytes": [
                        3,
                        0,
                        0,
                        0,
                        0,
                        0
                      ],
                      "Name": "Power-Off_Retract_Count",
                      "Type": 3,
                      "ValueRaw": 3
                    },
                    "193": {
                      "Id": 193,
                      "Flags": 50,
                      "Current": 100,
                      "Worst": 100,
                      "VendorBytes": [
                        15,
                        0,
                        0,
                        0,
                        0,
                        0
                      ],
                      "Name": "Load_Cycle_Count",
                      "Type": 3,
                      "ValueRaw": 15
                    },
                    "194": {
                      "Id": 194,
                      "Flags": 34,
                      "Current": 100,
                      "Worst": 100,
                      "VendorBytes": [
                        32,
                        0,
                        21,
                        0,
                        46,
                        0
                      ],
                      "Name": "Temperature_Celsius",
                      "Type": 18,
                      "ValueRaw": 197569871904
                    },
                    "196": {
                      "Id": 196,
                      "Flags": 50,
                      "Current": 100,
                      "Worst": 100,
                      "VendorBytes": [
                        0,
                        0,
                        0,
                        0,
                        0,
                        0
                      ],
                      "Name": "Reallocated_Event_Count",
                      "Type": 9,
                      "ValueRaw": 0
                    },
                    "197": {
                      "Id": 197,
                      "Flags": 50,
                      "Current": 100,
                      "Worst": 100,
                      "VendorBytes": [
                        0,
                        0,
                        0,
                        0,
                        0,
                        0
                      ],
                      "Name": "Current_Pending_Sector",
                      "Type": 3,
                      "ValueRaw": 0
                    },
                    "198": {
                      "Id": 198,
                      "Flags": 48,
                      "Current": 100,
                      "Worst": 100,
                      "VendorBytes": [
                        0,
                        0,
                        0,
                        0,
                        0,
                        0
                      ],
                      "Name": "Offline_Uncorrectable",
                      "Type": 3,
                      "ValueRaw": 0
                    },
                    "199": {
                      "Id": 199,
                      "Flags": 50,
                      "Current": 200,
                      "Worst": 200,
                      "VendorBytes": [
                        0,
                        0,
                        0,
                        0,
                        0,
                        0
                      ],
                      "Name": "UDMA_CRC_Error_Count",
                      "Type": 3,
                      "ValueRaw": 0
                    },
                    "220": {
                      "Id": 220,
                      "Flags": 2,
                      "Current": 100,
                      "Worst": 100,
                      "VendorBytes": [
                        1,
                        0,
                        12,
                        5,
                        0,
                        0
                      ],
                      "Name": "Disk_Shift",
                      "Type": 3,
                      "ValueRaw": 84672513
                    },
                    "222": {
                      "Id": 222,
                      "Flags": 50,
                      "Current": 60,
                      "Worst": 60,
                      "VendorBytes": [
                        61,
                        63,
                        0,
                        0,
                        0,
                        0
                      ],
                      "Name": "Loaded_Hours",
                      "Type": 3,
                      "ValueRaw": 16189
                    },
                    "223": {
                      "Id": 223,
                      "Flags": 50,
                      "Current": 100,
                      "Worst": 100,
                      "VendorBytes": [
                        0,
                        0,
                        0,
                        0,
                        0,
                        0
                      ],
                      "Name": "Load_Retry_Count",
                      "Type": 3,
                      "ValueRaw": 0
                    },
                    "224": {
                      "Id": 224,
                      "Flags": 34,
                      "Current": 100,
                      "Worst": 100,
                      "VendorBytes": [
                        0,
                        0,
                        0,
                        0,
                        0,
                        0
                      ],
                      "Name": "Load_Friction",
                      "Type": 3,
                      "ValueRaw": 0
                    },
                    "226": {
                      "Id": 226,
                      "Flags": 38,
                      "Current": 100,
                      "Worst": 100,
                      "VendorBytes": [
                        11,
                        2,
                        0,
                        0,
                        0,
                        0
                      ],
                      "Name": "Load-in_Time",
                      "Type": 3,
                      "ValueRaw": 523
                    },
                    "240": {
                      "Id": 240,
                      "Flags": 1,
                      "Current": 100,
                      "Worst": 100,
                      "VendorBytes": [
                        0,
                        0,
                        0,
                        0,
                        0,
                        0
                      ],
                      "Name": "Head_Flying_Hours",
                      "Type": 11,
                      "ValueRaw": 0
                    }
                  }
                },
                "Thresholds": {
                  "Revnumber": 16,
                  "Thresholds": {
                    "1": 50,
                    "2": 50,
                    "3": 1,
                    "4": 0,
                    "5": 10,
                    "7": 50,
                    "8": 50,
                    "9": 0,
                    "10": 30,
                    "12": 0,
                    "23": 75,
                    "24": 75,
                    "191": 0,
                    "192": 0,
                    "193": 0,
                    "194": 0,
                    "196": 0,
                    "197": 0,
                    "198": 0,
                    "199": 0,
                    "220": 0,
                    "222": 0,
                    "223": 0,
                    "224": 0,
                    "226": 0,
                    "240": 1
                  }
                }
              }
            },
            {
              "name": "/dev/sdh",
              "type": "disk",
              "size": 16000900661248,
              "rota": true,
              "serial": "ZR7024CW",
              "wwn": "0x5000c500dc26f457",
              "vendor": "ATA     ",
              "model": "ST16000NM001J-2T",
              "rev": "SS02",
              "children": [
                {
                  "name": "/dev/sdh1",
                  "type": "part",
                  "size": 16000899595776,
                  "rota": true,
                  "wwn": "0x5000c500dc26f457",
                  "fstype": "linux_raid_member",
                  "partuuid": "290d0373-e4a4-4841-88f5-a20be9139329",
                  "children": [
                    {
                      "name": "/dev/md2",
                      "type": "raid6",
                      "size": 128006110576640,
                      "rota": true,
                      "fstype": "crypto_LUKS",
                      "children": [
                        {
                          "name": "/dev/mapper/securehdd",
                          "type": "crypt",
                          "size": 128006093799424,
                          "rota": true,
                          "mountpoint": "/mnt/secure",
                          "fstype": "ext4",
                          "children": [],
                          "usage": 110085240667504,
                          "smart": {
                            "Temperature": 0,
                            "Read": 0,
                            "Written": 0,
                            "PowerOnHours": 0,
                            "PowerCycles": 0,
                            "AdditionalData": null,
                            "Thresholds": null
                          }
                        }
                      ],
                      "usage": 0,
                      "smart": {
                        "Temperature": 0,
                        "Read": 0,
                        "Written": 0,
                        "PowerOnHours": 0,
                        "PowerCycles": 0,
                        "AdditionalData": null,
                        "Thresholds": null
                      }
                    }
                  ],
                  "usage": 0,
                  "smart": {
                    "Temperature": 0,
                    "Read": 0,
                    "Written": 0,
                    "PowerOnHours": 0,
                    "PowerCycles": 0,
                    "AdditionalData": null,
                    "Thresholds": null
                  }
                }
              ],
              "usage": 0,
              "smart": {
                "Temperature": 33,
                "Read": 14882882757714,
                "Written": 208164125266,
                "PowerOnHours": 19922,
                "PowerCycles": 4,
                "AdditionalData": {
                  "Version": 10,
                  "Attrs": {
                    "1": {
                      "Id": 1,
                      "Flags": 15,
                      "Current": 79,
                      "Worst": 64,
                      "VendorBytes": [
                        244,
                        246,
                        36,
                        5,
                        0,
                        0
                      ],
                      "Name": "Raw_Read_Error_Rate",
                      "Type": 3,
                      "ValueRaw": 86308596
                    },
                    "3": {
                      "Id": 3,
                      "Flags": 3,
                      "Current": 95,
                      "Worst": 95,
                      "VendorBytes": [
                        0,
                        0,
                        0,
                        0,
                        0,
                        0
                      ],
                      "Name": "Spin_Up_Time",
                      "Type": 10,
                      "ValueRaw": 0
                    },
                    "4": {
                      "Id": 4,
                      "Flags": 50,
                      "Current": 100,
                      "Worst": 100,
                      "VendorBytes": [
                        4,
                        0,
                        0,
                        0,
                        0,
                        0
                      ],
                      "Name": "Start_Stop_Count",
                      "Type": 3,
                      "ValueRaw": 4
                    },
                    "5": {
                      "Id": 5,
                      "Flags": 51,
                      "Current": 100,
                      "Worst": 100,
                      "VendorBytes": [
                        0,
                        0,
                        0,
                        0,
                        0,
                        0
                      ],
                      "Name": "Reallocated_Sector_Ct",
                      "Type": 9,
                      "ValueRaw": 0
                    },
                    "7": {
                      "Id": 7,
                      "Flags": 15,
                      "Current": 88,
                      "Worst": 60,
                      "VendorBytes": [
                        211,
                        255,
                        132,
                        33,
                        0,
                        0
                      ],
                      "Name": "Seek_Error_Rate",
                      "Type": 3,
                      "ValueRaw": 562364371
                    },
                    "9": {
                      "Id": 9,
                      "Flags": 50,
                      "Current": 78,
                      "Worst": 78,
                      "VendorBytes": [
                        210,
                        77,
                        0,
                        0,
                        0,
                        0
                      ],
                      "Name": "Power_On_Hours",
                      "Type": 11,
                      "ValueRaw": 19922
                    },
                    "10": {
                      "Id": 10,
                      "Flags": 19,
                      "Current": 100,
                      "Worst": 100,
                      "VendorBytes": [
                        0,
                        0,
                        0,
                        0,
                        0,
                        0
                      ],
                      "Name": "Spin_Retry_Count",
                      "Type": 3,
                      "ValueRaw": 0
                    },
                    "12": {
                      "Id": 12,
                      "Flags": 50,
                      "Current": 100,
                      "Worst": 100,
                      "VendorBytes": [
                        4,
                        0,
                        0,
                        0,
                        0,
                        0
                      ],
                      "Name": "Power_Cycle_Count",
                      "Type": 3,
                      "ValueRaw": 4
                    },
                    "18": {
                      "Id": 18,
                      "Flags": 11,
                      "Current": 100,
                      "Worst": 100,
                      "VendorBytes": [
                        0,
                        0,
                        0,
                        0,
                        0,
                        0
                      ],
                      "Name": "",
                      "Type": 0,
                      "ValueRaw": 0
                    },
                    "187": {
                      "Id": 187,
                      "Flags": 50,
                      "Current": 100,
                      "Worst": 100,
                      "VendorBytes": [
                        0,
                        0,
                        0,
                        0,
                        0,
                        0
                      ],
                      "Name": "Reported_Uncorrect",
                      "Type": 3,
                      "ValueRaw": 0
                    },
                    "188": {
                      "Id": 188,
                      "Flags": 50,
                      "Current": 100,
                      "Worst": 100,
                      "VendorBytes": [
                        0,
                        0,
                        0,
                        0,
                        0,
                        0
                      ],
                      "Name": "Command_Timeout",
                      "Type": 3,
                      "ValueRaw": 0
                    },
                    "190": {
                      "Id": 190,
                      "Flags": 34,
                      "Current": 67,
                      "Worst": 47,
                      "VendorBytes": [
                        33,
                        0,
                        30,
                        40,
                        0,
                        0
                      ],
                      "Name": "Airflow_Temperature_Cel",
                      "Type": 18,
                      "ValueRaw": 673054753
                    },
                    "192": {
                      "Id": 192,
                      "Flags": 50,
                      "Current": 100,
                      "Worst": 100,
                      "VendorBytes": [
                        1,
                        0,
                        0,
                        0,
                        0,
                        0
                      ],
                      "Name": "Power-Off_Retract_Count",
                      "Type": 3,
                      "ValueRaw": 1
                    },
                    "193": {
                      "Id": 193,
                      "Flags": 50,
                      "Current": 100,
                      "Worst": 100,
                      "VendorBytes": [
                        86,
                        3,
                        0,
                        0,
                        0,
                        0
                      ],
                      "Name": "Load_Cycle_Count",
                      "Type": 3,
                      "ValueRaw": 854
                    },
                    "194": {
                      "Id": 194,
                      "Flags": 34,
                      "Current": 33,
                      "Worst": 40,
                      "VendorBytes": [
                        33,
                        0,
                        0,
                        0,
                        23,
                        0
                      ],
                      "Name": "Temperature_Celsius",
                      "Type": 18,
                      "ValueRaw": 98784247841
                    },
                    "197": {
                      "Id": 197,
                      "Flags": 18,
                      "Current": 100,
                      "Worst": 100,
                      "VendorBytes": [
                        0,
                        0,
                        0,
                        0,
                        0,
                        0
                      ],
                      "Name": "Current_Pending_Sector",
                      "Type": 3,
                      "ValueRaw": 0
                    },
                    "198": {
                      "Id": 198,
                      "Flags": 16,
                      "Current": 100,
                      "Worst": 100,
                      "VendorBytes": [
                        0,
                        0,
                        0,
                        0,
                        0,
                        0
                      ],
                      "Name": "Offline_Uncorrectable",
                      "Type": 3,
                      "ValueRaw": 0
                    },
                    "199": {
                      "Id": 199,
                      "Flags": 62,
                      "Current": 200,
                      "Worst": 200,
                      "VendorBytes": [
                        0,
                        0,
                        0,
                        0,
                        0,
                        0
                      ],
                      "Name": "UDMA_CRC_Error_Count",
                      "Type": 3,
                      "ValueRaw": 0
                    },
                    "200": {
                      "Id": 200,
                      "Flags": 35,
                      "Current": 100,
                      "Worst": 100,
                      "VendorBytes": [
                        0,
                        0,
                        0,
                        0,
                        0,
                        0
                      ],
                      "Name": "Multi_Zone_Error_Rate",
                      "Type": 3,
                      "ValueRaw": 0
                    },
                    "240": {
                      "Id": 240,
                      "Flags": 0,
                      "Current": 100,
                      "Worst": 253,
                      "VendorBytes": [
                        123,
                        77,
                        0,
                        0,
                        52,
                        177
                      ],
                      "Name": "Head_Flying_Hours",
                      "Type": 11,
                      "ValueRaw": 194836896435579
                    },
                    "241": {
                      "Id": 241,
                      "Flags": 0,
                      "Current": 100,
                      "Worst": 253,
                      "VendorBytes": [
                        82,
                        122,
                        140,
                        119,
                        48,
                        0
                      ],
                      "Name": "Total_LBAs_Written",
                      "Type": 3,
                      "ValueRaw": 208164125266
                    },
                    "242": {
                      "Id": 242,
                      "Flags": 0,
                      "Current": 100,
                      "Worst": 253,
                      "VendorBytes": [
                        82,
                        164,
                        240,
                        48,
                        137,
                        13
                      ],
                      "Name": "Total_LBAs_Read",
                      "Type": 3,
                      "ValueRaw": 14882882757714
                    }
                  }
                },
                "Thresholds": {
                  "Revnumber": 1,
                  "Thresholds": {
                    "1": 44,
                    "3": 0,
                    "4": 20,
                    "5": 10,
                    "7": 45,
                    "9": 0,
                    "10": 97,
                    "12": 20,
                    "18": 50,
                    "187": 0,
                    "188": 0,
                    "190": 0,
                    "192": 0,
                    "193": 0,
                    "194": 0,
                    "197": 0,
                    "198": 0,
                    "199": 0,
                    "200": 1,
                    "240": 0,
                    "241": 0,
                    "242": 0
                  }
                }
              }
            },
            {
              "name": "/dev/sdi",
              "type": "disk",
              "size": 16000900661248,
              "rota": true,
              "serial": "ZR702532",
              "wwn": "0x5000c500dc2719bd",
              "vendor": "ATA     ",
              "model": "ST16000NM001J-2T",
              "rev": "SS02",
              "children": [
                {
                  "name": "/dev/sdi1",
                  "type": "part",
                  "size": 16000899595776,
                  "rota": true,
                  "wwn": "0x5000c500dc2719bd",
                  "fstype": "linux_raid_member",
                  "partuuid": "c9c933d9-9b04-244a-9a7c-63418a69a082",
                  "children": [
                    {
                      "name": "/dev/md2",
                      "type": "raid6",
                      "size": 128006110576640,
                      "rota": true,
                      "fstype": "crypto_LUKS",
                      "children": [
                        {
                          "name": "/dev/mapper/securehdd",
                          "type": "crypt",
                          "size": 128006093799424,
                          "rota": true,
                          "mountpoint": "/mnt/secure",
                          "fstype": "ext4",
                          "children": [],
                          "usage": 110085240667504,
                          "smart": {
                            "Temperature": 0,
                            "Read": 0,
                            "Written": 0,
                            "PowerOnHours": 0,
                            "PowerCycles": 0,
                            "AdditionalData": null,
                            "Thresholds": null
                          }
                        }
                      ],
                      "usage": 0,
                      "smart": {
                        "Temperature": 0,
                        "Read": 0,
                        "Written": 0,
                        "PowerOnHours": 0,
                        "PowerCycles": 0,
                        "AdditionalData": null,
                        "Thresholds": null
                      }
                    }
                  ],
                  "usage": 0,
                  "smart": {
                    "Temperature": 0,
                    "Read": 0,
                    "Written": 0,
                    "PowerOnHours": 0,
                    "PowerCycles": 0,
                    "AdditionalData": null,
                    "Thresholds": null
                  }
                }
              ],
              "usage": 0,
              "smart": {
                "Temperature": 39,
                "Read": 21318532469494,
                "Written": 208457973128,
                "PowerOnHours": 19922,
                "PowerCycles": 4,
                "AdditionalData": {
                  "Version": 10,
                  "Attrs": {
                    "1": {
                      "Id": 1,
                      "Flags": 15,
                      "Current": 77,
                      "Worst": 64,
                      "VendorBytes": [
                        154,
                        131,
                        184,
                        2,
                        0,
                        0
                      ],
                      "Name": "Raw_Read_Error_Rate",
                      "Type": 3,
                      "ValueRaw": 45646746
                    },
                    "3": {
                      "Id": 3,
                      "Flags": 3,
                      "Current": 96,
                      "Worst": 96,
                      "VendorBytes": [
                        0,
                        0,
                        0,
                        0,
                        0,
                        0
                      ],
                      "Name": "Spin_Up_Time",
                      "Type": 10,
                      "ValueRaw": 0
                    },
                    "4": {
                      "Id": 4,
                      "Flags": 50,
                      "Current": 100,
                      "Worst": 100,
                      "VendorBytes": [
                        4,
                        0,
                        0,
                        0,
                        0,
                        0
                      ],
                      "Name": "Start_Stop_Count",
                      "Type": 3,
                      "ValueRaw": 4
                    },
                    "5": {
                      "Id": 5,
                      "Flags": 51,
                      "Current": 100,
                      "Worst": 100,
                      "VendorBytes": [
                        0,
                        0,
                        0,
                        0,
                        0,
                        0
                      ],
                      "Name": "Reallocated_Sector_Ct",
                      "Type": 9,
                      "ValueRaw": 0
                    },
                    "7": {
                      "Id": 7,
                      "Flags": 15,
                      "Current": 87,
                      "Worst": 60,
                      "VendorBytes": [
                        20,
                        235,
                        53,
                        32,
                        0,
                        0
                      ],
                      "Name": "Seek_Error_Rate",
                      "Type": 3,
                      "ValueRaw": 540404500
                    },
                    "9": {
                      "Id": 9,
                      "Flags": 50,
                      "Current": 78,
                      "Worst": 78,
                      "VendorBytes": [
                        210,
                        77,
                        0,
                        0,
                        0,
                        0
                      ],
                      "Name": "Power_On_Hours",
                      "Type": 11,
                      "ValueRaw": 19922
                    },
                    "10": {
                      "Id": 10,
                      "Flags": 19,
                      "Current": 100,
                      "Worst": 100,
                      "VendorBytes": [
                        0,
                        0,
                        0,
                        0,
                        0,
                        0
                      ],
                      "Name": "Spin_Retry_Count",
                      "Type": 3,
                      "ValueRaw": 0
                    },
                    "12": {
                      "Id": 12,
                      "Flags": 50,
                      "Current": 100,
                      "Worst": 100,
                      "VendorBytes": [
                        4,
                        0,
                        0,
                        0,
                        0,
                        0
                      ],
                      "Name": "Power_Cycle_Count",
                      "Type": 3,
                      "ValueRaw": 4
                    },
                    "18": {
                      "Id": 18,
                      "Flags": 11,
                      "Current": 100,
                      "Worst": 100,
                      "VendorBytes": [
                        0,
                        0,
                        0,
                        0,
                        0,
                        0
                      ],
                      "Name": "",
                      "Type": 0,
                      "ValueRaw": 0
                    },
                    "187": {
                      "Id": 187,
                      "Flags": 50,
                      "Current": 100,
                      "Worst": 100,
                      "VendorBytes": [
                        0,
                        0,
                        0,
                        0,
                        0,
                        0
                      ],
                      "Name": "Reported_Uncorrect",
                      "Type": 3,
                      "ValueRaw": 0
                    },
                    "188": {
                      "Id": 188,
                      "Flags": 50,
                      "Current": 100,
                      "Worst": 100,
                      "VendorBytes": [
                        0,
                        0,
                        0,
                        0,
                        0,
                        0
                      ],
                      "Name": "Command_Timeout",
                      "Type": 3,
                      "ValueRaw": 0
                    },
                    "190": {
                      "Id": 190,
                      "Flags": 34,
                      "Current": 61,
                      "Worst": 48,
                      "VendorBytes": [
                        39,
                        0,
                        32,
                        43,
                        0,
                        0
                      ],
                      "Name": "Airflow_Temperature_Cel",
                      "Type": 18,
                      "ValueRaw": 723517479
                    },
                    "192": {
                      "Id": 192,
                      "Flags": 50,
                      "Current": 100,
                      "Worst": 100,
                      "VendorBytes": [
                        1,
                        0,
                        0,
                        0,
                        0,
                        0
                      ],
                      "Name": "Power-Off_Retract_Count",
                      "Type": 3,
                      "ValueRaw": 1
                    },
                    "193": {
                      "Id": 193,
                      "Flags": 50,
                      "Current": 100,
                      "Worst": 100,
                      "VendorBytes": [
                        85,
                        3,
                        0,
                        0,
                        0,
                        0
                      ],
                      "Name": "Load_Cycle_Count",
                      "Type": 3,
                      "ValueRaw": 853
                    },
                    "194": {
                      "Id": 194,
                      "Flags": 34,
                      "Current": 39,
                      "Worst": 43,
                      "VendorBytes": [
                        39,
                        0,
                        0,
                        0,
                        23,
                        0
                      ],
                      "Name": "Temperature_Celsius",
                      "Type": 18,
                      "ValueRaw": 98784247847
                    },
                    "197": {
                      "Id": 197,
                      "Flags": 18,
                      "Current": 100,
                      "Worst": 100,
                      "VendorBytes": [
                        0,
                        0,
                        0,
                        0,
                        0,
                        0
                      ],
                      "Name": "Current_Pending_Sector",
                      "Type": 3,
                      "ValueRaw": 0
                    },
                    "198": {
                      "Id": 198,
                      "Flags": 16,
                      "Current": 100,
                      "Worst": 100,
                      "VendorBytes": [
                        0,
                        0,
                        0,
                        0,
                        0,
                        0
                      ],
                      "Name": "Offline_Uncorrectable",
                      "Type": 3,
                      "ValueRaw": 0
                    },
                    "199": {
                      "Id": 199,
                      "Flags": 62,
                      "Current": 200,
                      "Worst": 200,
                      "VendorBytes": [
                        0,
                        0,
                        0,
                        0,
                        0,
                        0
                      ],
                      "Name": "UDMA_CRC_Error_Count",
                      "Type": 3,
                      "ValueRaw": 0
                    },
                    "200": {
                      "Id": 200,
                      "Flags": 35,
                      "Current": 100,
                      "Worst": 100,
                      "VendorBytes": [
                        0,
                        0,
                        0,
                        0,
                        0,
                        0
                      ],
                      "Name": "Multi_Zone_Error_Rate",
                      "Type": 3,
                      "ValueRaw": 0
                    },
                    "240": {
                      "Id": 240,
                      "Flags": 0,
                      "Current": 100,
                      "Worst": 253,
                      "VendorBytes": [
                        132,
                        77,
                        0,
                        0,
                        68,
                        83
                      ],
                      "Name": "Head_Flying_Hours",
                      "Type": 11,
                      "ValueRaw": 91551522901380
                    },
                    "241": {
                      "Id": 241,
                      "Flags": 0,
                      "Current": 100,
                      "Worst": 253,
                      "VendorBytes": [
                        136,
                        61,
                        16,
                        137,
                        48,
                        0
                      ],
                      "Name": "Total_LBAs_Written",
                      "Type": 3,
                      "ValueRaw": 208457973128
                    },
                    "242": {
                      "Id": 242,
                      "Flags": 0,
                      "Current": 100,
                      "Worst": 253,
                      "VendorBytes": [
                        246,
                        18,
                        142,
                        155,
                        99,
                        19
                      ],
                      "Name": "Total_LBAs_Read",
                      "Type": 3,
                      "ValueRaw": 21318532469494
                    }
                  }
                },
                "Thresholds": {
                  "Revnumber": 1,
                  "Thresholds": {
                    "1": 44,
                    "3": 0,
                    "4": 20,
                    "5": 10,
                    "7": 45,
                    "9": 0,
                    "10": 97,
                    "12": 20,
                    "18": 50,
                    "187": 0,
                    "188": 0,
                    "190": 0,
                    "192": 0,
                    "193": 0,
                    "194": 0,
                    "197": 0,
                    "198": 0,
                    "199": 0,
                    "200": 1,
                    "240": 0,
                    "241": 0,
                    "242": 0
                  }
                }
              }
            },
            {
              "name": "/dev/sdj",
              "type": "disk",
              "size": 16000900661248,
              "rota": true,
              "serial": "ZR7025EQ",
              "wwn": "0x5000c500dc26f17b",
              "vendor": "ATA     ",
              "model": "ST16000NM001J-2T",
              "rev": "SS02",
              "children": [
                {
                  "name": "/dev/sdj1",
                  "type": "part",
                  "size": 16000899595776,
                  "rota": true,
                  "wwn": "0x5000c500dc26f17b",
                  "fstype": "linux_raid_member",
                  "partuuid": "495c80e3-ae49-4245-86ff-171e0c67af50",
                  "children": [
                    {
                      "name": "/dev/md2",
                      "type": "raid6",
                      "size": 128006110576640,
                      "rota": true,
                      "fstype": "crypto_LUKS",
                      "children": [
                        {
                          "name": "/dev/mapper/securehdd",
                          "type": "crypt",
                          "size": 128006093799424,
                          "rota": true,
                          "mountpoint": "/mnt/secure",
                          "fstype": "ext4",
                          "children": [],
                          "usage": 110085240667504,
                          "smart": {
                            "Temperature": 0,
                            "Read": 0,
                            "Written": 0,
                            "PowerOnHours": 0,
                            "PowerCycles": 0,
                            "AdditionalData": null,
                            "Thresholds": null
                          }
                        }
                      ],
                      "usage": 0,
                      "smart": {
                        "Temperature": 0,
                        "Read": 0,
                        "Written": 0,
                        "PowerOnHours": 0,
                        "PowerCycles": 0,
                        "AdditionalData": null,
                        "Thresholds": null
                      }
                    }
                  ],
                  "usage": 0,
                  "smart": {
                    "Temperature": 0,
                    "Read": 0,
                    "Written": 0,
                    "PowerOnHours": 0,
                    "PowerCycles": 0,
                    "AdditionalData": null,
                    "Thresholds": null
                  }
                }
              ],
              "usage": 0,
              "smart": {
                "Temperature": 39,
                "Read": 13148204322133,
                "Written": 212229522374,
                "PowerOnHours": 19922,
                "PowerCycles": 4,
                "AdditionalData": {
                  "Version": 10,
                  "Attrs": {
                    "1": {
                      "Id": 1,
                      "Flags": 15,
                      "Current": 83,
                      "Worst": 64,
                      "VendorBytes": [
                        140,
                        96,
                        220,
                        12,
                        0,
                        0
                      ],
                      "Name": "Raw_Read_Error_Rate",
                      "Type": 3,
                      "ValueRaw": 215769228
                    },
                    "3": {
                      "Id": 3,
                      "Flags": 3,
                      "Current": 96,
                      "Worst": 96,
                      "VendorBytes": [
                        0,
                        0,
                        0,
                        0,
                        0,
                        0
                      ],
                      "Name": "Spin_Up_Time",
                      "Type": 10,
                      "ValueRaw": 0
                    },
                    "4": {
                      "Id": 4,
                      "Flags": 50,
                      "Current": 100,
                      "Worst": 100,
                      "VendorBytes": [
                        4,
                        0,
                        0,
                        0,
                        0,
                        0
                      ],
                      "Name": "Start_Stop_Count",
                      "Type": 3,
                      "ValueRaw": 4
                    },
                    "5": {
                      "Id": 5,
                      "Flags": 51,
                      "Current": 100,
                      "Worst": 100,
                      "VendorBytes": [
                        0,
                        0,
                        0,
                        0,
                        0,
                        0
                      ],
                      "Name": "Reallocated_Sector_Ct",
                      "Type": 9,
                      "ValueRaw": 0
                    },
                    "7": {
                      "Id": 7,
                      "Flags": 15,
                      "Current": 88,
                      "Worst": 60,
                      "VendorBytes": [
                        88,
                        63,
                        56,
                        36,
                        0,
                        0
                      ],
                      "Name": "Seek_Error_Rate",
                      "Type": 3,
                      "ValueRaw": 607666008
                    },
                    "9": {
                      "Id": 9,
                      "Flags": 50,
                      "Current": 78,
                      "Worst": 78,
                      "VendorBytes": [
                        210,
                        77,
                        0,
                        0,
                        0,
                        0
                      ],
                      "Name": "Power_On_Hours",
                      "Type": 11,
                      "ValueRaw": 19922
                    },
                    "10": {
                      "Id": 10,
                      "Flags": 19,
                      "Current": 100,
                      "Worst": 100,
                      "VendorBytes": [
                        0,
                        0,
                        0,
                        0,
                        0,
                        0
                      ],
                      "Name": "Spin_Retry_Count",
                      "Type": 3,
                      "ValueRaw": 0
                    },
                    "12": {
                      "Id": 12,
                      "Flags": 50,
                      "Current": 100,
                      "Worst": 100,
                      "VendorBytes": [
                        4,
                        0,
                        0,
                        0,
                        0,
                        0
                      ],
                      "Name": "Power_Cycle_Count",
                      "Type": 3,
                      "ValueRaw": 4
                    },
                    "18": {
                      "Id": 18,
                      "Flags": 11,
                      "Current": 100,
                      "Worst": 100,
                      "VendorBytes": [
                        0,
                        0,
                        0,
                        0,
                        0,
                        0
                      ],
                      "Name": "",
                      "Type": 0,
                      "ValueRaw": 0
                    },
                    "187": {
                      "Id": 187,
                      "Flags": 50,
                      "Current": 100,
                      "Worst": 100,
                      "VendorBytes": [
                        0,
                        0,
                        0,
                        0,
                        0,
                        0
                      ],
                      "Name": "Reported_Uncorrect",
                      "Type": 3,
                      "ValueRaw": 0
                    },
                    "188": {
                      "Id": 188,
                      "Flags": 50,
                      "Current": 100,
                      "Worst": 100,
                      "VendorBytes": [
                        0,
                        0,
                        0,
                        0,
                        0,
                        0
                      ],
                      "Name": "Command_Timeout",
                      "Type": 3,
                      "ValueRaw": 0
                    },
                    "190": {
                      "Id": 190,
                      "Flags": 34,
                      "Current": 61,
                      "Worst": 48,
                      "VendorBytes": [
                        39,
                        0,
                        33,
                        45,
                        0,
                        0
                      ],
                      "Name": "Airflow_Temperature_Cel",
                      "Type": 18,
                      "ValueRaw": 757137447
                    },
                    "192": {
                      "Id": 192,
                      "Flags": 50,
                      "Current": 100,
                      "Worst": 100,
                      "VendorBytes": [
                        1,
                        0,
                        0,
                        0,
                        0,
                        0
                      ],
                      "Name": "Power-Off_Retract_Count",
                      "Type": 3,
                      "ValueRaw": 1
                    },
                    "193": {
                      "Id": 193,
                      "Flags": 50,
                      "Current": 100,
                      "Worst": 100,
                      "VendorBytes": [
                        85,
                        3,
                        0,
                        0,
                        0,
                        0
                      ],
                      "Name": "Load_Cycle_Count",
                      "Type": 3,
                      "ValueRaw": 853
                    },
                    "194": {
                      "Id": 194,
                      "Flags": 34,
                      "Current": 39,
                      "Worst": 45,
                      "VendorBytes": [
                        39,
                        0,
                        0,
                        0,
                        23,
                        0
                      ],
                      "Name": "Temperature_Celsius",
                      "Type": 18,
                      "ValueRaw": 98784247847
                    },
                    "197": {
                      "Id": 197,
                      "Flags": 18,
                      "Current": 100,
                      "Worst": 100,
                      "VendorBytes": [
                        0,
                        0,
                        0,
                        0,
                        0,
                        0
                      ],
                      "Name": "Current_Pending_Sector",
                      "Type": 3,
                      "ValueRaw": 0
                    },
                    "198": {
                      "Id": 198,
                      "Flags": 16,
                      "Current": 100,
                      "Worst": 100,
                      "VendorBytes": [
                        0,
                        0,
                        0,
                        0,
                        0,
                        0
                      ],
                      "Name": "Offline_Uncorrectable",
                      "Type": 3,
                      "ValueRaw": 0
                    },
                    "199": {
                      "Id": 199,
                      "Flags": 62,
                      "Current": 200,
                      "Worst": 200,
                      "VendorBytes": [
                        0,
                        0,
                        0,
                        0,
                        0,
                        0
                      ],
                      "Name": "UDMA_CRC_Error_Count",
                      "Type": 3,
                      "ValueRaw": 0
                    },
                    "200": {
                      "Id": 200,
                      "Flags": 35,
                      "Current": 100,
                      "Worst": 100,
                      "VendorBytes": [
                        0,
                        0,
                        0,
                        0,
                        0,
                        0
                      ],
                      "Name": "Multi_Zone_Error_Rate",
                      "Type": 3,
                      "ValueRaw": 0
                    },
                    "240": {
                      "Id": 240,
                      "Flags": 0,
                      "Current": 100,
                      "Worst": 253,
                      "VendorBytes": [
                        131,
                        77,
                        0,
                        0,
                        111,
                        247
                      ],
                      "Name": "Head_Flying_Hours",
                      "Type": 11,
                      "ValueRaw": 272056113450371
                    },
                    "241": {
                      "Id": 241,
                      "Flags": 0,
                      "Current": 100,
                      "Worst": 253,
                      "VendorBytes": [
                        6,
                        132,
                        221,
                        105,
                        49,
                        0
                      ],
                      "Name": "Total_LBAs_Written",
                      "Type": 3,
                      "ValueRaw": 212229522438
                    },
                    "242": {
                      "Id": 242,
                      "Flags": 0,
                      "Current": 100,
                      "Worst": 253,
                      "VendorBytes": [
                        85,
                        77,
                        12,
                        78,
                        245,
                        11
                      ],
                      "Name": "Total_LBAs_Read",
                      "Type": 3,
                      "ValueRaw": 13148204322133
                    }
                  }
                },
                "Thresholds": {
                  "Revnumber": 1,
                  "Thresholds": {
                    "1": 44,
                    "3": 0,
                    "4": 20,
                    "5": 10,
                    "7": 45,
                    "9": 0,
                    "10": 97,
                    "12": 20,
                    "18": 50,
                    "187": 0,
                    "188": 0,
                    "190": 0,
                    "192": 0,
                    "193": 0,
                    "194": 0,
                    "197": 0,
                    "198": 0,
                    "199": 0,
                    "200": 1,
                    "240": 0,
                    "241": 0,
                    "242": 0
                  }
                }
              }
            },
            {
              "name": "/dev/nvme1n1",
              "type": "disk",
              "size": 960197124096,
              "rota": false,
              "serial": "S437NC0R203374",
              "wwn": "eui.34333730522033740025384300000001",
              "model": "SAMSUNG MZQLB960HAJR-00007",
              "children": [
                {
                  "name": "/dev/nvme1n1p1",
                  "type": "part",
                  "size": 1073741824,
                  "rota": false,
                  "wwn": "eui.34333730522033740025384300000001",
                  "fstype": "linux_raid_member",
                  "partuuid": "d8c72f27-01",
                  "children": [
                    {
                      "name": "/dev/md0",
                      "type": "raid1",
                      "size": 1071644672,
                      "rota": false,
                      "mountpoint": "/boot",
                      "fstype": "ext4",
                      "children": [],
                      "usage": 192896040,
                      "smart": {
                        "Temperature": 0,
                        "Read": 0,
                        "Written": 0,
                        "PowerOnHours": 0,
                        "PowerCycles": 0,
                        "AdditionalData": null,
                        "Thresholds": null
                      }
                    }
                  ],
                  "usage": 0,
                  "smart": {
                    "Temperature": 0,
                    "Read": 0,
                    "Written": 0,
                    "PowerOnHours": 0,
                    "PowerCycles": 0,
                    "AdditionalData": null,
                    "Thresholds": null
                  }
                },
                {
                  "name": "/dev/nvme1n1p2",
                  "type": "part",
                  "size": 959121285120,
                  "rota": false,
                  "wwn": "eui.34333730522033740025384300000001",
                  "fstype": "linux_raid_member",
                  "partuuid": "d8c72f27-02",
                  "children": [
                    {
                      "name": "/dev/md1",
                      "type": "raid1",
                      "size": 958985994240,
                      "rota": false,
                      "fstype": "crypto_LUKS",
                      "children": [
                        {
                          "name": "/dev/mapper/luks-7fdaaaa1-98a7-4bc4-9500-0cef98526994",
                          "type": "crypt",
                          "size": 958969217024,
                          "rota": false,
                          "mountpoint": "/",
                          "fstype": "ext4",
                          "children": [],
                          "usage": 498663992852,
                          "smart": {
                            "Temperature": 0,
                            "Read": 0,
                            "Written": 0,
                            "PowerOnHours": 0,
                            "PowerCycles": 0,
                            "AdditionalData": null,
                            "Thresholds": null
                          }
                        }
                      ],
                      "usage": 0,
                      "smart": {
                        "Temperature": 0,
                        "Read": 0,
                        "Written": 0,
                        "PowerOnHours": 0,
                        "PowerCycles": 0,
                        "AdditionalData": null,
                        "Thresholds": null
                      }
                    }
                  ],
                  "usage": 0,
                  "smart": {
                    "Temperature": 0,
                    "Read": 0,
                    "Written": 0,
                    "PowerOnHours": 0,
                    "PowerCycles": 0,
                    "AdditionalData": null,
                    "Thresholds": null
                  }
                }
              ],
              "usage": 0,
              "smart": {
                "Temperature": 39,
                "Read": 56242301,
                "Written": 307977320,
                "PowerOnHours": 19920,
                "PowerCycles": 5,
                "AdditionalData": {
                  "CritWarning": 0,
                  "Temperature": 312,
                  "AvailSpare": 100,
                  "SpareThresh": 10,
                  "PercentUsed": 3,
                  "EnduranceCritWarning": 0,
                  "DataUnitsRead": {
                    "Val": [
                      56242301,
                      0
                    ]
                  },
                  "DataUnitsWritten": {
                    "Val": [
                      307977320,
                      0
                    ]
                  },
                  "HostReads": {
                    "Val": [
                      389059177,
                      0
                    ]
                  },
                  "HostWrites": {
                    "Val": [
                      1473420844,
                      0
                    ]
                  },
                  "CtrlBusyTime": {
                    "Val": [
                      3276,
                      0
                    ]
                  },
                  "PowerCycles": {
                    "Val": [
                      5,
                      0
                    ]
                  },
                  "PowerOnHours": {
                    "Val": [
                      19920,
                      0
                    ]
                  },
                  "UnsafeShutdowns": {
                    "Val": [
                      1,
                      0
                    ]
                  },
                  "MediaErrors": {
                    "Val": [
                      0,
                      0
                    ]
                  },
                  "NumErrLogEntries": {
                    "Val": [
                      13,
                      0
                    ]
                  },
                  "WarningTempTime": 0,
                  "CritCompTime": 0,
                  "TempSensor": [
                    312,
                    317,
                    322,
                    0,
                    0,
                    0,
                    0,
                    0
                  ],
                  "ThermalTransitionCount": [
                    0,
                    0
                  ],
                  "ThermalManagementTime": [
                    0,
                    0
                  ]
                },
                "Thresholds": {
                  "VendorID": 5197,
                  "Ssvid": 5197,
                  "SerialNumberRaw": [
                    83,
                    52,
                    51,
                    55,
                    78,
                    67,
                    48,
                    82,
                    50,
                    48,
                    51,
                    51,
                    55,
                    52,
                    32,
                    32,
                    32,
                    32,
                    32,
                    32
                  ],
                  "ModelNumberRaw": [
                    83,
                    65,
                    77,
                    83,
                    85,
                    78,
                    71,
                    32,
                    77,
                    90,
                    81,
                    76,
                    66,
                    57,
                    54,
                    48,
                    72,
                    65,
                    74,
                    82,
                    45,
                    48,
                    48,
                    48,
                    48,
                    55,
                    32,
                    32,
                    32,
                    32,
                    32,
                    32,
                    32,
                    32,
                    32,
                    32,
                    32,
                    32,
                    32,
                    32
                  ],
                  "FirmwareRevRaw": [
                    69,
                    68,
                    65,
                    53,
                    53,
                    48,
                    50,
                    81
                  ],
                  "Rab": 2,
                  "IEEE": [
                    56,
                    37,
                    0
                  ],
                  "Cmic": 0,
                  "Mdts": 9,
                  "Cntlid": 4,
                  "Ver": 66048,
                  "Rtd3r": 8000000,
                  "Rtd3e": 8000000,
                  "Oaes": 0,
                  "Ctratt": 0,
                  "Rrls": 0,
                  "CntrlType": 0,
                  "Fguid": [
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0
                  ],
                  "Crdt1": 0,
                  "Crdt2": 0,
                  "Crdt3": 0,
                  "Nvmsr": 0,
                  "Vwci": 0,
                  "Mec": 0,
                  "Oacs": 15,
                  "Acl": 7,
                  "Aerl": 3,
                  "Frmw": 23,
                  "Lpa": 3,
                  "Elpe": 63,
                  "Npss": 0,
                  "Avscc": 1,
                  "Apsta": 0,
                  "Wctemp": 360,
                  "Cctemp": 361,
                  "Mtfa": 0,
                  "Hmpre": 0,
                  "Hmmin": 0,
                  "Tnvmcap": {
                    "Val": [
                      960197124096,
                      0
                    ]
                  },
                  "Unvmcap": {
                    "Val": [
                      0,
                      0
                    ]
                  },
                  "Rpmbs": 0,
                  "Edstt": 0,
                  "Dsto": 0,
                  "Fwug": 0,
                  "Kas": 0,
                  "Hctma": 0,
                  "Mntmt": 0,
                  "Mxtmt": 0,
                  "Sanicap": 0,
                  "Sqes": 102,
                  "Cqes": 68,
                  "Nn": 1,
                  "Oncs": 31,
                  "Fuses": 0,
                  "Fna": 4,
                  "Vwc": 0,
                  "Awun": 1023,
                  "Awupf": 7,
                  "Nvscc": 1,
                  "Acwu": 0,
                  "Sgls": 0,
                  "Psd": [
                    {
                      "MaxPower": 1060,
                      "Flags": 0,
                      "EntryLat": 0,
                      "ExitLat": 0,
                      "ReadThroughput": 0,
                      "ReadLatency": 0,
                      "WriteThroughput": 0,
                      "WriteLatency": 0,
                      "IdlePower": 0,
                      "IdleScale": 0,
                      "ActivePower": 0,
                      "ActiveWorkScale": 0
                    },
                    {
                      "MaxPower": 0,
                      "Flags": 0,
                      "EntryLat": 0,
                      "ExitLat": 0,
                      "ReadThroughput": 0,
                      "ReadLatency": 0,
                      "WriteThroughput": 0,
                      "WriteLatency": 0,
                      "IdlePower": 0,
                      "IdleScale": 0,
                      "ActivePower": 0,
                      "ActiveWorkScale": 0
                    },
                    {
                      "MaxPower": 0,
                      "Flags": 0,
                      "EntryLat": 0,
                      "ExitLat": 0,
                      "ReadThroughput": 0,
                      "ReadLatency": 0,
                      "WriteThroughput": 0,
                      "WriteLatency": 0,
                      "IdlePower": 0,
                      "IdleScale": 0,
                      "ActivePower": 0,
                      "ActiveWorkScale": 0
                    },
                    {
                      "MaxPower": 0,
                      "Flags": 0,
                      "EntryLat": 0,
                      "ExitLat": 0,
                      "ReadThroughput": 0,
                      "ReadLatency": 0,
                      "WriteThroughput": 0,
                      "WriteLatency": 0,
                      "IdlePower": 0,
                      "IdleScale": 0,
                      "ActivePower": 0,
                      "ActiveWorkScale": 0
                    },
                    {
                      "MaxPower": 0,
                      "Flags": 0,
                      "EntryLat": 0,
                      "ExitLat": 0,
                      "ReadThroughput": 0,
                      "ReadLatency": 0,
                      "WriteThroughput": 0,
                      "WriteLatency": 0,
                      "IdlePower": 0,
                      "IdleScale": 0,
                      "ActivePower": 0,
                      "ActiveWorkScale": 0
                    },
                    {
                      "MaxPower": 0,
                      "Flags": 0,
                      "EntryLat": 0,
                      "ExitLat": 0,
                      "ReadThroughput": 0,
                      "ReadLatency": 0,
                      "WriteThroughput": 0,
                      "WriteLatency": 0,
                      "IdlePower": 0,
                      "IdleScale": 0,
                      "ActivePower": 0,
                      "ActiveWorkScale": 0
                    },
                    {
                      "MaxPower": 0,
                      "Flags": 0,
                      "EntryLat": 0,
                      "ExitLat": 0,
                      "ReadThroughput": 0,
                      "ReadLatency": 0,
                      "WriteThroughput": 0,
                      "WriteLatency": 0,
                      "IdlePower": 0,
                      "IdleScale": 0,
                      "ActivePower": 0,
                      "ActiveWorkScale": 0
                    },
                    {
                      "MaxPower": 0,
                      "Flags": 0,
                      "EntryLat": 0,
                      "ExitLat": 0,
                      "ReadThroughput": 0,
                      "ReadLatency": 0,
                      "WriteThroughput": 0,
                      "WriteLatency": 0,
                      "IdlePower": 0,
                      "IdleScale": 0,
                      "ActivePower": 0,
                      "ActiveWorkScale": 0
                    },
                    {
                      "MaxPower": 0,
                      "Flags": 0,
                      "EntryLat": 0,
                      "ExitLat": 0,
                      "ReadThroughput": 0,
                      "ReadLatency": 0,
                      "WriteThroughput": 0,
                      "WriteLatency": 0,
                      "IdlePower": 0,
                      "IdleScale": 0,
                      "ActivePower": 0,
                      "ActiveWorkScale": 0
                    },
                    {
                      "MaxPower": 0,
                      "Flags": 0,
                      "EntryLat": 0,
                      "ExitLat": 0,
                      "ReadThroughput": 0,
                      "ReadLatency": 0,
                      "WriteThroughput": 0,
                      "WriteLatency": 0,
                      "IdlePower": 0,
                      "IdleScale": 0,
                      "ActivePower": 0,
                      "ActiveWorkScale": 0
                    },
                    {
                      "MaxPower": 0,
                      "Flags": 0,
                      "EntryLat": 0,
                      "ExitLat": 0,
                      "ReadThroughput": 0,
                      "ReadLatency": 0,
                      "WriteThroughput": 0,
                      "WriteLatency": 0,
                      "IdlePower": 0,
                      "IdleScale": 0,
                      "ActivePower": 0,
                      "ActiveWorkScale": 0
                    },
                    {
                      "MaxPower": 0,
                      "Flags": 0,
                      "EntryLat": 0,
                      "ExitLat": 0,
                      "ReadThroughput": 0,
                      "ReadLatency": 0,
                      "WriteThroughput": 0,
                      "WriteLatency": 0,
                      "IdlePower": 0,
                      "IdleScale": 0,
                      "ActivePower": 0,
                      "ActiveWorkScale": 0
                    },
                    {
                      "MaxPower": 0,
                      "Flags": 0,
                      "EntryLat": 0,
                      "ExitLat": 0,
                      "ReadThroughput": 0,
                      "ReadLatency": 0,
                      "WriteThroughput": 0,
                      "WriteLatency": 0,
                      "IdlePower": 0,
                      "IdleScale": 0,
                      "ActivePower": 0,
                      "ActiveWorkScale": 0
                    },
                    {
                      "MaxPower": 0,
                      "Flags": 0,
                      "EntryLat": 0,
                      "ExitLat": 0,
                      "ReadThroughput": 0,
                      "ReadLatency": 0,
                      "WriteThroughput": 0,
                      "WriteLatency": 0,
                      "IdlePower": 0,
                      "IdleScale": 0,
                      "ActivePower": 0,
                      "ActiveWorkScale": 0
                    },
                    {
                      "MaxPower": 0,
                      "Flags": 0,
                      "EntryLat": 0,
                      "ExitLat": 0,
                      "ReadThroughput": 0,
                      "ReadLatency": 0,
                      "WriteThroughput": 0,
                      "WriteLatency": 0,
                      "IdlePower": 0,
                      "IdleScale": 0,
                      "ActivePower": 0,
                      "ActiveWorkScale": 0
                    },
                    {
                      "MaxPower": 0,
                      "Flags": 0,
                      "EntryLat": 0,
                      "ExitLat": 0,
                      "ReadThroughput": 0,
                      "ReadLatency": 0,
                      "WriteThroughput": 0,
                      "WriteLatency": 0,
                      "IdlePower": 0,
                      "IdleScale": 0,
                      "ActivePower": 0,
                      "ActiveWorkScale": 0
                    },
                    {
                      "MaxPower": 0,
                      "Flags": 0,
                      "EntryLat": 0,
                      "ExitLat": 0,
                      "ReadThroughput": 0,
                      "ReadLatency": 0,
                      "WriteThroughput": 0,
                      "WriteLatency": 0,
                      "IdlePower": 0,
                      "IdleScale": 0,
                      "ActivePower": 0,
                      "ActiveWorkScale": 0
                    },
                    {
                      "MaxPower": 0,
                      "Flags": 0,
                      "EntryLat": 0,
                      "ExitLat": 0,
                      "ReadThroughput": 0,
                      "ReadLatency": 0,
                      "WriteThroughput": 0,
                      "WriteLatency": 0,
                      "IdlePower": 0,
                      "IdleScale": 0,
                      "ActivePower": 0,
                      "ActiveWorkScale": 0
                    },
                    {
                      "MaxPower": 0,
                      "Flags": 0,
                      "EntryLat": 0,
                      "ExitLat": 0,
                      "ReadThroughput": 0,
                      "ReadLatency": 0,
                      "WriteThroughput": 0,
                      "WriteLatency": 0,
                      "IdlePower": 0,
                      "IdleScale": 0,
                      "ActivePower": 0,
                      "ActiveWorkScale": 0
                    },
                    {
                      "MaxPower": 0,
                      "Flags": 0,
                      "EntryLat": 0,
                      "ExitLat": 0,
                      "ReadThroughput": 0,
                      "ReadLatency": 0,
                      "WriteThroughput": 0,
                      "WriteLatency": 0,
                      "IdlePower": 0,
                      "IdleScale": 0,
                      "ActivePower": 0,
                      "ActiveWorkScale": 0
                    },
                    {
                      "MaxPower": 0,
                      "Flags": 0,
                      "EntryLat": 0,
                      "ExitLat": 0,
                      "ReadThroughput": 0,
                      "ReadLatency": 0,
                      "WriteThroughput": 0,
                      "WriteLatency": 0,
                      "IdlePower": 0,
                      "IdleScale": 0,
                      "ActivePower": 0,
                      "ActiveWorkScale": 0
                    },
                    {
                      "MaxPower": 0,
                      "Flags": 0,
                      "EntryLat": 0,
                      "ExitLat": 0,
                      "ReadThroughput": 0,
                      "ReadLatency": 0,
                      "WriteThroughput": 0,
                      "WriteLatency": 0,
                      "IdlePower": 0,
                      "IdleScale": 0,
                      "ActivePower": 0,
                      "ActiveWorkScale": 0
                    },
                    {
                      "MaxPower": 0,
                      "Flags": 0,
                      "EntryLat": 0,
                      "ExitLat": 0,
                      "ReadThroughput": 0,
                      "ReadLatency": 0,
                      "WriteThroughput": 0,
                      "WriteLatency": 0,
                      "IdlePower": 0,
                      "IdleScale": 0,
                      "ActivePower": 0,
                      "ActiveWorkScale": 0
                    },
                    {
                      "MaxPower": 0,
                      "Flags": 0,
                      "EntryLat": 0,
                      "ExitLat": 0,
                      "ReadThroughput": 0,
                      "ReadLatency": 0,
                      "WriteThroughput": 0,
                      "WriteLatency": 0,
                      "IdlePower": 0,
                      "IdleScale": 0,
                      "ActivePower": 0,
                      "ActiveWorkScale": 0
                    },
                    {
                      "MaxPower": 0,
                      "Flags": 0,
                      "EntryLat": 0,
                      "ExitLat": 0,
                      "ReadThroughput": 0,
                      "ReadLatency": 0,
                      "WriteThroughput": 0,
                      "WriteLatency": 0,
                      "IdlePower": 0,
                      "IdleScale": 0,
                      "ActivePower": 0,
                      "ActiveWorkScale": 0
                    },
                    {
                      "MaxPower": 0,
                      "Flags": 0,
                      "EntryLat": 0,
                      "ExitLat": 0,
                      "ReadThroughput": 0,
                      "ReadLatency": 0,
                      "WriteThroughput": 0,
                      "WriteLatency": 0,
                      "IdlePower": 0,
                      "IdleScale": 0,
                      "ActivePower": 0,
                      "ActiveWorkScale": 0
                    },
                    {
                      "MaxPower": 0,
                      "Flags": 0,
                      "EntryLat": 0,
                      "ExitLat": 0,
                      "ReadThroughput": 0,
                      "ReadLatency": 0,
                      "WriteThroughput": 0,
                      "WriteLatency": 0,
                      "IdlePower": 0,
                      "IdleScale": 0,
                      "ActivePower": 0,
                      "ActiveWorkScale": 0
                    },
                    {
                      "MaxPower": 0,
                      "Flags": 0,
                      "EntryLat": 0,
                      "ExitLat": 0,
                      "ReadThroughput": 0,
                      "ReadLatency": 0,
                      "WriteThroughput": 0,
                      "WriteLatency": 0,
                      "IdlePower": 0,
                      "IdleScale": 0,
                      "ActivePower": 0,
                      "ActiveWorkScale": 0
                    },
                    {
                      "MaxPower": 0,
                      "Flags": 0,
                      "EntryLat": 0,
                      "ExitLat": 0,
                      "ReadThroughput": 0,
                      "ReadLatency": 0,
                      "WriteThroughput": 0,
                      "WriteLatency": 0,
                      "IdlePower": 0,
                      "IdleScale": 0,
                      "ActivePower": 0,
                      "ActiveWorkScale": 0
                    },
                    {
                      "MaxPower": 0,
                      "Flags": 0,
                      "EntryLat": 0,
                      "ExitLat": 0,
                      "ReadThroughput": 0,
                      "ReadLatency": 0,
                      "WriteThroughput": 0,
                      "WriteLatency": 0,
                      "IdlePower": 0,
                      "IdleScale": 0,
                      "ActivePower": 0,
                      "ActiveWorkScale": 0
                    },
                    {
                      "MaxPower": 0,
                      "Flags": 0,
                      "EntryLat": 0,
                      "ExitLat": 0,
                      "ReadThroughput": 0,
                      "ReadLatency": 0,
                      "WriteThroughput": 0,
                      "WriteLatency": 0,
                      "IdlePower": 0,
                      "IdleScale": 0,
                      "ActivePower": 0,
                      "ActiveWorkScale": 0
                    },
                    {
                      "MaxPower": 0,
                      "Flags": 0,
                      "EntryLat": 0,
                      "ExitLat": 0,
                      "ReadThroughput": 0,
                      "ReadLatency": 0,
                      "WriteThroughput": 0,
                      "WriteLatency": 0,
                      "IdlePower": 0,
                      "IdleScale": 0,
                      "ActivePower": 0,
                      "ActiveWorkScale": 0
                    }
                  ],
                  "Vs": [
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    88,
                    78,
                    85,
                    83,
                    82,
                    71,
                    50,
                    57,
                    18,
                    4,
                    32,
                    23,
                    1,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    7,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0
                  ]
                }
              }
            },
            {
              "name": "/dev/nvme0n1",
              "type": "disk",
              "size": 960197124096,
              "rota": false,
              "serial": "S437NC0R203369",
              "wwn": "eui.34333730522033690025384300000001",
              "model": "SAMSUNG MZQLB960HAJR-00007",
              "children": [
                {
                  "name": "/dev/nvme0n1p1",
                  "type": "part",
                  "size": 1073741824,
                  "rota": false,
                  "wwn": "eui.34333730522033690025384300000001",
                  "fstype": "linux_raid_member",
                  "partuuid": "4d0ca7bb-01",
                  "children": [
                    {
                      "name": "/dev/md0",
                      "type": "raid1",
                      "size": 1071644672,
                      "rota": false,
                      "mountpoint": "/boot",
                      "fstype": "ext4",
                      "children": [],
                      "usage": 192896040,
                      "smart": {
                        "Temperature": 0,
                        "Read": 0,
                        "Written": 0,
                        "PowerOnHours": 0,
                        "PowerCycles": 0,
                        "AdditionalData": null,
                        "Thresholds": null
                      }
                    }
                  ],
                  "usage": 0,
                  "smart": {
                    "Temperature": 0,
                    "Read": 0,
                    "Written": 0,
                    "PowerOnHours": 0,
                    "PowerCycles": 0,
                    "AdditionalData": null,
                    "Thresholds": null
                  }
                },
                {
                  "name": "/dev/nvme0n1p2",
                  "type": "part",
                  "size": 959121285120,
                  "rota": false,
                  "wwn": "eui.34333730522033690025384300000001",
                  "fstype": "linux_raid_member",
                  "partuuid": "4d0ca7bb-02",
                  "children": [
                    {
                      "name": "/dev/md1",
                      "type": "raid1",
                      "size": 958985994240,
                      "rota": false,
                      "fstype": "crypto_LUKS",
                      "children": [
                        {
                          "name": "/dev/mapper/luks-7fdaaaa1-98a7-4bc4-9500-0cef98526994",
                          "type": "crypt",
                          "size": 958969217024,
                          "rota": false,
                          "mountpoint": "/",
                          "fstype": "ext4",
                          "children": [],
                          "usage": 498663992852,
                          "smart": {
                            "Temperature": 0,
                            "Read": 0,
                            "Written": 0,
                            "PowerOnHours": 0,
                            "PowerCycles": 0,
                            "AdditionalData": null,
                            "Thresholds": null
                          }
                        }
                      ],
                      "usage": 0,
                      "smart": {
                        "Temperature": 0,
                        "Read": 0,
                        "Written": 0,
                        "PowerOnHours": 0,
                        "PowerCycles": 0,
                        "AdditionalData": null,
                        "Thresholds": null
                      }
                    }
                  ],
                  "usage": 0,
                  "smart": {
                    "Temperature": 0,
                    "Read": 0,
                    "Written": 0,
                    "PowerOnHours": 0,
                    "PowerCycles": 0,
                    "AdditionalData": null,
                    "Thresholds": null
                  }
                }
              ],
              "usage": 0,
              "smart": {
                "Temperature": 39,
                "Read": 62900445,
                "Written": 305206161,
                "PowerOnHours": 19920,
                "PowerCycles": 5,
                "AdditionalData": {
                  "CritWarning": 0,
                  "Temperature": 312,
                  "AvailSpare": 100,
                  "SpareThresh": 10,
                  "PercentUsed": 3,
                  "EnduranceCritWarning": 0,
                  "DataUnitsRead": {
                    "Val": [
                      62900445,
                      0
                    ]
                  },
                  "DataUnitsWritten": {
                    "Val": [
                      305206161,
                      0
                    ]
                  },
                  "HostReads": {
                    "Val": [
                      461630596,
                      0
                    ]
                  },
                  "HostWrites": {
                    "Val": [
                      1470621522,
                      0
                    ]
                  },
                  "CtrlBusyTime": {
                    "Val": [
                      3226,
                      0
                    ]
                  },
                  "PowerCycles": {
                    "Val": [
                      5,
                      0
                    ]
                  },
                  "PowerOnHours": {
                    "Val": [
                      19920,
                      0
                    ]
                  },
                  "UnsafeShutdowns": {
                    "Val": [
                      1,
                      0
                    ]
                  },
                  "MediaErrors": {
                    "Val": [
                      0,
                      0
                    ]
                  },
                  "NumErrLogEntries": {
                    "Val": [
                      13,
                      0
                    ]
                  },
                  "WarningTempTime": 0,
                  "CritCompTime": 0,
                  "TempSensor": [
                    312,
                    315,
                    321,
                    0,
                    0,
                    0,
                    0,
                    0
                  ],
                  "ThermalTransitionCount": [
                    0,
                    0
                  ],
                  "ThermalManagementTime": [
                    0,
                    0
                  ]
                },
                "Thresholds": {
                  "VendorID": 5197,
                  "Ssvid": 5197,
                  "SerialNumberRaw": [
                    83,
                    52,
                    51,
                    55,
                    78,
                    67,
                    48,
                    82,
                    50,
                    48,
                    51,
                    51,
                    54,
                    57,
                    32,
                    32,
                    32,
                    32,
                    32,
                    32
                  ],
                  "ModelNumberRaw": [
                    83,
                    65,
                    77,
                    83,
                    85,
                    78,
                    71,
                    32,
                    77,
                    90,
                    81,
                    76,
                    66,
                    57,
                    54,
                    48,
                    72,
                    65,
                    74,
                    82,
                    45,
                    48,
                    48,
                    48,
                    48,
                    55,
                    32,
                    32,
                    32,
                    32,
                    32,
                    32,
                    32,
                    32,
                    32,
                    32,
                    32,
                    32,
                    32,
                    32
                  ],
                  "FirmwareRevRaw": [
                    69,
                    68,
                    65,
                    53,
                    53,
                    48,
                    50,
                    81
                  ],
                  "Rab": 2,
                  "IEEE": [
                    56,
                    37,
                    0
                  ],
                  "Cmic": 0,
                  "Mdts": 9,
                  "Cntlid": 4,
                  "Ver": 66048,
                  "Rtd3r": 8000000,
                  "Rtd3e": 8000000,
                  "Oaes": 0,
                  "Ctratt": 0,
                  "Rrls": 0,
                  "CntrlType": 0,
                  "Fguid": [
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0
                  ],
                  "Crdt1": 0,
                  "Crdt2": 0,
                  "Crdt3": 0,
                  "Nvmsr": 0,
                  "Vwci": 0,
                  "Mec": 0,
                  "Oacs": 15,
                  "Acl": 7,
                  "Aerl": 3,
                  "Frmw": 23,
                  "Lpa": 3,
                  "Elpe": 63,
                  "Npss": 0,
                  "Avscc": 1,
                  "Apsta": 0,
                  "Wctemp": 360,
                  "Cctemp": 361,
                  "Mtfa": 0,
                  "Hmpre": 0,
                  "Hmmin": 0,
                  "Tnvmcap": {
                    "Val": [
                      960197124096,
                      0
                    ]
                  },
                  "Unvmcap": {
                    "Val": [
                      0,
                      0
                    ]
                  },
                  "Rpmbs": 0,
                  "Edstt": 0,
                  "Dsto": 0,
                  "Fwug": 0,
                  "Kas": 0,
                  "Hctma": 0,
                  "Mntmt": 0,
                  "Mxtmt": 0,
                  "Sanicap": 0,
                  "Sqes": 102,
                  "Cqes": 68,
                  "Nn": 1,
                  "Oncs": 31,
                  "Fuses": 0,
                  "Fna": 4,
                  "Vwc": 0,
                  "Awun": 1023,
                  "Awupf": 7,
                  "Nvscc": 1,
                  "Acwu": 0,
                  "Sgls": 0,
                  "Psd": [
                    {
                      "MaxPower": 1060,
                      "Flags": 0,
                      "EntryLat": 0,
                      "ExitLat": 0,
                      "ReadThroughput": 0,
                      "ReadLatency": 0,
                      "WriteThroughput": 0,
                      "WriteLatency": 0,
                      "IdlePower": 0,
                      "IdleScale": 0,
                      "ActivePower": 0,
                      "ActiveWorkScale": 0
                    },
                    {
                      "MaxPower": 0,
                      "Flags": 0,
                      "EntryLat": 0,
                      "ExitLat": 0,
                      "ReadThroughput": 0,
                      "ReadLatency": 0,
                      "WriteThroughput": 0,
                      "WriteLatency": 0,
                      "IdlePower": 0,
                      "IdleScale": 0,
                      "ActivePower": 0,
                      "ActiveWorkScale": 0
                    },
                    {
                      "MaxPower": 0,
                      "Flags": 0,
                      "EntryLat": 0,
                      "ExitLat": 0,
                      "ReadThroughput": 0,
                      "ReadLatency": 0,
                      "WriteThroughput": 0,
                      "WriteLatency": 0,
                      "IdlePower": 0,
                      "IdleScale": 0,
                      "ActivePower": 0,
                      "ActiveWorkScale": 0
                    },
                    {
                      "MaxPower": 0,
                      "Flags": 0,
                      "EntryLat": 0,
                      "ExitLat": 0,
                      "ReadThroughput": 0,
                      "ReadLatency": 0,
                      "WriteThroughput": 0,
                      "WriteLatency": 0,
                      "IdlePower": 0,
                      "IdleScale": 0,
                      "ActivePower": 0,
                      "ActiveWorkScale": 0
                    },
                    {
                      "MaxPower": 0,
                      "Flags": 0,
                      "EntryLat": 0,
                      "ExitLat": 0,
                      "ReadThroughput": 0,
                      "ReadLatency": 0,
                      "WriteThroughput": 0,
                      "WriteLatency": 0,
                      "IdlePower": 0,
                      "IdleScale": 0,
                      "ActivePower": 0,
                      "ActiveWorkScale": 0
                    },
                    {
                      "MaxPower": 0,
                      "Flags": 0,
                      "EntryLat": 0,
                      "ExitLat": 0,
                      "ReadThroughput": 0,
                      "ReadLatency": 0,
                      "WriteThroughput": 0,
                      "WriteLatency": 0,
                      "IdlePower": 0,
                      "IdleScale": 0,
                      "ActivePower": 0,
                      "ActiveWorkScale": 0
                    },
                    {
                      "MaxPower": 0,
                      "Flags": 0,
                      "EntryLat": 0,
                      "ExitLat": 0,
                      "ReadThroughput": 0,
                      "ReadLatency": 0,
                      "WriteThroughput": 0,
                      "WriteLatency": 0,
                      "IdlePower": 0,
                      "IdleScale": 0,
                      "ActivePower": 0,
                      "ActiveWorkScale": 0
                    },
                    {
                      "MaxPower": 0,
                      "Flags": 0,
                      "EntryLat": 0,
                      "ExitLat": 0,
                      "ReadThroughput": 0,
                      "ReadLatency": 0,
                      "WriteThroughput": 0,
                      "WriteLatency": 0,
                      "IdlePower": 0,
                      "IdleScale": 0,
                      "ActivePower": 0,
                      "ActiveWorkScale": 0
                    },
                    {
                      "MaxPower": 0,
                      "Flags": 0,
                      "EntryLat": 0,
                      "ExitLat": 0,
                      "ReadThroughput": 0,
                      "ReadLatency": 0,
                      "WriteThroughput": 0,
                      "WriteLatency": 0,
                      "IdlePower": 0,
                      "IdleScale": 0,
                      "ActivePower": 0,
                      "ActiveWorkScale": 0
                    },
                    {
                      "MaxPower": 0,
                      "Flags": 0,
                      "EntryLat": 0,
                      "ExitLat": 0,
                      "ReadThroughput": 0,
                      "ReadLatency": 0,
                      "WriteThroughput": 0,
                      "WriteLatency": 0,
                      "IdlePower": 0,
                      "IdleScale": 0,
                      "ActivePower": 0,
                      "ActiveWorkScale": 0
                    },
                    {
                      "MaxPower": 0,
                      "Flags": 0,
                      "EntryLat": 0,
                      "ExitLat": 0,
                      "ReadThroughput": 0,
                      "ReadLatency": 0,
                      "WriteThroughput": 0,
                      "WriteLatency": 0,
                      "IdlePower": 0,
                      "IdleScale": 0,
                      "ActivePower": 0,
                      "ActiveWorkScale": 0
                    },
                    {
                      "MaxPower": 0,
                      "Flags": 0,
                      "EntryLat": 0,
                      "ExitLat": 0,
                      "ReadThroughput": 0,
                      "ReadLatency": 0,
                      "WriteThroughput": 0,
                      "WriteLatency": 0,
                      "IdlePower": 0,
                      "IdleScale": 0,
                      "ActivePower": 0,
                      "ActiveWorkScale": 0
                    },
                    {
                      "MaxPower": 0,
                      "Flags": 0,
                      "EntryLat": 0,
                      "ExitLat": 0,
                      "ReadThroughput": 0,
                      "ReadLatency": 0,
                      "WriteThroughput": 0,
                      "WriteLatency": 0,
                      "IdlePower": 0,
                      "IdleScale": 0,
                      "ActivePower": 0,
                      "ActiveWorkScale": 0
                    },
                    {
                      "MaxPower": 0,
                      "Flags": 0,
                      "EntryLat": 0,
                      "ExitLat": 0,
                      "ReadThroughput": 0,
                      "ReadLatency": 0,
                      "WriteThroughput": 0,
                      "WriteLatency": 0,
                      "IdlePower": 0,
                      "IdleScale": 0,
                      "ActivePower": 0,
                      "ActiveWorkScale": 0
                    },
                    {
                      "MaxPower": 0,
                      "Flags": 0,
                      "EntryLat": 0,
                      "ExitLat": 0,
                      "ReadThroughput": 0,
                      "ReadLatency": 0,
                      "WriteThroughput": 0,
                      "WriteLatency": 0,
                      "IdlePower": 0,
                      "IdleScale": 0,
                      "ActivePower": 0,
                      "ActiveWorkScale": 0
                    },
                    {
                      "MaxPower": 0,
                      "Flags": 0,
                      "EntryLat": 0,
                      "ExitLat": 0,
                      "ReadThroughput": 0,
                      "ReadLatency": 0,
                      "WriteThroughput": 0,
                      "WriteLatency": 0,
                      "IdlePower": 0,
                      "IdleScale": 0,
                      "ActivePower": 0,
                      "ActiveWorkScale": 0
                    },
                    {
                      "MaxPower": 0,
                      "Flags": 0,
                      "EntryLat": 0,
                      "ExitLat": 0,
                      "ReadThroughput": 0,
                      "ReadLatency": 0,
                      "WriteThroughput": 0,
                      "WriteLatency": 0,
                      "IdlePower": 0,
                      "IdleScale": 0,
                      "ActivePower": 0,
                      "ActiveWorkScale": 0
                    },
                    {
                      "MaxPower": 0,
                      "Flags": 0,
                      "EntryLat": 0,
                      "ExitLat": 0,
                      "ReadThroughput": 0,
                      "ReadLatency": 0,
                      "WriteThroughput": 0,
                      "WriteLatency": 0,
                      "IdlePower": 0,
                      "IdleScale": 0,
                      "ActivePower": 0,
                      "ActiveWorkScale": 0
                    },
                    {
                      "MaxPower": 0,
                      "Flags": 0,
                      "EntryLat": 0,
                      "ExitLat": 0,
                      "ReadThroughput": 0,
                      "ReadLatency": 0,
                      "WriteThroughput": 0,
                      "WriteLatency": 0,
                      "IdlePower": 0,
                      "IdleScale": 0,
                      "ActivePower": 0,
                      "ActiveWorkScale": 0
                    },
                    {
                      "MaxPower": 0,
                      "Flags": 0,
                      "EntryLat": 0,
                      "ExitLat": 0,
                      "ReadThroughput": 0,
                      "ReadLatency": 0,
                      "WriteThroughput": 0,
                      "WriteLatency": 0,
                      "IdlePower": 0,
                      "IdleScale": 0,
                      "ActivePower": 0,
                      "ActiveWorkScale": 0
                    },
                    {
                      "MaxPower": 0,
                      "Flags": 0,
                      "EntryLat": 0,
                      "ExitLat": 0,
                      "ReadThroughput": 0,
                      "ReadLatency": 0,
                      "WriteThroughput": 0,
                      "WriteLatency": 0,
                      "IdlePower": 0,
                      "IdleScale": 0,
                      "ActivePower": 0,
                      "ActiveWorkScale": 0
                    },
                    {
                      "MaxPower": 0,
                      "Flags": 0,
                      "EntryLat": 0,
                      "ExitLat": 0,
                      "ReadThroughput": 0,
                      "ReadLatency": 0,
                      "WriteThroughput": 0,
                      "WriteLatency": 0,
                      "IdlePower": 0,
                      "IdleScale": 0,
                      "ActivePower": 0,
                      "ActiveWorkScale": 0
                    },
                    {
                      "MaxPower": 0,
                      "Flags": 0,
                      "EntryLat": 0,
                      "ExitLat": 0,
                      "ReadThroughput": 0,
                      "ReadLatency": 0,
                      "WriteThroughput": 0,
                      "WriteLatency": 0,
                      "IdlePower": 0,
                      "IdleScale": 0,
                      "ActivePower": 0,
                      "ActiveWorkScale": 0
                    },
                    {
                      "MaxPower": 0,
                      "Flags": 0,
                      "EntryLat": 0,
                      "ExitLat": 0,
                      "ReadThroughput": 0,
                      "ReadLatency": 0,
                      "WriteThroughput": 0,
                      "WriteLatency": 0,
                      "IdlePower": 0,
                      "IdleScale": 0,
                      "ActivePower": 0,
                      "ActiveWorkScale": 0
                    },
                    {
                      "MaxPower": 0,
                      "Flags": 0,
                      "EntryLat": 0,
                      "ExitLat": 0,
                      "ReadThroughput": 0,
                      "ReadLatency": 0,
                      "WriteThroughput": 0,
                      "WriteLatency": 0,
                      "IdlePower": 0,
                      "IdleScale": 0,
                      "ActivePower": 0,
                      "ActiveWorkScale": 0
                    },
                    {
                      "MaxPower": 0,
                      "Flags": 0,
                      "EntryLat": 0,
                      "ExitLat": 0,
                      "ReadThroughput": 0,
                      "ReadLatency": 0,
                      "WriteThroughput": 0,
                      "WriteLatency": 0,
                      "IdlePower": 0,
                      "IdleScale": 0,
                      "ActivePower": 0,
                      "ActiveWorkScale": 0
                    },
                    {
                      "MaxPower": 0,
                      "Flags": 0,
                      "EntryLat": 0,
                      "ExitLat": 0,
                      "ReadThroughput": 0,
                      "ReadLatency": 0,
                      "WriteThroughput": 0,
                      "WriteLatency": 0,
                      "IdlePower": 0,
                      "IdleScale": 0,
                      "ActivePower": 0,
                      "ActiveWorkScale": 0
                    },
                    {
                      "MaxPower": 0,
                      "Flags": 0,
                      "EntryLat": 0,
                      "ExitLat": 0,
                      "ReadThroughput": 0,
                      "ReadLatency": 0,
                      "WriteThroughput": 0,
                      "WriteLatency": 0,
                      "IdlePower": 0,
                      "IdleScale": 0,
                      "ActivePower": 0,
                      "ActiveWorkScale": 0
                    },
                    {
                      "MaxPower": 0,
                      "Flags": 0,
                      "EntryLat": 0,
                      "ExitLat": 0,
                      "ReadThroughput": 0,
                      "ReadLatency": 0,
                      "WriteThroughput": 0,
                      "WriteLatency": 0,
                      "IdlePower": 0,
                      "IdleScale": 0,
                      "ActivePower": 0,
                      "ActiveWorkScale": 0
                    },
                    {
                      "MaxPower": 0,
                      "Flags": 0,
                      "EntryLat": 0,
                      "ExitLat": 0,
                      "ReadThroughput": 0,
                      "ReadLatency": 0,
                      "WriteThroughput": 0,
                      "WriteLatency": 0,
                      "IdlePower": 0,
                      "IdleScale": 0,
                      "ActivePower": 0,
                      "ActiveWorkScale": 0
                    },
                    {
                      "MaxPower": 0,
                      "Flags": 0,
                      "EntryLat": 0,
                      "ExitLat": 0,
                      "ReadThroughput": 0,
                      "ReadLatency": 0,
                      "WriteThroughput": 0,
                      "WriteLatency": 0,
                      "IdlePower": 0,
                      "IdleScale": 0,
                      "ActivePower": 0,
                      "ActiveWorkScale": 0
                    },
                    {
                      "MaxPower": 0,
                      "Flags": 0,
                      "EntryLat": 0,
                      "ExitLat": 0,
                      "ReadThroughput": 0,
                      "ReadLatency": 0,
                      "WriteThroughput": 0,
                      "WriteLatency": 0,
                      "IdlePower": 0,
                      "IdleScale": 0,
                      "ActivePower": 0,
                      "ActiveWorkScale": 0
                    }
                  ],
                  "Vs": [
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    88,
                    78,
                    85,
                    83,
                    82,
                    71,
                    50,
                    57,
                    18,
                    4,
                    32,
                    23,
                    1,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    7,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0,
                    0
                  ]
                }
              }
            }
          ],
          "status": "OK"
        })},
        500
      );
    });
  },

  smartDef: () => {
    return new Promise((resolve, reject) => {
      setTimeout(() => {
        resolve({
          "data": {
            "ATA": {
              "1": {
                "display_name": "Read Error Rate",
                "ideal": "low",
                "critical": false,
                "description": "(Vendor specific raw value.) Stores data related to the rate of hardware read errors that occurred when reading data from a disk surface. The raw value has different structure for different vendors and is often not meaningful as a decimal number.",
                "display_type": "normalized"
              },
              "2": {
                "display_name": "Throughput Performance",
                "ideal": "high",
                "critical": false,
                "description": "Overall (general) throughput performance of a hard disk drive. If the value of this attribute is decreasing there is a high probability that there is a problem with the disk.",
                "display_type": "normalized"
              },
              "3": {
                "display_name": "Spin-Up Time",
                "ideal": "low",
                "critical": false,
                "description": "Average time of spindle spin up (from zero RPM to fully operational [milliseconds]).",
                "observed_thresholds": [
                  {
                    "low": 78,
                    "high": 96,
                    "annual_failure_rate": 0.11452195377351217,
                    "error_interval": [
                      0.10591837762295722,
                      0.12363823501915781
                    ]
                  },
                  {
                    "low": 96,
                    "high": 114,
                    "annual_failure_rate": 0.040274562840558074,
                    "error_interval": [
                      0.03465055611002801,
                      0.046551312468303144
                    ]
                  },
                  {
                    "low": 114,
                    "high": 132,
                    "annual_failure_rate": 0.009100406705780476,
                    "error_interval": [
                      0.006530608971356785,
                      0.012345729280075591
                    ]
                  },
                  {
                    "low": 132,
                    "high": 150,
                    "annual_failure_rate": 0.008561351734020232,
                    "error_interval": [
                      0.004273795939256936,
                      0.015318623141355509
                    ]
                  },
                  {
                    "low": 150,
                    "high": 168,
                    "annual_failure_rate": 0.015780508262068848,
                    "error_interval": [
                      0.005123888078524015,
                      0.03682644215646287
                    ]
                  },
                  {
                    "low": 168,
                    "high": 186,
                    "annual_failure_rate": 0.05262688124794024,
                    "error_interval": [
                      0.0325768689524594,
                      0.08044577830285578
                    ]
                  },
                  {
                    "low": 186,
                    "high": 204,
                    "annual_failure_rate": 0.01957419424036038,
                    "error_interval": [
                      0.0023705257325185624,
                      0.0707087198669825
                    ]
                  },
                  {
                    "low": 204,
                    "high": 222,
                    "annual_failure_rate": 0.026050959960031404,
                    "error_interval": [
                      0.0006595532020744994,
                      0.1451466588889228
                    ]
                  }
                ],
                "display_type": "normalized"
              },
              "4": {
                "display_name": "Start/Stop Count",
                "ideal": "",
                "critical": false,
                "description": "A tally of spindle start/stop cycles. The spindle turns on, and hence the count is increased, both when the hard disk is turned on after having before been turned entirely off (disconnected from power source) and when the hard disk returns from having previously been put to sleep mode.",
                "observed_thresholds": [
                  {
                    "low": 0,
                    "high": 13,
                    "annual_failure_rate": 0.01989335424860646,
                    "error_interval": [
                      0.016596548909440657,
                      0.023653263230617408
                    ]
                  },
                  {
                    "low": 13,
                    "high": 26,
                    "annual_failure_rate": 0.03776935438256488,
                    "error_interval": [
                      0.03310396052098642,
                      0.04290806173460437
                    ]
                  },
                  {
                    "low": 26,
                    "high": 39,
                    "annual_failure_rate": 0.11022223828187004,
                    "error_interval": [
                      0.09655110535164119,
                      0.12528657238811672
                    ]
                  },
                  {
                    "low": 39,
                    "high": 52,
                    "annual_failure_rate": 0.16289995457762474,
                    "error_interval": [
                      0.13926541653588131,
                      0.18939614504497515
                    ]
                  },
                  {
                    "low": 52,
                    "high": 65,
                    "annual_failure_rate": 0.19358212432279714,
                    "error_interval": [
                      0.15864522253849073,
                      0.23392418181765526
                    ]
                  },
                  {
                    "low": 65,
                    "high": 78,
                    "annual_failure_rate": 0.1157094940074447,
                    "error_interval": [
                      0.07861898732346269,
                      0.16424039052527728
                    ]
                  },
                  {
                    "low": 78,
                    "high": 91,
                    "annual_failure_rate": 0.12262136155304391,
                    "error_interval": [
                      0.0670382394080032,
                      0.20573780888032978
                    ]
                  },
                  {
                    "low": 91,
                    "high": 104,
                    "annual_failure_rate": 0,
                    "error_interval": [
                      0,
                      0
                    ]
                  }
                ],
                "display_type": "raw"
              },
              "5": {
                "display_name": "Reallocated Sectors Count",
                "ideal": "low",
                "critical": true,
                "description": "Count of reallocated sectors. The raw value represents a count of the bad sectors that have been found and remapped.Thus, the higher the attribute value, the more sectors the drive has had to reallocate. This value is primarily used as a metric of the life expectancy of the drive; a drive which has had any reallocations at all is significantly more likely to fail in the immediate months.",
                "observed_thresholds": [
                  {
                    "low": 0,
                    "high": 0,
                    "annual_failure_rate": 0.025169175350572493,
                    "error_interval": [
                      0.022768612038746357,
                      0.027753988579272894
                    ]
                  },
                  {
                    "low": 1,
                    "high": 4,
                    "annual_failure_rate": 0.027432608477803388,
                    "error_interval": [
                      0.010067283827589948,
                      0.05970923963096652
                    ]
                  },
                  {
                    "low": 4,
                    "high": 16,
                    "annual_failure_rate": 0.07501976284584981,
                    "error_interval": [
                      0.039944864177334186,
                      0.12828607921150972
                    ]
                  },
                  {
                    "low": 16,
                    "high": 70,
                    "annual_failure_rate": 0.23589260654405794,
                    "error_interval": [
                      0.1643078435800227,
                      0.32806951196017664
                    ]
                  },
                  {
                    "low": 70,
                    "high": 260,
                    "annual_failure_rate": 0.36193219378600433,
                    "error_interval": [
                      0.2608488901774093,
                      0.4892271827875412
                    ]
                  },
                  {
                    "low": 260,
                    "high": 1100,
                    "annual_failure_rate": 0.5676621428968173,
                    "error_interval": [
                      0.4527895568499355,
                      0.702804359408436
                    ]
                  },
                  {
                    "low": 1100,
                    "high": 4500,
                    "annual_failure_rate": 1.5028253400346423,
                    "error_interval": [
                      1.2681757596263297,
                      1.768305221795894
                    ]
                  },
                  {
                    "low": 4500,
                    "high": 17000,
                    "annual_failure_rate": 2.0659987547404763,
                    "error_interval": [
                      1.6809790460512237,
                      2.512808045182302
                    ]
                  },
                  {
                    "low": 17000,
                    "high": 70000,
                    "annual_failure_rate": 1.7755385684503124,
                    "error_interval": [
                      1.2796520259849835,
                      2.400012341226441
                    ]
                  }
                ],
                "display_type": "raw"
              },
              "6": {
                "display_name": "Read Channel Margin",
                "ideal": "",
                "critical": false,
                "description": "Margin of a channel while reading data. The function of this attribute is not specified.",
                "display_type": "normalized"
              },
              "7": {
                "display_name": "Seek Error Rate",
                "ideal": "",
                "critical": false,
                "description": "(Vendor specific raw value.) Rate of seek errors of the magnetic heads. If there is a partial failure in the mechanical positioning system, then seek errors will arise. Such a failure may be due to numerous factors, such as damage to a servo, or thermal widening of the hard disk. The raw value has different structure for different vendors and is often not meaningful as a decimal number.",
                "display_type": "normalized"
              },
              "8": {
                "display_name": "Seek Time Performance",
                "ideal": "high",
                "critical": false,
                "description": "Average performance of seek operations of the magnetic heads. If this attribute is decreasing, it is a sign of problems in the mechanical subsystem.",
                "display_type": "normalized"
              },
              "9": {
                "display_name": "Power-On Hours",
                "ideal": "",
                "critical": false,
                "description": "Count of hours in power-on state. The raw value of this attribute shows total count of hours (or minutes, or seconds, depending on manufacturer) in power-on state. By default, the total expected lifetime of a hard disk in perfect condition is defined as 5 years (running every day and night on all days). This is equal to 1825 days in 24/7 mode or 43800 hours. On some pre-2005 drives, this raw value may advance erratically and/or \"wrap around\" (reset to zero periodically).",
                "display_type": "normalized"
              },
              "10": {
                "display_name": "Spin Retry Count",
                "ideal": "low",
                "critical": true,
                "description": "Count of retry of spin start attempts. This attribute stores a total count of the spin start attempts to reach the fully operational speed (under the condition that the first attempt was unsuccessful). An increase of this attribute value is a sign of problems in the hard disk mechanical subsystem.",
                "observed_thresholds": [
                  {
                    "low": 0,
                    "high": 0,
                    "annual_failure_rate": 0.05459827163896099,
                    "error_interval": [
                      0.05113785787727033,
                      0.05823122757702782
                    ]
                  },
                  {
                    "low": 0,
                    "high": 80,
                    "annual_failure_rate": 0.5555555555555556,
                    "error_interval": [
                      0.014065448880161053,
                      3.095357439410498
                    ]
                  }
                ],
                "display_type": "raw"
              },
              "11": {
                "display_name": "Recalibration Retries or Calibration Retry Count",
                "ideal": "low",
                "critical": false,
                "description": "This attribute indicates the count that recalibration was requested (under the condition that the first attempt was unsuccessful). An increase of this attribute value is a sign of problems in the hard disk mechanical subsystem.",
                "observed_thresholds": [
                  {
                    "low": 0,
                    "high": 0,
                    "annual_failure_rate": 0.04658866433672694,
                    "error_interval": [
                      0.03357701137320878,
                      0.06297433993055492
                    ]
                  },
                  {
                    "low": 0,
                    "high": 80,
                    "annual_failure_rate": 0.5555555555555556,
                    "error_interval": [
                      0.014065448880161053,
                      3.095357439410498
                    ]
                  },
                  {
                    "low": 80,
                    "high": 160,
                    "annual_failure_rate": 0,
                    "error_interval": [
                      0,
                      0
                    ]
                  },
                  {
                    "low": 160,
                    "high": 240,
                    "annual_failure_rate": 0,
                    "error_interval": [
                      0,
                      0
                    ]
                  },
                  {
                    "low": 240,
                    "high": 320,
                    "annual_failure_rate": 0,
                    "error_interval": [
                      0,
                      0
                    ]
                  },
                  {
                    "low": 320,
                    "high": 400,
                    "annual_failure_rate": 0,
                    "error_interval": [
                      0,
                      0
                    ]
                  },
                  {
                    "low": 400,
                    "high": 480,
                    "annual_failure_rate": 0,
                    "error_interval": [
                      0,
                      0
                    ]
                  },
                  {
                    "low": 480,
                    "high": 560,
                    "annual_failure_rate": 0,
                    "error_interval": [
                      0,
                      0
                    ]
                  }
                ],
                "display_type": "raw"
              },
              "12": {
                "display_name": "Power Cycle Count",
                "ideal": "low",
                "critical": false,
                "description": "This attribute indicates the count of full hard disk power on/off cycles.",
                "display_type": "raw"
              },
              "13": {
                "display_name": "Soft Read Error Rate",
                "ideal": "low",
                "critical": false,
                "description": "Uncorrected read errors reported to the operating system.",
                "display_type": "normalized"
              },
              "22": {
                "display_name": "Current Helium Level",
                "ideal": "high",
                "critical": false,
                "description": "Specific to He8 drives from HGST. This value measures the helium inside of the drive specific to this manufacturer. It is a pre-fail attribute that trips once the drive detects that the internal environment is out of specification.",
                "display_type": "normalized"
              },
              "170": {
                "display_name": "Available Reserved Space",
                "ideal": "",
                "critical": false,
                "description": "See attribute E8.",
                "display_type": "normalized"
              },
              "171": {
                "display_name": "SSD Program Fail Count",
                "ideal": "",
                "critical": false,
                "description": "(Kingston) The total number of flash program operation failures since the drive was deployed.[33] Identical to attribute 181.",
                "display_type": "normalized"
              },
              "172": {
                "display_name": "SSD Erase Fail Count",
                "ideal": "",
                "critical": false,
                "description": "(Kingston) Counts the number of flash erase failures. This attribute returns the total number of Flash erase operation failures since the drive was deployed. This attribute is identical to attribute 182.",
                "display_type": "normalized"
              },
              "173": {
                "display_name": "SSD Wear Leveling Count",
                "ideal": "",
                "critical": false,
                "description": "Counts the maximum worst erase count on any block.",
                "display_type": "normalized"
              },
              "174": {
                "display_name": "Unexpected Power Loss Count",
                "ideal": "",
                "critical": false,
                "description": "Also known as \"Power-off Retract Count\" per conventional HDD terminology. Raw value reports the number of unclean shutdowns, cumulative over the life of an SSD, where an \"unclean shutdown\" is the removal of power without STANDBY IMMEDIATE as the last command (regardless of PLI activity using capacitor power). Normalized value is always 100.",
                "display_type": ""
              },
              "175": {
                "display_name": "Power Loss Protection Failure",
                "ideal": "",
                "critical": false,
                "description": "Last test result as microseconds to discharge cap, saturated at its maximum value. Also logs minutes since last test and lifetime number of tests. Raw value contains the following data:     Bytes 0-1: Last test result as microseconds to discharge cap, saturates at max value. Test result expected in range 25 <= result <= 5000000, lower indicates specific error code. Bytes 2-3: Minutes since last test, saturates at max value.Bytes 4-5: Lifetime number of tests, not incremented on power cycle, saturates at max value. Normalized value is set to one on test failure or 11 if the capacitor has been tested in an excessive temperature condition, otherwise 100.",
                "display_type": "normalized"
              },
              "176": {
                "display_name": "Erase Fail Count",
                "ideal": "",
                "critical": false,
                "description": "S.M.A.R.T. parameter indicates a number of flash erase command failures.",
                "display_type": "normalized"
              },
              "177": {
                "display_name": "Wear Range Delta",
                "ideal": "",
                "critical": false,
                "description": "Delta between most-worn and least-worn Flash blocks. It describes how good/bad the wearleveling of the SSD works on a more technical way. ",
                "display_type": "normalized"
              },
              "179": {
                "display_name": "Used Reserved Block Count Total",
                "ideal": "",
                "critical": false,
                "description": "Pre-Fail attribute used at least in Samsung devices.",
                "display_type": "normalized"
              },
              "180": {
                "display_name": "Unused Reserved Block Count Total",
                "ideal": "",
                "critical": false,
                "description": "\"Pre-Fail\" attribute used at least in HP devices. ",
                "display_type": "normalized"
              },
              "181": {
                "display_name": "Program Fail Count Total",
                "ideal": "",
                "critical": false,
                "description": "Total number of Flash program operation failures since the drive was deployed.",
                "display_type": "normalized"
              },
              "182": {
                "display_name": "Erase Fail Count",
                "ideal": "",
                "critical": false,
                "description": "\"Pre-Fail\" Attribute used at least in Samsung devices.",
                "display_type": "normalized"
              },
              "183": {
                "display_name": "SATA Downshift Error Count or Runtime Bad Block",
                "ideal": "low",
                "critical": false,
                "description": "Western Digital, Samsung or Seagate attribute: Either the number of downshifts of link speed (e.g. from 6Gbit/s to 3Gbit/s) or the total number of data blocks with detected, uncorrectable errors encountered during normal operation. Although degradation of this parameter can be an indicator of drive aging and/or potential electromechanical problems, it does not directly indicate imminent drive failure.",
                "observed_thresholds": [
                  {
                    "low": 0,
                    "high": 0,
                    "annual_failure_rate": 0.09084549203210031,
                    "error_interval": [
                      0.08344373475686712,
                      0.09872777224842152
                    ]
                  },
                  {
                    "low": 1,
                    "high": 2,
                    "annual_failure_rate": 0.05756065656498585,
                    "error_interval": [
                      0.04657000847949464,
                      0.07036491775108872
                    ]
                  },
                  {
                    "low": 2,
                    "high": 4,
                    "annual_failure_rate": 0.6193088626208925,
                    "error_interval": [
                      0.41784508895529787,
                      0.8841019099092139
                    ]
                  },
                  {
                    "low": 4,
                    "high": 8,
                    "annual_failure_rate": 0.5533447034299792,
                    "error_interval": [
                      0.31628430884775033,
                      0.8985971312402635
                    ]
                  },
                  {
                    "low": 8,
                    "high": 16,
                    "annual_failure_rate": 0.3882388694727245,
                    "error_interval": [
                      0.21225380267814295,
                      0.6513988534774338
                    ]
                  },
                  {
                    "low": 16,
                    "high": 35,
                    "annual_failure_rate": 0.37116708385481856,
                    "error_interval": [
                      0.19763084005134446,
                      0.6347070173754686
                    ]
                  },
                  {
                    "low": 35,
                    "high": 70,
                    "annual_failure_rate": 0.2561146752205292,
                    "error_interval": [
                      0.10297138269895259,
                      0.5276941165819332
                    ]
                  },
                  {
                    "low": 70,
                    "high": 130,
                    "annual_failure_rate": 0.40299684542586756,
                    "error_interval": [
                      0.16202563309223209,
                      0.8303275247667772
                    ]
                  },
                  {
                    "low": 130,
                    "high": 260,
                    "annual_failure_rate": 0,
                    "error_interval": [
                      0,
                      0
                    ]
                  }
                ],
                "display_type": "raw"
              },
              "184": {
                "display_name": "End-to-End error",
                "ideal": "low",
                "critical": true,
                "description": "This attribute is a part of Hewlett-Packard\"s SMART IV technology, as well as part of other vendors\" IO Error Detection and Correction schemas, and it contains a count of parity errors which occur in the data path to the media via the drive\"s cache RAM",
                "observed_thresholds": [
                  {
                    "low": 93,
                    "high": 94,
                    "annual_failure_rate": 1.631212012870933,
                    "error_interval": [
                      1.055634407303844,
                      2.407990716767714
                    ]
                  },
                  {
                    "low": 94,
                    "high": 95,
                    "annual_failure_rate": 0,
                    "error_interval": [
                      0,
                      0
                    ]
                  },
                  {
                    "low": 95,
                    "high": 96,
                    "annual_failure_rate": 0,
                    "error_interval": [
                      0,
                      0
                    ]
                  },
                  {
                    "low": 96,
                    "high": 97,
                    "annual_failure_rate": 0,
                    "error_interval": [
                      0,
                      0
                    ]
                  },
                  {
                    "low": 97,
                    "high": 97,
                    "annual_failure_rate": 0,
                    "error_interval": [
                      0,
                      0
                    ]
                  },
                  {
                    "low": 97,
                    "high": 98,
                    "annual_failure_rate": 1.8069306930693072,
                    "error_interval": [
                      0.04574752432804858,
                      10.067573453924245
                    ]
                  },
                  {
                    "low": 98,
                    "high": 99,
                    "annual_failure_rate": 0.8371559633027523,
                    "error_interval": [
                      0.10138347095016888,
                      3.0240951820174824
                    ]
                  },
                  {
                    "low": 99,
                    "high": 100,
                    "annual_failure_rate": 0.09334816849865138,
                    "error_interval": [
                      0.08689499010435861,
                      0.10015372448181788
                    ]
                  }
                ],
                "display_type": "normalized"
              },
              "185": {
                "display_name": "Head Stability",
                "ideal": "",
                "critical": false,
                "description": "Western Digital attribute.",
                "display_type": "normalized"
              },
              "186": {
                "display_name": "Induced Op-Vibration Detection",
                "ideal": "",
                "critical": false,
                "description": "Western Digital attribute.",
                "display_type": "normalized"
              },
              "187": {
                "display_name": "Reported Uncorrectable Errors",
                "ideal": "low",
                "critical": true,
                "description": "The count of errors that could not be recovered using hardware ECC (see attribute 195).",
                "observed_thresholds": [
                  {
                    "low": 0,
                    "high": 0,
                    "annual_failure_rate": 0.028130798308190524,
                    "error_interval": [
                      0.024487830609364304,
                      0.032162944988161336
                    ]
                  },
                  {
                    "low": 1,
                    "high": 1,
                    "annual_failure_rate": 0.33877621175661743,
                    "error_interval": [
                      0.22325565823630591,
                      0.4929016016666955
                    ]
                  },
                  {
                    "low": 1,
                    "high": 3,
                    "annual_failure_rate": 0.24064820598237213,
                    "error_interval": [
                      0.14488594021076606,
                      0.3758019832614595
                    ]
                  },
                  {
                    "low": 3,
                    "high": 6,
                    "annual_failure_rate": 0.5014425058387142,
                    "error_interval": [
                      0.3062941096766342,
                      0.7744372808405151
                    ]
                  },
                  {
                    "low": 6,
                    "high": 11,
                    "annual_failure_rate": 0.38007108544136836,
                    "error_interval": [
                      0.2989500188963677,
                      0.4764223967570595
                    ]
                  },
                  {
                    "low": 11,
                    "high": 20,
                    "annual_failure_rate": 0.5346094598348444,
                    "error_interval": [
                      0.40595137663302483,
                      0.6911066985735377
                    ]
                  },
                  {
                    "low": 20,
                    "high": 35,
                    "annual_failure_rate": 0.8428063943161636,
                    "error_interval": [
                      0.6504601819243522,
                      1.0742259350903411
                    ]
                  },
                  {
                    "low": 35,
                    "high": 65,
                    "annual_failure_rate": 1.4429071005017484,
                    "error_interval": [
                      1.1405581860945952,
                      1.8008133631629157
                    ]
                  },
                  {
                    "low": 65,
                    "high": 120,
                    "annual_failure_rate": 1.6190935390549661,
                    "error_interval": [
                      1.0263664163011208,
                      2.4294352761068576
                    ]
                  }
                ],
                "display_type": "raw"
              },
              "188": {
                "display_name": "Command Timeout",
                "ideal": "low",
                "critical": true,
                "description": "The count of aborted operations due to HDD timeout. Normally this attribute value should be equal to zero.",
                "observed_thresholds": [
                  {
                    "low": 0,
                    "high": 100,
                    "annual_failure_rate": 0.024893587674442153,
                    "error_interval": [
                      0.020857343769186413,
                      0.0294830350167543
                    ]
                  },
                  {
                    "low": 100,
                    "high": 13000000000,
                    "annual_failure_rate": 0.10044174089362015,
                    "error_interval": [
                      0.0812633664077498,
                      0.1227848196758574
                    ]
                  },
                  {
                    "low": 13000000000,
                    "high": 26000000000,
                    "annual_failure_rate": 0.334030592234279,
                    "error_interval": [
                      0.2523231196342665,
                      0.4337665082489293
                    ]
                  },
                  {
                    "low": 26000000000,
                    "high": 39000000000,
                    "annual_failure_rate": 0.36724705400842445,
                    "error_interval": [
                      0.30398009356575617,
                      0.4397986538328568
                    ]
                  },
                  {
                    "low": 39000000000,
                    "high": 52000000000,
                    "annual_failure_rate": 0.29848155926978354,
                    "error_interval": [
                      0.2509254838615984,
                      0.35242890006477073
                    ]
                  },
                  {
                    "low": 52000000000,
                    "high": 65000000000,
                    "annual_failure_rate": 0.2203079701535098,
                    "error_interval": [
                      0.18366082845676174,
                      0.26212468677179274
                    ]
                  },
                  {
                    "low": 65000000000,
                    "high": 78000000000,
                    "annual_failure_rate": 0.3018169948863018,
                    "error_interval": [
                      0.23779746376787655,
                      0.37776897542831006
                    ]
                  },
                  {
                    "low": 78000000000,
                    "high": 91000000000,
                    "annual_failure_rate": 0.32854928239235887,
                    "error_interval": [
                      0.2301118782147336,
                      0.4548506948185028
                    ]
                  },
                  {
                    "low": 91000000000,
                    "high": 104000000000,
                    "annual_failure_rate": 0.28488916640649387,
                    "error_interval": [
                      0.1366154288236293,
                      0.5239213202729072
                    ]
                  }
                ],
                "display_type": "transformed"
              },
              "189": {
                "display_name": "High Fly Writes",
                "ideal": "low",
                "critical": false,
                "description": "HDD manufacturers implement a flying height sensor that attempts to provide additional protections for write operations by detecting when a recording head is flying outside its normal operating range. If an unsafe fly height condition is encountered, the write process is stopped, and the information is rewritten or reallocated to a safe region of the hard drive. This attribute indicates the count of these errors detected over the lifetime of the drive.",
                "observed_thresholds": [
                  {
                    "low": 0,
                    "high": 0,
                    "annual_failure_rate": 0.09070551401946862,
                    "error_interval": [
                      0.08018892683853401,
                      0.10221801211956287
                    ]
                  },
                  {
                    "low": 1,
                    "high": 2,
                    "annual_failure_rate": 0.0844336097370013,
                    "error_interval": [
                      0.07299813695315267,
                      0.09715235540340669
                    ]
                  },
                  {
                    "low": 2,
                    "high": 5,
                    "annual_failure_rate": 0.07943219628781906,
                    "error_interval": [
                      0.06552176680630226,
                      0.09542233189887633
                    ]
                  },
                  {
                    "low": 5,
                    "high": 13,
                    "annual_failure_rate": 0.09208847603893404,
                    "error_interval": [
                      0.07385765060838133,
                      0.11345557807163456
                    ]
                  },
                  {
                    "low": 13,
                    "high": 30,
                    "annual_failure_rate": 0.18161161650924224,
                    "error_interval": [
                      0.13858879602902988,
                      0.23377015012749933
                    ]
                  },
                  {
                    "low": 30,
                    "high": 70,
                    "annual_failure_rate": 0.2678117886102384,
                    "error_interval": [
                      0.19044036194841887,
                      0.36610753129699186
                    ]
                  },
                  {
                    "low": 70,
                    "high": 150,
                    "annual_failure_rate": 0.26126480798826107,
                    "error_interval": [
                      0.15958733218826962,
                      0.4035023060905559
                    ]
                  },
                  {
                    "low": 150,
                    "high": 350,
                    "annual_failure_rate": 0.11337164155924832,
                    "error_interval": [
                      0.030889956621649995,
                      0.2902764300762812
                    ]
                  }
                ],
                "display_type": "raw"
              },
              "190": {
                "display_name": "Temperature Difference",
                "ideal": "",
                "critical": false,
                "description": "Value is equal to (100-temp. °C), allowing manufacturer to set a minimum threshold which corresponds to a maximum temperature. This also follows the convention of 100 being a best-case value and lower values being undesirable. However, some older drives may instead report raw Temperature (identical to 0xC2) or Temperature minus 50 here.",
                "display_type": "normalized"
              },
              "191": {
                "display_name": "G-sense Error Rate",
                "ideal": "low",
                "critical": false,
                "description": "The count of errors resulting from externally induced shock and vibration. ",
                "display_type": "normalized"
              },
              "192": {
                "display_name": "Power-off Retract Count",
                "ideal": "low",
                "critical": false,
                "description": "Number of power-off or emergency retract cycles.",
                "observed_thresholds": [
                  {
                    "low": 1,
                    "high": 2,
                    "annual_failure_rate": 0.02861098445412803,
                    "error_interval": [
                      0.022345416230915037,
                      0.036088863823297186
                    ]
                  },
                  {
                    "low": 2,
                    "high": 6,
                    "annual_failure_rate": 0.0738571777154862,
                    "error_interval": [
                      0.06406927746420421,
                      0.0847175264009771
                    ]
                  },
                  {
                    "low": 6,
                    "high": 16,
                    "annual_failure_rate": 0.11970378206823593,
                    "error_interval": [
                      0.10830059875098269,
                      0.13198105985656441
                    ]
                  },
                  {
                    "low": 16,
                    "high": 40,
                    "annual_failure_rate": 0.027266868552620425,
                    "error_interval": [
                      0.021131448605713823,
                      0.03462795920968522
                    ]
                  },
                  {
                    "low": 40,
                    "high": 100,
                    "annual_failure_rate": 0.011741682974559688,
                    "error_interval": [
                      0.00430899071133239,
                      0.025556700631152028
                    ]
                  },
                  {
                    "low": 100,
                    "high": 250,
                    "annual_failure_rate": 0.012659940134091309,
                    "error_interval": [
                      0.00607093338127348,
                      0.023282080653656938
                    ]
                  },
                  {
                    "low": 250,
                    "high": 650,
                    "annual_failure_rate": 0.01634692899031039,
                    "error_interval": [
                      0.009522688540043157,
                      0.026173016865409605
                    ]
                  },
                  {
                    "low": 650,
                    "high": 1600,
                    "annual_failure_rate": 0.005190074354440066,
                    "error_interval": [
                      0.0025908664180103293,
                      0.009286476666453648
                    ]
                  }
                ],
                "display_type": "raw"
              },
              "193": {
                "display_name": "Load Cycle Count",
                "ideal": "low",
                "critical": false,
                "description": "Count of load/unload cycles into head landing zone position.[45] Some drives use 225 (0xE1) for Load Cycle Count instead.",
                "display_type": "normalized"
              },
              "194": {
                "display_name": "Temperature",
                "ideal": "low",
                "critical": false,
                "description": "Indicates the device temperature, if the appropriate sensor is fitted. Lowest byte of the raw value contains the exact temperature value (Celsius degrees).",
                "transform_value_unit": "°C",
                "display_type": "transformed"
              },
              "195": {
                "display_name": "Hardware ECC Recovered",
                "ideal": "",
                "critical": false,
                "description": "(Vendor-specific raw value.) The raw value has different structure for different vendors and is often not meaningful as a decimal number.",
                "observed_thresholds": [
                  {
                    "low": 12,
                    "high": 24,
                    "annual_failure_rate": 0.31472916829975706,
                    "error_interval": [
                      0.15711166685282174,
                      0.5631374192486645
                    ]
                  },
                  {
                    "low": 24,
                    "high": 36,
                    "annual_failure_rate": 0.15250310197260136,
                    "error_interval": [
                      0.10497611828070175,
                      0.21417105521823687
                    ]
                  },
                  {
                    "low": 36,
                    "high": 48,
                    "annual_failure_rate": 0.2193119102723874,
                    "error_interval": [
                      0.16475385681835103,
                      0.28615447006525274
                    ]
                  },
                  {
                    "low": 48,
                    "high": 60,
                    "annual_failure_rate": 0.05672658497265746,
                    "error_interval": [
                      0.043182904776447234,
                      0.07317316161437043
                    ]
                  },
                  {
                    "low": 60,
                    "high": 72,
                    "annual_failure_rate": 0,
                    "error_interval": [
                      0,
                      0
                    ]
                  },
                  {
                    "low": 72,
                    "high": 84,
                    "annual_failure_rate": 0,
                    "error_interval": [
                      0,
                      0
                    ]
                  },
                  {
                    "low": 84,
                    "high": 96,
                    "annual_failure_rate": 0,
                    "error_interval": [
                      0,
                      0
                    ]
                  },
                  {
                    "low": 96,
                    "high": 108,
                    "annual_failure_rate": 0.04074570216566197,
                    "error_interval": [
                      0.001031591863615295,
                      0.22702052218047528
                    ]
                  }
                ],
                "display_type": "normalized"
              },
              "196": {
                "display_name": "Reallocation Event Count",
                "ideal": "low",
                "critical": true,
                "description": "Count of remap operations. The raw value of this attribute shows the total count of attempts to transfer data from reallocated sectors to a spare area. Both successful and unsuccessful attempts are counted.",
                "observed_thresholds": [
                  {
                    "low": 0,
                    "high": 0,
                    "annual_failure_rate": 0.007389855800729792,
                    "error_interval": [
                      0.005652654139732716,
                      0.009492578928212054
                    ]
                  },
                  {
                    "low": 1,
                    "high": 1,
                    "annual_failure_rate": 0.026558331312151347,
                    "error_interval": [
                      0.005476966404484466,
                      0.07761471429677293
                    ]
                  },
                  {
                    "low": 1,
                    "high": 2,
                    "annual_failure_rate": 0.02471894893674658,
                    "error_interval": [
                      0.0006258296027540169,
                      0.13772516847438018
                    ]
                  },
                  {
                    "low": 2,
                    "high": 4,
                    "annual_failure_rate": 0.03200912040691046,
                    "error_interval": [
                      0.0008104007642081744,
                      0.17834340416493005
                    ]
                  },
                  {
                    "low": 4,
                    "high": 7,
                    "annual_failure_rate": 0.043078012510326925,
                    "error_interval": [
                      0.001090640849081295,
                      0.24001532369794615
                    ]
                  },
                  {
                    "low": 7,
                    "high": 11,
                    "annual_failure_rate": 0.033843300880853036,
                    "error_interval": [
                      0.0008568381932559863,
                      0.18856280368036135
                    ]
                  },
                  {
                    "low": 11,
                    "high": 17,
                    "annual_failure_rate": 0.16979376647542252,
                    "error_interval": [
                      0.035015556653263225,
                      0.49620943874336304
                    ]
                  },
                  {
                    "low": 17,
                    "high": 27,
                    "annual_failure_rate": 0.059042381106438044,
                    "error_interval": [
                      0.0014948236677880642,
                      0.32896309247698113
                    ]
                  },
                  {
                    "low": 27,
                    "high": 45,
                    "annual_failure_rate": 0.24701105346266636,
                    "error_interval": [
                      0.050939617608142244,
                      0.721871118983972
                    ]
                  }
                ],
                "display_type": "raw"
              },
              "197": {
                "display_name": "Current Pending Sector Count",
                "ideal": "low",
                "critical": true,
                "description": "Count of \"unstable\" sectors (waiting to be remapped, because of unrecoverable read errors). If an unstable sector is subsequently read successfully, the sector is remapped and this value is decreased. Read errors on a sector will not remap the sector immediately (since the correct value cannot be read and so the value to remap is not known, and also it might become readable later); instead, the drive firmware remembers that the sector needs to be remapped, and will remap it the next time it\"s written.",
                "observed_thresholds": [
                  {
                    "low": 0,
                    "high": 0,
                    "annual_failure_rate": 0.025540791394761345,
                    "error_interval": [
                      0.023161777231213983,
                      0.02809784482748174
                    ]
                  },
                  {
                    "low": 1,
                    "high": 2,
                    "annual_failure_rate": 0.34196613799103254,
                    "error_interval": [
                      0.22723401523750225,
                      0.4942362818474496
                    ]
                  },
                  {
                    "low": 2,
                    "high": 6,
                    "annual_failure_rate": 0.6823772508117681,
                    "error_interval": [
                      0.41083568090070416,
                      1.0656166047061635
                    ]
                  },
                  {
                    "low": 6,
                    "high": 16,
                    "annual_failure_rate": 0.6108100007493069,
                    "error_interval": [
                      0.47336936083368364,
                      0.7757071095273286
                    ]
                  },
                  {
                    "low": 16,
                    "high": 40,
                    "annual_failure_rate": 0.9564879341127684,
                    "error_interval": [
                      0.7701044196378299,
                      1.174355230793638
                    ]
                  },
                  {
                    "low": 40,
                    "high": 100,
                    "annual_failure_rate": 1.6519989942167461,
                    "error_interval": [
                      1.328402276482456,
                      2.0305872327541317
                    ]
                  },
                  {
                    "low": 100,
                    "high": 250,
                    "annual_failure_rate": 2.5137741046831956,
                    "error_interval": [
                      1.9772427971560862,
                      3.1510376077891613
                    ]
                  },
                  {
                    "low": 250,
                    "high": 650,
                    "annual_failure_rate": 3.3203378817413904,
                    "error_interval": [
                      2.5883662702274406,
                      4.195047163573006
                    ]
                  },
                  {
                    "low": 650,
                    "high": 1600,
                    "annual_failure_rate": 3.133047210300429,
                    "error_interval": [
                      1.1497731080460096,
                      6.819324775707182
                    ]
                  }
                ],
                "display_type": "raw"
              },
              "198": {
                "display_name": "(Offline) Uncorrectable Sector Count",
                "ideal": "low",
                "critical": true,
                "description": "The total count of uncorrectable errors when reading/writing a sector. A rise in the value of this attribute indicates defects of the disk surface and/or problems in the mechanical subsystem.",
                "observed_thresholds": [
                  {
                    "low": 0,
                    "high": 0,
                    "annual_failure_rate": 0.028675322159886437,
                    "error_interval": [
                      0.026159385510707116,
                      0.03136793218577656
                    ]
                  },
                  {
                    "low": 0,
                    "high": 2,
                    "annual_failure_rate": 0.8135764944275583,
                    "error_interval": [
                      0.40613445471964466,
                      1.4557130815309443
                    ]
                  },
                  {
                    "low": 2,
                    "high": 4,
                    "annual_failure_rate": 1.1173469387755102,
                    "error_interval": [
                      0.5773494680315332,
                      1.9517802404552516
                    ]
                  },
                  {
                    "low": 4,
                    "high": 6,
                    "annual_failure_rate": 1.3558692421991083,
                    "error_interval": [
                      0.4402470522980859,
                      3.1641465148237544
                    ]
                  },
                  {
                    "low": 6,
                    "high": 8,
                    "annual_failure_rate": 0.7324414715719062,
                    "error_interval": [
                      0.15104704003805655,
                      2.140504796291604
                    ]
                  },
                  {
                    "low": 8,
                    "high": 10,
                    "annual_failure_rate": 0.5777213677766163,
                    "error_interval": [
                      0.43275294849366835,
                      0.7556737733062419
                    ]
                  },
                  {
                    "low": 10,
                    "high": 12,
                    "annual_failure_rate": 1.7464114832535886,
                    "error_interval": [
                      0.47583835092536914,
                      4.471507017371231
                    ]
                  },
                  {
                    "low": 12,
                    "high": 14,
                    "annual_failure_rate": 2.6449275362318843,
                    "error_interval": [
                      0.3203129951758959,
                      9.554387676519005
                    ]
                  },
                  {
                    "low": 14,
                    "high": 16,
                    "annual_failure_rate": 0.796943231441048,
                    "error_interval": [
                      0.5519063550198366,
                      1.113648286331181
                    ]
                  }
                ],
                "display_type": "raw"
              },
              "199": {
                "display_name": "UltraDMA CRC Error Count",
                "ideal": "low",
                "critical": false,
                "description": "The count of errors in data transfer via the interface cable as determined by ICRC (Interface Cyclic Redundancy Check).",
                "observed_thresholds": [
                  {
                    "low": 0,
                    "high": 1,
                    "annual_failure_rate": 0.04068379316116366,
                    "error_interval": [
                      0.037534031558106425,
                      0.04402730201866553
                    ]
                  },
                  {
                    "low": 1,
                    "high": 2,
                    "annual_failure_rate": 0.1513481259734218,
                    "error_interval": [
                      0.12037165605991791,
                      0.18786293065527596
                    ]
                  },
                  {
                    "low": 2,
                    "high": 4,
                    "annual_failure_rate": 0.16849758722418978,
                    "error_interval": [
                      0.12976367397863445,
                      0.2151676572000481
                    ]
                  },
                  {
                    "low": 4,
                    "high": 8,
                    "annual_failure_rate": 0.15385127340491614,
                    "error_interval": [
                      0.10887431782430312,
                      0.21117289306426648
                    ]
                  },
                  {
                    "low": 8,
                    "high": 16,
                    "annual_failure_rate": 0.14882894050104387,
                    "error_interval": [
                      0.09631424312463635,
                      0.2197008753522735
                    ]
                  },
                  {
                    "low": 16,
                    "high": 35,
                    "annual_failure_rate": 0.20878219917249793,
                    "error_interval": [
                      0.14086447304552446,
                      0.29804957135975
                    ]
                  },
                  {
                    "low": 35,
                    "high": 70,
                    "annual_failure_rate": 0.13742940270409038,
                    "error_interval": [
                      0.06860426267470295,
                      0.24589916335290812
                    ]
                  },
                  {
                    "low": 70,
                    "high": 130,
                    "annual_failure_rate": 0.22336578581363,
                    "error_interval": [
                      0.11150339549604707,
                      0.39966309081252904
                    ]
                  },
                  {
                    "low": 130,
                    "high": 260,
                    "annual_failure_rate": 0.18277416124186283,
                    "error_interval": [
                      0.07890890989692058,
                      0.3601379610272007
                    ]
                  }
                ],
                "display_type": "raw"
              },
              "200": {
                "display_name": "Multi-Zone Error Rate",
                "ideal": "low",
                "critical": false,
                "description": "The count of errors found when writing a sector. The higher the value, the worse the disk\"s mechanical condition is.",
                "display_type": "normalized"
              },
              "201": {
                "display_name": "Soft Read Error Rate",
                "ideal": "low",
                "critical": true,
                "description": "Count indicates the number of uncorrectable software read errors.",
                "display_type": "normalized"
              },
              "202": {
                "display_name": "Data Address Mark errors",
                "ideal": "low",
                "critical": false,
                "description": "Count of Data Address Mark errors (or vendor-specific).",
                "display_type": "normalized"
              },
              "203": {
                "display_name": "Run Out Cancel",
                "ideal": "low",
                "critical": false,
                "description": "The number of errors caused by incorrect checksum during the error correction.",
                "display_type": "normalized"
              },
              "204": {
                "display_name": "Soft ECC Correction",
                "ideal": "low",
                "critical": false,
                "description": "Count of errors corrected by the internal error correction software.",
                "display_type": ""
              },
              "205": {
                "display_name": "Thermal Asperity Rate",
                "ideal": "low",
                "critical": false,
                "description": "Count of errors due to high temperature.",
                "display_type": "normalized"
              },
              "206": {
                "display_name": "Flying Height",
                "ideal": "",
                "critical": false,
                "description": "Height of heads above the disk surface. If too low, head crash is more likely; if too high, read/write errors are more likely.",
                "display_type": "normalized"
              },
              "207": {
                "display_name": "Spin High Current",
                "ideal": "low",
                "critical": false,
                "description": "Amount of surge current used to spin up the drive.",
                "display_type": "normalized"
              },
              "208": {
                "display_name": "Spin Buzz",
                "ideal": "",
                "critical": false,
                "description": "Count of buzz routines needed to spin up the drive due to insufficient power.",
                "display_type": "normalized"
              },
              "209": {
                "display_name": "Offline Seek Performance",
                "ideal": "",
                "critical": false,
                "description": "Drive\"s seek performance during its internal tests.",
                "display_type": "normalized"
              },
              "210": {
                "display_name": "Vibration During Write",
                "ideal": "",
                "critical": false,
                "description": "Found in Maxtor 6B200M0 200GB and Maxtor 2R015H1 15GB disks.",
                "display_type": "normalized"
              },
              "211": {
                "display_name": "Vibration During Write",
                "ideal": "",
                "critical": false,
                "description": "A recording of a vibration encountered during write operations.",
                "display_type": "normalized"
              },
              "212": {
                "display_name": "Shock During Write",
                "ideal": "",
                "critical": false,
                "description": "A recording of shock encountered during write operations.",
                "display_type": "normalized"
              },
              "220": {
                "display_name": "Disk Shift",
                "ideal": "low",
                "critical": false,
                "description": "Distance the disk has shifted relative to the spindle (usually due to shock or temperature). Unit of measure is unknown.",
                "display_type": "normalized"
              },
              "221": {
                "display_name": "G-Sense Error Rate",
                "ideal": "low",
                "critical": false,
                "description": "The count of errors resulting from externally induced shock and vibration.",
                "display_type": "normalized"
              },
              "222": {
                "display_name": "Loaded Hours",
                "ideal": "",
                "critical": false,
                "description": "Time spent operating under data load (movement of magnetic head armature).",
                "display_type": "normalized"
              },
              "223": {
                "display_name": "Load/Unload Retry Count",
                "ideal": "",
                "critical": false,
                "description": "Count of times head changes position.",
                "display_type": "normalized"
              },
              "224": {
                "display_name": "Load Friction",
                "ideal": "low",
                "critical": false,
                "description": "Resistance caused by friction in mechanical parts while operating.",
                "display_type": "normalized"
              },
              "225": {
                "display_name": "Load/Unload Cycle Count",
                "ideal": "low",
                "critical": false,
                "description": "Total count of load cycles Some drives use 193 (0xC1) for Load Cycle Count instead. See Description for 193 for significance of this number. ",
                "display_type": "normalized"
              },
              "226": {
                "display_name": "Load \"In\"-time",
                "ideal": "",
                "critical": false,
                "description": "Total time of loading on the magnetic heads actuator (time not spent in parking area).",
                "display_type": "normalized"
              },
              "227": {
                "display_name": "Torque Amplification Count",
                "ideal": "low",
                "critical": false,
                "description": "Count of attempts to compensate for platter speed variations.[66]",
                "display_type": ""
              },
              "228": {
                "display_name": "Power-Off Retract Cycle",
                "ideal": "low",
                "critical": false,
                "description": "The number of power-off cycles which are counted whenever there is a \"retract event\" and the heads are loaded off of the media such as when the machine is powered down, put to sleep, or is idle.",
                "display_type": ""
              },
              "230": {
                "display_name": "GMR Head Amplitude ",
                "ideal": "",
                "critical": false,
                "description": "Amplitude of \"thrashing\" (repetitive head moving motions between operations).",
                "display_type": "normalized"
              },
              "231": {
                "display_name": "Life Left",
                "ideal": "",
                "critical": false,
                "description": "Indicates the approximate SSD life left, in terms of program/erase cycles or available reserved blocks. A normalized value of 100 represents a new drive, with a threshold value at 10 indicating a need for replacement. A value of 0 may mean that the drive is operating in read-only mode to allow data recovery.",
                "display_type": "normalized"
              },
              "232": {
                "display_name": "Endurance Remaining",
                "ideal": "",
                "critical": false,
                "description": "Number of physical erase cycles completed on the SSD as a percentage of the maximum physical erase cycles the drive is designed to endure.",
                "display_type": "normalized"
              },
              "233": {
                "display_name": "Media Wearout Indicator",
                "ideal": "",
                "critical": false,
                "description": "Intel SSDs report a normalized value from 100, a new drive, to a minimum of 1. It decreases while the NAND erase cycles increase from 0 to the maximum-rated cycles.",
                "display_type": "normalized"
              },
              "234": {
                "display_name": "Average erase count",
                "ideal": "",
                "critical": false,
                "description": "Decoded as: byte 0-1-2 = average erase count (big endian) and byte 3-4-5 = max erase count (big endian).",
                "display_type": "normalized"
              },
              "235": {
                "display_name": "Good Block Count",
                "ideal": "",
                "critical": false,
                "description": "Decoded as: byte 0-1-2 = good block count (big endian) and byte 3-4 = system (free) block count.",
                "display_type": "normalized"
              },
              "240": {
                "display_name": "Head Flying Hours",
                "ideal": "",
                "critical": false,
                "description": "Time spent during the positioning of the drive heads.[15][71] Some Fujitsu drives report the count of link resets during a data transfer.",
                "display_type": "normalized"
              },
              "241": {
                "display_name": "Total LBAs Written",
                "ideal": "",
                "critical": false,
                "description": "Total count of LBAs written.",
                "display_type": "normalized"
              },
              "242": {
                "display_name": "Total LBAs Read",
                "ideal": "",
                "critical": false,
                "description": "Total count of LBAs read.Some S.M.A.R.T. utilities will report a negative number for the raw value since in reality it has 48 bits rather than 32.",
                "display_type": "normalized"
              },
              "243": {
                "display_name": "Total LBAs Written Expanded",
                "ideal": "",
                "critical": false,
                "description": "The upper 5 bytes of the 12-byte total number of LBAs written to the device. The lower 7 byte value is located at attribute 0xF1.",
                "display_type": "normalized"
              },
              "244": {
                "display_name": "Total LBAs Read Expanded",
                "ideal": "",
                "critical": false,
                "description": "The upper 5 bytes of the 12-byte total number of LBAs read from the device. The lower 7 byte value is located at attribute 0xF2.",
                "display_type": "normalized"
              },
              "249": {
                "display_name": "NAND Writes (1GiB)",
                "ideal": "",
                "critical": false,
                "description": "Total NAND Writes. Raw value reports the number of writes to NAND in 1 GB increments.",
                "display_type": "normalized"
              },
              "250": {
                "display_name": "Read Error Retry Rate",
                "ideal": "low",
                "critical": false,
                "description": "Count of errors while reading from a disk.",
                "display_type": "normalized"
              },
              "251": {
                "display_name": "Minimum Spares Remaining",
                "ideal": "",
                "critical": false,
                "description": "The Minimum Spares Remaining attribute indicates the number of remaining spare blocks as a percentage of the total number of spare blocks available.",
                "display_type": "normalized"
              },
              "252": {
                "display_name": "Newly Added Bad Flash Block",
                "ideal": "",
                "critical": false,
                "description": "The Newly Added Bad Flash Block attribute indicates the total number of bad flash blocks the drive detected since it was first initialized in manufacturing.",
                "display_type": "normalized"
              },
              "254": {
                "display_name": "Free Fall Protection",
                "ideal": "low",
                "critical": false,
                "description": "Count of \"Free Fall Events\" detected.",
                "display_type": "normalized"
              }
            },
            "NVME": {
              "available_spare": {
                "display_name": "Available Spare",
                "ideal": "high",
                "critical": true,
                "description": "Contains a normalized percentage (0 to 100%) of the remaining spare capacity available.",
                "display_type": ""
              },
              "controller_busy_time": {
                "display_name": "Controller Busy Time",
                "ideal": "",
                "critical": false,
                "description": "Contains the amount of time the controller is busy with I/O commands. The controller is busy when there is a command outstanding to an I/O Queue (specifically, a command was issued via an I/O Submission Queue Tail doorbell write and the corresponding completion queue entry has not been posted yet to the associated I/O Completion Queue). This value is reported in minutes.",
                "display_type": ""
              },
              "critical_comp_time": {
                "display_name": "Critical CompTime",
                "ideal": "",
                "critical": false,
                "description": "Contains the amount of time in minutes that the controller is operational and the Composite Temperature is greater the Critical Composite Temperature Threshold (CCTEMP) field in the Identify Controller data structure.",
                "display_type": ""
              },
              "critical_warning": {
                "display_name": "Critical Warning",
                "ideal": "low",
                "critical": true,
                "description": "This field indicates critical warnings for the state of the controller. Each bit corresponds to a critical warning type; multiple bits may be set. If a bit is cleared to ‘0’, then that critical warning does not apply. Critical warnings may result in an asynchronous event notification to the host. Bits in this field represent the current associated state and are not persistent.",
                "display_type": ""
              },
              "data_units_read": {
                "display_name": "Data Units Read",
                "ideal": "",
                "critical": false,
                "description": "Contains the number of 512 byte data units the host has read from the controller; this value does not include metadata. This value is reported in thousands (i.e., a value of 1 corresponds to 1000 units of 512 bytes read) and is rounded up. When the LBA size is a value other than 512 bytes, the controller shall convert the amount of data read to 512 byte units.",
                "display_type": ""
              },
              "data_units_written": {
                "display_name": "Data Units Written",
                "ideal": "",
                "critical": false,
                "description": "Contains the number of 512 byte data units the host has written to the controller; this value does not include metadata. This value is reported in thousands (i.e., a value of 1 corresponds to 1000 units of 512 bytes written) and is rounded up. When the LBA size is a value other than 512 bytes, the controller shall convert the amount of data written to 512 byte units.",
                "display_type": ""
              },
              "host_reads": {
                "display_name": "Host Reads",
                "ideal": "",
                "critical": false,
                "description": "Contains the number of read commands completed by the controller",
                "display_type": ""
              },
              "host_writes": {
                "display_name": "Host Writes",
                "ideal": "",
                "critical": false,
                "description": "Contains the number of write commands completed by the controller",
                "display_type": ""
              },
              "media_errors": {
                "display_name": "Media Errors",
                "ideal": "low",
                "critical": true,
                "description": "Contains the number of occurrences where the controller detected an unrecovered data integrity error. Errors such as uncorrectable ECC, CRC checksum failure, or LBA tag mismatch are included in this field.",
                "display_type": ""
              },
              "num_err_log_entries": {
                "display_name": "Numb Err Log Entries",
                "ideal": "low",
                "critical": true,
                "description": "Contains the number of Error Information log entries over the life of the controller.",
                "display_type": ""
              },
              "percentage_used": {
                "display_name": "Percentage Used",
                "ideal": "low",
                "critical": true,
                "description": "Contains a vendor specific estimate of the percentage of NVM subsystem life used based on the actual usage and the manufacturer’s prediction of NVM life. A value of 100 indicates that the estimated endurance of the NVM in the NVM subsystem has been consumed, but may not indicate an NVM subsystem failure. The value is allowed to exceed 100. Percentages greater than 254 shall be represented as 255. This value shall be updated once per power-on hour (when the controller is not in a sleep state).",
                "display_type": ""
              },
              "power_cycles": {
                "display_name": "Power Cycles",
                "ideal": "",
                "critical": false,
                "description": "Contains the number of power cycles.",
                "display_type": ""
              },
              "power_on_hours": {
                "display_name": "Power on Hours",
                "ideal": "",
                "critical": false,
                "description": "Contains the number of power-on hours. Power on hours is always logging, even when in low power mode.",
                "display_type": ""
              },
              "temperature": {
                "display_name": "Temperature",
                "ideal": "",
                "critical": false,
                "description": "",
                "display_type": ""
              },
              "unsafe_shutdowns": {
                "display_name": "Unsafe Shutdowns",
                "ideal": "",
                "critical": false,
                "description": "Contains the number of unsafe shutdowns. This count is incremented when a shutdown notification (CC.SHN) is not received prior to loss of power.",
                "display_type": ""
              },
              "warning_temp_time": {
                "display_name": "Warning Temp Time",
                "ideal": "",
                "critical": false,
                "description": "Contains the amount of time in minutes that the controller is operational and the Composite Temperature is greater than or equal to the Warning Composite Temperature Threshold (WCTEMP) field and less than the Critical Composite Temperature Threshold (CCTEMP) field in the Identify Controller data structure.",
                "display_type": ""
              }
            }
          },
          "status": "OK"
        })},
        100
      );
    });
  },

  format({disk, format, password}, onProgress) {
    return new Promise((resolve, reject) => {
      setTimeout(() => {
        resolve({
          "status": "ok",
        })},
        2000
      );
    });
  }
};

const snapRAID = {
  create: (args) => {
    return new Promise((resolve, reject) => {
      setTimeout(() => {
        resolve({
          "status": "ok",
        })},
        2000
      );
    });
  },
  update: (name, args) => {
    return new Promise((resolve, reject) => {
      setTimeout(() => {
        resolve({
          "status": "ok",
        })},
        2000
      );
    });
  },
  delete: (name) => {
    return new Promise((resolve, reject) => {
      setTimeout(() => {
        resolve({
          "status": "ok",
        })},
        2000
      );
    });
  },
  list: (args) => {
    return new Promise((resolve, reject) => {
      setTimeout(() => {
        resolve({
          "data": [
            {
              "Name": "Storage Parity",
              "Enabled": true,
              "Data": {
                "disk0": "/var/mnt/sdc1",
                "disk1": "/var/mnt/sdd1"
              },
              "Parity": [
                "/var/mnt/sda1"
              ],
              "SyncCrontab": "0 0 2 * * *",
              "ScrubCrontab": "0 0 4 */2 * *",
              "CheckOnFix": false,
              "Status": "Self test...\nLoading state from /var/mnt/sdc1/snapraid.content...\nWARNING! Content file '/var/mnt/sdc1/snapraid.content' not found, attempting with another copy...\nLoading state from /var/mnt/sdd1/snapraid.content...\nNo content file found. Assuming empty.\nUsing 0 MiB of memory for the file-system.\nSnapRAID status report:\n\n   Files Fragmented Excess  Wasted  Used    Free  Use Name\n            Files  Fragments  GB      GB      GB\n       0       0       0     0.0       0       -   -  disk0\n       0       0       0     0.0       0       -   -  disk1\n --------------------------------------------------------------------------\n       0       0       0     0.0       0       0   0%\n\nWARNING! Free space info will be valid after the first sync.\nThe array is empty.\n"
            }
          ],
          "status": "OK"
        })},
        2000
      );
    });
  },
  sync: (name) => {
    return new Promise((resolve, reject) => {
      setTimeout(() => {
        resolve({
          "status": "ok",
        })},
        2000
      );
    });
  },
  fix: (name) => {
    return new Promise((resolve, reject) => {
      setTimeout(() => {
        resolve({
          "status": "ok",
        })},
        2000
      );
    });
  },
  enable: (name, enable) => {
    return new Promise((resolve, reject) => {
      setTimeout(() => {
        resolve({
          "status": "ok",
        })},
        2000
      );
    });
  },
  scrub: (name) => {
    return new Promise((resolve, reject) => {
      setTimeout(() => {
        resolve({
          "status": "ok",
        })},
        2000
      );
    });
  },
};

export {
  mounts,
  disks,
  snapRAID,
};