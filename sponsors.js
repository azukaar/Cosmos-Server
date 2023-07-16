const tokens = require('./tokens.json')
const sponsors = require('@jamesives/github-sponsors-readme-action')
const fs = require('fs')
const packageJson = require('./package.json')

let configuration = {
  token: tokens.github,
  file: './readme.md',
  marker: 'sponsors',
  template: '<a href="{url}"><img src="{avatar}" style="border-radius:48px" width="48" height="48" alt="{name}" title="{name}" /></a>',
}

const sponsorsGenerate = async () => {
  const sponsorsReturn = await (await (sponsors.getSponsors(configuration))).data.viewer
  const count = sponsorsReturn.sponsorshipsAsMaintainer.totalCount
  const sponsorsList = sponsorsReturn.sponsorshipsAsMaintainer.nodes

  let sentence = ``;

  if (count >= 1) {
    sentence = `<h3 align="center">Thanks to the sponsors:</h3>`
  } else if (count > 10) {
    sentence = `<h3 align="center">Thanks to the ${count} sponsors:</h3>`
  }

  // open the readme
  const readme = fs.readFileSync(configuration.file, 'utf8')

  // replace the sponsors marker with the new list
  const newReadme = readme.replace(
    new RegExp(`<!-- ${configuration.marker} -->[\\s\\S]*<!-- /${configuration.marker} -->`),
    `<!-- ${configuration.marker} -->\n${sentence}</br>\n<p align="center">${sponsorsList.map(sponsor => {
      return configuration.template
        .replace(/\{name\}/gi, sponsor.sponsorEntity.name)
        .replace(/\{avatar\}/gi, `https://avatars.githubusercontent.com/${sponsor.sponsorEntity.login}`)
        .replace(/\{url\}/gi, sponsor.sponsorEntity.url)
    }).join('\n')}\n</p><!-- /${configuration.marker} -->`
  )

  // write the new readme
  fs.writeFileSync(configuration.file, newReadme)
  
  // add to current commit by runnning shell command
  const { exec } = require('child_process')
  const e = exec('git add readme.md')
  e.stdout.on('data', (data) => {
    console.log(data)
  })
  e.stderr.on('data', (data) => {
    console.error(data)
  })

  console.log(`Sponsors updated!`)
}

function changelog() {
  // get the changes from last commit message
  const { execSync } = require('child_process')
  let commitMessage = execSync('git log -1 --pretty=%B').toString()

  if(!commitMessage.startsWith('[release]')) {
    console.log('Not a release commit, skipping changelog update')
    return
  }
  const version = packageJson.version

  commitMessage = commitMessage.replace('[release]', 'Version')
 
  // open changelog.md
  const changelog = fs.readFileSync('./changelog.md', 'utf8')
  
  // add the new changes to the top of the changelog
  const newChangelog = `## ${commitMessage}\n\n${changelog}`
  
  // write the new changelog
  fs.writeFileSync('./changelog.md', newChangelog)
  
  // add to current commit by runnning shell command
  const { exec } = require('child_process')
  const e = exec('git add changelog.md')
  e.stdout.on('data', (data) => {
    console.log(data)
  })
  e.stderr.on('data', (data) => {
    console.error(data)
  })
}

function refreshLicenceDate() {
  const year = new Date().getFullYear() + 3
  const month = new Date().getMonth() + 1
  const day = new Date().getDate()

  const licence = fs.readFileSync('./LICENCE', 'utf8')
  const newLicence = licence.replace(/(Change Date:\s+)(\d+-\d+-\d+)/, `$1${year}-${month}-${day}`)
  fs.writeFileSync('./LICENCE', newLicence)
}

sponsorsGenerate()
//refreshLicenceDate()
// changelog();