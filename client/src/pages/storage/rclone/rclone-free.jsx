import { Typography } from "@mui/material";
import PremiumSalesPage from "../../../utils/free"
import { useTranslation } from "react-i18next";
import { CosmosCollapse } from "../../config/users/formShortcuts";
import VMWarning from "../vmWarning";

const RemoteStorageSalesPage = ({containerized}) => {
  const { t, i18n } = useTranslation();

  return <>
  
  {containerized && <VMWarning />}

  <PremiumSalesPage salesKey="remote" extra={
    <CosmosCollapse title="Supported Storages">
      <table>
        <thead>
          <tr>
            <th>Name</th>
            <th >Hash</th>
            <th >ModTime</th>
            <th >Case Insensitive</th>
            <th >Duplicate Files</th>
            <th >MIME Type</th>
            <th >Metadata</th>
          </tr>
        </thead>
        <tbody>
          <tr>
            <td>1Fichier</td>
            <td >Whirlpool</td>
            <td >-</td>
            <td >No</td>
            <td >Yes</td>
            <td >R</td>
            <td >-</td>
          </tr>
          <tr>
            <td>Akamai Netstorage</td>
            <td >MD5, SHA256</td>
            <td >R/W</td>
            <td >No</td>
            <td >No</td>
            <td >R</td>
            <td >-</td>
          </tr>
          <tr>
            <td>Amazon S3 (or S3 compatible)</td>
            <td >MD5</td>
            <td >R/W</td>
            <td >No</td>
            <td >No</td>
            <td >R/W</td>
            <td >RWU</td>
          </tr>
          <tr>
            <td>Backblaze B2</td>
            <td >SHA1</td>
            <td >R/W</td>
            <td >No</td>
            <td >No</td>
            <td >R/W</td>
            <td >-</td>
          </tr>
          <tr>
            <td>Box</td>
            <td >SHA1</td>
            <td >R/W</td>
            <td >Yes</td>
            <td >No</td>
            <td >-</td>
            <td >-</td>
          </tr>
          <tr>
            <td>Citrix ShareFile</td>
            <td >MD5</td>
            <td >R/W</td>
            <td >Yes</td>
            <td >No</td>
            <td >-</td>
            <td >-</td>
          </tr>
          <tr>
            <td>Dropbox</td>
            <td >DBHASH ¹</td>
            <td >R</td>
            <td >Yes</td>
            <td >No</td>
            <td >-</td>
            <td >-</td>
          </tr>
          <tr>
            <td>Enterprise File Fabric</td>
            <td >-</td>
            <td >R/W</td>
            <td >Yes</td>
            <td >No</td>
            <td >R/W</td>
            <td >-</td>
          </tr>
          <tr>
            <td>Files.com</td>
            <td >MD5, CRC32</td>
            <td >DR/W</td>
            <td >Yes</td>
            <td >No</td>
            <td >R</td>
            <td >-</td>
          </tr>
          <tr>
            <td>FTP</td>
            <td >-</td>
            <td >R/W ¹⁰</td>
            <td >No</td>
            <td >No</td>
            <td >-</td>
            <td >-</td>
          </tr>
          <tr>
            <td>Gofile</td>
            <td >MD5</td>
            <td >DR/W</td>
            <td >No</td>
            <td >Yes</td>
            <td >R</td>
            <td >-</td>
          </tr>
          <tr>
            <td>Google Cloud Storage</td>
            <td >MD5</td>
            <td >R/W</td>
            <td >No</td>
            <td >No</td>
            <td >R/W</td>
            <td >-</td>
          </tr>
          <tr>
            <td>Google Drive</td>
            <td >MD5, SHA1, SHA256</td>
            <td >DR/W</td>
            <td >No</td>
            <td >Yes</td>
            <td >R/W</td>
            <td >DRWU</td>
          </tr>
          <tr>
            <td>Google Photos</td>
            <td >-</td>
            <td >-</td>
            <td >No</td>
            <td >Yes</td>
            <td >R</td>
            <td >-</td>
          </tr>
          <tr>
            <td>HDFS</td>
            <td >-</td>
            <td >R/W</td>
            <td >No</td>
            <td >No</td>
            <td >-</td>
            <td >-</td>
          </tr>
          <tr>
            <td>HiDrive</td>
            <td >HiDrive ¹²</td>
            <td >R/W</td>
            <td >No</td>
            <td >No</td>
            <td >-</td>
            <td >-</td>
          </tr>
          <tr>
            <td>HTTP</td>
            <td >-</td>
            <td >R</td>
            <td >No</td>
            <td >No</td>
            <td >R</td>
            <td >-</td>
          </tr>
          <tr>
            <td>Internet Archive</td>
            <td >MD5, SHA1, CRC32</td>
            <td >R/W ¹¹</td>
            <td >No</td>
            <td >No</td>
            <td >-</td>
            <td >RWU</td>
          </tr>
          <tr>
            <td>Jottacloud</td>
            <td >MD5</td>
            <td >R/W</td>
            <td >Yes</td>
            <td >No</td>
            <td >R</td>
            <td >RW</td>
          </tr>
          <tr>
            <td>Koofr</td>
            <td >MD5</td>
            <td >-</td>
            <td >Yes</td>
            <td >No</td>
            <td >-</td>
            <td >-</td>
          </tr>
          <tr>
            <td>Linkbox</td>
            <td >-</td>
            <td >R</td>
            <td >No</td>
            <td >No</td>
            <td >-</td>
            <td >-</td>
          </tr>
          <tr>
            <td>Mail.ru Cloud</td>
            <td >Mailru ⁶</td>
            <td >R/W</td>
            <td >Yes</td>
            <td >No</td>
            <td >-</td>
            <td >-</td>
          </tr>
          <tr>
            <td>Mega</td>
            <td >-</td>
            <td >-</td>
            <td >No</td>
            <td >Yes</td>
            <td >-</td>
            <td >-</td>
          </tr>
          <tr>
            <td>Memory</td>
            <td >MD5</td>
            <td >R/W</td>
            <td >No</td>
            <td >No</td>
            <td >-</td>
            <td >-</td>
          </tr>
          <tr>
            <td>Microsoft Azure Blob Storage</td>
            <td >MD5</td>
            <td >R/W</td>
            <td >No</td>
            <td >No</td>
            <td >R/W</td>
            <td >-</td>
          </tr>
          <tr>
            <td>Microsoft Azure Files Storage</td>
            <td >MD5</td>
            <td >R/W</td>
            <td >Yes</td>
            <td >No</td>
            <td >R/W</td>
            <td >-</td>
          </tr>
          <tr>
            <td>Microsoft OneDrive</td>
            <td >QuickXorHash ⁵</td>
            <td >DR/W</td>
            <td >Yes</td>
            <td >No</td>
            <td >R</td>
            <td >DRW</td>
          </tr>
          <tr>
            <td>OpenDrive</td>
            <td >MD5</td>
            <td >R/W</td>
            <td >Yes</td>
            <td >Partial ⁸</td>
            <td >-</td>
            <td >-</td>
          </tr>
          <tr>
            <td>OpenStack Swift</td>
            <td >MD5</td>
            <td >R/W</td>
            <td >No</td>
            <td >No</td>
            <td >R/W</td>
            <td >-</td>
          </tr>
          <tr>
            <td>Oracle Object Storage</td>
            <td >MD5</td>
            <td >R/W</td>
            <td >No</td>
            <td >No</td>
            <td >R/W</td>
            <td >-</td>
          </tr>
          <tr>
            <td>pCloud</td>
            <td >MD5, SHA1 ⁷</td>
            <td >R</td>
            <td >No</td>
            <td >No</td>
            <td >W</td>
            <td >-</td>
          </tr>
          <tr>
            <td>PikPak</td>
            <td >MD5</td>
            <td >R</td>
            <td >No</td>
            <td >No</td>
            <td >R</td>
            <td >-</td>
          </tr>
          <tr>
            <td>Pixeldrain</td>
            <td >SHA256</td>
            <td >R/W</td>
            <td >No</td>
            <td >No</td>
            <td >R</td>
            <td >RW</td>
          </tr>
          <tr>
            <td>premiumize.me</td>
            <td >-</td>
            <td >-</td>
            <td >Yes</td>
            <td >No</td>
            <td >R</td>
            <td >-</td>
          </tr>
          <tr>
            <td>put.io</td>
            <td >CRC-32</td>
            <td >R/W</td>
            <td >No</td>
            <td >Yes</td>
            <td >R</td>
            <td >-</td>
          </tr>
          <tr>
            <td>Proton Drive</td>
            <td >SHA1</td>
            <td >R/W</td>
            <td >No</td>
            <td >No</td>
            <td >R</td>
            <td >-</td>
          </tr>
          <tr>
            <td>QingStor</td>
            <td >MD5</td>
            <td >- ⁹</td>
            <td >No</td>
            <td >No</td>
            <td >R/W</td>
            <td >-</td>
          </tr>
          <tr>
            <td>Quatrix by Maytech</td>
            <td >-</td>
            <td >R/W</td>
            <td >No</td>
            <td >No</td>
            <td >-</td>
            <td >-</td>
          </tr>
          <tr>
            <td>Seafile</td>
            <td >-</td>
            <td >-</td>
            <td >No</td>
            <td >No</td>
            <td >-</td>
            <td >-</td>
          </tr>
          <tr>
            <td>SFTP</td>
            <td >MD5, SHA1 ²</td>
            <td >DR/W</td>
            <td >Depends</td>
            <td >No</td>
            <td >-</td>
            <td >-</td>
          </tr>
          <tr>
            <td>Sia</td>
            <td >-</td>
            <td >-</td>
            <td >No</td>
            <td >No</td>
            <td >-</td>
            <td >-</td>
          </tr>
          <tr>
            <td>SMB</td>
            <td >-</td>
            <td >R/W</td>
            <td >Yes</td>
            <td >No</td>
            <td >-</td>
            <td >-</td>
          </tr>
          <tr>
            <td>SugarSync</td>
            <td >-</td>
            <td >-</td>
            <td >No</td>
            <td >No</td>
            <td >-</td>
            <td >-</td>
          </tr>
          <tr>
            <td>Storj</td>
            <td >-</td>
            <td >R</td>
            <td >No</td>
            <td >No</td>
            <td >-</td>
            <td >-</td>
          </tr>
          <tr>
            <td>Uloz.to</td>
            <td >MD5, SHA256 ¹³</td>
            <td >-</td>
            <td >No</td>
            <td >Yes</td>
            <td >-</td>
            <td >-</td>
          </tr>
          <tr>
            <td>Uptobox</td>
            <td >-</td>
            <td >-</td>
            <td >No</td>
            <td >Yes</td>
            <td >-</td>
            <td >-</td>
          </tr>
          <tr>
            <td>WebDAV</td>
            <td >MD5, SHA1 ³</td>
            <td >R ⁴</td>
            <td >Depends</td>
            <td >No</td>
            <td >-</td>
            <td >-</td>
          </tr>
          <tr>
            <td>Yandex Disk</td>
            <td >MD5</td>
            <td >R/W</td>
            <td >No</td>
            <td >No</td>
            <td >R</td>
            <td >-</td>
          </tr>
          <tr>
            <td>Zoho WorkDrive</td>
            <td >-</td>
            <td >-</td>
            <td >No</td>
            <td >No</td>
            <td >-</td>
            <td >-</td>
          </tr>
          <tr>
            <td>The local filesystem</td>
            <td >All</td>
            <td >DR/W</td>
            <td >Depends</td>
            <td >No</td>
            <td >-</td>
            <td >DRWU</td>
          </tr>
        </tbody>
      </table>
    </CosmosCollapse>
  } />
  </>;
};

export default RemoteStorageSalesPage;