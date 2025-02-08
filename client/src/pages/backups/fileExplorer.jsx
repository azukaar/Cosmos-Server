// material-ui
import * as React from 'react';
import { Alert, Box, Button, Checkbox, CircularProgress, FormControl, FormHelperText, Grid, Icon, IconButton, InputLabel, List, ListItem, ListItemButton, MenuItem, NativeSelect, Select, Stack, TextField, Typography } from '@mui/material';
import Dialog from '@mui/material/Dialog';
import {TreeItem, TreeView} from '@mui/x-tree-view';
import DialogActions from '@mui/material/DialogActions';
import DialogContent from '@mui/material/DialogContent';
import DialogContentText from '@mui/material/DialogContentText';
import DialogTitle from '@mui/material/DialogTitle';
import { LoadingButton } from '@mui/lab';
import { useFormik, FormikProvider } from 'formik';
import * as Yup from 'yup';
import { useTranslation } from 'react-i18next';
import { FolderOpenOutlined, ClockCircleOutlined, DashboardOutlined, DeleteOutlined, DownOutlined, LockOutlined, SafetyOutlined, UpOutlined, ExpandOutlined, FolderFilled, FileFilled, FileOutlined, LoadingOutlined, FolderAddFilled, FolderViewOutlined, ReloadOutlined } from "@ant-design/icons";
import * as API from '../../api';
import {simplifyNumber} from '../dashboard/components/utils';
import { current } from '@reduxjs/toolkit';
import { json } from 'react-router';
import ResponsiveButton from '../../components/responseiveButton';
import RestoreDialog from './restoreDialog';

function pathIncluded(list, path, originalSource) {
  if(list.includes(path) || list.includes(originalSource)) {
    return true;
  }

  const pathParts = path.split('/');
  return list.some(p => {
    const parentParts = p.split('/');
    return parentParts.every((part, i) => pathParts[i] === part);
  });
}

function metaData(file) {
  // get latest change between mtime and ctime and return with dayjs time ago
  const mtime = file.mtime;
  const ctime = file.ctime;

  let latest;

  let res = "";

  if (file.size) {
    res += simplifyNumber(file.size, 'B');
    
    if(mtime && ctime) {
      res += " | ";
    }
  }

  if(mtime && ctime) {
    if (mtime > ctime) {
      latest = mtime;
    } else {
      latest = ctime;
    }

    res += new Date(latest).toLocaleString();
  }
  
  return res;
}


