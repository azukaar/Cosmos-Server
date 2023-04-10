const packageJson = require('./package.json');
const { execSync } = require('child_process');

function tagRelease() {
  const commitMessage = execSync('git log -1 --pretty=%B').toString()
  if(!commitMessage.toLocaleLowerCase().startsWith('[release]')) {
    console.log('Not a release commit, skipping tag update')
    return
  }

  const version = packageJson.version;
  const tag = `v${version}`;
  const message = `Release ${tag}`;

  const e = execSync(`git tag -f -a ${tag} -m "${message}"`);
  
  console.log(e.toString());
}

tagRelease();