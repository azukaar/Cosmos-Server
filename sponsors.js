const tokens = require('./tokens.json')
const sponsors = require('@jamesives/github-sponsors-readme-action')
const fs = require('fs').promises

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

  if (count === 1) {
    sentence = `<h3 align="center">Thanks to the sponsors:</h3>`
  } else if (count > 10) {
    sentence = `<h3 align="center">Thanks to the ${count} sponsors:</h3>`
  }

  // open the readme
  const readme = await fs.readFile(configuration.file, 'utf8')

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
  await fs.writeFile(configuration.file, newReadme)
  
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
sponsorsGenerate()