const BackupFileExplorer = ({ getFile, onSelect, selectedSnapshot, backupName, backup, className = '' }) => {
  const [files, setFiles] = React.useState({});
  const [directoriesAsTree, setDirectoriesAsTree] = React.useState({});
  const [selectedPath, setSelectedPath] = React.useState('');
  const [candidatePaths, setCandidatePaths] = React.useState([]);
  const { t } = useTranslation();
  
  function transformToTree(data) {
    const root = {
      path: Object.keys(data)[0],
      label: Object.keys(data)[0].split('/').pop(),
      opened: true,
      file: {
        type: "dir",
      },
      children: []
    };

    root.label = backup.Source

    function buildTree(node) {
      const path = node.path;
      if (data[path]) {
        for (const item of data[path]) {
          const childPath = `${path}/${item.name}`.replace(/^\/\//, '/');
          const child = {
            path: childPath,
            label: item.name,
            file: item,
          };

          if (data[childPath]) {
            child.children = [];
            buildTree(child);
          }

          node.children.push(child);
        }
      }
    }

    buildTree(root);
    return root;
  }

  React.useEffect(() => {
    updateFiles(backup.Source, true);
  }, [selectedSnapshot]);

  const updateFiles = async (path, forceNew) => {
    // If already loaded, just update the selection
    if (!forceNew && files[path]) {
      setSelectedPath(path);
      return;
    }

    if(forceNew) {
      setCandidatePaths([]);
    }

    try {
      const result = await getFile(path);
      if (!result.data) return;
      
      const metadata = result.data.shift();
      const files = result.data.filter(i => i.path != path);

      setFiles(current => {
        let newFiles = {
          ...current,
          [path]: files || []
        };
        if(forceNew) {
          newFiles = {
            [path]: files || []
          };
        }
        setDirectoriesAsTree(transformToTree(newFiles));
        return newFiles;
      });
      setSelectedPath(path);
    } catch (error) {
      console.error('Error loading files:', error);
    }
  };

  const handleToggleNode = (event, nodePath) => {
    nodePath = nodePath.split('@').pop();
    console.log(nodePath)
    if (nodePath === backup.Source) {
      updateFiles(nodePath);
      return;
    }

    // Find the parent path and filename
    let nodeParent = nodePath.split('/').slice(0, -1).join('/');
    nodeParent = nodeParent.replace(/\/$/, '');
    if (nodeParent === '') {
      nodeParent = '/';
    }
    const filename = nodePath.split('/').pop();

    if (files[nodeParent]) {
      const file = files[nodeParent].find(f => f.name === filename);
      if (file) {
        if (file.type == "dir") {
          updateFiles(nodePath);
        }
        onSelect && onSelect(nodePath);
        setSelectedPath(nodePath);
      }
    }

    return true;
  };

  const renderAllTree = (node) => {
    let nodeCounter = 0;

    const renderTree = (node) => {
      nodeCounter++;
      return (
        <TreeItem 
          key={node.path} 
          nodeId={selectedSnapshot + "@" + node.path}
          label={
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', width: '100%' }}>
              <Stack direction="row" spacing={2} alignItems="center">
                {nodeCounter > 1 ? <Checkbox style={{zIndex: 1000}} onClick={(e) => {
                    e.stopPropagation();
                    if(candidatePaths.some(path => node.path == path)) {
                      setCandidatePaths(candidatePaths.filter(path => node.path != path));
                    } else {
                      setCandidatePaths([...candidatePaths, node.path]);
                    }
                  }}
                  checked={pathIncluded(candidatePaths, node.path)}
                  disabled={pathIncluded(candidatePaths, node.path) && !candidatePaths.some(path => node.path == path)}
                  /> : <Checkbox style={{zIndex: 1000, opacity: 0, width: '1px', paddingLeft:0, paddingRight:0, margin: 0}} disabled  />}
                {node.label}
              </Stack>
              <div style={{opacity: 0.8}}>{metaData(node.file)}</div>
            </div>
          }
          icon={node.file && node.file.type == "dir" ? null : <FileOutlined className="h-4 w-4" />}
        >
          {Array.isArray(node.children)
            ? node.children.map((child) => renderTree(child))
            : (
              (node.file && node.file.type == "dir") ? 
              <TreeItem key={node.id + "load"} nodeId={(++nodeCounter).toString()} label={"Loading..."} icon={<LoadingOutlined />} />
              : null
            )}
        </TreeItem>
      );
    };

    return renderTree(node);
  }

  if(Object.keys(directoriesAsTree).length === 0) {
    return (
      <Stack spacing={2} justifyContent={'center'} alignItems={'center'}>
        <CircularProgress />
      </Stack>
    );
  }

  return (
    <Stack spacing={2} className={className}>
      <Stack direction="row" spacing={2} justifyContent="flex-start">
        <RestoreDialog originalSource={backup.Source} candidatePaths={candidatePaths} selectedSnapshot={selectedSnapshot} backupName={backupName} />
      </Stack>
      <TreeView
        onNodeSelect={handleToggleNode}
        defaultCollapseIcon={<FolderOpenOutlined className="h-4 w-4" />}
        defaultExpandIcon={<FolderFilled className="h-4 w-4" />}
      >
        {Object.keys(directoriesAsTree).length > 0 && renderAllTree(directoriesAsTree)}
      </TreeView>
    </Stack>
  );
};

export default BackupFileExplorer;