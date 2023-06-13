const dnsList = require('./client/src/utils/dns-list.json');

console.log(dnsList);

// for each make a request to https://go-acme.github.io/lego/dns/{dns}/#credentials

let finalList = {};

let i = 0;
(async () => {
  for(const dns of dnsList) {
    console.log(`Fetching ${dns} infos`)
    let result = (await (await fetch(`https://go-acme.github.io/lego/dns/${dns}/#credentials`)).text());
    result = result.split(`<h2 id="credentials">Credentials</h2>`)[1];
    result = result.split(`</table>`)[0] + `</table>`;
    let vars = result.match(/<code>(.*?)<\/code>/g);
    vars = vars.map(v => v.replace(/<\/?code>/g, ''));

    finalList[dns] = {
      name: dns,
      url: `https://go-acme.github.io/lego/dns/${dns}/#credentials`,
      docs: result,
      vars: vars
    }

    console.log(`${i++}/${dnsList.length} done`)
  }

  // save to file
  const fs = require('fs');
  fs.writeFileSync('./client/src/utils/dns-config.json', JSON.stringify(finalList, null, 2));
})();